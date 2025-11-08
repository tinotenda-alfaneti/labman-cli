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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	loginCmd.Flags().StringP("username", "u", "", "username for ssh login")
	loginCmd.Flags().StringP("password", "p", "", "password for ssh login")
}
