package cmd

import (
	"github.com/kemadev/ci-cd/tool/kemutil/iac"
	"github.com/spf13/cobra"
)

var iacCmd = &cobra.Command{
	Use:   "iac",
	Short: "Wrapper for IaC tasks",
	Long:  `Run everyday IaC tasks like initializing a Iac stack, deploying it, ...`,
	Args:  cobra.ExactArgs(1),
	PreRun: toggleDebug,
}

var iacInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Pulumi stack",
	Long:  `Initialize a Pulumi stack in the current directory, using a template`,
	RunE:  iac.Init,
	Args:  cobra.NoArgs,
	PreRun: toggleDebug,
}

func init() {
	rootCmd.AddCommand(iacCmd)
	iacCmd.AddCommand(iacInit)
	iacCmd.PersistentFlags().
		BoolVar(&iac.DebugEnabled, "iac-debug",  false, "Enable debug output for IaC commands")
	iacCmd.PersistentFlags().
		BoolVar(&iac.Refresh, "refresh",  false, "Refresh the IaC stack before updating")
}
