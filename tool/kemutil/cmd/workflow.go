/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/kemadev/ci-cd/tool/kemutil/workflow"
	"github.com/spf13/cobra"
)

// workflowCmd represents the workflow command
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Run workflows",
	Long:  `Run workflows that usually run in CI/CD pipelines, locally`,
	Args:  cobra.MinimumNArgs(1),
}

var workflowCiCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run CI workflows",
	Long:  `Run all CI pipelines`,
	RunE:  workflow.Ci,
	Args:  cobra.NoArgs,
}

var workflowCustomCmd = &cobra.Command{
	Use:   "custom",
	Short: "Run custom commands",
	Long:  `Run custom commands using the CI/CD runner`,
	RunE:  workflow.Custom,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCiCmd.Flags().BoolVar(&workflow.Hot, "hot", false, "Enable hot reload mode")
	workflowCiCmd.Flags().BoolVar(&workflow.Fix, "fix", false, "Enable fix mode")
	workflowCmd.AddCommand(workflowCiCmd)
	workflowCmd.AddCommand(workflowCustomCmd)
}
