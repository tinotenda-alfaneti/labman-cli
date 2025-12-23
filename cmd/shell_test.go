package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPrintShellHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printShellHelp()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedStrings := []string{
		"Available commands:",
		"cluster",
		"diag",
		"self",
		"session",
		"help",
		"exit",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("printShellHelp() output missing %q", expected)
		}
	}
}

func TestExecuteShellCommand(t *testing.T) {
	// Create test commands
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	subCmd := &cobra.Command{
		Use:   "sub",
		Short: "Sub command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	testCmd.AddCommand(subCmd)

	// Store original commands
	origClusterCmd := clusterCmd
	origDiagCmd := diagCmd
	origSelfCmd := selfCmd
	origSessionCmd := sessionCmd

	// Replace with test command
	clusterCmd = testCmd
	diagCmd = testCmd
	selfCmd = testCmd
	sessionCmd = testCmd

	// Restore original commands after test
	defer func() {
		clusterCmd = origClusterCmd
		diagCmd = origDiagCmd
		selfCmd = origSelfCmd
		sessionCmd = origSessionCmd
	}()

	tests := []struct {
		name    string
		parts   []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid cluster command",
			parts:   []string{"cluster"},
			wantErr: false,
		},
		{
			name:    "valid diag command",
			parts:   []string{"diag"},
			wantErr: false,
		},
		{
			name:    "valid self command",
			parts:   []string{"self"},
			wantErr: false,
		},
		{
			name:    "valid session command",
			parts:   []string{"session"},
			wantErr: false,
		},
		{
			name:    "valid command with subcommand",
			parts:   []string{"cluster", "sub"},
			wantErr: false,
		},
		{
			name:    "unknown command",
			parts:   []string{"unknown"},
			wantErr: true,
			errMsg:  "unknown command: unknown",
		},
		{
			name:    "invalid command",
			parts:   []string{"invalid", "args"},
			wantErr: true,
			errMsg:  "unknown command: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeShellCommand(tt.parts)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeShellCommand(%v) error = %v, wantErr %v", tt.parts, err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("executeShellCommand(%v) error = %v, expected to contain %q", tt.parts, err, tt.errMsg)
			}
		})
	}
}

func TestGetShellPassword_NonTerminal(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate non-terminal input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()

	os.Stdin = r

	// Write password to pipe
	testPassword := "testpassword123"
	go func() {
		defer w.Close()
		w.Write([]byte(testPassword + "\n"))
	}()

	cmd := &cobra.Command{}
	password, err := getShellPassword(cmd)
	if err != nil {
		t.Fatalf("getShellPassword() error = %v", err)
	}

	if password != testPassword {
		t.Errorf("getShellPassword() = %q, want %q", password, testPassword)
	}
}

func TestGetShellPassword_NonTerminal_WithWhitespace(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate non-terminal input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()

	os.Stdin = r

	// Write password with whitespace to pipe
	testPassword := "  testpassword123  "
	expected := "testpassword123"
	go func() {
		defer w.Close()
		w.Write([]byte(testPassword + "\n"))
	}()

	cmd := &cobra.Command{}
	password, err := getShellPassword(cmd)
	if err != nil {
		t.Fatalf("getShellPassword() error = %v", err)
	}

	if password != expected {
		t.Errorf("getShellPassword() = %q, want %q", password, expected)
	}
}

func TestJoinCommand_EdgeCases(t *testing.T) {
	// Test nil slices don't panic
	t.Run("variadic empty", func(t *testing.T) {
		result := joinCommand()
		if result != "" {
			t.Errorf("joinCommand() = %q, want empty string", result)
		}
	})

	// Test very long strings
	t.Run("long strings", func(t *testing.T) {
		longStr := strings.Repeat("a", 1000)
		result := joinCommand(longStr, longStr)
		expected := longStr + " " + longStr
		if result != expected {
			t.Errorf("joinCommand() length = %d, want %d", len(result), len(expected))
		}
	})
}
