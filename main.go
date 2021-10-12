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

//@todo(athenais): add ./... arg (recusive) (flagArgs?)

const (
	projectNameArg = "project-name"
	filePathArg    = "file-path"
	outputArg      = "output"
	orderArg       = "order"
	defaultOutput  = "file"
	StdOutput      = "stdout"
)

var projectName, filePath, output, order string

func init() {
	flag.StringVar(
		&filePath,
		filePathArg,
		"",
		"File path to fix imports(ex.: ./reviser/reviser.go). Required parameter.",
	)

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
		`Default is "std, alias, project, general". Imports can be sorted in whichever order between those for. Optional paramater.`,
	)

}

func printUsage() {
	if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
		log.Fatalf("failed to print usage: %s", err)
	}

	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if err := validateRequiredParam(filePath); err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	if err := validateOutputParam(output); err != nil {
		fmt.Printf("%s\n\n", err)
		os.Exit(1)
	}

	var isRecursive bool
	if len(flag.Args()) > 0 {
		if flag.Args()[0] == "./..." {
			isRecursive = true
		}
	}

	projectName, err := determineProjectName(projectName, filePath)
	if err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	orderSplitted := strings.Split(order, ",")

	if len(orderSplitted) < 3 {
		fmt.Printf("not enough arguments for flag order")
		os.Exit(1)
	}

	for _, order := range orderSplitted {
		if (order != stdPkg) && (order != aliasedPkg) && (order != projectPkg) && (order != generalPkg) {
			fmt.Printf("order flag must either be %s, %s, %s, or %s", stdPkg, aliasedPkg, projectPkg, generalPkg)
			os.Exit(1)
		}
	}

	if isRecursive {
		output = defaultOutput
		RunCommandRecursive(projectName, orderSplitted)
		os.Exit(0)
	}

	RunCommand(projectName, filePath, orderSplitted)
}

func RunCommandRecursive(projectName string, order []string) {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("hello\n")

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".go" {
			fmt.Printf("sorting imports from file : %v\n", path)
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
		log.Fatalf("%+v", errors.WithStack(err))
	}

	if output == StdOutput {
		fmt.Print(string(formattedOutput))
	} else if output == defaultOutput {
		if !hasChange {
			return
		}

		if err := ioutil.WriteFile(filePath, formattedOutput, 0644); err != nil {
			log.Fatalf("failed to write fixed result to file(%s): %+v", filePath, errors.WithStack(err))
		}
	}

}

func validateRequiredParam(filePath string) error {
	if filePath == "" {
		return errors.Errorf("-%s should be set", filePathArg)
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
