package myfilesystem

import "fmt"

func (fs MyFileSystem) PrintContent(node PseudoInode) {
	fs.ReadDataFromInodeFx(node, func(data []byte) {
		fmt.Print(string(data))
	})
}

func (fs MyFileSystem) PrintInfo(inode PseudoInode, item DirectoryItem) {
	stringz := ""
	addresses := fs.GetUsedClusterAddresses(inode)
	for _, address := range addresses {
		stringz += fmt.Sprintf("%d ", address)
	}
	fmt.Printf("%s - %d - %d - %s", item.GetName(), item.NodeID, inode.FileSize, stringz)
}
