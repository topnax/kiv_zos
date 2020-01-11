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
	signature               [signatureLength]rune
	volumeDescriptor        [volumeDescriptor]rune
	diskSize                Size
	clusterSize             Size
	clusterCount            ClusterCount
	dataBitmapStartAddress  Address
	inodeBitmapStartAddress Address
	inodeStartAddress       Address
	dataStartAddress        Address
}

func (superBlock SuperBlock) info() {
	log.Infoln("### SUPERBLOCK INFO ###")
	log.Infoln("ClusterSize:", superBlock.clusterSize)
	log.Infoln("ClusterCount:", superBlock.clusterCount)
	log.Infoln("DiskSize:", superBlock.diskSize)
	log.Infoln("Inode bitmap start address:", superBlock.inodeBitmapStartAddress)
	log.Infoln("Inode start address:", superBlock.inodeStartAddress)
	log.Infoln("Data bitmap start address:", superBlock.dataBitmapStartAddress)
	log.Infoln("Data start address:", superBlock.dataStartAddress)
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
