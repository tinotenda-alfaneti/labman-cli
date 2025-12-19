package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIHelpCommand(t *testing.T) {
	output, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("go run --help failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "labman connects to your homelab hosts over SSH and runs curated workflows") {
		t.Fatalf("help output did not describe the CLI: %s", output)
	}
}

func TestRootCommandShowsHelp(t *testing.T) {
	output, err := runCLI(t)
	if err != nil {
		t.Fatalf("go run with no args failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "labman connects to your homelab hosts over SSH and runs curated workflows") {
		t.Fatalf("expected root help output, got: %s", output)
	}
}

func TestClusterInfoFailsWithoutSession(t *testing.T) {
	output, err := runCLI(t, "cluster", "info")
	if err == nil {
		t.Fatalf("expected cluster info to fail without session")
	}

	if !strings.Contains(output, "failed to load session") {
		t.Fatalf("expected session error, got: %s", output)
	}
}

func TestLoginRequiresUsernameFlag(t *testing.T) {
	output, err := runCLI(t, "login", "example.com", "-p", "secret")
	if err == nil {
		t.Fatalf("expected login to fail when username flag missing")
	}

	if !strings.Contains(output, "username is required") {
		t.Fatalf("expected missing username error, got: %s", output)
	}
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	root := moduleRoot(t)
	cmdArgs := append([]string{"run", "./main.go"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = root

	tempHome, err := os.MkdirTemp("", "labman-cli-home-")
	if err != nil {
		t.Fatalf("create temp home: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tempHome)
	})

	cmd.Env = append(os.Environ(),
		"HOME="+tempHome,
		"USERPROFILE="+tempHome,
		"HOMEDRIVE=",
		"HOMEPATH=",
	)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("unable to locate go.mod from %s", dir)
		}
		dir = parent
	}
}
