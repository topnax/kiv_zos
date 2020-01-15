package myfilesystem

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"os"
	"unsafe"
)

const (
	inodeRatio  = 0.05
	clusterSize = 1024
)

func (fs *MyFileSystem) Format(desiredFsSize int) {
	log.Infof("About to format a volume of desiredFsSize %d bytes, %d kB, %d MB", desiredFsSize, desiredFsSize/1024, desiredFsSize/1024/1024)

	fs.SuperBlock = SuperBlock{
		Signature:        [8]rune{'k', 'r', 'a', 'l', 's', 't'},
		VolumeDescriptor: [251]rune{'m', 'y', 'f', 's'},
		ClusterSize:      clusterSize,
	}

	fs.SuperBlock.init(desiredFsSize)

	fs.SuperBlock.info()

	file, err := os.Create(fs.filePath)

	if err == nil {
		err := file.Truncate(int64(fs.SuperBlock.DiskSize))

		if err == nil {
			_, err = file.Seek(0, io.SeekStart)
			if err == nil {
				err = binary.Write(file, binary.LittleEndian, fs.SuperBlock)
				if err != nil {
					log.Errorf("Could not write SB at a File '%s'", fs.filePath)
					log.Error(err)
				} else {
					fs.File = file
				}
			} else {
				log.Errorf("Could not seek at a File at '%s' at SEEK_SET of 0", fs.filePath)
				log.Error(err)
			}
		} else {
			log.Errorf("Could not truncate a File at '%s' of size %d kB", fs.filePath, desiredFsSize/1024)
			log.Error(err)
		}
	} else {
		log.Errorf("Could not create a File at '%s' of size %d kB", fs.filePath, desiredFsSize/1024)
		log.Error(err)
	}
}

func (superBlock *SuperBlock) init(desiredFsSize int) {
	// maximal hypothetical inode block size, without inode bitmap
	maximalInodeBlockSize := Size(math.Floor(float64(Size(desiredFsSize)-Size(unsafe.Sizeof(SuperBlock{}))) * inodeRatio))

	// count of inodes that fit into the hypothetical maximal inode block size
	inodeCount := Size(math.Floor(float64(maximalInodeBlockSize / Size(unsafe.Sizeof(PseudoInode{})))))

	// real inode block size, including an inode bitmap
	inodeBlockSize := inodeCount * (Size(unsafe.Sizeof(PseudoInode{}) + 1))

	// maximal hypothetical cluster block size
	maximalClusterBlockSize := Size(desiredFsSize) - Size(unsafe.Sizeof(SuperBlock{})) - inodeBlockSize

	// count of clusters (including one bit for an entry in cluster bitmap) that fit into the hypothetical cluster block size
	clusterCount := math.Floor(float64(maximalClusterBlockSize / (superBlock.ClusterSize + 1)))

	// real cluster block size, including one bit for an entry in cluster bitmap
	clusterBlockSize := Size(clusterCount) * (superBlock.ClusterSize + 1)

	log.Infoln("Preview:")
	log.Infoln("Inodes count:", inodeCount)
	log.Infoln("Cluster count:", clusterCount)
	log.Infoln("Superblock size:", unsafe.Sizeof(SuperBlock{}))
	log.Infoln("Inode area size:", inodeBlockSize)
	log.Infoln("Cluster area size:", clusterBlockSize)
	total := Size(unsafe.Sizeof(SuperBlock{})) + inodeBlockSize + clusterBlockSize
	log.Infoln("Total:", total)
	log.Infoln("Desired:", desiredFsSize)
	log.Infoln("Diff:", Size(desiredFsSize)-total)

	/*
		SUPERBLOCK | INODE BITMAP | INODES | DATA BITMAP | DATA
	*/

	superBlock.DiskSize = total
	superBlock.ClusterCount = Size(clusterCount)

	superBlock.InodeBitmapStartAddress = Address(unsafe.Sizeof(SuperBlock{}))
	superBlock.InodeStartAddress = superBlock.InodeBitmapStartAddress + Address(inodeCount)

	superBlock.DataBitmapStartAddress = superBlock.InodeStartAddress + Address(inodeCount*Size(unsafe.Sizeof(PseudoInode{})))
	superBlock.DataStartAddress = superBlock.DataBitmapStartAddress + Address(clusterCount)

	log.Infoln("Calculated inode count:", superBlock.InodeCount())

}
