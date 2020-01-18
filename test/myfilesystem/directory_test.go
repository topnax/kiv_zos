package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
	"unsafe"
)

func TestSomething(t *testing.T) {
	if false {
		t.Error(unsafe.Sizeof(myfilesystem.DirectoryItem{}))
	}
}
