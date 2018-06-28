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
	os.MkdirAll(dirPath, 0777)

	os.Chdir(dirPath)
	return os.Create(paths[len(paths)-1])
}

// Remove remove file in the given path
func Remove(path string) error {
	return os.Remove(path)
}

// Append read bytes from r, append it to given file in path
func Append(path string, r io.Reader) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 1024)
	for n, err := r.Read(buf); err == nil; {
		nw, e := f.Write(buf)
		if e != nil {
			return e
		}

		if nw != n {
			logger.Sugar.Infof("failed write %x, read %d bytes, write %d bytes", buf, n, nw)
			return ErrWriteFailed
		}
	}

	return nil
}
