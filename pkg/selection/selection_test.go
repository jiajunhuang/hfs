package selection

import (
	"testing"
)

func TestRandomSelection(t *testing.T) {
	itself := "192.168.1.2"
	avalable := []string{"192.168.1.1", itself, "192.168.1.3", "192.168.1.4"}

	if len(Random(avalable, itself, 0)) != 0 {
		t.Fatalf("should not get any node return")
	}

	if len(Random(avalable, itself, 1)) != 0 {
		t.Fatalf("should not get any node return")
	}

	var result []string
	result = Random(avalable, itself, 2)
	for _, r := range result {
		if r == itself {
			t.Fatalf("should not select itself")
		}
	}

	result = Random(avalable, itself, 3)
	for _, r := range result {
		if r == itself {
			t.Fatalf("should not select itself")
		}
	}

	result = Random(avalable, itself, 4)
	for _, r := range result {
		if r == itself {
			t.Fatalf("should not select itself")
		}
	}

	result = Random(avalable, itself, 5)
	if len(result) != 3 {
		t.Fatalf("result should be 3 elems, but got: %s", result)
	}
	for _, r := range result {
		if r == itself {
			t.Fatalf("should not select itself")
		}
	}
}
