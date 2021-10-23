package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Project-Centurion/ordino/sorter"
)

//@todo : add a 5th optional package with a specified path
//@todo : add a yml config file

const (
	projectNameArg = "project-name"
	filePathArg    = "file-path"
	outputArg      = "output"
	orderArg       = "order"
	defaultOutput  = "file"
	StdOutput      = "stdout"
	RecursiveArg   = "./..."
	ColorGreen     = "\033[32m"
	ColorRed       = "\033[31m"
	ColorReset     = "\033[0m"
	FilePathUsage  = `either provide "filepath/to/directory" or "./..." as an unnamed arg to sort recursively`
)

var projectName, output, order string

func init() {
	flag.StringVar(
		&projectName,
		projectNameArg,
		"",
		"Your project name(ex.: github.com/Project-Centurion/ordo). Optional parameter.",
	)

	flag.StringVar(
		&output,
		outputArg,
		defaultOutput,
		`Can be "file" or "stdout". Whether to write the formatted content back to the file or to stdout. Optional parameter.`,
	)

	flag.StringVar(
		&order,
		orderArg,
		"std,alias,project,general",
		`Default is "std, alias, project, general". Imports can be sorted in whichever order between those four, "alias" is optional. Optional paramater.`,
	)

}

func printUsage() {
	if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
		log.Fatalf("failed to print usage: %s", err)
	}

	flag.PrintDefaults()
	fmt.Printf("file-path : %s", FilePathUsage)
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println(colorError(fmt.Sprintf("No file provided : %s", FilePathUsage)))
		os.Exit(1)
	}

	if len(args) > 1 {
		fmt.Println(colorError(fmt.Sprintf(`Too much unflagged arguments defined: %s`, FilePathUsage)))
		os.Exit(1)
	}

	filePath := args[0]

	var isRecursive bool

	if flag.Args()[0] == RecursiveArg {
		isRecursive = true
		path, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}

		filePath = path
	}

	if !isRecursive {
		if err := validateSinglePathParam(filePath); err != nil {
			fmt.Println(colorError(err.Error()))
			os.Exit(1)
		}
	}

	if err := validateOutputParam(output); err != nil {
		fmt.Println(colorError(err.Error()))
		printUsage()
		os.Exit(1)
	}

	projectName, err := determineProjectName(projectName, filePath)
	if err != nil {
		fmt.Println(colorError(err.Error()))
		printUsage()
		os.Exit(1)
	}

	orderSplitted := strings.Split(order, ",")

	if len(orderSplitted) < 3 {
		printError("exited: not enough arguments for flag order")
	}

	for _, order := range orderSplitted {
		if (order != sorter.StdPkg) && (order != sorter.AliasedPkg) && (order != sorter.ProjectPkg) && (order != sorter.GeneralPkg) {
			fmt.Println(colorError(fmt.Sprintf("exited: order flag must either be %s, %s, %s, or %s", sorter.StdPkg, sorter.AliasedPkg, sorter.ProjectPkg, sorter.GeneralPkg)))
			os.Exit(1)
		}
	}

	if isRecursive {
		output = defaultOutput
		runCommandRecursive(projectName, filePath, orderSplitted)
		os.Exit(0)
	}

	runCommand(projectName, filePath, orderSplitted)
	os.Exit(0)
}
