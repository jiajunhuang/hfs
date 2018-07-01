package chunkserver

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/jiajunhuang/hfs/pkg/logger"
)

func StartWorker(name, ip string, etcdClient *clientv3.Client) {
	kvClient := clientv3.NewKV(etcdClient)

	for {
		lease := clientv3.NewLease(etcdClient)
		grantResp, err := lease.Grant(context.TODO(), 10)
		if err != nil {
			logger.Sugar.Errorf("failed to grant lease: %s", err)
			continue
		}
		_, err = kvClient.Put(context.Background(), "/workers/"+ip, name, clientv3.WithLease(grantResp.ID))
		if err != nil {
			logger.Sugar.Errorf("failed to put %s to %s: %s", ip, name, err)
		} else {
			logger.Sugar.Infof("refresh ip %s to worker %s in KV %+v", ip, name, kvClient)
		}
		time.Sleep(time.Second * 3)
	}
}

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

	StartWorker("idea", "127.0.0.1", etcdClient)
}
