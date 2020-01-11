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
		if fs.Load() {
			log.Infoln("Filesystem correctly loaded")
		} else {
			log.Infoln("Filesystem not loaded, please format first.")
		}
		scanner := bufio.NewScanner(os.Stdin)
		fsApp.Terminate(func(_ int) {})
		kingpin.CommandLine.Terminate(func(_ int) {})
		for scanner.Scan() {
			parseArgs(strings.Split(strings.Trim(scanner.Text(), " "), " "), fs)
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

	if unitStr == "kB" {
		fs.Format(sizeNum * 1024)
	}
}

func parseArgs(args []string, fs filesystem.FileSystem) {
	parseResult := kingpin.MustParse(fsApp.Parse(args))

	if parseResult != formatCommand.FullCommand() && !fs.IsLoaded() {
		log.Warnln("Filesystem not formatted. Use format command first please")
		return
	}
	switch parseResult {
	case formatCommand.FullCommand():
		Format(fs, *formatDesiredSize)
	case cpCommand.FullCommand():
		log.Infof("cp: '%s' '%s'", *cpSrc, *cpDst)
	case mvCommand.FullCommand():
		fmt.Println("mv: ", *mvSrc, *mvDst)
	case rmCommand.FullCommand():
		fmt.Println("rm: ", *rmTarget)
	case exitCommand.FullCommand():
		exit(fs)
	default:
		fmt.Println("Unknown command... Try --help")
	}
}

func exit(fs filesystem.FileSystem) {
	fs.Close()
	fmt.Println("Exitting... :)")
	os.Exit(0)
}
