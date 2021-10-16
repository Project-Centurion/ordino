package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/incu6us/goimports-reviser/v2/pkg/module"
	"github.com/pkg/errors"
)

//@todo : add a 5th optional package with a specified path

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
	FilePathUsage  = `either provide "filepath/to/directory" or "./..." to sort recursively`
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
	fmt.Printf("unamed arg : %s", FilePathUsage)
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
		fmt.Println(colorError("exited: not enough arguments for flag order"))
		os.Exit(1)
	}

	for _, order := range orderSplitted {
		if (order != stdPkg) && (order != aliasedPkg) && (order != projectPkg) && (order != generalPkg) {
			fmt.Println(colorError(fmt.Sprintf("exited: order flag must either be %s, %s, %s, or %s", stdPkg, aliasedPkg, projectPkg, generalPkg)))
			os.Exit(1)
		}
	}

	if isRecursive {
		output = defaultOutput
		RunCommandRecursive(projectName, filePath, orderSplitted)
		os.Exit(0)
	}

	RunCommand(projectName, filePath, orderSplitted)
	os.Exit(0)
}

func colorError(error string) string {
	return fmt.Sprintf("%s%s%s", ColorRed, error, ColorReset)
}

func colorWorked(success string) string {
	return fmt.Sprintf("%s%s%s", ColorGreen, success, ColorReset)
}

func RunCommandRecursive(projectName, path string, order []string) {

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(colorError(err.Error()))
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".go" {
			RunCommand(projectName, path, order)
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}

func RunCommand(projectName, filePath string, orderSplitted []string) {

	formattedOutput, hasChange, err := Execute(projectName, filePath, orderSplitted)
	if err != nil {
		log.Fatalf(colorError(fmt.Sprintf("%+v", errors.WithStack(err))))
	}

	if output == StdOutput {
		fmt.Print(string(formattedOutput))
	} else if output == defaultOutput {
		if !hasChange {
			return
		}

		fmt.Println(colorWorked(fmt.Sprintf("	imports sorted: %v", filePath)))

		if err := ioutil.WriteFile(filePath, formattedOutput, 0644); err != nil {
			log.Fatalf(colorError(fmt.Sprintf("failed to write fixed result to file(%s): %+v", filePath, errors.WithStack(err))))
		}
	}

}

func validateSinglePathParam(filePath string) error {
	if filePath == "" {
		return errors.Errorf("-%s should be set", filePathArg)
	}

	if _, err := os.Stat(filePath); err != nil {
		return err
	}

	if filepath.Ext(filePath) != ".go" {
		return errors.Errorf("%s is not a go file", filePath)
	}

	return nil
}

func validateOutputParam(output string) error {
	if output == "" {
		return nil
	}
	if output != StdOutput && output != defaultOutput {
		return errors.Errorf(`output does not have to be set but can either be "%s" or "%s". Default : "%s"`, defaultOutput, StdOutput, defaultOutput)
	}

	return nil
}

func determineProjectName(projectName, filePath string) (string, error) {
	if projectName == "" {

		projectRootPath, err := module.GoModRootPath(filePath)
		if err != nil {
			return "", err
		}

		moduleName, err := module.Name(projectRootPath)
		if err != nil {
			return "", err
		}

		return moduleName, nil
	}

	return projectName, nil

}
