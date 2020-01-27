package filesystem

type FileSystem interface {
	Format(size int)
	FilePath(filePath string)
	Load() bool
	IsLoaded() bool
	Close()
	PrintCurrentPath()
	CurrentPath() string
	CreateNewDirectory(name string)
	ChangeDirectory(path string)
	ListDirectoryContent(name string)
	Info(path string)
	Print(path string)
	RemoveDirectory(path string)
}
