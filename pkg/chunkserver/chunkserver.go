package chunkserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/google/uuid"
	"github.com/jiajunhuang/hfs/pb"
	"github.com/jiajunhuang/hfs/pkg/config"
	"github.com/jiajunhuang/hfs/pkg/files"
	"github.com/jiajunhuang/hfs/pkg/hfsclient"
	"github.com/jiajunhuang/hfs/pkg/logger"
	"github.com/jiajunhuang/hfs/pkg/selection"
	"github.com/jiajunhuang/hfs/pkg/utils"
	"google.golang.org/grpc"
)

var (
	ErrFailedWrite     = errors.New("failed to write file or chunk")
	ErrFailedWriteMeta = errors.New("failed to sync metadata of file or chunk")
	ErrFailedGetFile   = errors.New("failed to get file or chunk")
	ErrFileNotExist    = errors.New("file or chunk not exist")
	ErrAlreadyExist    = errors.New("file or chunk already exist")
)

type ChunkServer struct {
	name       string
	addr       string
	etcdClient *clientv3.Client
}

func (s *ChunkServer) CreateFile(stream pb.ChunkServer_CreateFileServer) error {
	var file = pb.File{
		UUID: uuid.New().String(),
		// FileName
		// Size
		ReplicaNum: 1,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
		// Chunks
	}
	var size int64
	var kvClient = clientv3.NewKV(s.etcdClient)

	for {
		fileChunkData, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Sugar.Errorf("failed to receive chunk: %s", err)
			return ErrFailedWrite
		}
		file.FileName = fileChunkData.Msg
		dataSize := int64(len(fileChunkData.Data))
		size += dataSize

		c := pb.Chunk{
			UUID:     uuid.New().String(),
			Size:     int64(config.ChunkSize), // for now
			Used:     dataSize,
			Replicas: []string{s.name},
			FileUUID: file.UUID,
		}

		chunkPath := config.ChunkBasePath + c.UUID
		zeros := make([]byte, config.ChunkSize-len(fileChunkData.Data))
		data := append(fileChunkData.Data, zeros...)
		if err := files.Append(chunkPath, bytes.NewReader(data)); err != nil {
			logger.Sugar.Errorf("failed to write data into chunk %s: %s", c.UUID, err)
			return ErrFailedWrite
		}

		// sync metadata
		v, err := utils.ToJSONString(c)
		if err != nil {
			logger.Sugar.Errorf("failed to sync metadata of chunk %s", c.UUID)
			return ErrFailedWriteMeta
		}
		_, err = kvClient.Put(context.Background(), chunkPath, v)
		if err != nil {
			logger.Sugar.Errorf("failed to sync metadata of chunk %s", c.UUID)
			return ErrFailedWriteMeta
		}
		file.Chunks = append(file.Chunks, &c)
	}

	// sync metadata of file
	file.Size = size
	v, err := utils.ToJSONString(file)
	if err != nil {
		logger.Sugar.Errorf("failed to sync metadata of file %s", file.UUID)
		return ErrFailedWriteMeta
	}
	filePath := config.FileBasePath + file.UUID
	_, err = kvClient.Put(context.Background(), filePath, v)
	if err != nil {
		logger.Sugar.Errorf("failed to sync metadata of chunk %s", file.UUID)
		return ErrFailedWriteMeta
	}

	logger.Sugar.Infof("file %s created", file.UUID)
	return stream.SendAndClose(&pb.CreateFileResponse{Code: 0, File: &file})
}

func (s *ChunkServer) RemoveFile(ctx context.Context, file *pb.File) (*pb.GenericResponse, error) {
	kvClient := clientv3.NewKV(s.etcdClient)
	filePath := config.FileBasePath + file.UUID

	resp, err := kvClient.Get(context.Background(), filePath)
	if err != nil {
		logger.Sugar.Errorf("failed to get metadata of file %s", filePath)
		return nil, ErrFailedGetFile
	}

	if resp.Count == 0 {
		return nil, ErrFileNotExist
	} else if resp.Count != 1 {
		logger.Sugar.Errorf("bad metadata of file %s: %+v", filePath, resp)
		return nil, ErrFailedGetFile
	}

	if err := json.Unmarshal(resp.Kvs[0].Value, &file); err != nil {
		return nil, err
	}
	chunks := file.Chunks

	for _, c := range chunks {
		// remove all replicas of chunk
		for _, dialURL := range c.Replicas {
			// get gRPC ready
			conn, err := grpc.Dial(dialURL, grpc.WithInsecure(), grpc.WithMaxMsgSize(config.GRPCMaxMsgSize))
			if err != nil {
				logger.Sugar.Fatalf("failed to connect to grpc server %s: %s", config.GRPCAddr, err)
			}
			defer conn.Close()

			grpcClient := pb.NewChunkServerClient(conn)
			chunkUUID := c.UUID
			if err := hfsclient.Delete(grpcClient, chunkUUID); err != nil {
				logger.Sugar.Errorf("failed to delete chunk %s of node %s", chunkUUID, dialURL)
			} else {
				logger.Sugar.Infof("chunk %s of node %s delete success!", chunkUUID, dialURL)
			}
		}
	}

	if _, err := kvClient.Delete(context.Background(), config.FileBasePath+file.UUID); err != nil {
		logger.Sugar.Errorf("failed to delete metadata of file %s: %s", file.UUID, err)
	}

	logger.Sugar.Infof("file %s removed", file.UUID)
	return &pb.GenericResponse{Code: 0, Msg: "success"}, nil
}

func (s *ChunkServer) CreateChunk(ctx context.Context, file *pb.FileChunkData) (*pb.GenericResponse, error) {
	chunkUUID := file.Msg
	chunkPath := config.ChunkBasePath + chunkUUID

	f, err := files.Create(chunkPath)
	if err != nil {
		logger.Sugar.Errorf("failed to create chunk %s: %s", chunkPath, err)
		return nil, ErrFailedWrite
	}
	f.Close()

	if err := files.Append(chunkPath, bytes.NewReader(file.Data)); err != nil {
		logger.Sugar.Errorf("failed to create chunk %s: %s", chunkUUID, err)
		return nil, ErrFailedWrite
	}
	logger.Sugar.Infof("chunk %s has been create", chunkUUID)

	return &pb.GenericResponse{Code: 0, Msg: chunkUUID}, nil
}

func (s *ChunkServer) ReadFile(req *pb.ReadFileRequest, stream pb.ChunkServer_ReadFileServer) error {
	kvClient := clientv3.NewKV(s.etcdClient)
	filePath := config.FileBasePath + req.FileUUID

	resp, err := kvClient.Get(context.Background(), filePath)
	if err != nil {
		logger.Sugar.Errorf("failed to get metadata of file %s", filePath)
		return ErrFailedGetFile
	}

	if resp.Count == 0 {
		return ErrFileNotExist
	} else if resp.Count != 1 {
		logger.Sugar.Errorf("bad metadata of file %s: %+v", filePath, resp)
		return ErrFailedGetFile
	}

	var file pb.File
	if err := json.Unmarshal(resp.Kvs[0].Value, &file); err != nil {
		return err
	}
	chunks := file.Chunks

	for i, c := range chunks {
		// read chunk from local file system. TODO: read it from one of it's replica
		chunkPath := config.ChunkBasePath + c.UUID
		f, err := os.Open(chunkPath)
		if err != nil {
			logger.Sugar.Errorf("failed to read %dth chunk %s: %s", i, c.UUID, err)
			return err
		}

		buf := make([]byte, config.ChunkSize)
		for {
			_, err := f.Read(buf)
			if err == io.EOF {
				break
			}
			// write it to stream
			stream.Send(&pb.FileChunkData{Data: buf[:c.Used], Msg: file.FileName})
		}
	}

	logger.Sugar.Infof("file %s readed", req.FileUUID)
	return nil
}

func (s *ChunkServer) KeepAlive() {
	kvClient := clientv3.NewKV(s.etcdClient)

	for {
		lease := clientv3.NewLease(s.etcdClient)
		grantResp, err := lease.Grant(context.TODO(), 10)
		if err != nil {
			logger.Sugar.Errorf("failed to grant lease: %s", err)
			continue
		}
		_, err = kvClient.Put(context.Background(), config.WorkerBasePath+s.name, s.addr, clientv3.WithLease(grantResp.ID))
		if err != nil {
			logger.Sugar.Errorf("failed to put %s to %s: %s", s.name, s.addr, err)
		} else {
			logger.Sugar.Infof("refresh ip %s to worker %s in KV %+v", s.name, s.addr, kvClient)
		}
		time.Sleep(time.Second * 7)
	}
}

func (s *ChunkServer) SyncChunk(chunkUUID string) {
	// get metadata of chunk
	chunk, err := utils.GetChunkMeta(s.etcdClient, chunkUUID)
	if err != nil {
		logger.Sugar.Errorf("failed to sync chunk %s: %s", chunkUUID, err)
		return
	}

	// get metadata of file
	file, err := utils.GetFileMeta(s.etcdClient, chunk.FileUUID)
	if err != nil {
		logger.Sugar.Errorf("failed to sync chunk %s: %s", chunkUUID, err)
		return
	}

	// get workers
	workers, err := utils.GetWorkersMeta(s.etcdClient)
	if err != nil {
		logger.Sugar.Errorf("failed to sync chunk %s: %s", chunkUUID, err)
		return
	}

	syncTo := selection.Random(workers, s.name, file.ReplicaNum)
	if len(syncTo) == 0 {
		logger.Sugar.Warnf("do not find any scheduable node for chunk %s, so quit", chunkUUID)
		return
	}

	succeed := []string{}
	for _, node := range syncTo {
		dialURL, err := utils.GetWorkerAddr(s.etcdClient, node)
		if err != nil {
			logger.Sugar.Errorf("failed to get IP of worker %s: %s", node, err)
			continue
		}
		// get gRPC ready
		conn, err := grpc.Dial(dialURL, grpc.WithInsecure(), grpc.WithMaxMsgSize(config.GRPCMaxMsgSize))
		if err != nil {
			logger.Sugar.Fatalf("failed to connect to grpc server %s: %s", config.GRPCAddr, err)
		}
		defer conn.Close()

		grpcClient := pb.NewChunkServerClient(conn)
		if err := hfsclient.UploadChunk(grpcClient, chunkUUID); err != nil {
			logger.Sugar.Errorf("failed to sync chunk %s to node %s: %s", chunkUUID, node, err)
			continue
		}

		logger.Sugar.Infof("chunk %s sync to node %s success!", chunkUUID, node)
		succeed = append(succeed, node)
	}

	if len(succeed) < 1 {
		logger.Sugar.Infof("chunk %s sync failed!", chunkUUID)
		return
	}

	// TODO: sync metadata of chunk
	// NOTE: here should start a transaction? for data safe
	chunk.Replicas = append(chunk.Replicas, succeed...)
	v, err := utils.ToJSONString(chunk)
	if err != nil {
		logger.Sugar.Fatalf("failed to save metadata of chunk %s: %s", chunkUUID, err)
	}
	_, err = s.etcdClient.Put(context.Background(), config.ChunkBasePath+chunkUUID, v)
	if err != nil {
		logger.Sugar.Fatalf("failed to save metadata of chunk %s: %s", chunkUUID, err)
	}

	logger.Sugar.Infof("metadata of chunk %s updated!", chunkUUID)
}

func (s *ChunkServer) ChunkWatcher() {
	chunkChan := s.etcdClient.Watch(context.Background(), config.ChunkBasePath, clientv3.WithPrefix())

	for resp := range chunkChan {
		for _, ev := range resp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				logger.Sugar.Infof("chunk %s added: %s, start to sync\n", ev.Kv.Key, ev.Kv.Value)
				chunk := pb.Chunk{}
				if err := json.Unmarshal(ev.Kv.Value, &chunk); err != nil {
					logger.Sugar.Errorf("failed to unmarshal chunk %s: %s", ev.Kv.Key, err)
					continue
				}

				if len(chunk.Replicas) != 1 {
					logger.Sugar.Warnf("chunk.Replicas: %s is more than 1, %s will not responsible for sync it", chunk.Replicas, s.name)
					continue
				} else {
					if chunk.Replicas[0] != s.name {
						logger.Sugar.Infof("chunk %s is not created at %s, so it will not responsible for sync it", chunk.UUID, s.name)
					} else {
						go s.SyncChunk(chunk.UUID)
					}
				}
			case mvccpb.DELETE:
				logger.Sugar.Infof("chunk %s deleted: %s\n", ev.Kv.Key, ev.Kv.Value)
			default:
				logger.Sugar.Fatalf("watcher: should not be here")
			}
		}
	}
}

// StartChunkServer works as it's name
func StartChunkServer() {
	etcdClient, err := clientv3.New(
		clientv3.Config{
			Endpoints:   config.EtcdEndpoints,
			DialTimeout: 2 * time.Second,
		},
	)

	if err != nil {
		logger.Sugar.Fatalf("failed to connect to etcd: %s", err)
	}

	defer etcdClient.Close()

	chunkServer := ChunkServer{config.ChunkServerName, config.ChunkServerAddr, etcdClient}
	go chunkServer.KeepAlive()
	go chunkServer.ChunkWatcher()

	// grpc server
	lis, err := net.Listen("tcp", config.GRPCAddr)
	if err != nil {
		logger.Sugar.Fatalf("failed to listen: %s", err)
	}

	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(config.GRPCMaxMsgSize), grpc.MaxSendMsgSize(config.GRPCMaxMsgSize))
	pb.RegisterChunkServerServer(grpcServer, &chunkServer)
	logger.Sugar.Infof("listen at %s", config.GRPCAddr)
	grpcServer.Serve(lis)
}
