package gocache

import "testing"

func TestSizeOf(t *testing.T) {
	var testStruct = struct {
		name  string
		value string
	}{
		name:  "test",
		value: "test",
	}

	size1 := sizeOf(testStruct)
	t.Log(size1)
}
