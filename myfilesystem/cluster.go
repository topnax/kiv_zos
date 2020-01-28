package myfilesystem

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"unsafe"
)

// finds a free cluster id. if no id is in the cached, new search for ids is made
func (fs *MyFileSystem) FindFreeClusterID() ID {
	if fs.freeClusterIdIndex < len(fs.freeClusterIds)-1 {
		id := fs.freeClusterIds[fs.freeClusterIdIndex]
		fs.freeClusterIdIndex++
		return id
	} else {
		fs.freeClusterIds = fs.FindFreeBitsInBitmap(-1, fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize(), fs.SuperBlock.ClusterCount)
		fs.freeClusterIdIndex = 0

		if len(fs.freeClusterIds) > 0 {
			id := fs.freeClusterIds[0]
			fs.freeClusterIdIndex++
			return id
		}
	}
	return -1
}

// adds a cluster of data, returns the id of the cluster that was created
func (fs *MyFileSystem) AddCluster(bytes [ClusterSize]byte) ID {
	freeID := fs.FindFreeClusterID()
	log.Infof("Free id=%d", freeID)
	if freeID != -1 {
		// mark in data block bitmap
		fs.SetInBitmap(true, int32(freeID), fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())

		// write the actual inode at its address
		fs.SetClusterAt(freeID, bytes)
		return freeID
	}
	log.Warnln("No free cluster found")
	return -1
}

// sets the cluster at the given id
func (fs *MyFileSystem) SetClusterAt(id ID, data [ClusterSize]byte) {
	clusterAddress := fs.GetClusterAddress(id)

	log.Infof("SetClusterAt writing to address=%d for ID of %d", clusterAddress, id)
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

// returns the data of a cluster by cluster id
func (fs *MyFileSystem) GetClusterDataAt(id ID) [ClusterSize]byte {
	clusterAddress := fs.GetClusterAddress(id)
	return fs.GetClusterDataAtAddress(clusterAddress)
}

// returns the data of a cluster by cluster address
func (fs *MyFileSystem) GetClusterDataAtAddress(address Address) [ClusterSize]byte {
	_, err := fs.File.Seek(int64(address), io.SeekStart)

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

// clears the cluster selected by id. Free's it's ID in the cluster bitmap
func (fs *MyFileSystem) ClearClusterById(id ID) {
	fs.SetInBitmap(false, int32(id), fs.SuperBlock.ClusterBitmapStartAddress, fs.SuperBlock.ClusterBitmapSize())
}

// returns a cluster address by ID
func (fs *MyFileSystem) GetClusterAddress(id ID) Address {
	return fs.SuperBlock.ClusterStartAddress + Address(Size(id)*Size(ClusterSize))
}

// returns a cluster ID by a cluster address
func (fs *MyFileSystem) GetClusterId(address Address) ID {
	return ID((address - fs.SuperBlock.ClusterStartAddress) / Address(ClusterSize))
}

// creates a cluster by ID
func (fs *MyFileSystem) GetCluster(id ID) Cluster {
	clusterAddress := fs.GetClusterAddress(id)
	return Cluster{
		fs:      fs,
		id:      id,
		address: clusterAddress,
	}
}

// writes an ID to the cluster at the given index
func (cluster Cluster) WriteId(id ID, idIndex ID) {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)
	log.Infof("WriteId seeking to %d, about to write %d at ID=%d", cluster.address, id, idIndex)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to Start of cluster of ID %d", cluster.id))
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

// writes an address to the cluster at the given index
func (cluster Cluster) WriteAddress(address Address, addressId ID) {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)
	log.Infof("WriteAddress seeking to %d, about to write %d at ID=%d", cluster.address, address, addressId)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to Start of cluster of ID %d", cluster.id))
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

// writes data to a cluster with a certain offset
func (cluster *Cluster) WriteDataOffset(data []byte, offset int) {
	if len(data)+offset > ClusterSize {
		panic("Cannot write... Data + offset > ClusterSize")
	}
	_, err := cluster.fs.File.Seek(int64(cluster.address)+int64(offset), io.SeekStart)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to Start of cluster of ID %d", cluster.id))
	}

	_, err = cluster.fs.File.Write(data[:])

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not write"))
	}
}

// writes data to the given cluster
func (cluster *Cluster) WriteData(data [ClusterSize]byte) {
	cluster.WriteDataOffset(data[:], 0)
}

// reads data from the cluster
func (cluster Cluster) ReadData() [ClusterSize]byte {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to Start of cluster of ID %d", cluster.id))
	}

	var data [ClusterSize]byte
	_, err = cluster.fs.File.Read(data[:])

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not read"))
	}

	return data
}

// reads address at the given ID from the cluster
func (cluster Cluster) ReadAddress(index ID) Address {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)
	log.Infof("ReadAddress seeking to: %d", cluster.address)
	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to Start of cluster of ID %d", cluster.id))
	}

	_, err = cluster.fs.File.Seek(int64(unsafe.Sizeof(Address(0)))*int64(index), io.SeekCurrent)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to Start of cluster of ID %d and indirect index of =%d", cluster.id, index))
	}

	var foundId Address
	err = binary.Read(cluster.fs.File, binary.LittleEndian, &foundId)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not read"))
	}

	return foundId
}

// reads ID at the given ID fro the cluster
func (cluster Cluster) ReadId(index ID) ID {
	_, err := cluster.fs.File.Seek(int64(cluster.address), io.SeekStart)
	log.Infof("ReadId seeking to: %d", cluster.address)
	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to Start of cluster of ID %d", cluster.id))
	}

	_, err = cluster.fs.File.Seek(int64(unsafe.Sizeof(ID(0)))*int64(index), io.SeekCurrent)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprintf("could not seek to Start of cluster of ID %d and index of =%d", cluster.id, index))
	}

	var foundId ID
	err = binary.Read(cluster.fs.File, binary.LittleEndian, &foundId)

	if err != nil {
		log.Error(err)
		panic(fmt.Sprint("could not read"))
	}

	return foundId
}

// returns the used cluster count based by the given size
func GetUsedClusterCount(size Size) Size {
	currCount := size / ClusterSize
	currRemainder := size % ClusterSize
	if currRemainder > 0 {
		currCount++
	}
	return currCount
}
