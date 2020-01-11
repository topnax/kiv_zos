package myfilesystem

import (
	log "github.com/sirupsen/logrus"
)

const (
	maxFileNameLength = 20
	signatureLength   = 8
	volumeDescriptor  = 251
)

type Address int32
type NodeID int32
type Size int32
type ClusterCount int32
type ReferenceCounter int8

type SuperBlock struct {
	Signature               [signatureLength]rune
	VolumeDescriptor        [volumeDescriptor]rune
	DiskSize                Size
	ClusterSize             Size
	ClusterCount            ClusterCount
	DataBitmapStartAddress  Address
	InodeBitmapStartAddress Address
	InodeStartAddress       Address
	DataStartAddress        Address
}

func (superBlock SuperBlock) info() {
	log.Infoln("### SUPERBLOCK INFO ###")
	log.Infoln("ClusterSize:", superBlock.ClusterSize)
	log.Infoln("ClusterCount:", superBlock.ClusterCount)
	log.Infoln("DiskSize:", superBlock.DiskSize)
	log.Infoln("Inode bitmap start address:", superBlock.InodeBitmapStartAddress)
	log.Infoln("Inode start address:", superBlock.InodeStartAddress)
	log.Infoln("Data bitmap start address:", superBlock.DataBitmapStartAddress)
	log.Infoln("Data start address:", superBlock.DataStartAddress)
}

type PseudoInode struct {
	isDirectory bool
	references  ReferenceCounter
	fileSize    Size
	direct1     Address
	direct2     Address
	direct3     Address
	direct4     Address
	direct5     Address
	indirect1   NodeID
	indirect2   NodeID
}

type DirectoryItem struct {
	nodeID NodeID
	name   [maxFileNameLength]rune
}
