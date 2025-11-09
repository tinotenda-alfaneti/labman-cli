package remote

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestHostKeyCallback(t *testing.T) {
	generateHostEntry := func(t *testing.T) string {
		t.Helper()

		pub, _, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatalf("generate ed25519 key: %v", err)
		}

		sshPub, err := ssh.NewPublicKey(pub)
		if err != nil {
			t.Fatalf("convert ed25519 key: %v", err)
		}

		return knownhosts.Line([]string{"example.com"}, sshPub)
	}

	writeKnownHosts := func(t *testing.T, path string) {
		t.Helper()

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("prepare known_hosts dir: %v", err)
		}

		if err := os.WriteFile(path, []byte(generateHostEntry(t)), 0o600); err != nil {
			t.Fatalf("write known_hosts: %v", err)
		}
	}

	t.Run("uses provided path when supplied", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "known_hosts")
		writeKnownHosts(t, path)

		callback, err := hostKeyCallback(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if callback == nil {
			t.Fatalf("expected callback, got nil")
		}
	})

	t.Run("defaults to home directory when path empty", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		t.Setenv("HOMEDRIVE", "")
		t.Setenv("HOMEPATH", "")

		path := filepath.Join(home, ".ssh", "known_hosts")
		writeKnownHosts(t, path)

		callback, err := hostKeyCallback("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if callback == nil {
			t.Fatalf("expected callback, got nil")
		}
	})

	t.Run("returns error when known hosts file missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing_known_hosts")

		callback, err := hostKeyCallback(path)
		if err == nil {
			t.Fatalf("expected error %q, got nil", "load known hosts")
		}
		if !strings.Contains(err.Error(), "load known hosts") {
			t.Fatalf("expected error to contain %q, got %v", "load known hosts", err)
		}
		if callback != nil {
			t.Fatalf("expected nil callback on error, got %v", callback)
		}
	})
}
