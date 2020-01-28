package myfilesystem

import (
	"github.com/sirupsen/logrus"
	"unsafe"
)

const (
	directAddresses = 5 // number of direct addresses in a node
	FileTooLarge    = -3
	NoIndirect      = -2
	FirstIndirect   = -1
)

// returns a cluster path by cluster index.
// When the second returned value is NoIndirect, the first value represents index of the direct pointer to the cluster.
// When the second returned value is FirstIndirect, the first value represents the index of the pointer in the first indirect cluster.
// when the second returned value is not one of as stated above, the first value then represents the index of the pointer
// in the indirect cluster which is index is the second value
func (fs *MyFileSystem) GetClusterPath(index int) (int, int) {
	const addressesPerCluster = ClusterSize / int(unsafe.Sizeof(Address(0)))
	if index < directAddresses {
		return index, NoIndirect
	} else if index < addressesPerCluster+directAddresses {
		return index - directAddresses, FirstIndirect
	} else {
		indirectId := (index - addressesPerCluster - directAddresses) / addressesPerCluster
		if indirectId >= addressesPerCluster {
			return FileTooLarge, FileTooLarge
		} else {
			return (index - addressesPerCluster - directAddresses) % addressesPerCluster, indirectId
		}
	}
}

// reads the data of the given node, each cluster's data is passed to the function specified in the fx parameter
func (fs *MyFileSystem) ReadDataFromInodeFx(inode PseudoInode, fx func(data []byte) bool) {
	clusters := inode.FileSize / ClusterSize

	for i := 0; Size(i) < clusters; i++ {
		read := fs.ReadDataFromInodeAt(inode, i)
		if !fx(read[:]) {
			return
		}
	}

	remainder := inode.FileSize % ClusterSize

	if remainder > 0 {
		read := fs.ReadDataFromInodeAt(inode, int(clusters))
		fx(read[:remainder])
	}
}

// reads the whole content of inode and returns it as byte array
func (fs *MyFileSystem) ReadDataFromInode(inode PseudoInode) []byte {
	bytes := []byte{}

	fs.ReadDataFromInodeFx(inode, func(data []byte) bool {
		bytes = append(bytes, data...)
		return true
	})

	return bytes
}

// writes the given data to the cluster of the given id
func (fs *MyFileSystem) WriteDataToInode(inodeId ID, data []byte) bool {
	inode := fs.GetInodeAt(inodeId)

	clusters := len(data) / ClusterSize

	for i := 0; i < clusters; i++ {
		var bytes [ClusterSize]byte
		copy(bytes[:], data[i*ClusterSize:(i+1)*ClusterSize])
		id := fs.AddDataToInode(bytes, &inode, inodeId, i)
		if id < 0 {
			return false
		}
	}

	remainder := len(data) % ClusterSize

	if remainder > 0 {
		var bytes [ClusterSize]byte
		copy(bytes[:], data[clusters*ClusterSize:(clusters*ClusterSize)+remainder])
		id := fs.AddDataToInode(bytes, &inode, inodeId, clusters)
		if id < 0 {
			return false
		}
	}

	inode = fs.GetInodeAt(inodeId)
	inode.FileSize = Size(len(data))

	fs.SetInodeAt(inodeId, inode)
	return true
}

// reads the one cluster of data from the given inode
func (fs *MyFileSystem) ReadDataFromInodeAt(inode PseudoInode, clusterId int) [ClusterSize]byte {
	clusterIndex, indirectIndex := fs.GetClusterPath(clusterId)

	if indirectIndex == NoIndirect {
		return fs.ReadFromDirect(inode, clusterIndex)
	} else if indirectIndex == FirstIndirect {
		return fs.ReadFromFirstIndirect(inode, clusterIndex)
	} else if indirectIndex != FileTooLarge {
		return fs.ReadFromSecondIndirect(inode, clusterIndex, indirectIndex)
	} else {
		panic("Could not read out of the bounds")
	}
}

// returns the cluster address based cluster index
func (fs *MyFileSystem) GetClusterAddressByIndex(inode PseudoInode, clusterId int) Address {
	clusterIndex, indirectIndex := fs.GetClusterPath(clusterId)

	if indirectIndex == NoIndirect {
		return fs.GetDirectClusterAddressByIndex(inode, clusterIndex)
	} else if indirectIndex == FirstIndirect {
		return fs.GetIndirectClusterAddressByIndex(inode, clusterIndex)
	} else if indirectIndex != FileTooLarge {
		return fs.GetSecondIndirectClusterAddressByIndex(inode, indirectIndex, clusterIndex)
	} else {
		panic("Could not read out of the bounds")
	}
}

// reads one cluster of data from the inode based on the given cluster index
func (fs *MyFileSystem) ReadFromDirect(inode PseudoInode, clusterIndex int) [ClusterSize]byte {
	return fs.GetClusterDataAtAddress(fs.GetDirectClusterAddressByIndex(inode, clusterIndex))
}

// returns the address of a cluster based on the index and a direct pointer
func (fs *MyFileSystem) GetDirectClusterAddressByIndex(inode PseudoInode, clusterIndex int) Address {
	inodeDirectPtrs := []Address{
		inode.Direct1,
		inode.Direct2,
		inode.Direct3,
		inode.Direct4,
		inode.Direct5,
	}

	return inodeDirectPtrs[clusterIndex]
}

// reads data from the first indirect pointer based on the given index
func (fs *MyFileSystem) ReadFromFirstIndirect(inode PseudoInode, clusterIndex int) [ClusterSize]byte {
	return fs.GetClusterDataAtAddress(fs.GetIndirectClusterAddressByIndex(inode, clusterIndex))
}

// gets the address of the cluster based on the pointer in the first indirect point
func (fs *MyFileSystem) GetIndirectClusterAddressByIndex(inode PseudoInode, clusterIndex int) Address {
	return fs.GetCluster(inode.Indirect1).ReadAddress(ID(clusterIndex))
}

// reads data from the second indirect pointer based on the given index
func (fs *MyFileSystem) ReadFromSecondIndirect(inode PseudoInode, clusterIndex int, indirectIndex int) [ClusterSize]byte {
	return fs.GetClusterDataAtAddress(fs.GetSecondIndirectClusterAddressByIndex(inode, indirectIndex, clusterIndex))
}

// gets the address of the cluster based on the pointer in the second indirect point
func (fs *MyFileSystem) GetSecondIndirectClusterAddressByIndex(inode PseudoInode, indirectIndex int, clusterIndex int) Address {
	return fs.GetCluster(fs.GetCluster(inode.Indirect2).ReadId(ID(indirectIndex))).ReadAddress(ID(clusterIndex))
}

// adds one cluster of data to the given inode
func (fs *MyFileSystem) AddDataToInode(data [ClusterSize]byte, inode *PseudoInode, inodeId ID, clusterId int) ID {
	clusterIndex, indirectIndex := fs.GetClusterPath(clusterId)

	result := ID(-1)
	if indirectIndex == NoIndirect {
		result = fs.WriteToDirect(inode, clusterIndex, data)
	} else if indirectIndex == FirstIndirect {
		result = fs.WriteDataToTheFirstIndirectCluster(inode, clusterIndex, data)
	} else if indirectIndex != FileTooLarge {
		result = fs.WriteDataToSecondIndirectCluster(inode, clusterIndex, indirectIndex, data)
	} else {
		return -1
	}

	if indirectIndex == NoIndirect || (indirectIndex == FirstIndirect && clusterIndex == 0) || (indirectIndex == 0 || clusterIndex == 0) {
		fs.SetInodeAt(inodeId, *inode)
	}

	return result
}

// writes data to the direct pointer
func (fs *MyFileSystem) WriteToDirect(inode *PseudoInode, clusterIndex int, data [ClusterSize]byte) ID {
	inodeDirectPtrs := []*Address{
		&inode.Direct1,
		&inode.Direct2,
		&inode.Direct3,
		&inode.Direct4,
		&inode.Direct5,
	}

	clusterId := fs.AddCluster(data)

	if clusterId > -1 {
		*(inodeDirectPtrs[clusterIndex]) = fs.GetClusterAddress(clusterId)
	}

	return clusterId
}

// writes data to the a cluster that is accessed by the an indirect pointer
func (fs *MyFileSystem) WriteDataToIndirectCluster(cluster Cluster, clusterIndex int, data [ClusterSize]byte) ID {
	clusterId := fs.AddCluster(data)
	if clusterId > -1 {
		cluster.WriteAddress(fs.GetClusterAddress(clusterId), ID(clusterIndex))
	}

	return clusterId
}

// writes data to the a cluster that is accessed by the first indirect pointer
func (fs *MyFileSystem) WriteDataToTheFirstIndirectCluster(inode *PseudoInode, clusterIndex int, data [ClusterSize]byte) ID {
	var cluster Cluster
	if clusterIndex == 0 {
		// if first indirect cluster was not yet created
		id := fs.AddCluster([ClusterSize]byte{})
		if id > -1 {
			// create a new cluster and assign it
			cluster = fs.GetCluster(id)
			inode.Indirect1 = id
		} else {
			return id
		}
	} else {
		// find the assigned cluster
		cluster = fs.GetCluster(inode.Indirect1)
	}
	return fs.WriteDataToIndirectCluster(cluster, clusterIndex, data)
}

// writes data to the second indirect cluster
func (fs *MyFileSystem) WriteDataToSecondIndirectCluster(inode *PseudoInode, clusterIndex int, indirectIndex int, data [ClusterSize]byte) ID {
	var secondIndirectCluster Cluster
	if clusterIndex == 0 && indirectIndex == 0 {
		// the second indirect cluster was not created yet
		id := fs.AddCluster([ClusterSize]byte{})

		if id > -1 {
			inode.Indirect2 = id
			secondIndirectCluster = fs.GetCluster(id)
		} else {
			return id
		}
	} else {
		secondIndirectCluster = fs.GetCluster(inode.Indirect2)
	}

	logrus.Infof("SecondIndirectCluster ID is %d", secondIndirectCluster.id)

	var indirectCluster Cluster
	if clusterIndex == 0 {
		// if first indirect cluster was not yet created
		id := fs.AddCluster([ClusterSize]byte{})
		if id > -1 {
			// create a new cluster and assign it
			indirectCluster = fs.GetCluster(id)
			logrus.Infof("Indirectcluster in 2 found %d", id)
			secondIndirectCluster.WriteId(id, ID(indirectIndex))
		} else {
			return id
		}
	} else {
		indirectCluster = fs.GetCluster(secondIndirectCluster.ReadId(ID(indirectIndex)))
	}

	logrus.Infof("IndirectCluster ID is %d", indirectCluster.id)

	return fs.WriteDataToIndirectCluster(indirectCluster, clusterIndex, data)
}

// shrinks the data of the inode, removing all allocated clusters
func (fs *MyFileSystem) ShrinkInodeData(inode *PseudoInode, inodeId ID, desiredSize Size) {
	tbr := GetClusterCountToBeRemoved(inode.FileSize, desiredSize)
	clusterCount := GetUsedClusterCount(inode.FileSize)
	for i := 0; i < tbr; i++ {
		// each address that is to be removed, free the cluster
		address := fs.GetClusterAddressByIndex(*inode, int(clusterCount)-i-1)
		fs.ClearClusterById(fs.GetClusterId(address))
	}
}

// computes the count of clusters to be removed based on the current and target size
func GetClusterCountToBeRemoved(currentSize Size, targetSize Size) int {
	currCount := GetUsedClusterCount(currentSize)
	tgtCount := GetUsedClusterCount(targetSize)
	return int(currCount - tgtCount)
}

// returns an array of addresses that the inode uses to store data
func (fs MyFileSystem) GetUsedClusterAddresses(inode PseudoInode) []Address {
	var ids []Address
	count := GetUsedClusterCount(inode.FileSize)
	for i := 0; i < int(count); i++ {
		ids = append(ids, fs.GetClusterAddressByIndex(inode, i))
	}
	return ids
}
