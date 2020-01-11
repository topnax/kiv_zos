package filesystem

type FileSystem interface {
	Format(size int)
	FilePath(filePath string)
	Load() bool
	IsLoaded() bool
	Close()
}
