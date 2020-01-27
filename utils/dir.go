package utils

import "strings"

const (
	FolderSeparator = "/"
)

func GetDirNames(path string) []string {
	dirs := []string{}
	if path[0] == '/' {
		dirs = append(dirs, "/")
	}

	return append(dirs, strings.Split(path, FolderSeparator)...)
}
