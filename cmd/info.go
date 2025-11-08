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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		
		infoClient, err := remote.LoadSession()
		if err != nil {
			return fmt.Errorf("failed to load session: %v", err)
		}

		if !infoClient.IsConnected() {
			return fmt.Errorf("failed to connect to server")
		}

		output, err := infoClient.Run("kubectl cluster-info dump")

		if err != nil {
			return fmt.Errorf("failed to run command: %s", err)
		}

		fmt.Printf("------Command output-----\n%s\n-------------------------\n", output)

		defer infoClient.Close()

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(infoCmd)
}
