package remote

var current *SSHSession

func SetCurrent(s *SSHSession) { current = s }
func Current() *SSHSession { return current }
func IsActive() bool { return current != nil && current.IsConnected() }