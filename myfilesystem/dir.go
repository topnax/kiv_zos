package myfilesystem

import (
	log "github.com/sirupsen/logrus"
	"unsafe"
)

type ReadOrder struct {
	ClusterId ID
	Start     int
	Bytes     int
}

func GetReadOrder(offset int, read int) []ReadOrder {
	clusterId := ID(offset / ClusterSize)

	log.Infof("Computed cid %d", clusterId)

	overflow := (offset%ClusterSize)+read > ClusterSize
	log.Infof("overflow : %v", overflow)

	if !overflow {
		return []ReadOrder{{
			ClusterId: clusterId,
			Start:     offset % ClusterSize,
			Bytes:     read,
		}}
	} else {
		return []ReadOrder{{
			ClusterId: clusterId,
			Start:     offset % ClusterSize,
			Bytes:     ClusterSize - (offset % ClusterSize),
		}, {
			ClusterId: clusterId + 1,
			Start:     0,
			Bytes:     read - (ClusterSize - (offset % ClusterSize)),
		}}
	}
}

func (fs *MyFileSystem) AddDirItem(item DirectoryItem, node PseudoInode) {
	if node.IsDirectory {
		fs.AppendDirItem(item, node)
	} else {
		log.Warnf("Trying to add a directory item to an inode that is not a directory")
	}
}

func (fs *MyFileSystem) AppendDirItem(item DirectoryItem, node PseudoInode) {
	if node.IsDirectory {

	} else {
		log.Warnf("Trying to add a directory item to an inode that is not a directory")
	}
}

func (fs MyFileSystem) GetDirItemsCount(node PseudoInode) Size {
	return node.FileSize / Size(unsafe.Sizeof(DirectoryItem{}))
}
