package myfilesystem

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"kiv_zos/utils"
	"os"
)

func (fs *MyFileSystem) CopyIn(src string, dst string) {
	fs.VisitDirectoryByPathAndExecute(dst, func() {
		if fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), GetTargetName(dst)).NodeID == -1 {
			file, err := os.Open(src)
			if err == nil {
				first := true
				id := ID(-1)
				node := PseudoInode{}
				clusterIndex := 0
				var bytes [ClusterSize]byte
				for {
					read, err := file.Read(bytes[:])
					if err == nil {
						if read > 0 {
							if first {
								id = fs.AddInode(node)
								first = false
							}
							if id >= 0 {
								fs.AddDataToInode(bytes, &node, id, clusterIndex)
								node.FileSize += Size(read)
							}
						} else {
							break
						}
					} else {
						logrus.Warn(err)
						if first {
							utils.PrintError(fmt.Sprintf("Could not read file '%s!", src))
						}
						break
					}
					clusterIndex++
					if !first {
						fs.SetInodeAt(id, node)
						utils.PrintSuccess(fmt.Sprintf("Successfully copied a file of length %d bytes (%d kB)", node.FileSize, node.FileSize/1024))
						fs.AddDirItem(DirectoryItem{
							NodeID: id,
							Name:   NameToDirName(GetTargetName(dst)),
						}, fs.currentInodeID)
					}
				}
			} else {
				utils.PrintError(fmt.Sprintf("Could not find a file in the real FS at '%s'", src))
			}
		} else {
			utils.PrintError(fmt.Sprintf("File '%s' already exists at '%s'", GetTargetName(dst), dst))
		}
	}, func() {
		utils.PrintError(fmt.Sprintf("'%s' destination path not found", dst))
	})
}

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
