package remote

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

const keyringService = "labman"

type SSHSession struct {
	Client   *ssh.Client
	Host     string
	User     string
	Password string
}

type SSHSessionConfig struct {
	Host    string    `yaml:"host"`
	User    string    `yaml:"user"`
	Timeout time.Time `yaml:"timeout"`
}

func NewSSHSession(host, user, password string) (*SSHSession, error) {

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSASHA256,
			ssh.KeyAlgoRSASHA512,
			ssh.KeyAlgoRSA,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
		},
		Timeout: 30 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &SSHSession{
		Client:   client,
		Host:     host,
		User:     user,
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

func (s *SSHSession) RunStream(cmd string, w io.Writer) error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()
	go io.Copy(w, stderr)

	if err := session.Start(cmd); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Fprintln(w, scanner.Text())
	}

	if err := session.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *SSHSession) Close() error {
	if s.Client != nil {
		return s.Client.Close()
	}
	return nil
}

func (s *SSHSession) IsConnected() bool {
	if s.Client == nil {
		return false
	}

	session, err := s.Client.NewSession()
	if err != nil {
		return false
	}
	session.Close()
	return true
}
