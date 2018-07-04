package chunkserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
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
		CreatedAt:  int64(time.Now().Second()),
		UpdatedAt:  int64(time.Now().Second()),
		// Chunks
	}
	var size int64
	var kvClient = clientv3.NewKV(s.etcdClient)

	for {
		fileChunkData, err := stream.Recv()
		if err == io.EOF {
			break
		}
		file.FileName = fileChunkData.FileName
		dataSize := int64(len(fileChunkData.Data))
		size += dataSize

		c := pb.Chunk{
			UUID:     uuid.New().String(),
			Size:     dataSize, // for now
			Used:     dataSize,
			Replicas: []string{s.name},
			FileUUID: file.UUID,
		}

		chunkPath := config.ChunkBasePath + c.UUID
		if err := files.Append(chunkPath, bytes.NewReader(fileChunkData.Data)); err != nil {
			logger.Sugar.Errorf("failed to write data into chunk %s", c.UUID)
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
		time.Sleep(time.Second * 3)
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

	grpcServer := grpc.NewServer()
	pb.RegisterChunkServerServer(grpcServer, &chunkServer)
	logger.Sugar.Infof("listen at %s", config.GRPCAddr)
	grpcServer.Serve(lis)
}
