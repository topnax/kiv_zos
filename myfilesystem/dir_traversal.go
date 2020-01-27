package myfilesystem

import "strings"

const (
	FolderSeparator = "/"
)

func (fs *MyFileSystem) ChangeToDirectoryByPath(path string) bool {
	return fs.ChangeToDirectoryByNames(GetDirNames(path))
}

func (fs *MyFileSystem) ChangeToDirectoryByNames(dirNames []string) bool {

	fallbackNodeId := fs.currentInodeID

	for _, name := range dirNames {
		if fs.ChangeToDirectoryByName(name) == -1 {
			fs.currentInodeID = fallbackNodeId
			return false
		}
	}

	return true
}

func (fs *MyFileSystem) VisitDirectoryByPathAndExecute(path string, sfx func(), efx func()) {
	dirNames := GetDirNames(path)
	dirNames = dirNames[:len(dirNames)-1]
	fallbackNodeId := fs.currentInodeID

	for _, name := range dirNames {
		if fs.ChangeToDirectoryByName(name) == -1 {
			efx()
			fs.currentInodeID = fallbackNodeId
			return
		}
	}
	sfx()
	fs.currentInodeID = fallbackNodeId
}

func (fs *MyFileSystem) ChangeToDirectoryByName(name string) ID {

	if name == "/" {
		fs.currentInodeID = 0
		return 0
	}

	items := fs.ReadDirItems(fs.currentInodeID)

	item := fs.FindDirItemByName(items, name)

	if item.NodeID != -1 {
		fs.currentInodeID = item.NodeID
	}

	return item.NodeID
}

func GetDirNames(path string) []string {
	dirs := strings.Split(path, FolderSeparator)

	if path[0] == '/' {
		dirs[0] = "/"
	}

	//if len(dirs) == 0 {
	//	return [] string {path}
	//}

	return dirs
}
