package myfilesystem

import (
	log "github.com/sirupsen/logrus"
	"unsafe"
)

const (
	maxFileNameLength = 15
	signatureLength   = 8
	volumeDescriptor  = 251
)

type Address int32
type ID int32
type Size int32
type ReferenceCounter int8

type SuperBlock struct {
	Signature                 [signatureLength]rune
	VolumeDescriptor          [volumeDescriptor]rune
	DiskSize                  Size
	ClusterSize               Size
	ClusterCount              Size
	ClusterBitmapStartAddress Address
	InodeBitmapStartAddress   Address
	InodeStartAddress         Address
	ClusterStartAddress       Address
}

func (superBlock SuperBlock) info() {
	log.Infoln("### SUPERBLOCK INFO ###")
	log.Infoln("ClusterSize:", superBlock.ClusterSize)
	log.Infoln("ClusterCount:", superBlock.ClusterCount)
	log.Infoln("DiskSize:", superBlock.DiskSize)
	log.Infoln("Inode bitmap start address:", superBlock.InodeBitmapStartAddress)
	log.Infoln("Inode start address:", superBlock.InodeStartAddress)
	log.Infoln("Cluster bitmap start address:", superBlock.ClusterBitmapStartAddress)
	log.Infoln("Clust er start address:", superBlock.ClusterStartAddress)
}

func (superBlock SuperBlock) InodeCount() Size {
	return Size(superBlock.ClusterBitmapStartAddress-superBlock.InodeStartAddress) / Size(unsafe.Sizeof(PseudoInode{}))
}

func (superBlock SuperBlock) InodeBitmapSize() Size {
	return Size(superBlock.InodeStartAddress - superBlock.InodeBitmapStartAddress)
}

func (superBlock SuperBlock) ClusterBitmapSize() Size {
	return Size(superBlock.ClusterStartAddress - superBlock.ClusterBitmapStartAddress)
}

type PseudoInode struct {
	IsDirectory bool
	References  ReferenceCounter
	FileSize    Size
	Direct1     Address
	Direct2     Address
	Direct3     Address
	Direct4     Address
	Direct5     Address
	Indirect1   ID
	Indirect2   ID
}

type DirectoryItem struct {
	nodeID ID
	name   [maxFileNameLength]rune
}

type Cluster struct {
	fs      *MyFileSystem
	id      ID
	address Address
}
