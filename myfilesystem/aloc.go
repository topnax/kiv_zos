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

func (fs *MyFileSystem) AddDataToInode(data [ClusterSize]byte, inode PseudoInode, inodeId ID, clusterId int) ID {
	clusterIndex, indirectIndex := fs.GetClusterPath(clusterId)
	logrus.Infof("StartAddData Inode Indirect1 addr %d", inode.Indirect1)
	logrus.Infof("clusterIndex, indirectIndex %d <=> %d", clusterIndex, indirectIndex)
	result := ID(-1)
	if indirectIndex == NoIndirect {
		result = fs.WriteToDirect(&inode, clusterIndex, data)
	} else if indirectIndex == FirstIndirect {
		result = fs.WriteDataToTheFirstIndirectCluster(&inode, clusterIndex, data)
	} else if indirectIndex != FileTooLarge {
		result = fs.WriteDataToSecondIndirectCluster(&inode, clusterIndex, indirectIndex, data)
	} else {
		return -1
	}
	logrus.Infof("EndAddData Inode Indirect1 addr %d", inode.Indirect1)
	logrus.Infof("EndAddData Inode Indirect2 addr %d", inode.Indirect2)
	logrus.Infof("EndAddData Inode Indirect2 id %d", inodeId)
	fs.SetInodeAt(inodeId, inode)
	logrus.Infof("EndAddData ReadInode Indirect2 id %d", fs.GetInodeAt(inodeId).Indirect2)
	return result
}

func (fs MyFileSystem) WriteToDirect(inode *PseudoInode, clusterIndex int, data [1024]byte) ID {
	inodeDirectPtrs := []*Address{
		&inode.Direct1,
		&inode.Direct2,
		&inode.Direct3,
		&inode.Direct4,
		&inode.Direct5,
	}

	clusterId := fs.AddCluster(data)

	if clusterId != -1 {
		*(inodeDirectPtrs[clusterIndex]) = fs.GetClusterAddress(clusterId)
	}

	return clusterId
}

func (fs *MyFileSystem) WriteDataToIndirectCluster(cluster Cluster, clusterIndex int, data [1024]byte) ID {
	clusterId := fs.AddCluster(data)

	if clusterId > -1 {
		cluster.WriteAddress(fs.GetClusterAddress(clusterId), ID(clusterIndex))
	}

	return clusterId
}
func (fs *MyFileSystem) WriteDataToTheFirstIndirectCluster(inode *PseudoInode, clusterIndex int, data [1024]byte) ID {
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

func (fs MyFileSystem) WriteDataToSecondIndirectCluster(inode *PseudoInode, clusterIndex int, indirectIndex int, data [ClusterSize]byte) ID {
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

//func (fs *MyFileSystem) WriteDataToIndirectCluster(clusterIndex int, indirectPointer *ID, data [1024]byte) ID {
//	freeClusterId := fs.FindFreeClusterID()
//	if freeClusterId != -1 {
//		if clusterIndex == 0 {
//			// cluster was not created yet
//			*indirectPointer = fs.AddCluster([ClusterSize]byte{})
//			logrus.Printf("WriteDataToIndirectCluster created a new cluster of pointers %d", *indirectPointer)
//			freeClusterId = fs.FindFreeClusterID()
//		}
//
//		clusterId := fs.AddCluster(data)
//		if clusterId != -1 {
//			logrus.Infof("WriteDataToIndirectCluster written clusterID=%d index=%d to address %d", clusterId, clusterIndex, fs.GetClusterAddress(clusterId))
//			logrus.Infof("ID pointer %d", *indirectPointer)
//			fs.GetCluster(*indirectPointer).WriteAddress(fs.GetClusterAddress(clusterId), ID(clusterIndex))
//			logrus.Infof("ID pointer %d", *indirectPointer)
//			return clusterId
//		} else {
//			fs.ClearClusterById(*indirectPointer)
//		}
//	}
//	return -1
//}

func (fs *MyFileSystem) WriteDataToSecondIndirect(clusterIndex int, indirectIndex int, inode *PseudoInode, bytes [1024]byte) ID {

	secondIndirectClusterId := ID(-1)

	if clusterIndex == 0 && indirectIndex == 0 {
		secondIndirectClusterId = fs.AddCluster([ClusterSize]byte{})
		if secondIndirectClusterId > -1 {
			inode.Indirect2 = secondIndirectClusterId
		} else {
			panic("No space for indirect2")
		}
	}

	//fs.WriteDataToIndirectCluster(clusterIndex,)

	//logrus.Infof("WriteDataToSecondIndirect %d <=> %d", clusterIndex, indirectIndex)
	//if indirectIndex == 0 && clusterIndex == 0 {
	//	clusterId := fs.AddCluster([ClusterSize] byte{})
	//	if clusterId > -1 {
	//		inode.Indirect2 = clusterId
	//	} else {
	//		panic("No space for indirect2")
	//	}
	//}
	//
	//if clusterIndex == 0 {
	//	freeId := fs.FindFreeClusterID()
	//	fs.WriteDataToIndirectCluster(clusterIndex, &freeId, bytes)
	//	logrus.Infof("WriteDataToSecondIndirect about to write address")
	//	fs.GetCluster(inode.Indirect2).WriteAddress(fs.GetClusterAddress(freeId), ID(indirectIndex))
	//} else {
	//	indirectClusterId := fs.GetCluster(inode.Indirect2).ReadAddress(ID(indirectIndex))
	//	fs.WriteDataToIndirectCluster(clusterIndex, &indirectClusterId, bytes)
	//}

	//if indirectIndex == 0 && clusterIndex == 0 {
	//	// indirect lv2 was not created yet
	//	freeClusterId := fs.FindFreeClusterID()
	//	if freeClusterId != -1 {
	//		inode.Indirect2 =fs.AddCluster([ClusterSize]byte{})
	//		logrus.Infof("Indirect2 set at %d", inode.Indirect2)
	//	} else {
	//		return -1
	//	}
	//}
	//logrus.Info("WriteDataToSecondIndirect about to write to indirect")
	//
	//freeClusterId := fs.FindFreeClusterID()
	//if freeClusterId != -1 {
	//	indirectIndexId := freeClusterId
	//	if clusterIndex > 0 {
	//		indirectClusterID :=fs.GetCluster(inode.Indirect2).ReadAddress(ID(indirectIndex))
	//		println(indirectClusterID)
	//	}
	//	fs.WriteDataToIndirectCluster(clusterIndex, &indirectIndexId, bytes)
	//	fs.GetCluster(inode.Indirect2).WriteAddress(fs.GetClusterAddress(indirectIndexId), ID(indirectIndex))
	//}

	return -1
}
