package main

import (
	"kiv_zos/filesystem/app"
	"kiv_zos/myfilesystem"
	"os"
)

func main() {
	app.Main(os.Args[1:], &myfilesystem.MyFileSystem{})
}
