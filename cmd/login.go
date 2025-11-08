/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github/tinotenda-alfaneti/labman/internal/remote"

	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate to a server and cache the SSH session",
	Long: `Establishes an SSH connection to the specified host, verifies the host key,
and saves the credentials in the local keyring so other commands can reuse them.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		user, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		serverIP := args[0]

		fmt.Printf("Logging into server %s \n", serverIP)


		loginClient, err := remote.NewSSHSession(serverIP, user, password)

		if err != nil {
			return fmt.Errorf("failed to login to SSH session: %w", err)
		}

		if !loginClient.IsConnected() {
			return fmt.Errorf("failed to connect to server: %w", err)
		}

		err = loginClient.SaveSession()
		if err != nil {
			return fmt.Errorf("failed to save session: %s", err)
		}

		fmt.Printf("Login successful and session saved.")

		defer loginClient.Close()

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("username", "u", "", "username for ssh login")
	loginCmd.Flags().StringP("password", "p", "", "password for ssh login")
}
