package myfilesystem

import (
	"encoding/binary"
	"io"
	"kiv_zos/myfilesystem"
	"testing"
	"unsafe"
)

func TestFindFreeCluster(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(10 * 1024 * 1024)

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

	fs.GetBitInBitmap(5, fs.SuperBlock.ClusterBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterStartAddress-fs.SuperBlock.ClusterBitmapStartAddress))

	fs.Close()
}

func TestSetId(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddCluster([myfilesystem.ClusterSize]byte{})

	want := myfilesystem.ID(99)

	fs.GetCluster(id).WriteId(want, 0)

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(id)), io.SeekStart)

	var read myfilesystem.ID

	binary.Read(fs.File, binary.LittleEndian, &read)

	if read != want {
		t.Errorf("want=%d, read=%d", want, read)
	}

	want = myfilesystem.ID(150)

	fs.GetCluster(id).WriteId(want, 50)

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(id)), io.SeekStart)
	_, _ = fs.File.Seek(int64(unsafe.Sizeof(myfilesystem.ID(0))*50), io.SeekCurrent)

	binary.Read(fs.File, binary.LittleEndian, &read)

	if read != want {
		t.Errorf("want=%d, read=%d", want, read)
	}

	fs.Close()
}

func TestReadId(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddCluster([myfilesystem.ClusterSize]byte{})

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(id)), io.SeekStart)

	want := myfilesystem.ID(99)
	_ = binary.Write(fs.File, binary.LittleEndian, want)

	got := fs.GetCluster(id).ReadId(0)

	if want != got {
		t.Errorf("want=%d, got=%d", want, got)
	}

	fs.Close()
}

func TestSetAndReadId(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddCluster([myfilesystem.ClusterSize]byte{})

	fs.GetCluster(id).WriteId(99, 0)

	read := fs.GetCluster(id).ReadId(0)
	if read != 99 {
		t.Errorf("TestSetAndReadId failed want=%d, got=%d", 99, read)
	}

	fs.Close()
}

func TestSetAddress(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddCluster([myfilesystem.ClusterSize]byte{})

	want := myfilesystem.Address(99)

	fs.GetCluster(id).WriteAddress(want, 0)

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(id)), io.SeekStart)

	var read myfilesystem.Address

	binary.Read(fs.File, binary.LittleEndian, &read)

	if read != want {
		t.Errorf("want=%d, read=%d", want, read)
	}

	want = myfilesystem.Address(150)

	fs.GetCluster(id).WriteAddress(want, 50)

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(id)), io.SeekStart)
	_, _ = fs.File.Seek(int64(unsafe.Sizeof(myfilesystem.ID(0))*50), io.SeekCurrent)

	binary.Read(fs.File, binary.LittleEndian, &read)

	if read != want {
		t.Errorf("want=%d, read=%d", want, read)
	}

	fs.Close()
}

func TestReadAddress(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddCluster([myfilesystem.ClusterSize]byte{})

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(id)), io.SeekStart)

	want := myfilesystem.Address(99)
	_ = binary.Write(fs.File, binary.LittleEndian, want)

	got := fs.GetCluster(id).ReadAddress(0)

	if want != got {
		t.Errorf("want=%d, got=%d", want, got)
	}

	fs.Close()
}

func TestSetAndReadAddress(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddCluster([myfilesystem.ClusterSize]byte{})

	fs.GetCluster(id).WriteAddress(99, 0)

	read := fs.GetCluster(id).ReadAddress(0)
	if read != 99 {
		t.Errorf("TestSetAndReadAddress failed want=%d, got=%d", 99, read)
	}

	fs.Close()
}

func TestGetIdByAddress(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	for i := 0; i < 50; i++ {
		want := fs.AddCluster([myfilesystem.ClusterSize]byte{})
		got := fs.GetClusterId(fs.GetClusterAddress(want))
		if got != want {
			t.Errorf("Address to ID conversion failed. Want=%d, got=%d", want, got)
		}
	}

	fs.Close()
}
