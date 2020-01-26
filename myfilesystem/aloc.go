package myfilesystem

import (
	"github.com/sirupsen/logrus"
	"unsafe"
)

const (
	directAddresses = 5
	FileTooLarge    = -3
	NoIndirect      = -2
	FirstIndirect   = -1
)

func (fs *MyFileSystem) GetClusterPath(id int) (int, int) {
	const addressesPerCluster = ClusterSize / int(unsafe.Sizeof(Address(0)))
	if id < directAddresses {
		return id, NoIndirect
	} else if id < addressesPerCluster+directAddresses {
		return id - directAddresses, FirstIndirect
	} else {
		indirectId := (id - addressesPerCluster - directAddresses) / addressesPerCluster
		if indirectId >= addressesPerCluster {
			return FileTooLarge, FileTooLarge
		} else {
			return (id - addressesPerCluster - directAddresses) % addressesPerCluster, indirectId
		}
	}
}

func (fs *MyFileSystem) ReadDataFromInode(inode PseudoInode) []byte {
	bytes := []byte{}

	clusters := inode.FileSize / ClusterSize

	for i := 0; Size(i) < clusters; i++ {
		read := fs.ReadDataFromInodeAt(inode, i)
		bytes = append(bytes, read[:]...)
	}

	remainder := inode.FileSize % ClusterSize

	if remainder > 0 {
		read := fs.ReadDataFromInodeAt(inode, int(clusters))
		bytes = append(bytes, read[:remainder]...)
	}

	return bytes
}

func (fs *MyFileSystem) WriteDataToInode(inodeId ID, data []byte) {

	inode := fs.GetInodeAt(inodeId)

	clusters := len(data) / ClusterSize

	for i := 0; i < clusters; i++ {
		var bytes [ClusterSize]byte
		copy(bytes[:], data[i*ClusterSize:(i+1)*ClusterSize])
		fs.AddDataToInode(bytes, &inode, inodeId, i)
		logrus.Infof("#%d, inode=%v", i, inode)
		logrus.Infof("dir1=%v", inode.Direct1)
		logrus.Infof("dir2=%v", inode.Direct2)
	}

	remainder := len(data) % ClusterSize

	if remainder > 0 {
		var bytes [ClusterSize]byte
		copy(bytes[:], data[clusters*ClusterSize:(clusters*ClusterSize)+remainder])
		fs.AddDataToInode(bytes, &inode, inodeId, clusters)
	}

	inode = fs.GetInodeAt(inodeId)

	inode.FileSize = Size(len(data))

	logrus.Infof("inode=%v", inode)
	logrus.Infof("dir1=%v", inode.Direct1)
	logrus.Infof("dir2=%v", inode.Direct2)

	fs.SetInodeAt(inodeId, inode)
}

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

func (fs *MyFileSystem) GetClusterByIndex(inode PseudoInode, clusterId int) [ClusterSize]byte {
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

func (fs *MyFileSystem) ReadFromDirect(inode PseudoInode, clusterIndex int) [ClusterSize]byte {
	inodeDirectPtrs := []Address{
		inode.Direct1,
		inode.Direct2,
		inode.Direct3,
		inode.Direct4,
		inode.Direct5,
	}

	return fs.GetClusterDataAtAddress(inodeDirectPtrs[clusterIndex])
}

func (fs *MyFileSystem) ReadFromFirstIndirect(inode PseudoInode, clusterIndex int) [ClusterSize]byte {
	return fs.GetClusterDataAtAddress(fs.GetCluster(inode.Indirect1).ReadAddress(ID(clusterIndex)))
}

func (fs *MyFileSystem) ReadFromSecondIndirect(inode PseudoInode, clusterIndex int, indirectIndex int) [ClusterSize]byte {
	return fs.GetClusterDataAtAddress(fs.GetCluster(fs.GetCluster(inode.Indirect2).ReadId(ID(indirectIndex))).ReadAddress(ID(clusterIndex)))
}

func (fs *MyFileSystem) AddDataToInode(data [ClusterSize]byte, inode *PseudoInode, inodeId ID, clusterId int) ID {
	clusterIndex, indirectIndex := fs.GetClusterPath(clusterId)
	logrus.Infof("StartAddData Inode Indirect1 addr %d", inode.Indirect1)
	logrus.Infof("clusterIndex, indirectIndex %d <=> %d", clusterIndex, indirectIndex)
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
	logrus.Infof("EndAddData Inode Direct addr %d", inode.Indirect1)
	logrus.Infof("EndAddData Inode Indirect1 addr %d", inode.Indirect1)
	logrus.Infof("EndAddData Inode Indirect2 addr %d", inode.Indirect2)
	logrus.Infof("EndAddData Inode Indirect2 id %d", inodeId)
	if indirectIndex == NoIndirect || (indirectIndex == FirstIndirect && clusterIndex == 0) || (indirectIndex == 0 || clusterIndex == 0) {
		fs.SetInodeAt(inodeId, *inode)
	}
	logrus.Infof("EndAddData ReadInode Indirect2 id %d", fs.GetInodeAt(inodeId).Indirect2)
	return result
}

func (fs *MyFileSystem) WriteToDirect(inode *PseudoInode, clusterIndex int, data [ClusterSize]byte) ID {
	inodeDirectPtrs := []*Address{
		&inode.Direct1,
		&inode.Direct2,
		&inode.Direct3,
		&inode.Direct4,
		&inode.Direct5,
	}

	clusterId := fs.AddCluster(data)

	logrus.Infof("ClusterContent rn %b", fs.GetCluster(clusterId).ReadData())
	logrus.Infof("ADding id %d %d", clusterId, clusterIndex)

	if clusterId != -1 {
		*(inodeDirectPtrs[clusterIndex]) = fs.GetClusterAddress(clusterId)
	}

	return clusterId
}

func (fs *MyFileSystem) WriteDataToIndirectCluster(cluster Cluster, clusterIndex int, data [ClusterSize]byte) ID {
	logrus.Infof("test")
	clusterId := fs.AddCluster(data)
	logrus.Infof("WriteDataToIndirectCluster in found %d", clusterId)
	if clusterId > -1 {
		cluster.WriteAddress(fs.GetClusterAddress(clusterId), ID(clusterIndex))
	}

	return clusterId
}

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
			return ID(-1)
		}
	} else {
		// find the assigned cluster
		cluster = fs.GetCluster(inode.Indirect1)
	}
	return fs.WriteDataToIndirectCluster(cluster, clusterIndex, data)
}

func (fs *MyFileSystem) WriteDataToSecondIndirectCluster(inode *PseudoInode, clusterIndex int, indirectIndex int, data [ClusterSize]byte) ID {
	var secondIndirectCluster Cluster
	if clusterIndex == 0 && indirectIndex == 0 {
		// the second indirect cluster was not created yet
		id := fs.AddCluster([ClusterSize]byte{})

		if id > -1 {
			inode.Indirect2 = id
			secondIndirectCluster = fs.GetCluster(id)
		} else {
			return ID(-1)
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
			return ID(-1)
		}
	} else {
		indirectCluster = fs.GetCluster(secondIndirectCluster.ReadId(ID(indirectIndex)))
	}

	logrus.Infof("IndirectCluster ID is %d", indirectCluster.id)

	return fs.WriteDataToIndirectCluster(indirectCluster, clusterIndex, data)
}
