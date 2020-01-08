package main

import (
	"kiv_zos/filesystem/app"
	"os"
)

func main() {
	app.Main(os.Args[1:])
}
