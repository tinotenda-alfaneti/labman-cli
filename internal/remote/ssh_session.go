package remote

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"

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


func (s *SSHSession) IsConnected() bool {
	_, err := s.Run("echo connected")
	return err == nil
}

func (s *SSHSession) Close() error {
	return s.Client.Close()
}

