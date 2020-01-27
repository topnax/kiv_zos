package utils

import "fmt"

func PrintError(format string) {
	fmt.Printf("\x1b[31;1m" + format + "\x1b[0m\n")
}
func PrintSuccess(format string) {
	fmt.Printf("\x1b[32;1m" + format + "\x1b[0m\n")
}

func PrintHighlight(format string) {
	fmt.Printf("\x1b[35;1m" + format + "\x1b[0m\n")
}

func PrintBlue(content string) {
	fmt.Printf("\x1b[34;1m" + content + "\x1b[0m")
}
