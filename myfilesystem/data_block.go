package myfilesystem

import (
	log "github.com/sirupsen/logrus"
	"io"
)

func (fs *MyFileSystem) FindFreeDataBlockID() ID {
	return fs.FindFreeBitInBitmap(fs.SuperBlock.DataBitmapStartAddress, fs.SuperBlock.ClusterCount)
}

func (fs *MyFileSystem) AddDataBlock(bytes [clusterSize]byte) ID {
	freeID := fs.FindFreeDataBlockID()
	if freeID != -1 {
		// mark in data block bitmap
		fs.SetInBitmap(true, int32(freeID), fs.SuperBlock.DataBitmapStartAddress, fs.SuperBlock.DataBitmapSize())

		// write the actual inode at its address
		fs.SetDataAt(freeID, bytes)
		return freeID
	}
	log.Errorln("No free data block found")
	return -1
}

func (fs *MyFileSystem) SetDataAt(id ID, data [clusterSize]byte) {
	clusterAddress := fs.GetClusterAddress(id)

	_, err := fs.File.Seek(int64(clusterAddress), io.SeekStart)

	if err == nil {
		_, err = fs.File.Write(data[:])
		if err != nil {
			log.Error(err)
			panic("could not binary write")
		}
	} else {
		log.Error(err)
		panic("could not seek")
	}
}

func (fs *MyFileSystem) GetDataAt(id ID) [clusterSize]byte {
	inodeAddress := fs.GetClusterAddress(id)

	_, err := fs.File.Seek(int64(inodeAddress), io.SeekStart)

	if err == nil {
		data := [clusterSize]byte{}
		_, err = fs.File.Read(data[:])
		if err != nil {
			log.Error(err)
			panic("could not binary write")
		} else {
			return data
		}
	} else {
		log.Error(err)
		panic("could not seek")
	}
}

func (fs *MyFileSystem) ClearClusterById(id ID) {
	clusterAddress := fs.GetClusterAddress(id)

	fs.SetInBitmap(false, int32(id), clusterAddress, fs.SuperBlock.DataBitmapSize())
	_, err := fs.File.Seek(int64(clusterAddress), io.SeekStart)
	if err == nil {
		_, err = fs.File.Write(make([]byte, clusterSize))
		if err != nil {
			log.Error(err)
			panic("could write empty zeroes to clear an inode")
		}
	} else {
		log.Error(err)
		panic("could not seek")
	}
}

func (fs *MyFileSystem) GetClusterAddress(id ID) Address {
	return fs.SuperBlock.DataStartAddress + Address(Size(id)*Size(clusterSize))
}
