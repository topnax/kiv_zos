package myfilesystem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"kiv_zos/utils"
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

func (fs *MyFileSystem) AddDirItem(item DirectoryItem, nodeId ID) {
	node := fs.GetInodeAt(nodeId)
	if node.IsDirectory {
		fs.AppendDirItem(item, node, nodeId)
	} else {
		log.Warnf("Trying to add a directory item to an inode that is not a directory")
	}
}

func (fs *MyFileSystem) ReadDirItems(nodeId ID) []DirectoryItem {
	node := fs.GetInodeAt(nodeId)
	if node.IsDirectory {
		data := fs.ReadDataFromInode(node)

		buf := new(bytes.Buffer)

		log.Infof("read data %v", data)

		buf.Write(data)
		var items []DirectoryItem
		for i := 0; i < int(node.FileSize)/int(unsafe.Sizeof(DirectoryItem{})); i++ {
			var item DirectoryItem
			err := binary.Read(buf, binary.LittleEndian, &item)
			if err == nil {
				items = append(items, item)
			} else {
				panic("Could not read an directory item")
			}
		}
		return items
	} else {
		log.Warnf("Trying to read a directory item from an inode=%d that is not a directory", nodeId)
	}
	return []DirectoryItem{}
}

func (fs *MyFileSystem) AppendDirItem(item DirectoryItem, node PseudoInode, nodeId ID) ID {
	if node.IsDirectory {
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, item)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
			panic(err)
		}
		dirItemBytes := make([]byte, unsafe.Sizeof(DirectoryItem{}))
		_, err = buf.Read(dirItemBytes)
		if err != nil {
			log.Error(err)
			panic(err)
		}
		dirId := NextDirItemIndex(node)
		readOrders := GetReadOrder(int(dirId)*int(unsafe.Sizeof(DirectoryItem{})), int(unsafe.Sizeof(DirectoryItem{})))
		log.Infof("read dirItemBytes add=%v", dirItemBytes)
		written := 0
		for _, readOrder := range readOrders {
			clusterBytes := fs.ReadDataFromInodeAt(node, int(readOrder.ClusterId))
			log.Infof("curr clusterbytes=%v", clusterBytes)

			firstHalfBytes := clusterBytes[0:readOrder.Start]
			secondHalfBytes := clusterBytes[readOrder.Start+readOrder.Bytes:]

			log.Infof("first half bytes=%v readOrder.STart=%d", firstHalfBytes, readOrder.Start)

			write := append(firstHalfBytes, dirItemBytes[written:written+readOrder.Bytes]...)
			write = append(write, secondHalfBytes...)
			written += readOrder.Bytes

			var final [ClusterSize]byte
			copy(final[:], write)

			id := fs.AddDataToInode(final, &node, nodeId, int(readOrder.ClusterId))
			log.Infof("written @%d add=%v", id, final)
		}

		node.FileSize += Size(unsafe.Sizeof(DirectoryItem{}))
		fs.SetInodeAt(nodeId, node)
		return dirId
	} else {
		log.Warnf("Trying to add a directory item to an inode that is not a directory")
	}
	return -1
}

func (fs *MyFileSystem) RemoveDirItem(delete string, nodeId ID, removeData bool) bool {
	items := fs.ReadDirItems(nodeId)
	deleteIndex := -1
	for index, item := range items {
		// find delete to be deleted
		if item.Name == NameToDirName(delete) {
			deleteIndex = index
			break
		}
	}
	if deleteIndex == -1 {
		return false
	}

	nodeIdToBeDeleted := items[deleteIndex].NodeID
	nodeToBeDeleted := fs.GetInodeAt(items[deleteIndex].NodeID)

	if nodeToBeDeleted.IsDirectory {
		if len(fs.ReadDirItems(nodeIdToBeDeleted)) > 2 {
			utils.PrintError("Cannot delete a directory that is not empty")
			return false
		}
	}

	if removeData {
		fs.ShrinkInodeData(&nodeToBeDeleted, nodeIdToBeDeleted, 0)
	}
	if !fs.faultyMode {
		fs.ClearInodeById(nodeIdToBeDeleted)
	} else {
		nodeToBeDeleted.Direct2 = 0
		fs.SetInodeAt(nodeIdToBeDeleted, nodeToBeDeleted)
		utils.PrintHighlight(fmt.Sprintf("FAULTY MODE ENABLED FOR INODE OF ID=%d", nodeIdToBeDeleted))
	}

	if deleteIndex != len(items)-1 {
		items[deleteIndex] = items[len(items)-1]
	}

	items = items[:len(items)-1]

	fs.WriteDataToInode(nodeId, ItemsToBytes(items))
	node := fs.GetInodeAt(nodeId)
	fs.ShrinkInodeData(&node, nodeId, Size(len(items)*int(unsafe.Sizeof(DirectoryItem{}))))

	return true
}

func NextDirItemIndex(node PseudoInode) ID {
	if node.FileSize == 0 {
		return 0
	}
	return ID(node.FileSize) / ID(unsafe.Sizeof(DirectoryItem{}))
}

func (fs MyFileSystem) GetDirItemsCount(node PseudoInode) Size {
	return node.FileSize / Size(unsafe.Sizeof(DirectoryItem{}))
}

func ItemsToBytes(items []DirectoryItem) []byte {
	if len(items) <= 0 {
		log.Errorf("Trying to convert empty item array to bytes")
		return []byte{}
	}

	buf := new(bytes.Buffer)
	for _, item := range items {
		err := binary.Write(buf, binary.LittleEndian, item)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
			panic(err)
		}
	}

	dirItemBytes := make([]byte, int(unsafe.Sizeof(items[0]))*len(items))
	_, err := buf.Read(dirItemBytes)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	return dirItemBytes
}

func (fs MyFileSystem) ListDirectory(nodeId ID) {
	node := fs.GetInodeAt(nodeId)
	if node.IsDirectory {
		//utils.PrintSuccess(fmt.Sprintf("Items of %s", item.GetName()))
		items := fs.ReadDirItems(nodeId)

		for _, item := range items {
			node = fs.GetInodeAt(item.NodeID)
			var char string
			if node.IsDirectory {
				char = "+"
			} else {
				char = "-"
			}

			fmt.Printf("%s%s\n", char, item.GetName())
		}
	} else {
		log.Errorf("Trying to list directory of node=%d that is not a directory", nodeId)
	}
}

func (fs *MyFileSystem) NewDirectory(parentNodeId ID, name string, formatting bool) ID {

	if !formatting && fs.FindDirItemByName(fs.ReadDirItems(parentNodeId), name).NodeID != -1 {
		log.Warnf("Directory '%s' already exists in %s", name, fs.FindDirPath(parentNodeId))
		return -1
	}

	newNode := PseudoInode{
		IsDirectory: true,
	}
	newNodeId := fs.AddInode(newNode)

	// add a new dir item to the parent node
	if !formatting {
		fs.AddDirItem(DirectoryItem{
			NodeID: newNodeId,
			Name:   NameToDirName(name),
		}, parentNodeId)
	}

	fs.AddDirItem(DirectoryItem{
		NodeID: newNodeId,
		Name:   NameToDirName("."),
	}, newNodeId)

	fs.AddDirItem(DirectoryItem{
		NodeID: parentNodeId,
		Name:   NameToDirName(".."),
	}, newNodeId)

	return newNodeId
}

func NameToDirName(name string) [maxFileNameLength]byte {
	var dirNameBytes [maxFileNameLength]byte
	copy(dirNameBytes[:], name)
	return dirNameBytes
}

func (fs *MyFileSystem) CurrentPath() string {
	return fs.FindDirPath(fs.currentInodeID)
}

func (fs *MyFileSystem) FindDirPath(currentInodeId ID) string {
	path := ""
	item, parentId := fs.FindDirItemById(currentInodeId)
	path += item.GetName() + "/"
	for item.NodeID != 0 {
		itemx, pid := fs.FindDirItemById(parentId)
		item = itemx
		parentId = pid
		path = item.GetName() + "/" + path
	}

	return path
}

func (fs *MyFileSystem) FindDirItemById(currentInodeId ID) (DirectoryItem, ID) {
	if currentInodeId == 0 {
		return DirectoryItem{0, NameToDirName("")}, 0
	}
	node := fs.GetInodeAt(currentInodeId)

	if !node.IsDirectory {
		panic("FindDirNameById called on non-dir node")
	} else {
		items := fs.ReadDirItems(currentInodeId)

		return fs.FindDirItemByNodeId(fs.ReadDirItems(items[1].NodeID), currentInodeId), items[1].NodeID
	}
}

func (fs *MyFileSystem) FindDirItemByName(items []DirectoryItem, name string) DirectoryItem {
	for _, item := range items {
		if item.Name == NameToDirName(name) {
			return item
		}
	}
	return DirectoryItem{
		NodeID: -1,
	}
}

func (fs *MyFileSystem) FindDirItemByNodeId(items []DirectoryItem, id ID) DirectoryItem {
	for _, item := range items {
		if item.NodeID == id {
			return item
		}
	}
	panic("Dir item mot found by ID")
	return DirectoryItem{
		NodeID: -1,
	}
}
