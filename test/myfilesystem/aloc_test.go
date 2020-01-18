package myfilesystem

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"io"
	"kiv_zos/myfilesystem"
	"testing"
	"unsafe"
)

func TestGetClusterPath(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id, indirect := fs.GetClusterPath(0)
	if id != 0 || indirect != myfilesystem.NoIndirect {
		t.Errorf("GetClusterPath 0 failed want=%d %d, got=%d %d", 0, myfilesystem.NoIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(4)
	if id != 4 || indirect != myfilesystem.NoIndirect {
		t.Errorf("GetClusterPath 4 failed want=%d %d, got=%d %d", 4, myfilesystem.NoIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(5)
	if id != 0 || indirect != myfilesystem.FirstIndirect {
		t.Errorf("GetClusterPath 5 failed want=%d %d, got=%d %d", 6, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(6)
	if id != 1 || indirect != myfilesystem.FirstIndirect {
		t.Errorf("GetClusterPath 6 failed want=%d %d, got=%d %d", 6, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260)
	if id != 255 || indirect != myfilesystem.FirstIndirect {
		t.Errorf("GetClusterPath 260 failed want=%d %d, got=%d %d", 255, myfilesystem.FirstIndirect, id, indirect)
	}

	id, indirect = fs.GetClusterPath(261)
	if id != 0 || indirect != 0 {
		t.Errorf("GetClusterPath 261 failed want=%d %d, got=%d %d", 0, 0, id, indirect)
	}

	id, indirect = fs.GetClusterPath(262)
	if id != 1 || indirect != 0 {
		t.Errorf("GetClusterPath 262 failed want=%d %d, got=%d %d", 1, 0, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 255)
	if id != 254 || indirect != 0 {
		t.Errorf("GetClusterPath 260+255 failed want=%d %d, got=%d %d", 254, 0, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256)
	if id != 255 || indirect != 0 {
		t.Errorf("GetClusterPath 260+256 failed want=%d %d, got=%d %d", 255, 0, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 1)
	if id != 0 || indirect != 1 {
		t.Errorf("GetClusterPath 260+256+1 failed want=%d %d, got=%d %d", 0, 1, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 2)
	if id != 1 || indirect != 1 {
		t.Errorf("GetClusterPath 260+256+2 failed want=%d %d, got=%d %d", 1, 1, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 255)
	if id != 254 || indirect != 1 {
		t.Errorf("GetClusterPath 260+256+255 failed want=%d %d, got=%d %d", 254, 1, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 256)
	if id != 255 || indirect != 1 {
		t.Errorf("GetClusterPath 260+256+256 failed want=%d %d, got=%d %d", 255, 1, id, indirect)
	}

	id, indirect = fs.GetClusterPath(260 + 256 + 256 + 1)
	if id != 0 || indirect != 2 {
		t.Errorf("GetClusterPath 260+256+256+1 failed want=%d %d, got=%d %d", 0, 2, id, indirect)
	}

	id, indirect = fs.GetClusterPath(5 + 256 + 256*256)
	if id != myfilesystem.FileTooLarge || indirect != myfilesystem.FileTooLarge {
		t.Errorf("GetClusterPath 5+256+256*256 failed want=%d %d, got=%d %d", myfilesystem.FileTooLarge, myfilesystem.FileTooLarge, id, indirect)
	}

	fs.Close()
}

func TestSimpleAddData(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.SetInodeAt(0, myfilesystem.PseudoInode{
		IsDirectory: false,
		References:  0,
		FileSize:    0,
		Direct1:     0,
		Direct2:     0,
		Direct3:     0,
		Direct4:     0,
		Direct5:     0,
		Indirect1:   0,
		Indirect2:   0,
	})

	data := [myfilesystem.ClusterSize]byte{10, 11, 12, 13, 14, 0, 15}

	fs.AddDataToInode(data, fs.GetInodeAt(0), 0, 0)

	_, _ = fs.File.Seek(int64(fs.GetInodeAt(0).Direct1), io.SeekStart)

	read := [myfilesystem.ClusterSize]byte{}

	_, _ = fs.File.Read(read[:])

	if read != data {
		t.Errorf("Want=%b got=%b", data, read)
	}

	data = [myfilesystem.ClusterSize]byte{10, 11, 50, 9, 14, 0, 15}

	fs.AddDataToInode(data, fs.GetInodeAt(0), 0, 1)

	_, _ = fs.File.Seek(int64(fs.GetInodeAt(0).Direct2), io.SeekStart)

	read = [myfilesystem.ClusterSize]byte{}

	_, _ = fs.File.Read(read[:])

	if read != data {
		t.Errorf("Want=%b got=%b", data, read)
	}

	data = [myfilesystem.ClusterSize]byte{10, 11, 12, 13, 14, 0, 15}

	fs.AddDataToInode(data, fs.GetInodeAt(0), 0, 0)

	data = [myfilesystem.ClusterSize]byte{10, 11, 50, 9, 14, 0, 15}

	_, _ = fs.File.Seek(int64(fs.GetInodeAt(0).Direct2), io.SeekStart)

	read = [myfilesystem.ClusterSize]byte{}

	_, _ = fs.File.Read(read[:])

	if read != data {
		t.Errorf("Want=%b got=%b", data, read)
	}

	fs.Close()
}

func TestSimpleAddData2(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.SetInodeAt(0, myfilesystem.PseudoInode{
		IsDirectory: false,
		References:  0,
		FileSize:    0,
		Direct1:     0,
		Direct2:     0,
		Direct3:     0,
		Direct4:     0,
		Direct5:     0,
		Indirect1:   0,
		Indirect2:   0,
	})

	data := [myfilesystem.ClusterSize]byte{10, 11, 12, 13, 14, 0, 15}

	fs.AddDataToInode(data, fs.GetInodeAt(0), 0, 5)

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(fs.GetInodeAt(0).Indirect1)), io.SeekStart)

	var address myfilesystem.Address

	_ = binary.Read(fs.File, binary.LittleEndian, &address)

	logrus.Infof("Seeking to %d", address)
	_, _ = fs.File.Seek(int64(address), io.SeekStart)

	read := [myfilesystem.ClusterSize]byte{}

	_, _ = fs.File.Read(read[:])

	if read != data {
		t.Errorf("Want=%b got=%b", data, read)
	}

	fs.Close()
}

func TestSimpleAddData3(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.SetInodeAt(0, myfilesystem.PseudoInode{
		IsDirectory: false,
		References:  0,
		FileSize:    0,
		Direct1:     0,
		Direct2:     0,
		Direct3:     0,
		Direct4:     0,
		Direct5:     0,
		Indirect1:   0,
		Indirect2:   0,
	})

	data := [myfilesystem.ClusterSize]byte{10, 11, 12, 13, 14, 0, 15}

	fs.AddDataToInode(data, fs.GetInodeAt(0), 0, 0)
	fs.AddDataToInode(data, fs.GetInodeAt(0), 0, 260)

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(fs.GetInodeAt(0).Indirect1)), io.SeekStart)
	_, _ = fs.File.Seek(int64(255*unsafe.Sizeof(myfilesystem.Address(0))), io.SeekCurrent)

	var address myfilesystem.Address

	_ = binary.Read(fs.File, binary.LittleEndian, &address)

	logrus.Infof("Seeking to %d", address)
	_, _ = fs.File.Seek(int64(address), io.SeekStart)

	read := [myfilesystem.ClusterSize]byte{}

	_, _ = fs.File.Read(read[:])

	if read != data {
		t.Errorf("Want=%b got=%b", data, read)
	}

	fs.Close()
}

func TestSimpleAddData4(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.SetInodeAt(0, myfilesystem.PseudoInode{
		IsDirectory: false,
		References:  0,
		FileSize:    0,
		Direct1:     0,
		Direct2:     0,
		Direct3:     0,
		Direct4:     0,
		Direct5:     0,
		Indirect1:   0,
		Indirect2:   0,
	})

	fs.AddCluster([myfilesystem.ClusterSize]byte{})
	data1 := [myfilesystem.ClusterSize]byte{10, 11, 12, 13, 14, 0, 15}
	logrus.Warnf("FirstWrite")
	fs.AddDataToInode(data1, fs.GetInodeAt(0), 0, 261)
	data2 := [myfilesystem.ClusterSize]byte{10, 11, 12, 19, 10, 0, 15}
	data3 := [myfilesystem.ClusterSize]byte{10, 11, 12, 20, 10, 0, 15}
	data4 := [myfilesystem.ClusterSize]byte{10, 11, 12, 20, 10, 0, 99}
	data5 := [myfilesystem.ClusterSize]byte{10, 11, 12, 20, 17, 0, 99}
	logrus.Warnf("SecondWrite")
	fs.AddDataToInode(data2, fs.GetInodeAt(0), 0, 260+256)
	logrus.Warnf("ThirdWrite")
	fs.AddDataToInode(data3, fs.GetInodeAt(0), 0, 260+256+1)
	logrus.Warnf("ThirdWrite end")
	fs.AddDataToInode(data4, fs.GetInodeAt(0), 0, 260+256+256)
	fs.AddDataToInode(data5, fs.GetInodeAt(0), 0, 260+256+256+1)

	//
	if true {
		indirect2Cluster := fs.GetCluster(fs.GetInodeAt(0).Indirect2)

		id := indirect2Cluster.ReadId(0)

		address := fs.GetCluster(id).ReadAddress(0)

		logrus.Infof("Read and seeking to address %d", address)
		_, _ = fs.File.Seek(int64(address), io.SeekStart)

		read := [myfilesystem.ClusterSize]byte{}

		_, _ = fs.File.Read(read[:])

		if read != data1 {
			t.Errorf("Want=%b got=%b", data1, read)
		}
	}

	if true {
		indirect2Cluster := fs.GetCluster(fs.GetInodeAt(0).Indirect2)

		id := indirect2Cluster.ReadId(0)

		address := fs.GetCluster(id).ReadAddress(255)

		logrus.Infof("Read and seeking to address %d", address)
		_, _ = fs.File.Seek(int64(address), io.SeekStart)
		//_, _ = fs.File.Seek(int64(unsafe.Sizeof(myfilesystem.Address(0))), io.SeekCurrent)

		read := [myfilesystem.ClusterSize]byte{}

		_, _ = fs.File.Read(read[:])

		if read != data2 {
			t.Errorf("Want=%b got=%b", data2, read)
		}
	}

	if true {
		indirect2Cluster := fs.GetCluster(fs.GetInodeAt(0).Indirect2)

		id := indirect2Cluster.ReadId(1)

		address := fs.GetCluster(id).ReadAddress(0)

		logrus.Infof("Read and seeking to address %d", address)
		_, _ = fs.File.Seek(int64(address), io.SeekStart)

		read := [myfilesystem.ClusterSize]byte{}

		_, _ = fs.File.Read(read[:])

		if read != data3 {
			t.Errorf("Want=%b got=%b", data3, read)
		}
	}

	if true {
		indirect2Cluster := fs.GetCluster(fs.GetInodeAt(0).Indirect2)

		id := indirect2Cluster.ReadId(2)

		address := fs.GetCluster(id).ReadAddress(0)

		logrus.Infof("Read and seeking to address %d", address)
		_, _ = fs.File.Seek(int64(address), io.SeekStart)

		read := [myfilesystem.ClusterSize]byte{}

		_, _ = fs.File.Read(read[:])

		if read != data5 {
			t.Errorf("Want=%b got=%b", data5, read)
		}
	}

	fs.Close()
}
