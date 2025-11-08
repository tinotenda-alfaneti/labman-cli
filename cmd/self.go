/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github/tinotenda-alfaneti/labman/internal/remote"

	"github.com/spf13/cobra"
)

// selfCmd groups OS-level maintenance commands for the remote host.
var selfCmd = &cobra.Command{
	Use:   "self",
	Short: "Run maintenance tasks directly on the server",
	Long: `Use the self command for host-level operations such as updating packages,
cleaning disk space, or checking system status without touching the cluster.`,
	PersistentPreRunE: requireSession,
	PersistentPostRun: cleanupSession,
}

var selfInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show remote server system information",
	Long: `Fetches and displays system information from the remote host,
including OS version, kernel details, and hardware specs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := remote.Current()
		if client == nil {
			return fmt.Errorf("not connected to any server. Run 'labman login' first")
		}

		output, err := client.Run("uname -a && lsb_release -a && lscpu")
		if err != nil {
			return fmt.Errorf("failed to fetch system info: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Remote system information:\n%s\n", output)
		return nil
	},
}



func init() {
	rootCmd.AddCommand(selfCmd)
	selfCmd.AddCommand(selfInfoCmd)
}
