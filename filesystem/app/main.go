package app

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"kiv_zos/filesystem"
	"kiv_zos/utils"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func Main(args []string, fs filesystem.FileSystem) {
	log.SetLevel(log.WarnLevel)
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
		utils.PrintBlue(fmt.Sprintf("%s > ", fs.CurrentPath()))
		for scanner.Scan() {
			parseArgs(strings.Split(strings.Trim(scanner.Text(), " "), " "), fs)
			utils.PrintBlue(fmt.Sprintf("%s > ", fs.CurrentPath()))
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
	case cdCommand.FullCommand():
		fs.ChangeDirectory(*cdDirName)
	case formatCommand.FullCommand():
		Format(fs, *formatDesiredSize)
	case mvCommand.FullCommand():
		fs.Move(*mvSrc, *mvDst)
	case rmCommand.FullCommand():
		fs.Remove(*rmTarget)
	case pwdCommand.FullCommand():
		fs.PrintCurrentPath()
	case mkdirCommand.FullCommand():
		fs.CreateNewDirectory(*mkdirDirName)
	case lsCommand.FullCommand():
		fs.ListDirectoryContent(*lsDirName)
	case infoCommand.FullCommand():
		fs.Info(*infoTarget)
	case catCommand.FullCommand():
		fs.Print(*catDirName)
	case checkCommand.FullCommand():
		fs.ConsistencyCheck()
	case rmDirCommand.FullCommand():
		fs.Remove(*rmDirTarget)
	case cpCommand.FullCommand():
		fs.Copy(*cpSrc, *cpDst)
	case inCpCommand.FullCommand():
		fs.CopyIn(*inCpSrc, *inCpDst)
	case outCpCommand.FullCommand():
		fs.CopyOut(*outCpSrc, *outCpDst)
	case loadCommand.FullCommand():
		loadCommands(*loadFile, fs)
	case badRmCommand.FullCommand():
		fs.BadRemove(*badRmTarget)
	case exitCommand.FullCommand():
		exit(fs)
	default:
		fmt.Println("Unknown command... Try --help")
	}
}

func loadCommands(path string, fs filesystem.FileSystem) {
	file, err := os.Open(path)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			utils.PrintBlue(fmt.Sprintf("%s > ", fs.CurrentPath()))
			utils.PrintHighlight(scanner.Text())
			parseArgs(strings.Split(strings.Trim(scanner.Text(), " "), " "), fs)
		}
		utils.PrintSuccess("\nOK")
	} else {
		utils.PrintError(fmt.Sprintf("FILE NOT FOUND"))
	}
}

func exit(fs filesystem.FileSystem) {
	fs.Close()
	fmt.Println("Exitting... :)")
	os.Exit(0)
}
