/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github/tinotenda-alfaneti/labman/internal/remote"

	"errors"
	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Interact with the current homelab cluster",
	Long: `Cluster hosts all SSH-backed operations. Use it to authenticate,
inspect state, and run other management tasks against your remote server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cluster called")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "login" || cmd.HasParent() && cmd.Parent().Name() == "login" {
			return nil
		}

		if remote.Current() != nil {
			return nil
		}

		client, err := remote.LoadSession()
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		if !client.IsConnected() {
			_ = client.Close()
			return errors.New("not connected to any cluster. Please login first")
		}

		remote.SetCurrent(client)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		client := remote.Current()
		if client == nil {
			return
		}

		_ = client.Close()
		remote.SetCurrent(nil)
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
