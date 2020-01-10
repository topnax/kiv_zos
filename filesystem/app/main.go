package app

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"kiv_zos/filesystem"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func Main(args []string, fs filesystem.FileSystem) {
	args = append(args, "myfs")
	if len(args) > 0 {
		fs.FilePath(args[0])
		log.Infof("File path set to '%s'", args[0])
		if true {
			Format(fs, "5MB")
		} else {
			scanner := bufio.NewScanner(os.Stdin)
			fsApp.Terminate(func(_ int) {})
			kingpin.CommandLine.Terminate(func(_ int) {})
			for scanner.Scan() {
				parseArgs(strings.Split(strings.Trim(scanner.Text(), " "), " "))
			}
		}
	} else {
		log.Errorln("File path argument not supplied")
	}

}

func Format(fs filesystem.FileSystem, size string) {
	numberStr := ""
	unitStr := ""
	for _, c := range size {
		if !unicode.IsDigit(c) {
			unitStr += string(c)
		} else {
			numberStr += string(c)
		}
	}

	sizeNum, _ := strconv.Atoi(numberStr)
	if unitStr == "MB" {
		fs.Format(sizeNum * 1024 * 1024)
	}
}

func parseArgs(args []string) {
	switch kingpin.MustParse(fsApp.Parse(args)) {
	case cpCommand.FullCommand():
		log.Infof("cp: '%s' '%s'", *cpSrc, *cpDst)
	case mvCommand.FullCommand():
		fmt.Println("mv: ", *mvSrc, *mvDst)
	case rmCommand.FullCommand():
		fmt.Println("rm: ", *rmTarget)
	case exitCommand.FullCommand():
		exit()
	default:
		fmt.Println("Unknown command... Try --help")
	}
}

func exit() {
	fmt.Println("Exitting... :)")
	os.Exit(0)
}
