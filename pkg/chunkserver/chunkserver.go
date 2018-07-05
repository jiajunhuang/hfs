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
	"github.com/google/uuid"
	"github.com/jiajunhuang/hfs/pb"
	"github.com/jiajunhuang/hfs/pkg/config"
	"github.com/jiajunhuang/hfs/pkg/files"
	"github.com/jiajunhuang/hfs/pkg/logger"
	"github.com/jiajunhuang/hfs/pkg/utils"
	"google.golang.org/grpc"
)

var (
	ErrFailedWrite     = errors.New("failed to write")
	ErrFailedWriteMeta = errors.New("failed to sync metadata")
	ErrFailedGetFile   = errors.New("failed to get file")
	ErrFileNotExist    = errors.New("file not exist")
)

type ChunkServer struct {
	name       string
	ip         string
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
		file.FileName = fileChunkData.FileName
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
	return stream.SendAndClose(&pb.CreateFileResponse{Code: 0, File: &file})
}

func (s *ChunkServer) RemoveFile(ctx context.Context, file *pb.File) (*pb.GenericResponse, error) {
	return nil, nil
}

func (s *ChunkServer) AppendFile(stream pb.ChunkServer_AppendFileServer) error {
	return nil
}

func (s *ChunkServer) ReadFile(req *pb.ReadFileRequest, stream pb.ChunkServer_ReadFileServer) error {
	kvClient := clientv3.NewKV(s.etcdClient)
	filePath := config.FileBasePath + req.FileUUID

	// TODO: resp should be used
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
			stream.Send(&pb.FileChunkData{Data: buf[:c.Used], FileName: file.FileName})
		}
	}

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
		_, err = kvClient.Put(context.Background(), config.WorkerBasePath+s.name, s.ip, clientv3.WithLease(grantResp.ID))
		if err != nil {
			logger.Sugar.Errorf("failed to put %s to %s: %s", s.name, s.ip, err)
		} else {
			logger.Sugar.Infof("refresh ip %s to worker %s in KV %+v", s.name, s.ip, kvClient)
		}
		time.Sleep(time.Second * 7)
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

	chunkServer := ChunkServer{config.ChunkServerName, config.ChunkServerIPAddr, etcdClient}
	go chunkServer.KeepAlive()

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
