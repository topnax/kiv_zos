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

	n := fs.GetInodeAt(0)
	inode := &n

	fs.AddDataToInode(data, inode, 0, 0)

	_, _ = fs.File.Seek(int64(fs.GetInodeAt(0).Direct1), io.SeekStart)

	read := [myfilesystem.ClusterSize]byte{}

	_, _ = fs.File.Read(read[:])

	if read != data {
		t.Fatalf("Want=%b got=%b", data, read)
	}

	data = [myfilesystem.ClusterSize]byte{10, 11, 50, 9, 14, 0, 15}

	fs.AddDataToInode(data, inode, 0, 1)

	_, _ = fs.File.Seek(int64(fs.GetInodeAt(0).Direct2), io.SeekStart)

	read = [myfilesystem.ClusterSize]byte{}

	_, _ = fs.File.Read(read[:])

	if read != data {
		t.Fatalf("read2 failed: want=%b got=%b", data, read)
	}

	data = [myfilesystem.ClusterSize]byte{10, 11, 12, 13, 14, 0, 15}

	fs.AddDataToInode(data, inode, 0, 0)

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

	n := fs.GetInodeAt(0)
	inode := &n

	fs.AddDataToInode(data, inode, 0, 5)

	_, _ = fs.File.Seek(int64(fs.GetClusterAddress(inode.Indirect1)), io.SeekStart)

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

	n := fs.GetInodeAt(0)
	inode := &n

	fs.AddDataToInode(data, inode, 0, 0)
	fs.AddDataToInode(data, inode, 0, 260)

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

	n := fs.GetInodeAt(0)
	inode := &n

	fs.AddCluster([myfilesystem.ClusterSize]byte{})
	data1 := [myfilesystem.ClusterSize]byte{10, 11, 12, 13, 14, 0, 15}
	logrus.Warnf("FirstWrite")
	fs.AddDataToInode(data1, inode, 0, 261)
	data2 := [myfilesystem.ClusterSize]byte{10, 11, 12, 19, 10, 0, 15}
	data3 := [myfilesystem.ClusterSize]byte{10, 11, 12, 20, 10, 0, 15}
	data4 := [myfilesystem.ClusterSize]byte{10, 11, 12, 20, 10, 0, 99}
	//data5 := [myfilesystem.ClusterSize]byte{10, 0xFA, 12, 20, 17, 0, 99}
	data5 := [myfilesystem.ClusterSize]byte{1, 1, 1, 1, 1, 1, 1, 1}
	logrus.Warnf("SecondWrite")
	fs.AddDataToInode(data2, inode, 0, 260+256)
	logrus.Warnf("ThirdWrite")
	fs.AddDataToInode(data3, inode, 0, 260+256+1)
	logrus.Warnf("ThirdWrite end")
	fs.AddDataToInode(data4, inode, 0, 260+256+256)
	fs.AddDataToInode(data5, inode, 0, 260+256+256+1)

	if true {
		indirect2Cluster := fs.GetCluster(fs.GetInodeAt(0).Indirect2)

		id := indirect2Cluster.ReadId(2)

		address := fs.GetCluster(id).ReadAddress(0)

		logrus.Infof("data: %v", fs.GetClusterDataAtAddress(address))
		logrus.Infof("data: %v", fs.GetCluster(id).ReadAddress(0))
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

func TestAddAndReadData1(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	inode := myfilesystem.PseudoInode{
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
	}

	if true {
		fs.SetInodeAt(0, inode)

		want := [myfilesystem.ClusterSize]byte{10, 11, 12, 13, 14, 0, 15}

		fs.AddDataToInode(want, &inode, 0, 0)

		got := fs.ReadDataFromInodeAt(fs.GetInodeAt(0), 0)

		if want != got {
			t.Errorf("want=%b, got=%b", want, got)
		}
	}

	if true {
		want := [myfilesystem.ClusterSize]byte{10, 1, 4, 5, 14, 0, 11}

		fs.AddDataToInode(want, &inode, 0, 1)

		got := fs.ReadDataFromInodeAt(fs.GetInodeAt(0), 1)

		if want != got {
			t.Errorf("want=%b, got=%b", want, got)
		}
	}

	if true {
		want := [myfilesystem.ClusterSize]byte{10, 1, 66, 99, 14, 0, 11}

		fs.AddDataToInode(want, &inode, 0, 4)

		got := fs.ReadDataFromInodeAt(fs.GetInodeAt(0), 4)

		if want != got {
			t.Errorf("want=%b, got=%b", want, got)
		}
	}

	fs.Close()
}

func TestAddAndReadDataIndirect(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	inode := myfilesystem.PseudoInode{
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
	}

	id := fs.AddInode(inode)

	if true {
		for i := 0; i < 964; i++ {
			want := [myfilesystem.ClusterSize]byte{byte(i), 11, 12, 13, 14, 0, byte(i)}

			fs.AddDataToInode(want, &inode, id, i)

			got := fs.ReadDataFromInodeAt(fs.GetInodeAt(id), i)

			if want != got {
				t.Errorf("tc=%d, want=%b, got=%b", i, want, got)
			}
		}
	}

	fs.Close()
}

func TestAddAndReadDataIndirectLarge(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)

	fs := myfilesystem.NewMyFileSystem("lgfs")

	fs.Format(150 * 1024 * 1024)

	inode := myfilesystem.PseudoInode{
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
	}

	fs.SetInodeAt(0, inode)

	if true {
		for i := 0; i < 25*1024; i++ {
			want := [myfilesystem.ClusterSize]byte{byte(i), 11, 12, 13, 14, 0, byte(i)}
			fs.AddDataToInode(want, &inode, 0, i)

			got := fs.ReadDataFromInodeAt(fs.GetInodeAt(0), i)
			//
			if want != got {
				t.Errorf("tc=%d, want=%b, got=%b", i, want, got)
			}
		}
	}

	fs.Close()
}

func TestWriteDataToInode(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddInode(myfilesystem.PseudoInode{})

	var want []byte

	for i := 0; i < 1050; i++ {
		want = append(want, byte(i%255))
	}

	fs.WriteDataToInode(id, want)

	node := fs.GetInodeAt(id)

	if node.FileSize != myfilesystem.Size(len(want)) {
		t.Errorf("Invalid node size, want=%d, got=%d", len(want), node.FileSize)
	}

	read := fs.ReadDataFromInode(node)

	if len(read) != len(want) {
		t.Errorf("Invalid read data size, want=%d, got=%d", len(want), len(read))
	} else {
		for i := 0; i < len(read); i++ {
			if read[i] != want[i] {
				t.Errorf("content comparision at=%d failed want=%b, got=%b", i, want[i], read[i])
			}
		}
	}

	fs.Close()
}

func TestGetClusterIdsToBeRemoved(t *testing.T) {
	got := myfilesystem.GetClusterCountToBeRemoved((myfilesystem.ClusterSize)+40, (myfilesystem.ClusterSize))
	want := 1

	if got != want {
		t.Errorf("Got incorrect cluster cound to be removed. Want=%d, got=%d", want, got)
	}
}

func TestGetClusterIdsToBeRemoved2(t *testing.T) {
	got := myfilesystem.GetClusterCountToBeRemoved((myfilesystem.ClusterSize)*5+40, myfilesystem.ClusterSize*2+500)
	want := 3

	if got != want {
		t.Errorf("Got incorrect cluster cound to be removed. Want=%d, got=%d", want, got)
	}
}

func TestGetClusterIdsToBeRemoved3(t *testing.T) {
	got := myfilesystem.GetClusterCountToBeRemoved((myfilesystem.ClusterSize)*5+500, myfilesystem.ClusterSize*5+10)
	want := 0

	if got != want {
		t.Errorf("Got incorrect cluster cound to be removed. Want=%d, got=%d", want, got)
	}
}

func TestShrinkData(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddInode(myfilesystem.PseudoInode{})

	var want []byte

	for i := 0; i < 2500; i++ {
		want = append(want, byte(i%255))
	}

	fs.WriteDataToInode(id, want)

	node := fs.GetInodeAt(id)

	fs.ShrinkInodeData(&node, id, 2000)

	if fs.GetInBitmap(2, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("The second cluster ID should be free")
	}

	fs.Close()
}

func TestShrinkDataWhole(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddInode(myfilesystem.PseudoInode{})

	var want []byte

	for i := 0; i < 2500; i++ {
		want = append(want, byte(i%255))
	}

	fs.WriteDataToInode(id, want)

	node := fs.GetInodeAt(id)

	fs.ShrinkInodeData(&node, id, 0)

	if fs.GetInBitmap(0, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("The first cluster ID should be free")
	}
	if fs.GetInBitmap(1, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("The second cluster ID should be free")
	}
	if fs.GetInBitmap(3, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("The third cluster ID should be free")
	}

	fs.Close()
}

func TestGetUsedClusterAddresses(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	id := fs.AddInode(myfilesystem.PseudoInode{})
	//dirItem := myfilesystem.DirectoryItem{
	//	NodeID: 99,
	//	Name:   [12]rune{'s', 'o', 'u', 'b'},
	//}
	inode := fs.GetInodeAt(id)
	var addresses []myfilesystem.Address
	for i := 0; i < 15; i++ {
		addresses = append(addresses, fs.GetClusterAddress(fs.AddDataToInode([myfilesystem.ClusterSize]byte{1, 1, 123, 1, 1}, &inode, 0, i)))
	}
	inode.FileSize = myfilesystem.Size(15 * myfilesystem.ClusterSize)

	got := fs.GetUsedClusterAddresses(inode)

	for index, address := range addresses {
		if got[index] != address {
			t.Errorf("UsedClusterAddresses failed. Want=%d, got=%d", address, got[index])
		}
	}

	//fs.PrintInfo(inode, dirItem)

	fs.Close()
}
