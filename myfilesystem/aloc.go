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

func (fs *MyFileSystem) AddDataToInode(data [ClusterSize]byte, inode *PseudoInode, clusterId int) ID {
	clusterIndex, indirectIndex := fs.GetClusterPath(clusterId)

	if indirectIndex == NoIndirect {
		return fs.WriteToDirect(inode, clusterIndex, data)
	} else if indirectIndex == FirstIndirect {
		return fs.WriteDataToIndirect(clusterIndex, &inode.Indirect1, data)

	} else if indirectIndex != FileTooLarge {
		return fs.WriteDataToSecondIndirect(clusterIndex, indirectIndex, inode, data)
	} else {
		return -1
	}

}

func (fs *MyFileSystem) WriteToDirect(inode *PseudoInode, clusterIndex int, data [1024]byte) ID {
	inodeDirectPtrs := []*Address{
		&inode.Direct1,
		&inode.Direct2,
		&inode.Direct3,
		&inode.Direct4,
		&inode.Direct5,
	}
	freeClusterId := fs.FindFreeClusterID()
	if freeClusterId != -1 {
		*(inodeDirectPtrs[clusterIndex]) = fs.GetClusterAddress(freeClusterId)
		fs.SetClusterAt(freeClusterId, data)
		return freeClusterId
	} else {
		logrus.Error("WriteToDirect could not find a free cluster ID")
		return -1
	}
}

func (fs *MyFileSystem) WriteDataToIndirect(clusterIndex int, indirectPointer *ID, data [1024]byte) ID {
	freeClusterId := fs.FindFreeClusterID()
	if freeClusterId != -1 {
		if clusterIndex == 0 {
			// cluster was not created yet
			*indirectPointer = freeClusterId
			fs.SetClusterAt(freeClusterId, [ClusterSize]byte{})
			freeClusterId = fs.FindFreeClusterID()
		}
		if freeClusterId != -1 {
			freeClusterAddress := fs.GetClusterAddress(freeClusterId)
			fs.GetCluster(*indirectPointer).WriteAddress(freeClusterAddress, ID(clusterIndex))
			fs.SetClusterAt(freeClusterId, data)
			return freeClusterId
		} else {
			fs.ClearClusterById(*indirectPointer)
		}
	}
	return -1
}

func (fs *MyFileSystem) WriteDataToSecondIndirect(clusterIndex int, indirectIndex int, inode *PseudoInode, bytes [1024]byte) ID {

	if indirectIndex == 0 {
		// indirect lv2 was not created yet
		freeClusterId := fs.FindFreeClusterID()
		if freeClusterId != -1 {
			inode.Indirect2 = freeClusterId
			fs.SetClusterAt(freeClusterId, [ClusterSize]byte{})
			freeClusterId = fs.FindFreeClusterID()
		} else {
			return -1
		}
	}

	freeClusterId := fs.FindFreeClusterID()
	if freeClusterId != -1 {
		indirectIndexId := freeClusterId
		fs.WriteDataToIndirect(clusterIndex, &indirectIndexId, bytes)
		fs.GetCluster(inode.Indirect2).WriteAddress(fs.GetClusterAddress(indirectIndexId), ID(indirectIndex))
	}

	return -1
}
