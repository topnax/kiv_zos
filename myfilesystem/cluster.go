package myfilesystem

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"unsafe"
)

func (fs *MyFileSystem) FindFreeClusterID() ID {
	return fs.FindFreeBitInBitmap(fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterCount)
}

func (fs *MyFileSystem) AddCluster(bytes [ClusterSize]byte) ID {
	freeID := fs.FindFreeClusterID()
	if freeID != -1 {
		// mark in data block bitmap
		fs.SetInBitmap(true, int32(freeID), fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())

		// write the actual inode at its address
		fs.SetClusterAt(freeID, bytes)
		return freeID
	}
	log.Errorln("No free cluster found")
	return -1
}

func (fs *MyFileSystem) SetClusterAt(id ID, data [ClusterSize]byte) {
	clusterAddress := fs.GetClusterAddress(id)

	log.Infof("SetClusterAt writing to address=%d for ID of %d", clusterAddress, id)
	log.Debugf("written %b", data)
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

func (fs *MyFileSystem) GetClusterDataAt(id ID) [ClusterSize]byte {
	inodeAddress := fs.GetClusterAddress(id)

	_, err := fs.File.Seek(int64(inodeAddress), io.SeekStart)

	if err == nil {
		data := [ClusterSize]byte{}
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
	fs.SetInBitmap(false, int32(id), clusterAddress, fs.SuperBlock.ClusterBitmapSize())
}

func (fs *MyFileSystem) GetClusterAddress(id ID) Address {
	return fs.SuperBlock.ClusterStartAddress + Address(Size(id)*Size(ClusterSize))
}

func (fs MyFileSystem) GetCluster(id ID) Cluster {
	clusterAddress := fs.GetClusterAddress(id)
	return Cluster{
		fs:      &fs,
		id:      id,
		address: clusterAddress,
	}
}

func (cluster Cluster) WriteId(id ID, idIndex ID) {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)
	log.Infof("WriteId seeking to %d, about to write %d at ID=%d", cluster.address, id, idIndex)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not seek to start of cluster of ID %d", cluster.id))
	}

	_, err = cluster.fs.File.Seek(int64(unsafe.Sizeof(ID(0)))*int64(idIndex), io.SeekCurrent)

	log.Infof("WriteId seeking from curr %d, about to write %d at ID=%d", int64(unsafe.Sizeof(ID(0)))*int64(idIndex), id, idIndex)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not seek to indirect address address"))
	}

	err = binary.Write(cluster.fs.File, binary.LittleEndian, id)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not do binary write of"))
	}
}

func (cluster Cluster) WriteAddress(address Address, addressId ID) {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)
	log.Infof("WriteAddress seeking to %d, about to write %d at ID=%d", cluster.address, address, addressId)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not seek to start of cluster of ID %d", cluster.id))
	}

	_, err = cluster.fs.File.Seek(int64(unsafe.Sizeof(Address(0)))*int64(addressId), io.SeekCurrent)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not seek to indirect address address"))
	}

	err = binary.Write(cluster.fs.File, binary.LittleEndian, address)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not do binary write of"))
	}
}

func (cluster *Cluster) WriteData(data [ClusterSize]byte) {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not seek to start of cluster of ID %d", cluster.id))
	}

	_, err = cluster.fs.File.Write(data[:])

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not write"))
	}
}

func (cluster Cluster) ReadData() [ClusterSize]byte {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not seek to start of cluster of ID %d", cluster.id))
	}

	var data [ClusterSize]byte
	_, err = cluster.fs.File.Read(data[:])

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not read"))
	}

	return data
}

func (cluster Cluster) ReadAddress(index ID) Address {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)
	log.Infof("ReadAddress seeking to: %d", cluster.address)
	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to start of cluster of ID %d", cluster.id))
	}

	_, err = cluster.fs.File.Seek(int64(unsafe.Sizeof(Address(0)))*int64(index), io.SeekCurrent)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to start of cluster of ID %d and indirect index of =%d", cluster.id, index))
	}

	var foundId Address
	err = binary.Read(cluster.fs.File, binary.LittleEndian, &foundId)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not read"))
	}

	return foundId
}

func (cluster Cluster) ReadId(index ID) ID {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)
	log.Infof("ReadId seeking to: %d", cluster.address)
	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to start of cluster of ID %d", cluster.id))
	}

	_, err = cluster.fs.File.Seek(int64(unsafe.Sizeof(ID(0)))*int64(index), io.SeekCurrent)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to start of cluster of ID %d and index of =%d", cluster.id, index))
	}

	var foundId ID
	err = binary.Read(cluster.fs.File, binary.LittleEndian, &foundId)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not read"))
	}

	return foundId
}
