package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/otiszv/render/domain"
	"github.com/otiszv/render/domain/common"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate some resources",
	Long:  "validate some resources",
}

var (
	files []string

	dir string
)
var validateDefinitionCmd = &cobra.Command{
	Use:          "definition",
	Short:        "validate resource definition",
	Long:         "validate resource definition",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return validate(files, dir)
	},
}

func validate(files []string, dir string) error {
	errs := common.Errors{}

	if dir != "" {
		var err error
		files, err = getFilelist(dir)
		if err != nil {
			return err
		}
	}

	if len(files) == 0 {
		return errors.New("no file need to validate")
	}

	for _, file := range files {
		kube := domain.Kubernete{}
		err := kube.LoadFromFile(file)
		if err != nil {
			fmt.Printf("×\t %s\n", file)
			fmt.Printf("\t %s\n", err.Error())
			errs = append(errs, err)
		}

		err = kube.ValidateDefinition()
		if err != nil {
			fmt.Printf("×\t %s\n", file)
			fmt.Printf("\t %s\n", err.Error())
			errs = append(errs, err)
		}

		fmt.Printf("√\t %s\n", file)
	}

	if len(errs) > 0 {
		return errors.New("template definition validation is not pass")
	}

	return nil
}

func getFilelist(dir string) ([]string, error) {
	var (
		err error
	)

	_, err = os.Stat(path.Join(dir, "definition"))
	if err != nil {
		return []string{}, err
	}

	var pathes = []string{}

	err = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		if !strings.HasSuffix(f.Name(), ".yaml") {
			return nil
		}

		pathes = append(pathes, path)
		return nil
	})

	if err != nil {
		return []string{}, err
	}
	return pathes, nil
}

func init() {

	validateDefinitionCmd.Flags().StringArrayVarP(
		&files,
		"file", "f", []string{}, "provider the file path that want to be validate",
	)

	validateDefinitionCmd.Flags().StringVarP(
		&dir,
		"dir", "d", "", "provider the pipeline template repository directory that want to be validate",
	)

	validateCmd.AddCommand(validateDefinitionCmd)
	RootCmd.AddCommand(validateCmd)
}
