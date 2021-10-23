package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Project-Centurion/ordino/sorter"

	"github.com/pkg/errors"
)

func runCommandRecursive(projectName, path string, order []string) {

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(colorError(err.Error()))
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".go" {
			runCommand(projectName, path, order)
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}

func runCommand(projectName, filePath string, orderSplitted []string) {

	formattedOutput, hasChange, err := sorter.Execute(projectName, filePath, orderSplitted)
	if err != nil {
		colorError(fmt.Sprintf("Error writing on file %s : %+v", filePath, err))
	}

	if output == StdOutput {
		fmt.Print(string(formattedOutput))
	} else if output == defaultOutput {
		if !hasChange {
			return
		}

		fmt.Println(colorWorked(fmt.Sprintf("	imports sorted: %v", filePath)))

		if err := ioutil.WriteFile(filePath, formattedOutput, 0600); err != nil {
			log.Fatalf(colorError(fmt.Sprintf("failed to write fixed result to file(%s): %+v", filePath, errors.WithStack(err))))
		}
	}

}
