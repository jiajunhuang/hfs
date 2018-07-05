package main

import (
	"fmt"
	"os"

	"github.com/jiajunhuang/hfs/pb"
	"github.com/jiajunhuang/hfs/pkg/config"
	"github.com/jiajunhuang/hfs/pkg/hfsclient"
	"github.com/jiajunhuang/hfs/pkg/logger"
	"google.golang.org/grpc"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	defer logger.Logger.Sync()
	conn, err := grpc.Dial(config.GRPCAddr, grpc.WithInsecure(), grpc.WithMaxMsgSize(config.GRPCMaxMsgSize))
	if err != nil {
		logger.Sugar.Fatalf("failed to connect to grpc server %s: %s", config.GRPCAddr, err)
	}
	defer conn.Close()

	grpcClient := pb.NewChunkServerClient(conn)

	app := cli.NewApp()
	app.Name = "hfsclient"
	app.Usage = "cli for Huang's Distributed File System"
	app.Commands = []cli.Command{
		{
			Name:  "upload",
			Usage: "upload file",
			Action: func(c *cli.Context) error {
				filePath := c.Args().First()
				if filePath == "" {
					fmt.Printf("Usage: $ hfsclient upload <filepath>\n")
					return nil
				}

				if err := hfsclient.Upload(grpcClient, filePath); err != nil {
					fmt.Printf("failed to upload: %s\n", err)
				}

				return nil
			},
		},
		{
			Name:  "download",
			Usage: "download file",
			Action: func(c *cli.Context) error {
				fileUUID := c.Args().First()
				if fileUUID == "" {
					fmt.Printf("Usage: $ hfsclient download <fileuuid>\n")
					return nil
				}

				if err := hfsclient.Download(grpcClient, fileUUID); err != nil {
					fmt.Printf("failed to download: %s\n", err)
				}

				return nil
			},
		},
		{
			Name:  "delete",
			Usage: "delete file",
			Action: func(c *cli.Context) error {
				fileUUID := c.Args().First()
				if fileUUID == "" {
					fmt.Printf("Usage: $ hfsclient download <fileuuid>\n")
					return nil
				}

				if err := hfsclient.Delete(grpcClient, fileUUID); err != nil {
					fmt.Printf("failed to download: %s\n", err)
				}

				return nil
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		logger.Sugar.Fatal(err)
	}
}
