package files

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/jiajunhuang/hfs/pkg/logger"
)

// error definitions
var (
	ErrWriteFailed = errors.New("failed to write same length bytes as read")
)

// Create create a file at path, with all the directory
// same as mkdir -p && touch
func Create(path string) (*os.File, error) {
	paths := strings.Split(path, "/")

	dirPath := strings.Join(paths[:len(paths)-1], "/")
	os.MkdirAll(dirPath, 0700)

	logger.Sugar.Debugf("create file: %s", path)
	return os.Create(path)
}

// Remove remove file in the given path
func Remove(path string) error {
	logger.Sugar.Debugf("remove file: %s", path)
	return os.Remove(path)
}

// Append read bytes from r, append it to given file in path
func Append(path string, r io.Reader) error {
	logger.Sugar.Debugf("append file: %s", path)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
	if err != nil {
		return err
	}
	defer f.Close()

	bufSize := 1024
	buf := make([]byte, bufSize)

	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		nw, e := f.Write(buf[:n])
		logger.Sugar.Debugf("Append: read %d bytes, write %d bytes, read err: %s, write err: %s", n, nw, err, e)
		if e != nil {
			return e
		}
	}

	return nil
}
