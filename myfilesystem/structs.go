package myfilesystem

import (
	log "github.com/sirupsen/logrus"
	"strings"
	"unsafe"
)

const (
	maxFileNameLength = 12
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
	log.Infoln("Inode bitmap Start address:", superBlock.InodeBitmapStartAddress)
	log.Infoln("Inode Start address:", superBlock.InodeStartAddress)
	log.Infoln("Cluster bitmap Start address:", superBlock.ClusterBitmapStartAddress)
	log.Infoln("Clust er Start address:", superBlock.ClusterStartAddress)
}

// calculates the inode count
func (superBlock SuperBlock) InodeCount() Size {
	return Size(superBlock.ClusterBitmapStartAddress-superBlock.InodeStartAddress) / Size(unsafe.Sizeof(PseudoInode{}))
}

// calculates the inode bitmap size
func (superBlock SuperBlock) InodeBitmapSize() Size {
	return Size(superBlock.InodeStartAddress - superBlock.InodeBitmapStartAddress)
}

// calculates the cluster bitmap size
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
	NodeID ID
	Name   [maxFileNameLength]byte
}

// returns the dir item name
func (dirItem DirectoryItem) GetName() string {
	str := string(dirItem.Name[:])
	str = strings.Replace(str, string(byte(0)), "", -1)
	return str
}

type Cluster struct {
	fs      *MyFileSystem
	id      ID
	address Address
}
