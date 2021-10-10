package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/incu6us/goimports-reviser/v2/pkg/module"
	"github.com/pkg/errors"
)

const (
	projectNameArg = "project-name"
	filePathArg    = "file-path"
	outputArg      = "output"
)

var projectName, filePath, output string

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
		"file",
		`Can be "file" or "stdout". Whether to write the formatted content back to the file or to stdout. Optional parameter.`,
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

	projectName, err := determineProjectName(projectName, filePath)
	if err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	formattedOutput, hasChange, err := Execute(projectName, filePath)
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}

	if output == "stdout" {
		fmt.Print(string(formattedOutput))
	} else if output == "file" {
		if !hasChange {
			return
		}

		if err := ioutil.WriteFile(filePath, formattedOutput, 0644); err != nil {
			log.Fatalf("failed to write fixed result to file(%s): %+v", filePath, errors.WithStack(err))
		}
	} else {
		log.Fatalf(`invalid output "%s" specified`, output)
	}
}

func validateRequiredParam(filePath string) error {
	if filePath == "" {
		return errors.Errorf("-%s should be set", filePathArg)
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
