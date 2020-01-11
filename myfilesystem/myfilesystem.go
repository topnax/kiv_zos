package myfilesystem

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"os"
)

type MyFileSystem struct {
	filePath     string
	file         *os.File
	superBlock   SuperBlock
	currentInode PseudoInode
}

func (fs *MyFileSystem) Close() {
	if fs.file != nil {
		_ = fs.file.Close()
	}
}

func (fs *MyFileSystem) IsLoaded() bool {
	return fs.file != nil
}

func (fs *MyFileSystem) Load() bool {
	file, err := os.Open(fs.filePath)
	log.Infof("About to load a filesystem at path of '%s'", fs.filePath)
	if err == nil {
		_, err = file.Seek(0, os.SEEK_SET)
		if err == nil {
			var block SuperBlock
			err = binary.Read(file, binary.LittleEndian, &block)
			if err == nil {
				fs.superBlock = block
				fs.file = file
				fs.superBlock.info()
				return true
			} else {
				log.Errorf("Binary read of a superblock failed, probably broken superblock")
				log.Error(err)
			}
		} else {
			log.Errorf("Could not seek at a file at '%s' at SEEK_SET of 0", fs.filePath)
			log.Error(err)
		}

	} else {
		log.Errorf("Could not find an existing filesystem in file at '%s'", fs.filePath)
	}
	return false
}

func (fs *MyFileSystem) FilePath(filePath string) {
	fs.filePath = filePath
}

func NewMyFileSystem(filePath string) MyFileSystem {
	return MyFileSystem{
		filePath: filePath,
	}
}
