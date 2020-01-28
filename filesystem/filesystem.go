package filesystem

type FileSystem interface {
	Format(size int)
	SetFilePath(filePath string)
	Load() bool
	IsLoaded() bool
	Close()
	PrintCurrentPath()
	CurrentPath() string
	CreateNewDirectory(name string)
	ChangeDirectory(path string)
	ConsistencyCheck()
	ListDirectoryContent(name string)
	Info(path string)
	BadRemove(path string)
	Print(path string)
	Copy(src string, dst string)
	CopyIn(src string, dst string)
	CopyOut(src string, dst string)
	Move(src string, dst string)
	Remove(path string)
}
