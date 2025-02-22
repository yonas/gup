package cmd

import (
	"fmt"
	"os"

	"github.com/nao1215/gup/internal/config"
	"github.com/nao1215/gup/internal/goutil"
	"github.com/nao1215/gup/internal/print"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the binary names under $GOPATH/bin and their path info. to gup.conf.",
	Long: `Export the binary names under $GOPATH/bin and their path info. to gup.conf.

Use the export subcommand if you want to install the same golang
binaries across multiple systems. By default, this sub-command 
exports the file to $HOME/.config/gup/gup.conf. After you have 
placed gup.conf in the same path hierarchy on another system,
you execute import subcommand. gup start the installation 
according to the contents of gup.conf.`,
	Run: func(cmd *cobra.Command, args []string) {
		OsExit(export(cmd, args))
	},
}

func init() {
	exportCmd.Flags().BoolP("output", "o", false, "print command path information at STDOUT")
	rootCmd.AddCommand(exportCmd)
}

func export(cmd *cobra.Command, args []string) int {
	if err := goutil.CanUseGoCmd(); err != nil {
		print.Err(fmt.Errorf("%s: %w", "you didn't install golang", err))
		return 1
	}

	output, err := cmd.Flags().GetBool("output")
	if err != nil {
		print.Err(fmt.Errorf("%s: %w", "can not parse command line argument (--output)", err))
		return 1
	}

	pkgs, err := getPackageInfo()
	if err != nil {
		print.Err(err)
		return 1
	}
	pkgs = validPkgInfo(pkgs)

	if len(pkgs) == 0 {
		print.Err("no package information")
		return 1
	}

	if output {
		err = outputConfig(pkgs)
	} else {
		err = writeConfigFile(pkgs)
	}
	if err != nil {
		print.Err(err)
		return 1
	}
	return 0
}

func writeConfigFile(pkgs []goutil.Package) error {
	if err := os.MkdirAll(config.DirPath(), 0775); err != nil {
		return fmt.Errorf("%s: %w", "can not make config directory", err)
	}

	file, err := os.Create(config.FilePath())
	if err != nil {
		return fmt.Errorf("%s %s: %w", "can't update", config.FilePath(), err)
	}
	defer file.Close()

	if err := config.WriteConfFile(file, pkgs); err != nil {
		return err
	}
	print.Info("Export " + config.FilePath())
	return nil
}

func outputConfig(pkgs []goutil.Package) error {
	return config.WriteConfFile(os.Stdout, pkgs)
}

func validPkgInfo(pkgs []goutil.Package) []goutil.Package {
	result := []goutil.Package{}
	for _, v := range pkgs {
		if v.ImportPath == "" {
			print.Warn("can't get '" + v.Name + "'package path information. old go version binary")
			continue
		}
		result = append(result, goutil.Package{Name: v.Name, ImportPath: v.ImportPath, Version: v.Version})
	}
	return result
}
