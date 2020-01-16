package myfilesystem

import "unsafe"

const (
	directAddresses = 5
	FileTooLarge    = -3
	NoIndirect      = -2
	FirstIndirect   = -1
)

func (fs *MyFileSystem) GetClusterPath(id int) (int, int) {
	const addressesPerCluster = ClusterSize / int(unsafe.Sizeof(Address(0)))
	if id < directAddresses {
		return id, NoIndirect
	} else if id < addressesPerCluster+directAddresses {
		return id - directAddresses, FirstIndirect
	} else {
		indirectId := (id - addressesPerCluster - directAddresses) / addressesPerCluster
		if indirectId >= addressesPerCluster {
			return FileTooLarge, FileTooLarge
		} else {
			return (id - addressesPerCluster - directAddresses) % addressesPerCluster, indirectId
		}
	}
}
