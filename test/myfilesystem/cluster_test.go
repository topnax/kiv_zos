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

		loaded := fs.GetClusterDataAt(myfilesystem.ID(i))

		if data != loaded {
			t.Errorf("Loaded and created are not equal at i=%d", i)
		}
	}

	fs.Close()
}

func TestClearCluster(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.SetClusterAt(myfilesystem.ID(5), [myfilesystem.ClusterSize]byte{10, 10, 10, 12, 15, 18})
	fs.ClearInodeById(5)

	fs.GetInBitmap(5, fs.SuperBlock.ClusterBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterStartAddress-fs.SuperBlock.ClusterBitmapStartAddress))

	fs.Close()
}
