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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cluster called")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		if cmd.Name() == "login" || cmd.HasParent() && cmd.Parent().Name() == "login" {
        	return nil
    	}

		client, err := remote.LoadSession()
		if err != nil {
			return fmt.Errorf("failed to load session: %v", err)
		}

		if !client.IsConnected() {
			return errors.New("not connected to any cluster. Please login first")
		}

		remote.SetCurrent(client)

		return nil

	
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
