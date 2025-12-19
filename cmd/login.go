/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tinotenda-alfaneti/labman/internal/config"
	"github.com/tinotenda-alfaneti/labman/internal/remote"
	"golang.org/x/term"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login <host|alias>",
	Short: "Authenticate to a server and cache the SSH session",
	Long: `Establishes an SSH connection to the specified host, verifies the host key,
and saves the credentials in the local keyring so other commands can reuse them.

The host argument can be:
  - A configured host alias from ~/.labman/config.yaml
  - A direct IP address or hostname

Examples:
  labman login homelab-prod              # Use configured alias
  labman login 192.168.1.10 -u admin     # Direct IP with username
  labman login my-server --password-stdin < password.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner(cmd)

		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		// Resolve host (could be alias or direct IP/hostname)
		hostIdentifier := args[0]
		serverIP, defaultUsername, port, keyFile, err := cfg.ResolveHost(hostIdentifier)
		if err != nil {
			return fmt.Errorf("resolve host: %w", err)
		}

		// Get username (flag takes precedence over config)
		user, err := cmd.Flags().GetString("username")
		if err != nil {
			return fmt.Errorf("read username flag: %w", err)
		}
		if user == "" {
			user = defaultUsername
		}
		if user == "" {
			return fmt.Errorf("username is required (specify with -u or set in config)")
		}

		// Get password
		password, err := getPassword(cmd)
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		if password == "" {
			return fmt.Errorf("password is required")
		}

		// Show connection info
		connectionInfo := fmt.Sprintf("Connecting to %s as %s", serverIP, user)
		if port != 22 && port != 0 {
			connectionInfo = fmt.Sprintf("Connecting to %s:%d as %s", serverIP, port, user)
		}
		if keyFile != "" {
			connectionInfo += fmt.Sprintf("\nKey file: %s", keyFile)
		}
		printSection(cmd, "LOGIN", connectionInfo)

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

func getPassword(cmd *cobra.Command) (string, error) {
	// Check if password provided via stdin
	passStdin, _ := cmd.Flags().GetBool("password-stdin")
	if passStdin {
		passBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}
		return string(passBytes), nil
	}

	// Check if password provided via flag (insecure, but supported for backward compatibility)
	passFlag, _ := cmd.Flags().GetString("password")
	if passFlag != "" {
		fmt.Fprintln(os.Stderr, "Warning: Using --password flag is insecure. Use --password-stdin or interactive prompt instead.")
		return passFlag, nil
	}

	// Interactive prompt
	fmt.Fprint(os.Stderr, "Password: ")
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after password input
	if err != nil {
		return "", fmt.Errorf("read password from terminal: %w", err)
	}
	return string(passBytes), nil
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("username", "u", "", "username for ssh login (overrides config)")
	loginCmd.Flags().StringP("password", "p", "", "(insecure) password for ssh login - prefer interactive prompt or --password-stdin")
	loginCmd.Flags().Bool("password-stdin", false, "read password from stdin")
}
