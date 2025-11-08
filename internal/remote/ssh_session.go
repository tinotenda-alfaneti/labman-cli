package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"gopkg.in/yaml.v2"
)

const keyringService = "labman"

type SSHSession struct {
    Client  *ssh.Client
    Host    string
    User    string
	Password string
}

type SSHSessionConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Timeout time.Time `yaml:"timeout"`
}


func NewSSHSession(host, user, password string) (*SSHSession, error) {
	callback, err := hostKeyCallback("")
	if err != nil {
		return nil, fmt.Errorf("prepare host key callback: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: callback,
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoED25519,
		},
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &SSHSession{
		Client:  client,
		Host:    host,
		User:    user,
		Password: password,
	}, nil	
}


func (s *SSHSession) Run(cmd string) (string, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	return string(output), nil
}

func (s *SSHSession) SaveSession() error {
	
	createSessionFilePath, err := getSessionFilePath()
	if err != nil {
		return fmt.Errorf("failed to get session file path: %w", err)
	}
	return saveSessionToFile(createSessionFilePath, s)
}

func getSessionFilePath() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("resolve home directory: %w", err)
    }
    return filepath.Join(homeDir, ".labman", "sessions", "credentials.yaml"), nil
}

func saveSessionToFile(filePath string, session *SSHSession) error {

	sessionTimeout := time.Now().Add(1 * time.Minute)
	credentialKey := credentialsKey(session.Host, session.User)

	if err := keyring.Set(keyringService, credentialKey, session.Password); err != nil {
		return fmt.Errorf("store password in keyring: %w", err)
	}

	sessionData := SSHSessionConfig{
		Host:     session.Host,
		User:     session.User,
		Timeout:  sessionTimeout,
			
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

func (s *SSHSession) IsConnected() bool {
	_, err := s.Run("echo connected")
	return err == nil
}

func (s *SSHSession) Close() error {
	return s.Client.Close()
}

func hostKeyCallback(knownHostsPath string) (ssh.HostKeyCallback, error) {
	if knownHostsPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home directory: %w", err)
		}
		knownHostsPath = filepath.Join(homeDir, ".ssh", "known_hosts")
	}

	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("load known hosts: %w", err)
	}

	return callback, nil
}

func credentialsKey(host, user string) string {
	return fmt.Sprintf("%s@%s", user, host)
}
