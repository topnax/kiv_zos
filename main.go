package main

import (
	"kiv_zos/filesystem/app"
	"kiv_zos/myfilesystem"
	"os"
)

func main() {
	myfs := myfilesystem.MyFileSystem{}
	myfs.RealMode = true
	app.Main(os.Args[1:], &myfs)
}
