package myfilesystem

import "unsafe"

const (
	directAddresses = 5
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
		return (id - addressesPerCluster - directAddresses) / addressesPerCluster, (id + addressesPerCluster) % addressesPerCluster
	}
}
