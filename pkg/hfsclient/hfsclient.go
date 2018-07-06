package hfsclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jiajunhuang/hfs/pb"
	"github.com/jiajunhuang/hfs/pkg/config"
	"github.com/jiajunhuang/hfs/pkg/logger"
)

func Upload(client pb.ChunkServerClient, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}

	stream, err := client.CreateFile(context.Background())
	if err != nil {
		return err
	}
	buf := make([]byte, config.ChunkSize)
	filePaths := strings.Split(filePath, "/")
	fileName := filePaths[len(filePaths)-1]

	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err := stream.Send(&pb.FileChunkData{Data: buf[:n], Msg: fileName}); err != nil {
			logger.Sugar.Fatalf("failed to send chunk: %s", err)
		}
	}

	createFileResp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	fmt.Printf("file created, uuid is %s\n", createFileResp.File.UUID)

	return nil
}

func Download(client pb.ChunkServerClient, fileUUID string) error {
	fileName := ""

	stream, err := client.ReadFile(context.Background(), &pb.ReadFileRequest{FileUUID: fileUUID})
	if err != nil {
		return err
	}
	f, err := os.OpenFile(fileUUID, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logger.Sugar.Errorf("failed to open file %s: %s", fileUUID, err)
		return err
	}
	defer f.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Sugar.Debugf("failed to read from stream: %s", err)
			return err
		}

		_, err = f.Write(chunk.Data)
		if err != nil {
			logger.Sugar.Fatalf("failed to write chunk of file %s: %s", chunk.Msg, err)
		}

		fileName = chunk.Msg
	}

	fmt.Printf("file with UUID %s download successful! origin file name is %s\n", fileUUID, fileName)

	return nil
}

func Delete(client pb.ChunkServerClient, fileUUID string) error {
	file := pb.File{UUID: fileUUID}
	if _, err := client.RemoveFile(context.Background(), &file); err != nil {
		fmt.Printf("failed to delete file %s: %s\n", file.UUID, err)
		return err
	}

	return nil
}
