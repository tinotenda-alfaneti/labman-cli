package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tinotenda-alfaneti/labman/internal/config"
	"github.com/tinotenda-alfaneti/labman/internal/remote"
	"golang.org/x/term"
)

var shellCmd = &cobra.Command{
	Use:   "shell [host|alias]",
	Short: "Open an interactive labman command shell",
	Long: `Opens an interactive command shell where you can run labman commands without typing 'labman' each time.
This maintains a persistent SSH session and allows you to run multiple commands interactively.

The host argument is optional:
  - If omitted, uses the last cached session
  - Can be a configured host alias from ~/.labman/config.yaml
  - Can be a direct IP address or hostname

Once in the shell, you can run commands like:
  cluster status
  cluster logs
  diag bundle
  exit (or Ctrl+D to quit)

Examples:
  labman shell                    # Use cached session or prompt for default host
  labman shell homelab-prod       # Connect to configured alias`,
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := ensureSession(cmd, args)
		if err != nil {
			return err
		}
		defer session.Close()

		fmt.Printf("Connected to %s@%s\n", session.User, session.Host)
		fmt.Println("Interactive labman shell - type 'help' for commands, 'exit' to quit")
		fmt.Println()

		return runInteractiveShell(session)
	},
}

func ensureSession(cmd *cobra.Command, args []string) (*remote.SSHSession, error) {
	session, err := remote.LoadSession()
	if err == nil {
		return session, nil
	}

	cfg, loadErr := config.Load()
	if loadErr != nil {
		return nil, fmt.Errorf("load config: %w", loadErr)
	}

	var hostIdentifier string
	if len(args) > 0 {
		hostIdentifier = args[0]
	} else {
		if len(cfg.Hosts) == 0 {
			return nil, fmt.Errorf("no session found and no hosts configured. Run 'labman login <host>' first")
		}
		for alias := range cfg.Hosts {
			hostIdentifier = alias
			break
		}
		fmt.Printf("No active session. Using configured host: %s\n", hostIdentifier)
	}

	serverIP, defaultUsername, port, keyFile, resolveErr := cfg.ResolveHost(hostIdentifier)
	if resolveErr != nil {
		return nil, fmt.Errorf("resolve host: %w", resolveErr)
	}

	user, _ := cmd.Flags().GetString("username")
	if user == "" {
		user = defaultUsername
	}
	if user == "" {
		return nil, fmt.Errorf("username is required (specify with -u or set in config)")
	}

	password, passErr := getShellPassword(cmd)
	if passErr != nil {
		return nil, fmt.Errorf("read password: %w", passErr)
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	connectionInfo := fmt.Sprintf("Connecting to %s as %s", serverIP, user)
	if port != 22 && port != 0 {
		connectionInfo = fmt.Sprintf("Connecting to %s:%d as %s", serverIP, port, user)
	}
	if keyFile != "" {
		connectionInfo += fmt.Sprintf(" (with key: %s)", keyFile)
	}
	fmt.Println(connectionInfo)

	session, err = remote.NewSSHSession(serverIP, user, password)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if saveErr := remote.SaveSession(session); saveErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", saveErr)
	}

	return session, nil
}

func runInteractiveShell(session *remote.SSHSession) error {
	if err := resetConsoleMode(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to reset console: %v\n", err)
	}


	remote.SetCurrent(session)
	defer remote.SetCurrent(nil)

	reader := bufio.NewReader(os.Stdin)

	os.Stdout.Sync()
	os.Stderr.Sync()

	for {
		fmt.Fprint(os.Stdout, "labman> ")
		os.Stdout.Sync() // Force flush the prompt

		line, err := reader.ReadString('\n')
		if err != nil {
			// EOF (Ctrl+D) or error
			fmt.Println()
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle exit commands
		if line == "exit" || line == "quit" {
			break
		}

		// Handle help
		if line == "help" {
			printShellHelp()
			continue
		}

		// Parse command and args
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		// Execute the command
		if err := executeShellCommand(parts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	return nil
}

func printShellHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  cluster <subcommand>  - Cluster management (status, logs, backup, restart)")
	fmt.Println("  diag <subcommand>     - Diagnostics (bundle, connectivity, dns)")
	fmt.Println("  self <subcommand>     - Server maintenance (update, reboot)")
	fmt.Println("  session status        - Show session information")
	fmt.Println("  help                  - Show this help")
	fmt.Println("  exit                  - Exit the shell")
}

func executeShellCommand(parts []string) error {
	cmdName := parts[0]
	cmdArgs := parts[1:]

	// Find the command
	var targetCmd *cobra.Command
	switch cmdName {
	case "cluster":
		targetCmd = clusterCmd
	case "diag":
		targetCmd = diagCmd
	case "self":
		targetCmd = selfCmd
	case "session":
		targetCmd = sessionCmd
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmdName)
	}

	// Set args and execute directly without going through root
	targetCmd.SetArgs(cmdArgs)
	targetCmd.SetOut(os.Stdout)
	targetCmd.SetErr(os.Stderr)

	// Parse flags manually
	if err := targetCmd.ParseFlags(cmdArgs); err != nil {
		targetCmd.SetArgs(nil)
		return err
	}

	// Get the actual args after flag parsing
	actualArgs := targetCmd.Flags().Args()

	// Execute the command's RunE function directly if it has subcommands
	if targetCmd.HasSubCommands() && len(actualArgs) > 0 {
		subCmd, _, err := targetCmd.Find(actualArgs)
		if err != nil {
			targetCmd.SetArgs(nil)
			return err
		}

		// Set remaining args for subcommand
		subCmd.SetArgs(actualArgs[1:])
		subCmd.SetOut(os.Stdout)
		subCmd.SetErr(os.Stderr)

		if err := subCmd.ParseFlags(actualArgs[1:]); err != nil {
			targetCmd.SetArgs(nil)
			subCmd.SetArgs(nil)
			return err
		}

		if subCmd.RunE != nil {
			err = subCmd.RunE(subCmd, subCmd.Flags().Args())
		} else if subCmd.Run != nil {
			subCmd.Run(subCmd, subCmd.Flags().Args())
		} else {
			err = subCmd.Help()
		}

		subCmd.SetArgs(nil)
		targetCmd.SetArgs(nil)
		return err
	}

	// Execute the command directly
	var err error
	if targetCmd.RunE != nil {
		err = targetCmd.RunE(targetCmd, actualArgs)
	} else if targetCmd.Run != nil {
		targetCmd.Run(targetCmd, actualArgs)
	} else {
		err = targetCmd.Help()
	}

	// Reset for next use
	targetCmd.SetArgs(nil)
	return err
}

func getShellPassword(cmd *cobra.Command) (string, error) {
	// Check if stdin is a terminal
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Not a terminal, read from stdin directly
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}
		return strings.TrimSpace(password), nil
	}

	// Save the terminal state before reading password
	oldState, err := term.GetState(fd)
	if err != nil {
		return "", fmt.Errorf("get terminal state: %w", err)
	}

	// Interactive prompt using proper terminal handling
	fmt.Fprint(os.Stderr, "Password: ")
	passBytes, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr) // newline after password input

	// Explicitly restore terminal state
	if restoreErr := term.Restore(fd, oldState); restoreErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to restore terminal: %v\n", restoreErr)
	}

	if err != nil {
		return "", fmt.Errorf("read password from terminal: %w", err)
	}
	return string(passBytes), nil
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().StringP("username", "u", "", "username for ssh login (overrides config)")
}
