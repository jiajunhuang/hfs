package config

import (
	"os"
	"strings"
)

// configurations
var (
	GRPCAddr          = "127.0.0.1:8899"
	ChunkServerName   = "hfs-chunk"
	ChunkServerIPAddr = "127.0.0.1"

	EtcdEndpoints = []string{"127.0.0.1:2379"}

	FileBasePath   = "/hfs/files/"
	ChunkBasePath  = "/hfs/chunks/"
	WorkerBasePath = "/hfs/workers/"
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
	if v := os.Getenv("ChunkServerIPAddr"); v != "" {
		ChunkServerIPAddr = v
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
}
