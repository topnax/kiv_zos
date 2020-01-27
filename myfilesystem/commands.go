package myfilesystem

import (
	"fmt"
	"kiv_zos/utils"
	"strings"
)

func (fs MyFileSystem) PrintContent(node PseudoInode) {
	fs.ReadDataFromInodeFx(node, func(data []byte) {
		fmt.Print(string(data))
	})
}

func (fs MyFileSystem) Info(name string) {

}

func (fs MyFileSystem) PrintInfo(inode PseudoInode, item DirectoryItem) {
	stringz := ""
	addresses := fs.GetUsedClusterAddresses(inode)
	for _, address := range addresses {
		stringz += fmt.Sprintf("%d ", address)
	}
	fmt.Printf("%s - %d - %d - %s", item.GetName(), item.NodeID, inode.FileSize, stringz)
}

func (fs *MyFileSystem) PrintCurrentPath() {
	fmt.Println(fs.FindDirPath(fs.currentInodeID))
}

func (fs *MyFileSystem) CreateNewDirectory(name string) {
	dirNames := GetDirNames(name)
	if len(dirNames) > 1 {
		dirName := dirNames[len(dirNames)-1]
		fs.VisitDirectoryByNamesAndExecute(dirNames[:len(dirNames)-1], func() {
			fs.NewDirectory(fs.currentInodeID, dirName, false)
		}, func() {
			utils.PrintError(fmt.Sprintf("Cannot create folder '%s' at '%s/' because such path does not exist", dirName, strings.Join(dirNames[:len(dirNames)-1], "/")))
		})
	} else {
		fs.NewDirectory(fs.currentInodeID, name, false)
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
	if !fs.ChangeToDirectoryByPath(path) {
		utils.PrintError(fmt.Sprintf("'%s' not found", path))
	}
}
