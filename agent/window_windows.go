//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Reading the focused window on Windows.
//
// Three calls: which window has focus, what its title says, and which
// executable is behind it. The executable is what identifies the editor —
// window titles vary by version and by what is open, but Zed is always
// Zed.exe.
//
// No dependencies: the DLLs are loaded lazily through the standard library,
// which is available here because this file only builds on Windows.

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")

	procOpenProcess               = kernel32.NewProc("OpenProcess")
	procCloseHandle               = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
)

const (
	// Enough to ask a process its own name, and nothing else. The older
	// combination for this needed VM_READ, which reads like spyware in a
	// security prompt and is not necessary.
	processQueryLimitedInformation = 0x1000
	maxPath                        = 260
)

func focusedWindow() (app, title string, err error) {
	handle, _, _ := procGetForegroundWindow.Call()
	if handle == 0 {
		return "", "", fmt.Errorf("odakta pencere yok")
	}

	title = windowTitle(handle)

	var pid uint32
	procGetWindowThreadProcessID.Call(handle, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return "", title, nil
	}

	return processName(pid), title, nil
}

func windowTitle(handle uintptr) string {
	buf := make([]uint16, 512)
	n, _, _ := procGetWindowTextW.Call(handle, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:n])
}

// processName is the executable's own file name — "Code.exe", "idea64.exe".
// The rest of the path is dropped: which folder someone installed their editor
// into is not information this sends anywhere.
func processName(pid uint32) string {
	handle, _, _ := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if handle == 0 {
		return ""
	}
	defer procCloseHandle.Call(handle)

	buf := make([]uint16, maxPath)
	size := uint32(len(buf))
	ret, _, _ := procQueryFullProcessImageName.Call(
		handle, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return ""
	}

	full := syscall.UTF16ToString(buf[:size])
	for i := len(full) - 1; i >= 0; i-- {
		if full[i] == '\\' || full[i] == '/' {
			return full[i+1:]
		}
	}
	return full
}
