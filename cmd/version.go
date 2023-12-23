package cmd

import (
	"fmt"
	"github.com/bcdevtools/consvp/constants"
	"github.com/spf13/cobra"
	"runtime"
	"runtime/debug"
)

const (
	flagVersion     = "version"
	flagLongVersion = "long"
)

func versionHandler(cmd *cobra.Command, _ []string) {
	fmt.Println(constants.APP_NAME)
	fmt.Println(constants.GITHUB_PROJECT)

	printLongVersion := cmd.Flags().Changed(flagLongVersion)

	if printLongVersion {
		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			fmt.Println("Build dependencies:")
		}
		for _, dep := range buildInfo.Deps {
			if dep.Replace != nil {
				fmt.Printf("- %s@%s => %s@%s\n", dep.Path, dep.Version, dep.Replace.Path, dep.Replace.Version)
			} else {
				fmt.Printf("- %s@%s\n", dep.Path, dep.Version)
			}
		}
	}

	fmt.Printf("%-9s %s\n", "Version:", constants.VERSION)

	if printLongVersion {
		fmt.Printf("%-9s %s %s/%s\n", "Go:", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	}
}
