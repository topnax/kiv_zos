package app

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strings"
)

func Main(args []string) {
	scanner := bufio.NewScanner(os.Stdin)
	fsApp.Terminate(func(_ int) {})
	kingpin.CommandLine.Terminate(func(_ int) {})
	for scanner.Scan() {
		parseArgs(strings.Split(strings.Trim(scanner.Text(), " "), " "))
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
