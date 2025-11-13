package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionManager(t *testing.T) {
	setTempHome(t)

	secrets, restoreKeyring := stubKeyring()
	t.Cleanup(restoreKeyring)

	restoreNewSession := stubNewSSHSession()
	t.Cleanup(restoreNewSession)

	t.Run("saveSessionToFile stores session config", func(t *testing.T) {
		sessionFilePath, err := getSessionFilePath()
		if err != nil {
			t.Fatalf("get session file path: %v", err)
		}

		session := &SSHSession{
			Host:     "localhost",
			User:     "user",
			Password: "password",
		}

		if err := saveSessionToFile(sessionFilePath, session); err != nil {
			t.Fatalf("save session: %v", err)
		}
		t.Cleanup(func() { os.Remove(sessionFilePath) })

		data, err := loadSessionDataFromFile(sessionFilePath)
		if err != nil {
			t.Fatalf("load session data: %v", err)
		}

		if data.Host != session.Host {
			t.Fatalf("expected host %q, got %q", session.Host, data.Host)
		}

		if data.User != session.User {
			t.Fatalf("expected user %q, got %q", session.User, data.User)
		}

		if !data.Timeout.After(time.Now()) {
			t.Fatalf("expected timeout in the future, got %v", data.Timeout)
		}

		wantSecret := credentialsKey(session.Host, session.User)
		if got := secrets[wantSecret]; got != session.Password {
			t.Fatalf("expected password %q in keyring, got %q", session.Password, got)
		}
	})

	t.Run("LoadSession returns SSHSession built from saved data", func(t *testing.T) {
		sessionFilePath, err := getSessionFilePath()
		if err != nil {
			t.Fatalf("get session file path: %v", err)
		}

		session := &SSHSession{
			Host:     "example.com",
			User:     "admin",
			Password: "secret",
		}

		if err := saveSessionToFile(sessionFilePath, session); err != nil {
			t.Fatalf("save session: %v", err)
		}
		t.Cleanup(func() { os.Remove(sessionFilePath) })

		loaded, err := LoadSession()
		if err != nil {
			t.Fatalf("load session: %v", err)
		}

		if loaded.Host != session.Host {
			t.Fatalf("expected host %q, got %q", session.Host, loaded.Host)
		}

		if loaded.User != session.User {
			t.Fatalf("expected user %q, got %q", session.User, loaded.User)
		}

		if loaded.Password != session.Password {
			t.Fatalf("expected password %q, got %q", session.Password, loaded.Password)
		}
	})

	t.Run("SessionMetadata returns saved session config", func(t *testing.T) {
		sessionFilePath, err := getSessionFilePath()
		if err != nil {
			t.Fatalf("get session file path: %v", err)
		}

		session := &SSHSession{
			Host:     "metadata.example",
			User:     "me",
			Password: "topsecret",
		}

		if err := saveSessionToFile(sessionFilePath, session); err != nil {
			t.Fatalf("save session: %v", err)
		}
		t.Cleanup(func() { os.Remove(sessionFilePath) })

		meta, err := SessionMetadata()
		if err != nil {
			t.Fatalf("SessionMetadata: %v", err)
		}

		if meta.Host != session.Host || meta.User != session.User {
			t.Fatalf("expected %s/%s, got %s/%s", session.User, session.Host, meta.User, meta.Host)
		}
	})

	t.Run("DeleteSession removes session file and secrets", func(t *testing.T) {
		sessionFilePath, err := getSessionFilePath()
		if err != nil {
			t.Fatalf("get session file path: %v", err)
		}

		session := &SSHSession{
			Host:     "delete.example",
			User:     "deleter",
			Password: "secret",
		}

		if err := saveSessionToFile(sessionFilePath, session); err != nil {
			t.Fatalf("save session: %v", err)
		}

		if err := DeleteSession(); err != nil {
			t.Fatalf("DeleteSession: %v", err)
		}

		if _, err := os.Stat(sessionFilePath); !os.IsNotExist(err) {
			t.Fatalf("expected session file removed, stat err=%v", err)
		}

		if _, ok := secrets[credentialsKey(session.Host, session.User)]; ok {
			t.Fatalf("expected keyring secret removed")
		}
	})
}

func stubKeyring() (map[string]string, func()) {
	originalSet := keyringSet
	originalGet := keyringGet
	originalDelete := keyringDelete

	secrets := map[string]string{}

	keyringSet = func(service, user, password string) error {
		secrets[user] = password
		return nil
	}
	keyringGet = func(service, user string) (string, error) {
		secret, ok := secrets[user]
		if !ok {
			return "", fmt.Errorf("secret %s not found", user)
		}
		return secret, nil
	}
	keyringDelete = func(service, user string) error {
		delete(secrets, user)
		return nil
	}

	return secrets, func() {
		keyringSet = originalSet
		keyringGet = originalGet
		keyringDelete = originalDelete
	}
}

func stubNewSSHSession() func() {
	original := newSSHSession

	newSSHSession = func(host, user, password string) (*SSHSession, error) {
		return &SSHSession{
			Host:     host,
			User:     user,
			Password: password,
		}, nil
	}

	return func() {
		newSSHSession = original
	}
}

func setTempHome(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")

	sessionPath := filepath.Join(dir, ".labman", "sessions")
	if err := os.MkdirAll(sessionPath, 0o700); err != nil {
		t.Fatalf("prepare session directory: %v", err)
	}
}
