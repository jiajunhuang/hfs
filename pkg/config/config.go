package config

import (
	"os"
	"strconv"
	"strings"
)

// configurations
var (
	GRPCAddr        = "127.0.0.1:8899"
	ChunkServerName = "hfs-chunk"
	ChunkServerAddr = "127.0.0.1:8899"
	ChunkSize       = 1024 * 1024 * 64 // 64M
	GRPCMaxMsgSize  = ChunkSize + 4096 // 64M + 4K

	EtcdEndpoints = []string{"127.0.0.1:2379"}

	FileBasePath   = "/hfs/files/"
	ChunkBasePath  = "/hfs/chunks/"
	WorkerBasePath = "/hfs/workers/"

	ReplicaNum = 3
)

// Config contains configurations, it will read from process environment, rewrite it with
// values in process environment
func init() {
	if v := os.Getenv("GRPCAddr"); v != "" {
		GRPCAddr = v
	}
	if v := os.Getenv("ChunkServerName"); v != "" {
		ChunkServerName = v
	}
	if v := os.Getenv("ChunkServerAddr"); v != "" {
		ChunkServerAddr = v
	}
	if v := os.Getenv("EtcdEndpoints"); v != "" {
		EtcdEndpoints = strings.Split(v, ",")
	}
	if v := os.Getenv("FileBasePath"); v != "" {
		FileBasePath = v
	}
	if v := os.Getenv("ChunkBasePath"); v != "" {
		ChunkBasePath = v
	}
	if v := os.Getenv("WorkerBasePath"); v != "" {
		WorkerBasePath = v
	}
	if v := os.Getenv("ReplicaNum"); v != "" {
		ReplicaNum, _ = strconv.Atoi(v)
	}
}
