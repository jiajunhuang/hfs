package files

import (
	"os"
	"strings"
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
