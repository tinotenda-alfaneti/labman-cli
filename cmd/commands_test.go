package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmd(t *testing.T) {
	t.Run("has correct use", func(t *testing.T) {
		if rootCmd.Use != "labman" {
			t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "labman")
		}
	})

	t.Run("has short description", func(t *testing.T) {
		if rootCmd.Short == "" {
			t.Error("rootCmd.Short should not be empty")
		}
	})

	t.Run("has long description", func(t *testing.T) {
		if rootCmd.Long == "" {
			t.Error("rootCmd.Long should not be empty")
		}
	})

	t.Run("has config flag", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("config")
		if flag == nil {
			t.Error("rootCmd should have --config flag")
		}
	})
}

func TestClusterCmd(t *testing.T) {
	t.Run("has correct use", func(t *testing.T) {
		if clusterCmd.Use != "cluster" {
			t.Errorf("clusterCmd.Use = %q, want %q", clusterCmd.Use, "cluster")
		}
	})

	t.Run("has subcommands", func(t *testing.T) {
		if !clusterCmd.HasSubCommands() {
			t.Error("clusterCmd should have subcommands")
		}
	})

	t.Run("has info subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range clusterCmd.Commands() {
			if cmd.Use == "info" {
				found = true
				break
			}
		}
		if !found {
			t.Error("clusterCmd should have 'info' subcommand")
		}
	})

	t.Run("has status subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range clusterCmd.Commands() {
			if cmd.Use == "status" {
				found = true
				break
			}
		}
		if !found {
			t.Error("clusterCmd should have 'status' subcommand")
		}
	})
}

func TestDiagCmd(t *testing.T) {
	t.Run("has correct use", func(t *testing.T) {
		if diagCmd.Use != "diag" {
			t.Errorf("diagCmd.Use = %q, want %q", diagCmd.Use, "diag")
		}
	})

	t.Run("has bundle subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range diagCmd.Commands() {
			if cmd.Use == "bundle" {
				found = true
				break
			}
		}
		if !found {
			t.Error("diagCmd should have 'bundle' subcommand")
		}
	})
}

func TestSelfCmd(t *testing.T) {
	t.Run("has correct use", func(t *testing.T) {
		if selfCmd.Use != "self" {
			t.Errorf("selfCmd.Use = %q, want %q", selfCmd.Use, "self")
		}
	})

	t.Run("has subcommands", func(t *testing.T) {
		if !selfCmd.HasSubCommands() {
			t.Error("selfCmd should have subcommands")
		}
	})
}

func TestSessionCmd(t *testing.T) {
	t.Run("has correct use", func(t *testing.T) {
		if sessionCmd.Use != "session" {
			t.Errorf("sessionCmd.Use = %q, want %q", sessionCmd.Use, "session")
		}
	})

	t.Run("has status subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range sessionCmd.Commands() {
			if cmd.Use == "status" {
				found = true
				break
			}
		}
		if !found {
			t.Error("sessionCmd should have 'status' subcommand")
		}
	})

	t.Run("has drop subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range sessionCmd.Commands() {
			if cmd.Use == "drop" {
				found = true
				break
			}
		}
		if !found {
			t.Error("sessionCmd should have 'drop' subcommand")
		}
	})
}

func TestLoginCmd(t *testing.T) {
	t.Run("has correct use", func(t *testing.T) {
		if !strings.HasPrefix(loginCmd.Use, "login") {
			t.Errorf("loginCmd.Use = %q, want to start with %q", loginCmd.Use, "login")
		}
	})

	t.Run("requires exactly one arg", func(t *testing.T) {
		// loginCmd.Args should be cobra.ExactArgs(1)
		if loginCmd.Args == nil {
			t.Error("loginCmd should have Args validator")
		}
	})

	t.Run("has username flag", func(t *testing.T) {
		flag := loginCmd.Flags().Lookup("username")
		if flag == nil {
			t.Error("loginCmd should have --username flag")
		}
	})

	t.Run("has password-stdin flag", func(t *testing.T) {
		flag := loginCmd.Flags().Lookup("password-stdin")
		if flag == nil {
			t.Error("loginCmd should have --password-stdin flag")
		}
	})
}

func TestConfigCmd(t *testing.T) {
	t.Run("has correct use", func(t *testing.T) {
		if configCmd.Use != "config" {
			t.Errorf("configCmd.Use = %q, want %q", configCmd.Use, "config")
		}
	})

	t.Run("has subcommands", func(t *testing.T) {
		if !configCmd.HasSubCommands() {
			t.Error("configCmd should have subcommands")
		}
	})

	t.Run("has show subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Use == "show" {
				found = true
				break
			}
		}
		if !found {
			t.Error("configCmd should have 'show' subcommand")
		}
	})

	t.Run("has validate subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Use == "validate" {
				found = true
				break
			}
		}
		if !found {
			t.Error("configCmd should have 'validate' subcommand")
		}
	})

	t.Run("has init subcommand", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Use == "init" {
				found = true
				break
			}
		}
		if !found {
			t.Error("configCmd should have 'init' subcommand")
		}
	})
}

func TestShellCmd(t *testing.T) {
	t.Run("has correct use", func(t *testing.T) {
		if !strings.HasPrefix(shellCmd.Use, "shell") {
			t.Errorf("shellCmd.Use = %q, want to start with %q", shellCmd.Use, "shell")
		}
	})

	t.Run("has username flag", func(t *testing.T) {
		flag := shellCmd.Flags().Lookup("username")
		if flag == nil {
			t.Error("shellCmd should have --username flag")
		}
	})
}

func TestClusterWorkloadsCmd(t *testing.T) {
	t.Run("has namespace flag", func(t *testing.T) {
		flag := clusterWorkloadsCmd.Flags().Lookup("namespace")
		if flag == nil {
			t.Error("clusterWorkloadsCmd should have --namespace flag")
		}
	})

	t.Run("has selector flag", func(t *testing.T) {
		flag := clusterWorkloadsCmd.Flags().Lookup("selector")
		if flag == nil {
			t.Error("clusterWorkloadsCmd should have --selector flag")
		}
	})

	t.Run("has max-crash-pods flag", func(t *testing.T) {
		flag := clusterWorkloadsCmd.Flags().Lookup("max-crash-pods")
		if flag == nil {
			t.Error("clusterWorkloadsCmd should have --max-crash-pods flag")
		}
	})

	t.Run("has logs-tail flag", func(t *testing.T) {
		flag := clusterWorkloadsCmd.Flags().Lookup("logs-tail")
		if flag == nil {
			t.Error("clusterWorkloadsCmd should have --logs-tail flag")
		}
	})
}

func TestClusterBackupCmd(t *testing.T) {
	t.Run("has list flag", func(t *testing.T) {
		flag := clusterBackupCmd.Flags().Lookup("list")
		if flag == nil {
			t.Error("clusterBackupCmd should have --list flag")
		}
	})

	t.Run("has name flag", func(t *testing.T) {
		flag := clusterBackupCmd.Flags().Lookup("name")
		if flag == nil {
			t.Error("clusterBackupCmd should have --name flag")
		}
	})

	t.Run("has skip-velero flag", func(t *testing.T) {
		flag := clusterBackupCmd.Flags().Lookup("skip-velero")
		if flag == nil {
			t.Error("clusterBackupCmd should have --skip-velero flag")
		}
	})

	t.Run("has skip-etcd flag", func(t *testing.T) {
		flag := clusterBackupCmd.Flags().Lookup("skip-etcd")
		if flag == nil {
			t.Error("clusterBackupCmd should have --skip-etcd flag")
		}
	})
}

func TestClusterRestartCmd(t *testing.T) {
	t.Run("requires exactly one arg", func(t *testing.T) {
		if clusterRestartCmd.Args == nil {
			t.Error("clusterRestartCmd should have Args validator")
		}
	})

	t.Run("has type flag", func(t *testing.T) {
		flag := clusterRestartCmd.Flags().Lookup("type")
		if flag == nil {
			t.Error("clusterRestartCmd should have --type flag")
		}
	})

	t.Run("has wait flag", func(t *testing.T) {
		flag := clusterRestartCmd.Flags().Lookup("wait")
		if flag == nil {
			t.Error("clusterRestartCmd should have --wait flag")
		}
	})
}

func TestDiagBundleCmd(t *testing.T) {
	t.Run("has remote-path flag", func(t *testing.T) {
		flag := diagBundleCmd.Flags().Lookup("remote-path")
		if flag == nil {
			t.Error("diagBundleCmd should have --remote-path flag")
		}
	})

	t.Run("has stdout flag", func(t *testing.T) {
		flag := diagBundleCmd.Flags().Lookup("stdout")
		if flag == nil {
			t.Error("diagBundleCmd should have --stdout flag")
		}
	})
}

func TestSessionStatusCmd(t *testing.T) {
	t.Run("has drop flag", func(t *testing.T) {
		flag := sessionStatusCmd.Flags().Lookup("drop")
		if flag == nil {
			t.Error("sessionStatusCmd should have --drop flag")
		}
	})
}

func TestCommandStructure(t *testing.T) {
	t.Run("all commands have descriptions", func(t *testing.T) {
		commands := []*cobra.Command{
			rootCmd,
			clusterCmd,
			diagCmd,
			selfCmd,
			sessionCmd,
			loginCmd,
			configCmd,
			shellCmd,
		}

		for _, cmd := range commands {
			if cmd.Short == "" {
				t.Errorf("command %q missing Short description", cmd.Use)
			}
		}
	})
}
