//go:build windows
// +build windows

package cmd

import (
	"syscall"
	"testing"
	"unsafe"
)

func TestResetConsoleMode(t *testing.T) {
	// This test ensures resetConsoleMode doesn't panic and returns an error
	// when it should (e.g., when stdin is not a console)
	t.Run("can call without panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("resetConsoleMode() panicked: %v", r)
			}
		}()

		// Just call it - it may or may not error depending on the environment
		// The important thing is it doesn't panic
		_ = resetConsoleMode()
	})

	t.Run("sets correct console mode flags", func(t *testing.T) {
		handle := syscall.Handle(syscall.Stdin)

		// Get current mode
		var modeBefore uint32
		r1, _, _ := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&modeBefore)))
		if r1 == 0 {
			t.Skip("Not a console, skipping console mode test")
		}

		// Call resetConsoleMode
		err := resetConsoleMode()
		if err != nil {
			t.Fatalf("resetConsoleMode() error = %v", err)
		}

		// Get mode after reset
		var modeAfter uint32
		r1, _, _ = procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&modeAfter)))
		if r1 == 0 {
			t.Fatal("Failed to get console mode after reset")
		}

		// Check that the required flags are set
		requiredFlags := uint32(enableLineInput | enableEchoInput | enableProcessedInput)
		if (modeAfter & requiredFlags) != requiredFlags {
			t.Errorf("resetConsoleMode() didn't set required flags. Mode = 0x%X, want flags 0x%X",
				modeAfter, requiredFlags)
		}
	})
}

func TestConsoleConstants(t *testing.T) {
	// Test that constants are defined correctly
	tests := []struct {
		name     string
		constant uint32
		expected uint32
	}{
		{
			name:     "enableLineInput",
			constant: enableLineInput,
			expected: 0x0002,
		},
		{
			name:     "enableEchoInput",
			constant: enableEchoInput,
			expected: 0x0004,
		},
		{
			name:     "enableProcessedInput",
			constant: enableProcessedInput,
			expected: 0x0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = 0x%X, want 0x%X", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
