package utils

import (
	"context"
	"encoding/json"

	"github.com/coreos/etcd/clientv3"
	"github.com/jiajunhuang/hfs/pb"
	"github.com/jiajunhuang/hfs/pkg/config"
	"github.com/jiajunhuang/hfs/pkg/logger"
)

func ToJSONString(c interface{}) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func GetChunkMeta(etcdClient *clientv3.Client, chunkUUID string) (*pb.Chunk, error) {
	resp, err := etcdClient.Get(context.Background(), config.ChunkBasePath+chunkUUID)
	if err != nil {
		logger.Sugar.Errorf("failed to get metadata of chunk %s: %s", chunkUUID, err)
		return nil, err
	}

	if len(resp.Kvs) != 1 {
		logger.Sugar.Errorf("bad metadata of chunk %s: %s", chunkUUID, resp.Kvs)
		return nil, err
	}

	chunk := pb.Chunk{}
	if err := json.Unmarshal(resp.Kvs[0].Value, &chunk); err != nil {
		logger.Sugar.Errorf("failed to load metadata of chunk %s: %s", chunkUUID, err)
		return nil, err
	}

	return &chunk, nil
}

func GetWorkersMeta(etcdClient *clientv3.Client) ([]string, error) {
	resp, err := etcdClient.Get(context.Background(), config.WorkerBasePath, clientv3.WithPrefix())
	if err != nil {
		logger.Sugar.Errorf("failed to get metadata of workers: %s", err)
		return nil, err
	}

	workers := []string{}
	for _, kv := range resp.Kvs {
		workers = append(workers, string(kv.Value))
	}

	return workers, nil
}

func GetFileMeta(etcdClient *clientv3.Client, fileUUID string) (*pb.File, error) {
	resp, err := etcdClient.Get(context.Background(), config.FileBasePath+fileUUID)
	if err != nil {
		logger.Sugar.Errorf("failed to get metadata of file %s: %s", fileUUID, err)
		return nil, err
	}

	if len(resp.Kvs) != 1 {
		logger.Sugar.Errorf("bad metadata of file %s: %s", fileUUID, resp.Kvs)
		return nil, err
	}

	file := pb.File{}
	if err := json.Unmarshal(resp.Kvs[0].Value, &file); err != nil {
		logger.Sugar.Errorf("failed to load metadata of file %s: %s", fileUUID, err)
		return nil, err
	}

	return &file, nil
}
