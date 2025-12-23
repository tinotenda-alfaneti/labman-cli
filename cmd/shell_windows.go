//go:build windows
// +build windows

package cmd

import (
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

const (
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004
	enableProcessedInput = 0x0001
)

func resetConsoleMode() error {
	handle := syscall.Handle(syscall.Stdin)

	var mode uint32
	r1, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		return err
	}

	// Ensure standard input modes are enabled
	mode |= enableLineInput | enableEchoInput | enableProcessedInput

	r1, _, err = procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if r1 == 0 {
		return err
	}

	return nil
}
