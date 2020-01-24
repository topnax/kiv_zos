package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
	"unsafe"
)

func TestSomething(t *testing.T) {
	ds := 20 * unsafe.Sizeof(myfilesystem.DirectoryItem{})
	t.Errorf("Total size: %d", ds)
	t.Errorf("Sizeof DI: %d", unsafe.Sizeof(myfilesystem.DirectoryItem{}))
	t.Errorf("Clustersize: %d", myfilesystem.ClusterSize)
	t.Errorf("%d", ds/myfilesystem.ClusterSize)
	t.Errorf("%d", ds%myfilesystem.ClusterSize)
	cond := (ds%myfilesystem.ClusterSize) < unsafe.Sizeof(myfilesystem.DirectoryItem{}) && ds%myfilesystem.ClusterSize != 0
	t.Errorf("%v", cond)

	if cond {

		taken := (myfilesystem.ClusterSize / unsafe.Sizeof(myfilesystem.DirectoryItem{})) * unsafe.Sizeof(myfilesystem.DirectoryItem{})
		t.Errorf("Start at %d", taken)
		t.Errorf("Read %d", myfilesystem.ClusterSize-taken)
		t.Errorf("Then at next %d", ds%myfilesystem.ClusterSize)
	}
	//if false {
	//	t.Error(unsafe.Sizeof(myfilesystem.DirectoryItem{}))
	//}
}
