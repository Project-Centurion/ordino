package main

import (
	"fmt"
	"os"
)

func colorError(error string) string {
	return fmt.Sprintf("%s%s%s", ColorRed, error, ColorReset)
}

func colorWorked(success string) string {
	return fmt.Sprintf("%s%s%s", ColorGreen, success, ColorReset)
}

func printError(error string) {
	fmt.Println(colorError(error))
	printUsage()
	os.Exit(1)
}
