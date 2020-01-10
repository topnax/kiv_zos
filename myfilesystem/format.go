package myfilesystem

import (
	log "github.com/sirupsen/logrus"
	"math"
	"unsafe"
)

const (
	inodeRatio  = 0.05
	clusterSize = 1024
)

func (fs *MyFileSystem) Format(desiredFsSize int) {
	log.Infof("About to format a volume of desiredFsSize %d bytes, %d kB, %d MB", desiredFsSize, desiredFsSize/1024, desiredFsSize/1024/1024)

	fs.superBlock = SuperBlock{
		signature:        [8]rune{'k', 'r', 'a', 'l', 's', 't'},
		volumeDescriptor: [251]rune{'m', 'y', 'f', 's'},
		clusterSize:      clusterSize,
	}

	// maximal hypothetical inode block size, without inode bitmap
	maximalInodeBlockSize := Size(math.Floor(float64(Size(desiredFsSize)-Size(unsafe.Sizeof(fs.superBlock))) * inodeRatio))

	// count of inodes that fit into the hypothetical maximal inode block size
	inodeCount := Size(math.Floor(float64(maximalInodeBlockSize / Size(unsafe.Sizeof(PseudoInode{})))))

	// real inode block size, including an inode bitmap
	inodeBlockSize := Size(inodeCount) * (Size(unsafe.Sizeof(PseudoInode{}) + 1))

	// maximal hypothetical cluster block size
	maximalClusterBlockSize := Size(desiredFsSize) - Size(unsafe.Sizeof(fs.superBlock)) - inodeBlockSize

	// count of clusters (including one bit for an entry in cluster bitmap) that fit into the hypothetical cluster block size
	clusterCount := math.Floor(float64(maximalClusterBlockSize / (fs.superBlock.clusterSize + 1)))

	// real cluster block size, including one bit for an entry in cluster bitmap
	clusterBlockSize := Size(clusterCount) * (fs.superBlock.clusterSize + 1)

	log.Infoln("Preview:")
	log.Infoln("Inodes count:", Size(inodeCount))
	log.Infoln("Cluster count:", clusterCount)
	log.Infoln("Superblock size:", unsafe.Sizeof(fs.superBlock))
	log.Infoln("Inode area size:", inodeBlockSize)
	log.Infoln("Cluster area size:", clusterBlockSize)
	total := Size(unsafe.Sizeof(fs.superBlock)) + inodeBlockSize + clusterBlockSize
	log.Infoln("Total:", total)
	log.Infoln("Desired:", desiredFsSize)
	log.Infoln("Diff:", Size(desiredFsSize)-total)

	fs.superBlock.diskSize = total
	fs.superBlock.clusterCount = ClusterCount(clusterCount)

	fs.superBlock.inodeBitmapStartAddress = Address(unsafe.Sizeof(fs.superBlock))
	fs.superBlock.inodeStartAddress = fs.superBlock.inodeBitmapStartAddress + Address(inodeCount)

	fs.superBlock.dataBitmapStartAddress = fs.superBlock.inodeStartAddress + Address(inodeCount*Size(unsafe.Sizeof(PseudoInode{})))
	fs.superBlock.dataStartAddress = fs.superBlock.dataBitmapStartAddress + Address(clusterCount)

	fs.superBlock.info()

	// TODO add math above to a func of a superblock

}
