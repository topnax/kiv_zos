package myfilesystem

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type MyFileSystem struct {
	filePath           string
	File               *os.File
	SuperBlock         SuperBlock
	currentInodeID     ID
	freeClusterIds     []ID
	freeClusterIdIndex int
	faultyMode         bool
	RealMode           bool
}

// sets whether the fs is in a real mode
func (fs *MyFileSystem) SetRealMode(realMode bool) {
	fs.RealMode = realMode
}

// closes the file system
func (fs *MyFileSystem) Close() {
	if fs.File != nil {
		err := fs.File.Close()
		if err != nil {
			log.Error(err)
		} else {
			log.Info("File successfully closed...")
		}
	}
}

// returns true when the fs is loeaded
func (fs *MyFileSystem) IsLoaded() bool {
	return fs.File != nil
}

// loads the file system
func (fs *MyFileSystem) Load() bool {
	fs.File = nil
	file, err := os.OpenFile(fs.filePath, os.O_RDWR, os.ModePerm)
	log.Infof("About to load a filesystem at path of '%s'", fs.filePath)
	if err == nil {
		_, err = file.Seek(0, io.SeekStart)
		if err == nil {
			var block SuperBlock
			err = binary.Read(file, binary.LittleEndian, &block)
			if err == nil {
				fs.SuperBlock = block
				fs.File = file
				fs.SuperBlock.info()
				return true
			} else {
				log.Errorf("Binary read of a superblock failed, probably broken superblock")
				log.Error(err)
			}
		} else {
			log.Errorf("Could not seek at a File at '%s' at SEEK_SET of 0", fs.filePath)
			log.Error(err)
		}

	} else {
		log.Errorf("Could not find an existing filesystem in File at '%s'", fs.filePath)
	}
	return false
}

// sets the file path
func (fs *MyFileSystem) SetFilePath(filePath string) {
	fs.filePath = filePath
}

func NewMyFileSystem(filePath string) MyFileSystem {
	return MyFileSystem{
		filePath: filePath,
	}
}
