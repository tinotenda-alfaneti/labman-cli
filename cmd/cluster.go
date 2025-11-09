/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/tinotenda-alfaneti/labman/internal/remote"
	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Interact with the current homelab cluster",
	Long: `Cluster hosts all SSH-backed operations. Use it to authenticate,
inspect state, and run other management tasks against your remote server.`,
	Run: func(cmd *cobra.Command, args []string) {
		printBanner(cmd)
		printSection(cmd, "CLUSTER", "Use subcommands such as 'labman cluster info'")
	},
	PersistentPreRunE: requireSession,
	PersistentPostRun: cleanupSession,
}

var clusterInfoCmd = &cobra.Command{
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

		printSection(cmd, "CLUSTER INFO", output)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterInfoCmd)
}
