package myfilesystem

import (
	"kiv_zos/myfilesystem"
	"testing"
)

func TestGetAndSet(t *testing.T) {
	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.SetInBitmap(true, 2, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())
	fs.SetInBitmap(true, 3, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())
	fs.SetInBitmap(true, 4, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())
	fs.SetInBitmap(true, 6, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())
	fs.SetInBitmap(true, 18, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())
	fs.SetInBitmap(true, 25, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())
	fs.SetInBitmap(true, 26, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())
	fs.SetInBitmap(true, 27, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())

	if !fs.GetInBitmap(2, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("wanted=SET got=FALSE @2")
	}
	if !fs.GetInBitmap(3, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("wanted=SET got=FALSE @3")
	}
	if !fs.GetInBitmap(4, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("wanted=SET got=FALSE @4")
	}
	if !fs.GetInBitmap(6, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("wanted=SET got=FALSE @6")
	}
	if !fs.GetInBitmap(18, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("wanted=SET got=FALSE @18")
	}
	if !fs.GetInBitmap(25, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("wanted=SET got=FALSE @25")
	}
	if !fs.GetInBitmap(26, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("wanted=SET got=FALSE @26")
	}
	if !fs.GetInBitmap(27, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize()) {
		t.Errorf("wanted=SET got=FALSE @27")
	}

	fs.Close()
}

func TestFindFreeBitsInBitmap(t *testing.T) {
	ids := myfilesystem.FindFreeBitsInBitmap(16, []byte{0x0A, 0, 255, 0xD7})

	if len(ids) != 16 {
		t.Fatalf("The size of found free bits should be 16. got=%d", len(ids))
	}

	if ids[0] != 0 {
		t.Errorf("0th id should be 0, got=%d", ids[0])
	}
	if ids[1] != 1 {
		t.Errorf("1st id should be 1, got=%d", ids[1])
	}
	if ids[2] != 2 {
		t.Errorf("2d id should be 2, got=%d", ids[2])
	}
	if ids[3] != 3 {
		t.Errorf("3d id should be 3, got=%d", ids[3])
	}
	if ids[4] != 5 {
		t.Errorf("4th id should be 5, got=%d", ids[4])
	}
	if ids[5] != 7 {
		t.Errorf("5th id should be 7, got=%d", ids[5])
	}

	if ids[6] != 8 {
		t.Errorf("6th id should be 8, got=%d", ids[6])
	}
	if ids[7] != 9 {
		t.Errorf("7st id should be 9, got=%d", ids[7])
	}
	if ids[8] != 10 {
		t.Errorf("8th id should be 10, got=%d", ids[8])
	}
	if ids[9] != 11 {
		t.Errorf("9th id should be 11, got=%d", ids[9])
	}
	if ids[10] != 12 {
		t.Errorf("10th id should be 12, got=%d", ids[10])
	}
	if ids[11] != 13 {
		t.Errorf("11th id should be 13, got=%d", ids[11])
	}
	if ids[12] != 14 {
		t.Errorf("12th id should be 14, got=%d", ids[12])
	}
	if ids[13] != 15 {
		t.Errorf("13th id should be 15, got=%d", ids[13])
	}
	if ids[14] != 26 {
		t.Errorf("14th id should be 26, got=%d", ids[14])
	}
	if ids[15] != 28 {
		t.Errorf("15th id should be 28, got=%d", ids[15])
	}

}
