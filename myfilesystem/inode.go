package myfilesystem

import (
	log "github.com/sirupsen/logrus"
	"io"
	"kiv_zos/utils"
)

func (fs *MyFileSystem) FindFreeInodeID() NodeID {
	inodeId := NodeID(0)
	bytes := make([]byte, 1)
	_, _ = fs.File.Seek(int64(fs.SuperBlock.InodeBitmapStartAddress), io.SeekStart)

	inodeCount := fs.SuperBlock.InodeCount()

	for address := fs.SuperBlock.InodeBitmapStartAddress; address < fs.SuperBlock.InodeStartAddress; address += 8 {
		_, _ = fs.File.Read(bytes)
		for index := int8(0); index < 8; index++ {
			if !utils.HasBit(bytes[0], 7-index) {
				return inodeId
			}
			inodeId++
			if Size(inodeId) >= inodeCount {
				return -1
			}
		}
	}
	log.Warnf("Free Inode not found")
	return -1
}

func (fs *MyFileSystem) AddInode(inode PseudoInode) NodeID {
	freeInodeID := fs.FindFreeInodeID()
	if freeInodeID != -1 {
		// mark in inode bitmap

		// write the actual inode at its address
	}
	log.Errorln("No free inode found")
	return -1
}
