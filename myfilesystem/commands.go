package myfilesystem

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"kiv_zos/utils"
	"os"
	"strings"
)

func (fs *MyFileSystem) Copy(src string, dst string) {
	moved := false
	fallbackNodeId := fs.currentInodeID
	fs.VisitDirectoryByPathAndExecute(src, func() {
		srcTarget := GetTargetName(src)
		srcDirItem := fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), srcTarget)
		srcNodeId := srcDirItem.NodeID
		if srcNodeId != -1 {
			fs.currentInodeID = fallbackNodeId
			fs.VisitDirectoryByPathAndExecute(dst, func() {
				dstTarget := GetTargetName(dst)
				if strings.Trim(dstTarget, " ") == "" {
					dstTarget = srcTarget
				}
				if fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), dstTarget).NodeID != -1 {
					utils.PrintError(fmt.Sprintf("CANNOT OVERWRITE '%s' THAT EXISTS AT '%s'", dstTarget, dst))
				} else {
					dstNode := PseudoInode{}
					dstNodeId := fs.AddInode(dstNode)
					if dstNodeId < 0 {
						utils.PrintError("NOT ENOUGH INODES, ABORTING")
						return
					}
					fs.AddDirItem(DirectoryItem{
						NodeID: dstNodeId,
						Name:   NameToDirName(dstTarget),
					}, fs.currentInodeID)
					srcNode := fs.GetInodeAt(srcNodeId)
					clusterIndex := 0
					var clusterData [ClusterSize]byte
					fs.ReadDataFromInodeFx(srcNode, func(data []byte) bool {
						copy(clusterData[:], data)
						clusterId := fs.AddDataToInode(clusterData, &dstNode, dstNodeId, clusterIndex)
						if clusterId < 0 {
							utils.PrintError("NOT ENOUGH SPACE (CLUSTERS), ABORTING")
							fs.RemoveAtPath(GetTargetName(dst))
							moved = false
							return false
						}
						clusterIndex++
						dstNode.FileSize += Size(len(data))
						return true
					})
					fs.SetInodeAt(dstNodeId, dstNode)
					moved = true
				}
			}, func() {
				utils.PrintError("PATH NOT FOUND")
			})

			if moved {
				utils.PrintSuccess("OK")
			}
		} else {
			utils.PrintError("FILE NOT FOUND")
		}
	}, func() {
		utils.PrintError("FILE NOT FOUND")
	})
}

func (fs *MyFileSystem) Move(src string, dst string) {
	moved := false
	fallbackNodeId := fs.currentInodeID
	fs.VisitDirectoryByPathAndExecute(src, func() {
		srcTarget := GetTargetName(src)
		srcDirNodeId := fs.currentInodeID
		srcDirItem := fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), srcTarget)
		srcNodeId := srcDirItem.NodeID
		if srcNodeId != -1 {
			fs.currentInodeID = fallbackNodeId
			fs.VisitDirectoryByPathAndExecute(dst, func() {
				dstTarget := GetTargetName(dst)
				if strings.Trim(dstTarget, " ") == "" {
					dstTarget = srcTarget
				}
				if fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), dstTarget).NodeID != -1 {
					utils.PrintError(fmt.Sprintf("CANNOT OVERWRITE '%s' THAT EXISTS AT '%s'", dstTarget, dst))
				} else {
					dstNodeId := fs.AddInode(fs.GetInodeAt(srcNodeId))
					fs.AddDirItem(DirectoryItem{
						NodeID: dstNodeId,
						Name:   NameToDirName(dstTarget),
					}, fs.currentInodeID)
					moved = true
				}
			}, func() {
				utils.PrintError("PATH NOT FOUND")
			})

			if moved {
				fs.RemoveDirItem(srcTarget, srcDirNodeId, false)
				utils.PrintSuccess("OK")
			}
		} else {
			utils.PrintError("SOURCE NOT FOUND")
		}
	}, func() {
		utils.PrintError("SOURCE NOT FOUND")
	})
}

func (fs *MyFileSystem) CopyOut(src string, dst string) {
	fs.VisitDirectoryByPathAndExecute(src, func() {
		id := fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), GetTargetName(src)).NodeID
		if id != -1 {
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				file, err := os.Create(dst)
				if err == nil {
					fs.ReadDataFromInodeFx(fs.GetInodeAt(id), func(data []byte) bool {
						_, err = file.Write(data)
						if err != nil {
							logrus.Error(err)
							utils.PrintError(fmt.Sprintf("An error occurred while writing data to '%s' in the real fs from '%s'.", dst, src))
							return false
						}
						return true
					})
					utils.PrintSuccess("OK")
				} else {
					logrus.Error(err)
					utils.PrintError(fmt.Sprintf("An error occurred while opening '%s' in the real fs.", dst))
				}
			} else {
				utils.PrintError(fmt.Sprintf("'%s' already exists in the real fs. Please use a different file", dst))
			}
		} else {
			utils.PrintError("FILE NOT FOUND")
		}
	}, func() {
		utils.PrintError("FILE NOT FOUND")
	})
}

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
								if id < 0 {
									utils.PrintError("NOT ENOUGH INODES, ABORTING")
									return
								}
								first = false
							}
							clusterId := fs.AddDataToInode(bytes, &node, id, clusterIndex)

							if clusterId >= 0 {
								node.FileSize += Size(read)
							} else {
								utils.PrintError("NOT ENOUGH SPACE (CLUSTERS), ABORTING")
								fs.ShrinkInodeData(&node, id, 0)
								fs.ClearInodeById(id)
								return
							}
						} else {
							break
						}
					} else {
						if first {
							utils.PrintError(fmt.Sprintf("Could not read file '%s!", src))
						}
						break
					}
					clusterIndex++
				}
				if !first {
					fs.SetInodeAt(id, node)
					//utils.PrintSuccess(fmt.Sprintf("Successfully copied a file of length %d bytes (%d kB)", node.FileSize, node.FileSize/1024))
					utils.PrintSuccess("OK")
					fs.AddDirItem(DirectoryItem{
						NodeID: id,
						Name:   NameToDirName(GetTargetName(dst)),
					}, fs.currentInodeID)
				}
			} else {
				utils.PrintError("FILE NOT FOUND")
			}
		} else {
			utils.PrintError(fmt.Sprintf("FILE '%s' ALREADY EXISTS AT '%s'", GetTargetName(dst), dst))
		}
	}, func() {
		utils.PrintError("PATH NOT FOUND")
	})
}

func (fs MyFileSystem) Print(path string) {
	tgtName := GetTargetName(path)
	fs.VisitDirectoryByPathAndExecute(path, func() {
		fs.PrintContent(fs.GetInodeAt(fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), tgtName).NodeID))
	}, func() {
		utils.PrintError(fmt.Sprintf("FILE NOT FOUND", path))
	})
}

func (fs MyFileSystem) PrintContent(node PseudoInode) {
	fs.ReadDataFromInodeFx(node, func(data []byte) bool {
		fmt.Printf("%s", data)
		return true
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
		id := fs.FindDirItemByName(fs.ReadDirItems(fs.currentInodeID), tgtName).NodeID
		if id > -1 {
			fs.PrintInfo(id)
		} else {
			utils.PrintError("FILE NOT FOUND")
		}
	}, func() {
		utils.PrintError("FILE NOT FOUND")
	})
}

func (fs *MyFileSystem) Remove(path string) {
	if fs.RemoveAtPath(path) {
		utils.PrintSuccess("OK")
	} else {
		utils.PrintError(fmt.Sprintf("ITEM AT '%s' NOT FOUND", path))
	}
}
func (fs *MyFileSystem) RemoveAtPath(path string) bool {
	tgtName := GetTargetName(path)
	result := false
	fs.VisitDirectoryByPathAndExecute(path, func() {
		if fs.RemoveDirItem(tgtName, fs.currentInodeID, true) {
			result = true
		}
	}, func() {
	})
	return result
}

func (fs *MyFileSystem) BadRemove(path string) {
	tgtName := GetTargetName(path)
	fs.VisitDirectoryByPathAndExecute(path, func() {
		fs.faultyMode = true
		if fs.RemoveDirItem(tgtName, fs.currentInodeID, true) {
			utils.PrintSuccess("OK")
		} else {
			utils.PrintError("FILE NOT FOUND")
		}
		fs.faultyMode = false
	}, func() {
	})
}
