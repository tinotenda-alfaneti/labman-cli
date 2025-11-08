package remote

import (
	"fmt"
	"os"
	"time"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v2"
)

func LoadSession() (*SSHSession, error) {
	sessionFile, err := getSessionFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get session file path: %v", err)
	}

	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("session file does not exist: %v", err)
	}

	sessionData , err := loadSessionDataFromFile(sessionFile)

	if err != nil {
		return nil, fmt.Errorf("failed to load session data from file: %v", err)
	}

	if sessionData.Timeout.Before(time.Now()) {
		return nil, fmt.Errorf("session has expired: %v", err)
	}

	password, err := keyring.Get(keyringService, credentialsKey(sessionData.Host, sessionData.User))
	if err != nil {
		return nil, fmt.Errorf("failed to load password from keyring: %w", err)
	}

	return NewSSHSession(sessionData.Host, sessionData.User, password)

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
