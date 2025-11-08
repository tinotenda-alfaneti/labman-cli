/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github/tinotenda-alfaneti/labman/internal/remote"

	"github.com/spf13/cobra"
)

// infoCmd represents the login command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show Kubernetes cluster diagnostics",
	Long: `Runs 'kubectl cluster-info dump' on the remote host to print detailed cluster
state, making it easy to inspect components from your local terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		infoClient := remote.Current()
		if infoClient == nil {
			return fmt.Errorf("not connected to any cluster. Please login first")
		}

		output, err := infoClient.Run("kubectl cluster-info dump")
		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}

		fmt.Printf("------Command output-----\n%s\n-------------------------\n", output)

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(infoCmd)
}
