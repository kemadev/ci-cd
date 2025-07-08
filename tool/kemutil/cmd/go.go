package cmd

import (
	"github.com/kemadev/ci-cd/tool/kemutil/wgo"
	"github.com/spf13/cobra"
)

var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Wrapper for Go tasks",
	Long:  `Run everyday Go tasks like initializing a module, updating dependencies, ...`,
	Args:  cobra.ExactArgs(1),
	PreRun: toggleDebug,
}

var goInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Go module",
	Long: `Initialize a Go module in the current directory

Repository url and module path are used to form the module name`,
	RunE: wgo.Init,
	Args: cobra.NoArgs,
	PreRun: toggleDebug,
}

var goUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update all Go modules dependencies",
	Long:  `Update all Go modules dependencies found in the current directory and subdirectories`,
	RunE:  wgo.Update,
	Args:  cobra.NoArgs,
	PreRun: toggleDebug,
}

var goTidy = &cobra.Command{
	Use:   "tidy",
	Short: "Tidy all Go modules dependencies",
	Long:  `Tidy all Go modules dependencies found in the current directory and subdirectories`,
	RunE:  wgo.Tidy,
	Args:  cobra.NoArgs,
	PreRun: toggleDebug,
}

func init() {
	rootCmd.AddCommand(goCmd)
	goCmd.AddCommand(goInit)
	goCmd.AddCommand(goUpdate)
	goCmd.AddCommand(goTidy)
}
