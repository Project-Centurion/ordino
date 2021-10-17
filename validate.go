package main

import (
	"os"
	"path/filepath"

	"github.com/incu6us/goimports-reviser/v2/pkg/module"
	"github.com/pkg/errors"
)

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
