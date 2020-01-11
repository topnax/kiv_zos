package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"kiv_zos/utils"
	"os"
	"testing"
)

func TestSetInBitmap(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.SetInBitmap(true, 0, fs.SuperBlock.DataBitmapStartAddress, fs.SuperBlock.ClusterSize)
	fs.SetInBitmap(true, 10, fs.SuperBlock.DataBitmapStartAddress, fs.SuperBlock.ClusterSize)
	fs.SetInBitmap(true, 20, fs.SuperBlock.DataBitmapStartAddress, fs.SuperBlock.ClusterSize)

	_, _ = fs.File.Seek(int64(fs.SuperBlock.DataBitmapStartAddress), os.SEEK_SET)

	b := make([]byte, 1)
	_, _ = fs.File.Read(b)

	if !utils.HasBit(b[0], 7) {
		t.Errorf("Wanted SET, got EMPTY at 0 in %b", b[0])
	}

	_, _ = fs.File.Seek(int64(fs.SuperBlock.DataBitmapStartAddress)+1, os.SEEK_SET)

	b = make([]byte, 1)
	_, _ = fs.File.Read(b)

	if !utils.HasBit(b[0], 7-(10%8)) {
		t.Errorf("Wanted SET, got EMPTY at 10 in %b", b[0])
	}

	_, _ = fs.File.Seek(int64(fs.SuperBlock.DataBitmapStartAddress)+2, os.SEEK_SET)

	b = make([]byte, 1)
	_, _ = fs.File.Read(b)

	if !utils.HasBit(b[0], 7-(20%8)) {
		t.Errorf("Wanted SET, got EMPTY at 20 in %b", b[0])
	}

	fs.Close()
}

func TestGetInBitmap(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.File.Seek(int64(fs.SuperBlock.DataBitmapStartAddress), os.SEEK_SET)

	b := byte(0xA0)

	fs.File.Write([]byte{b})

	fs.File.Seek(int64(fs.SuperBlock.DataBitmapStartAddress)+2, os.SEEK_SET)

	b = byte(0xD0)

	fs.File.Write([]byte{b})

	if !fs.GetInBitmap(16, fs.SuperBlock.DataBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterCount)) {
		t.Error("Wanted SET, got EMPTY at 17")
	}

	if !fs.GetInBitmap(17, fs.SuperBlock.DataBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterCount)) {
		t.Error("Wanted SET, got EMPTY at 18")
	}

	if !fs.GetInBitmap(19, fs.SuperBlock.DataBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterCount)) {
		t.Error("Wanted SET, got EMPTY at 20")
	}

	fs.Close()
}

func TestGetByteByBitInBitmap(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.File.Seek(int64(fs.SuperBlock.DataBitmapStartAddress), os.SEEK_SET)

	b := byte(0xA0)

	fs.File.Write([]byte{b})

	rb := fs.GetByteByBitInBitmap(4, fs.SuperBlock.DataBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterCount))

	if rb != b {
		t.Errorf("Wanted bit=4 byte=%b, got byte=%b", b, rb)
	}

	fs.File.Seek(int64(fs.SuperBlock.DataBitmapStartAddress)+2, os.SEEK_SET)

	b = byte(0xEA)

	fs.File.Write([]byte{b})

	rb = fs.GetByteByBitInBitmap(18, fs.SuperBlock.DataBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterCount))

	if rb != b {
		t.Errorf("Wanted at bit=18 byte=%b, got byte=%b", b, rb)
	}

	fs.Close()
}
