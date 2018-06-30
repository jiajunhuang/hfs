package files

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	path := "./test/hello/world"
	defer os.RemoveAll("./test/")

	f, err := Create(path)
	if err != nil {
		t.Errorf("failed to create file: %s", err)
	}
	f.Close()

	f, err = os.Open(path)
	if err != nil {
		t.Errorf("failed to open file: %s", err)
	}
	f.Close()
}

func TestRemove(t *testing.T) {
	path := "./test/hello/world"
	defer os.RemoveAll("./test/")

	f, err := Create(path)
	if err != nil {
		t.Errorf("failed to create file: %s", err)
	}
	f.Close()

	err = Remove(path)
	if err != nil {
		t.Fatalf("failed to remove file: %s", err)
	}

	f, err = os.Open(path)
	defer f.Close()

	if err == nil {
		t.Errorf("should failed to open file but not")
	}
}

func TestAppend(t *testing.T) {
	// if file not exist
	path := "./world"
	data := "hello"
	defer os.Remove(path)

	Append(path, strings.NewReader(data))

	if b, err := ioutil.ReadFile(path); err != nil {
		t.Fatalf("should read file %s success but got error: %s", path, err)
	} else {
		if string(b) != data {
			t.Fatalf("bytes from file %s(%s) not equal to content %s", path, b, data)
		}
	}

	// append
	Append(path, strings.NewReader("world"))
	if b, err := ioutil.ReadFile(path); err != nil {
		t.Fatalf("should read file %s success but got error: %s", path, err)
	} else {
		if string(b) != "helloworld" {
			t.Fatalf("bytes from file %s(%s) not equal to content %s", path, b, data)
		}
	}
}
