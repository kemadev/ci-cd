package cmd

import (
	"github.com/kemadev/ci-cd/tool/kemutil/repotpl"
	"github.com/spf13/cobra"
)

var repotplCmd = &cobra.Command{
	Use:   "repotpl",
	Short: "Repository template helpers",
	Long:  `Run tasks related to repository templates, such as initializing or updating`,
	Args:  cobra.ExactArgs(1),
}

var repotplInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize repository template",
	Long:  `Initialize the repository template from upstream`,
	RunE:  repotpl.Init,
	Args:  cobra.NoArgs,
}

var repotplUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update repository template",
	Long:  `Update the repository template with the latest changes from upstream`,
	RunE:  repotpl.Update,
	Args:  cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(repotplCmd)
	repotplCmd.AddCommand(repotplInit)
	repotplCmd.AddCommand(repotplUpdate)
	repotplUpdate.Flags().
		BoolVar(&repotpl.SkipAnswered, "skip-answered", false, "Skip answered questions update")
}
