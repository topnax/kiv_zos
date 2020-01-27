package myfilesystem

import (
	"fmt"
	"kiv_zos/utils"
)

func (fs MyFileSystem) Print(path string) {
	tgtName := GetTargetName(path)
	fs.VisitDirectoryByPathAndExecute(path, func() {
		fs.PrintContent(fs.GetInodeAt(fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), tgtName).NodeID))
	}, func() {
		utils.PrintError(fmt.Sprintf("'%s' not found", path))
	})
}

func (fs MyFileSystem) PrintContent(node PseudoInode) {
	fs.ReadDataFromInodeFx(node, func(data []byte) {
		fmt.Printf("%s", data)
	})
	fmt.Print("\n")
}

func (fs MyFileSystem) PrintInfo(inodeId ID) {
	inode := fs.GetInodeAt(inodeId)
	stringz := ""
	addresses := fs.GetUsedClusterAddresses(inode)
	for _, address := range addresses {
		stringz += fmt.Sprintf("%d ", address)
	}
	item := fs.FindDirItemByNodeId(fs.ReadDirItems(fs.currentInodeID), inodeId)
	fmt.Printf("%s - %d - %d - %s\n", item.GetName(), item.NodeID, inode.FileSize, stringz)
}

func (fs *MyFileSystem) PrintCurrentPath() {
	fmt.Println(fs.FindDirPath(fs.currentInodeID))
}

func (fs *MyFileSystem) CreateNewDirectory(name string) {
	tgtName := GetTargetName(name)

	fs.VisitDirectoryByPathAndExecute(name, func() {
		fs.NewDirectory(fs.currentInodeID, tgtName, false)

	}, func() {
		utils.PrintError(fmt.Sprintf("Cannot create folder '%s' at '/%s/' because such path does not exist", tgtName, name))
	})
}

func GetTargetName(path string) string {
	dirNames := GetDirNames(path)
	if len(dirNames) == 0 {
		return path
	} else {
		return dirNames[len(dirNames)-1]
	}
}

func (fs MyFileSystem) ListDirectoryContent(name string) {
	items := fs.ReadDirItems(fs.currentInodeID)

	item := fs.FindDirItemByName(items, name)
	if item.NodeID != -1 {
		if fs.GetInodeAt(item.NodeID).IsDirectory {
			fs.ListDirectory(item.NodeID)
		} else {
			utils.PrintError(fmt.Sprintf("%s is not a directory", item.GetName()))
		}
	} else {
		utils.PrintError(fmt.Sprintf("'%s' not found", name))
	}
}

func (fs *MyFileSystem) ChangeDirectory(path string) {
	if path == FolderSeparator {
		fs.ChangeToDirectoryByName(path)
		return
	}
	if !fs.ChangeToDirectoryByPath(path) {
		utils.PrintError(fmt.Sprintf("'%s' not found", path))
	}
}

func (fs *MyFileSystem) Info(path string) {
	tgtName := GetTargetName(path)
	fs.VisitDirectoryByPathAndExecute(path, func() {
		fs.PrintInfo(fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), tgtName).NodeID)
	}, func() {
		utils.PrintError(fmt.Sprintf("'%s' not found", path))
	})
}

func (fs *MyFileSystem) RemoveDirectory(path string) {
	tgtName := GetTargetName(path)
	fs.VisitDirectoryByPathAndExecute(path, func() {
		fs.RemoveDirItem(tgtName, fs.currentInodeID)
	}, func() {
		utils.PrintError(fmt.Sprintf("'%s' not found", path))
	})
}
