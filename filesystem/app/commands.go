package app

import "gopkg.in/alecthomas/kingpin.v2"

var (
	fsApp = kingpin.New("fs", "A semester project of KIV/ZOS, Stanislav Král 2020")

	cpCommand = fsApp.Command("cp", "Copies a file")
	cpSrc     = cpCommand.Arg("src", "Source").Required().String()
	cpDst     = cpCommand.Arg("dst", "Destination").Required().String()

	mvCommand = fsApp.Command("mv", "Moves a file")
	mvSrc     = mvCommand.Arg("src", "Source").Required().String()
	mvDst     = mvCommand.Arg("dst", "Destination").Required().String()

	rmCommand = fsApp.Command("rm", "Removes a file")
	rmTarget  = rmCommand.Arg("target", "Target").Required().String()

	rmDirCommand = fsApp.Command("rmdir", "Removes a directory, but only, if it's empty")
	rmDirTarget  = rmDirCommand.Arg("target", "Target").Required().String()

	mkdirCommand = fsApp.Command("mkdir", "Creates a directory")
	mkdirDirName = mkdirCommand.Arg("dirname", "Directory name").Required().String()

	lsCommand = fsApp.Command("ls", "Prints directory content")
	lsDirName = lsCommand.Arg("dirname", "Directory name").Default(".").String()

	catCommand = fsApp.Command("cat", "Prints a files content")
	catDirName = catCommand.Arg("filename", "File name").Required().String()

	cdCommand = fsApp.Command("cd", "Changes the current directory to the specified one")
	cdDirName = cdCommand.Arg("dirname", "Directory name").Required().String()

	pwdCommand = fsApp.Command("pwd", "Prints the path to the current working directory")

	checkCommand = fsApp.Command("check", "Does a consistency check")

	testCommand = fsApp.Command("test", "Prints the path to the current working directory")

	infoCommand = fsApp.Command("info", "Prints information about the given file/directory")
	infoTarget  = infoCommand.Arg("target", "File or directory to be inspected").Required().String()

	inCpCommand = fsApp.Command("incp", "Copies a file from REAL fs to the PSEUDO one")
	inCpSrc     = inCpCommand.Arg("src", "Source file located on the REAL fs").Required().String()
	inCpDst     = inCpCommand.Arg("dst", "Destination on the PSEUDO fs").Required().String()

	outCpCommand = fsApp.Command("outcp", "Copies a file from REAL fs to the PSEUDO one")
	outCpSrc     = outCpCommand.Arg("src", "Source file located on the REAL fs").Required().String()
	outCpDst     = outCpCommand.Arg("dst", "Destination on the PSEUDO fs").Required().String()

	loadCommand = fsApp.Command("load", "Loads the given file and starts to execute the commands inside")
	loadFile    = loadCommand.Arg("file", "File containing commands to be executed").Required().String()

	formatCommand     = fsApp.Command("format", "Initiates the PSEUDO fs by creating a file by the given name and formats it")
	formatDesiredSize = formatCommand.Arg("size", "Size of the desired filesystem. Ex: '5MB'").Required().String()

	exitCommand = fsApp.Command("exit", "Exits the program :)")
)
