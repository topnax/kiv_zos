package myfilesystem

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io"
	"unsafe"
)

// finds an free inode ID
func (fs *MyFileSystem) FindFreeInodeID() ID {
	return fs.FindFreeBitInBitmap(fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeCount())
}

func (fs *MyFileSystem) AddInode(inode PseudoInode) ID {
	freeInodeID := fs.FindFreeInodeID()
	if freeInodeID != -1 {
		// mark in inode bitmap
		fs.SetInBitmap(true, int32(freeInodeID), fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize())

		// write the actual inode at its address
		fs.SetInodeAt(freeInodeID, inode)
		return freeInodeID
	}
	log.Warnln("No free inode found")
	return -1
}

// sets the given inode at the given ID
func (fs *MyFileSystem) SetInodeAt(id ID, inode PseudoInode) {
	inodeAddress := fs.GetInodeAddress(id)
	log.Infof("Setting an inode at address %d", inodeAddress)

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

// reads an inode at the given ID
func (fs *MyFileSystem) GetInodeAt(id ID) PseudoInode {
	inodeAddress := fs.GetInodeAddress(id)

	_, err := fs.File.Seek(int64(inodeAddress), io.SeekStart)

	if err == nil {
		inode := PseudoInode{}
		err = binary.Read(fs.File, binary.LittleEndian, &inode)
		if err != nil {
			log.Error(err)
			panic("could not binary read")
		} else {
			return inode
		}
	} else {
		log.Error(err)
		panic("could not seek")
	}
}

// clears an inode by ID
func (fs *MyFileSystem) ClearInodeById(id ID) {
	inodeAddress := fs.GetInodeAddress(id)

	fs.SetInBitmap(false, int32(id), fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize())
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

// converts inode ID to an inode address
func (fs *MyFileSystem) GetInodeAddress(id ID) Address {
	return fs.SuperBlock.InodeStartAddress + Address(Size(id)*Size(unsafe.Sizeof(PseudoInode{})))
}
