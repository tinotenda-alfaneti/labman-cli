package cmd

import (
	"errors"
	"fmt"

	"github.com/tinotenda-alfaneti/labman/internal/remote"
	"github.com/spf13/cobra"
)

// requireSession ensures remote.Current() holds a live SSH session before commands run.
func requireSession(cmd *cobra.Command, args []string) error {
	if shouldSkipSession(cmd) {
		return nil
	}

	current := remote.Current()
	if current != nil && current.IsConnected() {
		return nil
	}

	client, err := remote.LoadSession()
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	if !client.IsConnected() {
		_ = client.Close()
		remote.SetCurrent(nil)
		return errors.New("not connected to any server; run 'labman login' first")
	}

	remote.SetCurrent(client)
	return nil
}

// cleanupSession closes any shared session after the command finishes.
func cleanupSession(cmd *cobra.Command, args []string) {
	if shouldSkipSession(cmd) {
		return
	}

	current := remote.Current()
	if current == nil {
		return
	}

	_ = current.Close()
	remote.SetCurrent(nil)
}

func shouldSkipSession(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}

	if cmd.Name() == "login" {
		return true
	}

	parent := cmd.Parent()
	return parent != nil && parent.Name() == "login"
}
