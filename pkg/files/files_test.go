package files

import (
	"os"
	"testing"
)

func TestCreate(t *testing.T) {
	path := "./test/hello/world"
	defer os.RemoveAll("./test/")

	if _, err := Create(path); err != nil {
		t.Errorf("failed to create file: %s", err)
	}

	os.Chdir("./test/hello")
	defer os.Chdir("../../")
	if _, err := os.Open("world"); err != nil {
		t.Errorf("failed to open file: %s", err)
	}
}
