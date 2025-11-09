package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLIHelpCommand(t *testing.T) {
	cmd := exec.Command("go", "run", "./main.go", "--help")

	tempHome, err := os.MkdirTemp("", "labman-cli-home-")
	if err != nil {
		t.Fatalf("create temp home: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempHome)
	}()

	cmd.Env = append(os.Environ(),
		"HOME="+tempHome,
		"USERPROFILE="+tempHome,
		"HOMEDRIVE=",
		"HOMEPATH=",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run --help failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "labman connects to your homelab hosts over SSH and runs curated workflows") {
		t.Fatalf("help output did not describe the CLI: %s", output)
	}
}
