package chunkserver

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/jiajunhuang/hfs/pkg/logger"
)

type ChunkServer struct {
	name       string
	ip         string
	etcdClient *clientv3.Client
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
		_, err = kvClient.Put(context.Background(), "/workers/"+s.ip, s.name, clientv3.WithLease(grantResp.ID))
		if err != nil {
			logger.Sugar.Errorf("failed to put %s to %s: %s", s.ip, s.name, err)
		} else {
			logger.Sugar.Infof("refresh ip %s to worker %s in KV %+v", s.ip, s.name, kvClient)
		}
		time.Sleep(time.Second * 3)
	}
}

// StartChunkServer works as it's name
func StartChunkServer() {
	etcdClient, err := clientv3.New(
		clientv3.Config{
			Endpoints:   []string{"127.0.0.1:2379"},
			DialTimeout: 2 * time.Second,
		},
	)

	if err != nil {
		logger.Sugar.Fatalf("failed to connect to etcd: %s", err)
	}

	defer etcdClient.Close()

	chunkServer := ChunkServer{"idea", "127.0.0.1", etcdClient}
	chunkServer.KeepAlive()
}
