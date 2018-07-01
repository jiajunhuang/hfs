package main

import (
	"os"

	"github.com/jiajunhuang/hfs/pkg/chunkserver"
	"github.com/jiajunhuang/hfs/pkg/logger"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	defer logger.Logger.Sync()

	app := cli.NewApp()
	app.Name = "chunkserver"
	app.Usage = "Chunkserver for Huang's Distributed File System"
	app.Action = func(c *cli.Context) error {
		chunkserver.StartChunkServer()
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Sugar.Fatal(err)
	}
}
