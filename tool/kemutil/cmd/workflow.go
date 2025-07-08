/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/kemadev/ci-cd/tool/kemutil/workflow"
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Run workflows",
	Long:  `Run workflows that usually run in CI/CD pipelines, locally`,
	Args:  cobra.MinimumNArgs(1),
	PreRun: toggleDebug,
}

var workflowCiCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run CI workflows",
	Long:  `Run all CI pipelines`,
	RunE:  workflow.Ci,
	Args:  cobra.NoArgs,
	PreRun: toggleDebug,
}

var workflowCustomCmd = &cobra.Command{
	Use:   "custom",
	Short: "Run custom commands",
	Long:  `Run custom commands using the CI/CD runner`,
	RunE:  workflow.Custom,
	Args:  cobra.MinimumNArgs(1),
	PreRun: toggleDebug,
}

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.PersistentFlags().
		BoolVar(&workflow.Hot, "hot", false, "Enable hot reload mode")
	workflowCmd.PersistentFlags().
		BoolVar(&workflow.RunnerDebug, "runner-debug", false, "Enable debug mode for the CI/CD runner")
	workflowCmd.AddCommand(workflowCiCmd)
	workflowCmd.AddCommand(workflowCustomCmd)
	workflowCmd.PersistentFlags().BoolVar(&workflow.Fix, "fix", false, "Enable fix mode")
}
