package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tinotenda-alfaneti/labman/internal/remote"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Inspect or remove the cached SSH session",
	Long: `Shows details about the cached SSH session stored under ~/.labman/sessions
and provides helpers to discard the credentials when you want to rotate secrets.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sessionStatusCmd.RunE(cmd, args)
	},
}

var sessionStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display cached session host, user, TTL, and connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner(cmd)

		meta, err := remote.SessionMetadata()
		if err != nil {
			return fmt.Errorf("unable to read cached session: %w", err)
		}

		ttl := time.Until(meta.Timeout)
		if ttl < 0 {
			ttl = 0
		}

		connectivity := diagnoseConnectivity()
		body := &strings.Builder{}
		fmt.Fprintf(body, "Host          : %s\n", meta.Host)
		fmt.Fprintf(body, "User          : %s\n", meta.User)
		fmt.Fprintf(body, "Expires At    : %s\n", meta.Timeout.Format(time.RFC1123))
		fmt.Fprintf(body, "TTL Remaining : %s\n", formatTTL(ttl))
		fmt.Fprintf(body, "Connectivity  : %s\n", connectivity)
		printSection(cmd, "SESSION STATUS", body.String())

		shouldDrop, _ := cmd.Flags().GetBool("drop")
		if shouldDrop {
			if err := dropSession(); err != nil {
				return err
			}
			printSection(cmd, "SESSION DROP", "Removed cached credentials and session file.")
		}
		return nil
	},
}

var sessionDropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Delete the cached SSH session and keyring secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner(cmd)
		if err := dropSession(); err != nil {
			return err
		}

		printSection(cmd, "SESSION DROP", "Cached session removed. Run 'labman login' to create a new one.")
		return nil
	},
}

func diagnoseConnectivity() string {
	session, err := remote.LoadSession()
	if err != nil {
		return fmt.Sprintf("unavailable (%v)", err)
	}
	defer session.Close()

	if session.IsConnected() {
		return "connected"
	}

	return "disconnected"
}

func formatTTL(ttl time.Duration) string {
	if ttl <= 0 {
		return "expired"
	}

	hours := int(ttl.Hours())
	minutes := int(ttl.Minutes()) % 60
	seconds := int(ttl.Seconds()) % 60

	parts := []string{}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 && hours == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}
	return strings.Join(parts, " ")
}

func dropSession() error {
	if err := remote.DeleteSession(); err != nil {
		return fmt.Errorf("delete cached session: %w", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionStatusCmd)
	sessionCmd.AddCommand(sessionDropCmd)

	sessionStatusCmd.Flags().Bool("drop", false, "drop the cached session after showing its details")
}
