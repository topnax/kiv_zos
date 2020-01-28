package myfilesystem

import (
	"github.com/sirupsen/logrus"
	"kiv_zos/utils"
)

func (fs *MyFileSystem) ConsistencyCheck() {
	fs.CheckThatAllFilesBelongToADirectory()
}

func (fs *MyFileSystem) CheckThatAllFilesBelongToADirectory() {
	foundFiles := NewIDSet()
	fs.VisitDirectoryByPathAndExecute("/", func() {
		fs.AddAllFiles(foundFiles, 0)
	}, func() {

	})

	ids := fs.FindFreeBitsInBitmap(int(fs.SuperBlock.InodeCount()), fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize(), fs.SuperBlock.InodeCount())

	//logrus.Warnf("%d - %d - %d", len(ids), len(foundFiles.List)+1, fs.SuperBlock.InodeCount())
	//logrus.Warnf("%v ", fs.GetInBitmap(0, fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize()))
	//logrus.Warnf("%v ", fs.GetInBitmap(1, fs.SuperBlock.InodeBitmapStartAddress, fs.SuperBlock.InodeBitmapSize()))

	if len(ids)+len(foundFiles.List)+1 != +int(fs.SuperBlock.InodeCount()) {
		logrus.Errorf("%d total found nodes, but %d found used in the bitmap.", len(foundFiles.List)+1, len(ids))
		utils.PrintSuccess("FOUND ONE OR MORE FILES THAT DO NOT BELONG IN A DIRECTORY")
	} else {
		utils.PrintSuccess("OK - EVERY FILE BELONGS IN A DIRECTORY")
	}
}

func (fs *MyFileSystem) AddAllFiles(foundFiles *IDSet, nodeID ID) {
	items := fs.ReadDirItems(nodeID)
	for index, item := range items {
		if index > 1 {
			if fs.GetInodeAt(item.NodeID).IsDirectory {
				fs.AddAllFiles(foundFiles, item.NodeID)
			}
			foundFiles.Add(item)
		}
	}
}
