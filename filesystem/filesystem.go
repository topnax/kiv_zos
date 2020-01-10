package filesystem

type FileSystem interface {
	Format(size int)
	FilePath(filePath string)
}
