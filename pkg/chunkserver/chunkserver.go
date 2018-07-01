package chunkserver

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/jiajunhuang/hfs/pkg/logger"
)

// StartChunkServer works as it's name
func StartChunkServer() {
	etcdClient, err := clientv3.New(
		clientv3.Config{
			Endpoints:   []string{"http://127.0.0.1:2379"},
			DialTimeout: 2 * time.Second,
		},
	)

	if err != nil {
		logger.Sugar.Fatalf("failed to connect to etcd: %s", err)
	}

	defer etcdClient.Close()

	kvClient := clientv3.NewKV(etcdClient)
	resp, err := kvClient.Get(context.Background(), "foo")
	if err != nil {
		logger.Sugar.Errorf("failed to get %s: %s", "foo", err)
	}
	logger.Sugar.Infof("resp: %+v", resp)
}
