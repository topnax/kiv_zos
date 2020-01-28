package myfilesystem

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"kiv_zos/utils"
	"unsafe"
)

// does the basic checking of file system's consistency
func (fs *MyFileSystem) ConsistencyCheck() {
	fs.CheckThatAllFilesBelongToADirectory()
	fs.CheckThatFilesAreCorrectlyAllocated()
}

// checks whether files (nodes) are correctly allocated
// Does the filesize correspond with the pointers written in the direct/indirect clusters?
func (fs *MyFileSystem) CheckThatFilesAreCorrectlyAllocated() {
	foundFiles := NewIdSet()
	fs.AddAllFiles(foundFiles, 0)
	addressesPerCluster := ClusterSize / Size(unsafe.Sizeof(Address(0)))

	// directly pointed clusters
	count := Size(5)

	// first indirect cluster
	count += addressesPerCluster

	// second indirect clusters
	count += addressesPerCluster * addressesPerCluster

	freeIds := fs.FindFreeBitsInBitmap(int(fs.SuperBlock.InodeCount()), fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize(), fs.SuperBlock.InodeCount())

	freeIdSet := NewIdSet()

	for _, id := range freeIds {
		freeIdSet.Add(id)
	}

	foundFiles.Clear()
	for i := Size(0); i < fs.SuperBlock.InodeCount(); i++ {
		if !freeIdSet.Has(ID(i)) {
			foundFiles.Add(ID(i))
		}
	}

	//utils.PrintHighlight(fmt.Sprintf("Hypothetical maximal file size: %d", count*ClusterSize))

	for id := range foundFiles.List {
		node := fs.GetInodeAt(id)

		i := 0
		for ; i < int(count); i++ {
			if fs.GetClusterAddressByIndex(node, i) < 1 {
				break
			}
		}

		if GetUsedClusterCount(node.FileSize) != Size(i) {
			utils.PrintError(fmt.Sprintf("INODE OF ID=%d HAS DIFFERENT AMOUNT OF ALLOCATED CLUSTERS THAN STATED IN THE HEADER", id))
			return
		}
	}
	utils.PrintSuccess("OK - EACH INODE HAS CORRECT AMOUNT OF ALLOCATED CLUSTERS")
}

// traverses the file tree from the root node and finds all files. Then checks whether are inodes are inside some folder
func (fs *MyFileSystem) CheckThatAllFilesBelongToADirectory() {
	foundFiles := NewIdSet()

	if !fs.AddAllFiles(foundFiles, 0) {
		utils.PrintError("FOUND AT LEAST TWO DIRECTORY ITEMS THAT POINT TO THE SAME INODE")
		return
	}

	// finds all free inode ids
	ids := fs.FindFreeBitsInBitmap(int(fs.SuperBlock.InodeCount()), fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize(), fs.SuperBlock.InodeCount())

	if len(ids)+len(foundFiles.List)+1 != +int(fs.SuperBlock.InodeCount()) {
		logrus.Infof("%d total found nodes, but %d found used in the bitmap.", len(foundFiles.List)+1, int(fs.SuperBlock.InodeCount())-len(ids))
		utils.PrintError("FOUND ONE OR MORE FILES THAT DO NOT BELONG IN ANY DIRECTORY")
	} else {
		utils.PrintSuccess("OK - EVERY FILE BELONGS IN A DIRECTORY")
	}
}

// adds all files found in the given node id to the passed ID set
// returns false when a duplicate node is found
func (fs *MyFileSystem) AddAllFiles(foundFiles *IDSet, nodeID ID) bool {
	items := fs.ReadDirItems(nodeID)
	for index, item := range items {
		if index > 1 {
			if fs.GetInodeAt(item.NodeID).IsDirectory {
				fs.AddAllFiles(foundFiles, item.NodeID)
			}
			if foundFiles.Has(item.NodeID) {
				return false
			}
			foundFiles.Add(item.NodeID)
		}
	}
	return true
}
