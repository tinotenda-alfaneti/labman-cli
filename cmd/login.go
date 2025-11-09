/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/tinotenda-alfaneti/labman/internal/remote"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login <host>",
	Short: "Authenticate to a server and cache the SSH session",
	Long: `Establishes an SSH connection to the specified host, verifies the host key,
and saves the credentials in the local keyring so other commands can reuse them.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner(cmd)

		user, err := cmd.Flags().GetString("username")
		if err != nil {
			return fmt.Errorf("read username flag: %w", err)
		}
		if user == "" {
			return fmt.Errorf("username is required")
		}

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return fmt.Errorf("read password flag: %w", err)
		}
		if password == "" {
			return fmt.Errorf("password is required")
		}

		serverIP := args[0]
		printSection(cmd, "LOGIN", fmt.Sprintf("Connecting to %s as %s", serverIP, user))

		loginClient, err := remote.NewSSHSession(serverIP, user, password)
		if err != nil {
			return fmt.Errorf("failed to login to SSH session: %w", err)
		}
		defer loginClient.Close()

		if !loginClient.IsConnected() {
			return fmt.Errorf("failed to connect to server")
		}

		if err := remote.SaveSession(loginClient); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		printSection(cmd, "SUCCESS", "Login successful and session saved.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("username", "u", "", "username for ssh login")
	loginCmd.Flags().StringP("password", "p", "", "password for ssh login")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")
}
