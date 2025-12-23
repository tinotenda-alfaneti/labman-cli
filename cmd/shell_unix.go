//go:build !windows
// +build !windows

package cmd

func resetConsoleMode() error {
	// No-op on non-Windows platforms
	return nil
}
