package remote

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v2"
)

var (
	keyringSet    = keyring.Set
	keyringGet    = keyring.Get
	keyringDelete = keyring.Delete
	newSSHSession = NewSSHSession
)

const sessionTimeOut = 3600 * time.Minute

func LoadSession() (*SSHSession, error) {
	sessionFile, err := getSessionFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get session file path: %v", err)
	}

	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("session file does not exist: %v", err)
	}

	sessionData, err := loadSessionDataFromFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load session data from file: %v", err)
	}

	if sessionData.Timeout.Before(time.Now()) {
		return nil, fmt.Errorf("session has expired at %s", sessionData.Timeout.Format(time.RFC3339))
	}

	password, err := keyringGet(keyringService, credentialsKey(sessionData.Host, sessionData.User))
	if err != nil {
		return nil, fmt.Errorf("failed to load password from keyring: %w", err)
	}

	return newSSHSession(sessionData.Host, sessionData.User, password)
}

func SaveSession(s *SSHSession) error {
	createSessionFilePath, err := getSessionFilePath()
	if err != nil {
		return fmt.Errorf("failed to get session file path: %w", err)
	}
	return saveSessionToFile(createSessionFilePath, s)
}

func SessionMetadata() (SSHSessionConfig, error) {
	sessionFile, err := getSessionFilePath()
	if err != nil {
		return SSHSessionConfig{}, fmt.Errorf("failed to get session file path: %w", err)
	}

	if _, err := os.Stat(sessionFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SSHSessionConfig{}, fmt.Errorf("session file does not exist")
		}
		return SSHSessionConfig{}, fmt.Errorf("stat session file: %w", err)
	}

	return loadSessionDataFromFile(sessionFile)
}

func DeleteSession() error {
	sessionFile, err := getSessionFilePath()
	if err != nil {
		return fmt.Errorf("failed to get session file path: %w", err)
	}

	sessionData, err := loadSessionDataFromFile(sessionFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("session file does not exist")
		}
		return err
	}

	credentialKey := credentialsKey(sessionData.Host, sessionData.User)
	if err := keyringDelete(keyringService, credentialKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("remove credentials from keyring: %w", err)
	}

	if err := os.Remove(sessionFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("remove session file: %w", err)
	}

	return nil
}

func getSessionFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(homeDir, ".labman", "sessions", "credentials.yaml"), nil
}

func saveSessionToFile(filePath string, session *SSHSession) error {
	sessionTimeout := time.Now().Add(sessionTimeOut)
	credentialKey := credentialsKey(session.Host, session.User)

	if err := keyringSet(keyringService, credentialKey, session.Password); err != nil {
		return fmt.Errorf("store password in keyring: %w", err)
	}

	sessionData := SSHSessionConfig{
		Host:    session.Host,
		User:    session.User,
		Timeout: sessionTimeout,
	}
	err := os.MkdirAll(filepath.Dir(filePath), 0o700)
	if err != nil {
		return fmt.Errorf("failed to create session directory: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create session file: %v", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(sessionData); err != nil {
		return fmt.Errorf("error encoding YAML: %v", err)
	}

	return nil
}

func loadSessionDataFromFile(sessionFile string) (SSHSessionConfig, error) {
	var sessionConfig SSHSessionConfig

	file, err := os.Open(sessionFile)
	if err != nil {
		return SSHSessionConfig{}, fmt.Errorf("failed to open session file: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&sessionConfig); err != nil {
		return SSHSessionConfig{}, fmt.Errorf("error decoding YAML: %v", err)
	}

	return sessionConfig, nil
}
