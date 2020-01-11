package myfilesystem

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io"
	"kiv_zos/utils"
	"unsafe"
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
		fs.SetInBitmap(true, int32(freeInodeID), fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize())

		// write the actual inode at its address
		fs.SetInodeAt(freeInodeID, inode)
		return freeInodeID
	}
	log.Errorln("No free inode found")
	return -1
}

func (fs *MyFileSystem) SetInodeAt(id NodeID, inode PseudoInode) {
	inodeAddress := fs.GetInodeAddress(id)

	_, err := fs.File.Seek(int64(inodeAddress), io.SeekStart)

	if err == nil {
		err = binary.Write(fs.File, binary.LittleEndian, inode)
		if err != nil {
			log.Error(err)
			panic("could not binary write")
		}
	} else {
		log.Error(err)
		panic("could not seek")
	}
}

func (fs *MyFileSystem) GetInodeAt(id NodeID) PseudoInode {
	inodeAddress := fs.GetInodeAddress(id)

	_, err := fs.File.Seek(int64(inodeAddress), io.SeekStart)

	if err == nil {
		inode := PseudoInode{}
		err = binary.Read(fs.File, binary.LittleEndian, &inode)
		if err != nil {
			log.Error(err)
			panic("could not binary write")
		} else {
			return inode
		}
	} else {
		log.Error(err)
		panic("could not seek")
	}
}

func (fs *MyFileSystem) ClearInodeById(id NodeID) {
	inodeAddress := fs.GetInodeAddress(id)

	fs.SetInBitmap(false, int32(id), inodeAddress, fs.SuperBlock.InodeBitmapSize())
	_, err := fs.File.Seek(int64(inodeAddress), io.SeekStart)
	if err == nil {
		_, err = fs.File.Write(make([]byte, unsafe.Sizeof(PseudoInode{})))
		if err != nil {
			log.Error(err)
			panic("could write empty zeroes to clear an inode")
		}
	} else {
		log.Error(err)
		panic("could not seek")
	}
}

func (fs *MyFileSystem) GetInodeAddress(id NodeID) Address {
	return fs.SuperBlock.InodeStartAddress + Address(Size(id)*Size(unsafe.Sizeof(PseudoInode{})))
}
