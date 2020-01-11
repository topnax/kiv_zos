package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
)

func TestFindFreeInode(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	for i := 0; i < int(fs.SuperBlock.InodeCount()); i++ {
		id := fs.FindFreeInodeID()
		want := myfilesystem.NodeID(i)
		if id != want {
			t.Errorf("Expected different free inode id. wanted=%d got=%d", want, id)
			return
		}
		fs.SetInBitmap(true, int32(i), fs.SuperBlock.InodeBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.InodeStartAddress-fs.SuperBlock.InodeBitmapStartAddress))
	}

	fs.Close()
}
