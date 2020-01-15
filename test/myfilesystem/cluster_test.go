package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
)

func TestFindFreeCluster(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	for i := 0; i < int(fs.SuperBlock.ClusterCount); i++ {
		id := fs.FindFreeClusterID()
		want := myfilesystem.ID(i)
		if id != want {
			t.Errorf("Expected different free inode id. wanted=%d got=%d", want, id)
			return
		}
		fs.SetInBitmap(true, int32(i), fs.SuperBlock.ClusterBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterStartAddress-fs.SuperBlock.ClusterBitmapStartAddress))
	}

	fs.Close()
}

func TestSetAndGetCluster(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	for i := 0; i < 30; i++ {
		data := [myfilesystem.ClusterSize]byte{11, 12, 13, 31, 23, 43, 12, 51, 0xAA}

		fs.SetClusterAt(myfilesystem.ID(i), data)

		loaded := fs.GetClusterAt(myfilesystem.ID(i))

		if data != loaded {
			t.Errorf("Loaded and created are not equal at i=%d", i)
		}
	}

	fs.Close()
}

func TestClearCluster(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id, indirect := fs.GetClusterPath(0)
	if id != 0 && indirect != myfilesystem.NoIndirect {
		t.Errorf("ClearCluster failed want=%d %d, got=%d %d", 0, myfilesystem.NoIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(4)
	if id != 4 && indirect != myfilesystem.NoIndirect {
		t.Errorf("ClearCluster failed want=%d %d, got=%d %d", 4, myfilesystem.NoIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(5)
	if id != 5 && indirect != myfilesystem.FirstIndirect {
		t.Errorf("ClearCluster failed want=%d %d, got=%d %d", 6, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(6)
	if id != 6 && indirect != myfilesystem.FirstIndirect {
		t.Errorf("ClearCluster failed want=%d %d, got=%d %d", 6, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260)
	if id != 255 && indirect != myfilesystem.FirstIndirect {
		t.Errorf("ClearCluster failed want=%d %d, got=%d %d", 255, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(261)
	if id != 0 && indirect < 0 {
		t.Errorf("ClearCluster failed want=%d %d, got=%d %d", id, 0, id, indirect)
	}

	fs.Close()
}
