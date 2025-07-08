package cmd

import (
	"github.com/kemadev/ci-cd/tool/kemutil/pulumi"
	"github.com/spf13/cobra"
)

var iacCmd = &cobra.Command{
	Use:   "iac",
	Short: "Wrapper for IaC tasks",
	Long:  `Run everyday IaC tasks like initializing a Pulumi stack, deploying it, ...`,
	Args:  cobra.ExactArgs(1),
}

var iacInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Pulumi stack",
	Long:  `Initialize a Pulumi stack in the current directory, using a template`,
	RunE:  pulumi.Init,
	Args:  cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(iacCmd)
	iacCmd.AddCommand(iacInit)
	iacCmd.PersistentFlags().
		BoolVarP(&pulumi.DebugEnabled, "debug", "d", false, "Enable debug output for Pulumi commands")
	iacCmd.PersistentFlags().
		BoolVarP(&pulumi.Refresh, "refresh", "r", false, "Refresh the Pulumi stack before updating")
}
