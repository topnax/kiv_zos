package myfilesystem

type MyFileSystem struct {
	filePath     string
	superBlock   SuperBlock
	currentInode PseudoInode
}

func (fs *MyFileSystem) FilePath(filePath string) {
	fs.filePath = filePath
}

func NewMyFileSystem(filePath string) MyFileSystem {
	return MyFileSystem{
		filePath: filePath,
	}
}
