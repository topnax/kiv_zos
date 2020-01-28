package myfilesystem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"kiv_zos/utils"
	"unsafe"
)

// a struct that represents an order, used to select which bytes from which cluster should be read/written at
type IOOrder struct {
	ClusterId ID
	Start     int
	Bytes     int
}

// returns an array of orders, that will be used to write/read at the given offset
func GetIOOrder(offset int, bytes int) []IOOrder {
	clusterId := ID(offset / ClusterSize)
	overflow := (offset%ClusterSize)+bytes > ClusterSize

	if !overflow {
		return []IOOrder{{
			ClusterId: clusterId,
			Start:     offset % ClusterSize,
			Bytes:     bytes,
		}}
	} else {
		return []IOOrder{{
			ClusterId: clusterId,
			Start:     offset % ClusterSize,
			Bytes:     ClusterSize - (offset % ClusterSize),
		}, {
			ClusterId: clusterId + 1,
			Start:     0,
			Bytes:     bytes - (ClusterSize - (offset % ClusterSize)),
		}}
	}
}

// adds a directory item to the given node
func (fs *MyFileSystem) AddDirItem(item DirectoryItem, nodeId ID) {
	node := fs.GetInodeAt(nodeId)
	if node.IsDirectory {
		fs.AppendDirItem(item, node, nodeId)
	} else {
		log.Warnf("Trying to add a directory item to an inode that is not a directory")
	}
}

// reads directory items from the given node
func (fs *MyFileSystem) ReadDirItems(nodeId ID) []DirectoryItem {
	node := fs.GetInodeAt(nodeId)
	if node.IsDirectory {
		data := fs.ReadDataFromInode(node)

		buf := new(bytes.Buffer)

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

// appends a directory item to the given node
func (fs *MyFileSystem) AppendDirItem(item DirectoryItem, node PseudoInode, nodeId ID) ID {
	if node.IsDirectory {
		buf := new(bytes.Buffer)
		// convert item to binary
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
		ioOrders := GetIOOrder(int(dirId)*int(unsafe.Sizeof(DirectoryItem{})), int(unsafe.Sizeof(DirectoryItem{})))
		written := 0
		// split into io orders, correctly append dir items to existing clusters
		for _, ioOrder := range ioOrders {
			clusterBytes := fs.ReadDataFromInodeAt(node, int(ioOrder.ClusterId))

			// split cluster bytes
			firstHalfBytes := clusterBytes[0:ioOrder.Start]
			secondHalfBytes := clusterBytes[ioOrder.Start+ioOrder.Bytes:]

			log.Infof("first half bytes=%v ioOrder.Start=%d", firstHalfBytes, ioOrder.Start)

			write := append(firstHalfBytes, dirItemBytes[written:written+ioOrder.Bytes]...)
			write = append(write, secondHalfBytes...)
			written += ioOrder.Bytes

			var final [ClusterSize]byte
			copy(final[:], write)

			id := fs.AddDataToInode(final, &node, nodeId, int(ioOrder.ClusterId))
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

// removes a directory item from the given node
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

	// node to be deleted
	nodeIdToBeDeleted := items[deleteIndex].NodeID
	nodeToBeDeleted := fs.GetInodeAt(items[deleteIndex].NodeID)

	if nodeToBeDeleted.IsDirectory {
		if len(fs.ReadDirItems(nodeIdToBeDeleted)) > 2 {
			utils.PrintError("Cannot delete a directory that is not empty")
			return false
		}
	}

	if removeData {
		// remove file data by shrinking it's data part to 0 bytes
		fs.ShrinkInodeData(&nodeToBeDeleted, nodeIdToBeDeleted, 0)
	}
	if !fs.faultyMode {
		fs.ClearInodeById(nodeIdToBeDeleted)
	} else {
		// while in faulty mode, node is not deleted and it's second direct pointer is compromised
		nodeToBeDeleted.Direct2 = 0
		fs.SetInodeAt(nodeIdToBeDeleted, nodeToBeDeleted)
		utils.PrintHighlight(fmt.Sprintf("FAULTY MODE ENABLED FOR INODE OF ID=%d", nodeIdToBeDeleted))
	}

	// if the to be deleted directory item is not the last one, it is swapped with the last one
	if deleteIndex != len(items)-1 {
		items[deleteIndex] = items[len(items)-1]
	}

	// the last item is removed
	items = items[:len(items)-1]

	fs.WriteDataToInode(nodeId, ItemsToBytes(items))
	node := fs.GetInodeAt(nodeId)
	fs.ShrinkInodeData(&node, nodeId, Size(len(items)*int(unsafe.Sizeof(DirectoryItem{}))))

	return true
}

// returns the next available dir item index
func NextDirItemIndex(node PseudoInode) ID {
	if node.FileSize == 0 {
		return 0
	}
	return ID(node.FileSize) / ID(unsafe.Sizeof(DirectoryItem{}))
}

// returns the number of directory items in a node
func (fs MyFileSystem) GetDirItemsCount(node PseudoInode) Size {
	return node.FileSize / Size(unsafe.Sizeof(DirectoryItem{}))
}

// converts a slice of directory items to a slice of bytes
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

// lists the content of a directory given by a node
func (fs MyFileSystem) ListDirectory(nodeId ID) {
	node := fs.GetInodeAt(nodeId)
	if node.IsDirectory {
		items := fs.ReadDirItems(nodeId)

		for _, item := range items {
			node = fs.GetInodeAt(item.NodeID)
			var char string
			if node.IsDirectory {
				char = "+"
			} else {
				char = "-"
			}

			fmt.Printf("%s %s\n", char, item.GetName())
		}
	} else {
		log.Errorf("Trying to list directory of node=%d that is not a directory", nodeId)
	}
}

// creates a new directory ad the given parent node
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

// converts a name to an array of bytes
func NameToDirName(name string) [maxFileNameLength]byte {
	var dirNameBytes [maxFileNameLength]byte
	copy(dirNameBytes[:], name)
	return dirNameBytes
}

// returns the current path
func (fs *MyFileSystem) CurrentPath() string {
	return fs.FindDirPath(fs.currentInodeID)
}

// recursively creates the current directory path
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

// finds a directory item by a node id
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

// finds a directory item by a name
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

// finds a directory item by an ID
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
