package myfilesystem

import (
	"github.com/sirupsen/logrus"
	"kiv_zos/myfilesystem"
	"testing"
)

func TestDivision(t *testing.T) {
	for i := 0; i < 1025; i++ {
		logrus.Infof("Result %d", i/1024)
	}
}

func TestGetClusterPath(t *testing.T) {

	fs := myfilesystem.NewMyFileSystem("testfs")

	fs.Format(1 * 1024 * 1024)

	fs.SetClusterAt(myfilesystem.ID(5), [myfilesystem.ClusterSize]byte{10, 10, 10, 12, 15, 18})
	fs.ClearInodeById(5)

	fs.GetInBitmap(5, fs.SuperBlock.ClusterBitmapStartAddress, myfilesystem.Size(fs.SuperBlock.ClusterStartAddress-fs.SuperBlock.ClusterBitmapStartAddress))

	fs.Close()
}
