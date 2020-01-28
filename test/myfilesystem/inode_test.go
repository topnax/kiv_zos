package myfilesystem

import (
	"io"
	"kiv_zos/myfilesystem"
	"testing"
	"unsafe"
)

func TestFindFreeInode(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	for i := 0; i < int(fs.SuperBlock.InodeCount()); i++ {
		id := fs.FindFreeInodeID()
		want := myfilesystem.ID(i)
		if id != want {
			t.Errorf("Expected different free inode id. wanted=%d got=%d", want, id)
			return
		}
		fs.SetInBitmap(true, int32(i), fs.SuperBlock.InodeBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.InodeStartAddress-fs.SuperBlock.InodeBitmapStartAddress))
	}

	fs.Close()
}

func TestSetAndGetInode(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	for i := 0; i < 30; i++ {
		inode := myfilesystem.PseudoInode{
			IsDirectory: true,
			References:  myfilesystem.ReferenceCounter(i),
			FileSize:    10,
			Direct1:     10,
			Direct2:     20,
			Direct3:     30 * myfilesystem.Address(i),
			Direct4:     40,
			Direct5:     50,
			Indirect1:   60,
			Indirect2:   150,
		}

		fs.SetInodeAt(myfilesystem.ID(i), inode)

		loaded := fs.GetInodeAt(myfilesystem.ID(i))

		if inode != loaded {
			t.Errorf("Loaded and created are not equal at i=%d", i)
		}
	}

	fs.Close()
}

func TestClearInode(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	inode := myfilesystem.PseudoInode{
		IsDirectory: true,
		References:  myfilesystem.ReferenceCounter(10),
		FileSize:    10,
		Direct1:     10,
		Direct2:     20,
		Direct3:     30 * myfilesystem.Address(34),
		Direct4:     40,
		Direct5:     50,
		Indirect1:   60,
		Indirect2:   150,
	}

	fs.SetInodeAt(myfilesystem.ID(5), inode)
	fs.SetInBitmap(true, int32(5), fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize())

	if !fs.GetBitInBitmap(5, fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize()) {
		t.Errorf("Fifth bit should be set")
	}
	fs.ClearInodeById(5)
	if fs.GetBitInBitmap(5, fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize()) {
		t.Errorf("Fifth bit should not be set")
	}

	_, _ = fs.File.Seek(int64(fs.GetInodeAddress(5)), io.SeekStart)

	bytes := make([]byte, unsafe.Sizeof(myfilesystem.PseudoInode{}))

	fs.File.Read(bytes)

	for index, byte := range bytes {
		if byte != 0 {
			t.Errorf("inode not cleared at index=%d", index)
		}
	}

	fs.Close()
}
