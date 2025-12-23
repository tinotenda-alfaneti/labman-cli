//go:build !windows
// +build !windows

package cmd

import "testing"

func TestResetConsoleMode_Unix(t *testing.T) {
	t.Run("always returns nil on unix", func(t *testing.T) {
		err := resetConsoleMode()
		if err != nil {
			t.Errorf("resetConsoleMode() on Unix should always return nil, got %v", err)
		}
	})

	t.Run("doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("resetConsoleMode() panicked on Unix: %v", r)
			}
		}()

		resetConsoleMode()
	})

	t.Run("can be called multiple times", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := resetConsoleMode()
			if err != nil {
				t.Errorf("resetConsoleMode() call %d returned error: %v", i+1, err)
			}
		}
	})
}
