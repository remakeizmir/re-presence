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
	procCreateToolhelp32Snapshot  = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW           = kernel32.NewProc("Process32FirstW")
	procProcess32NextW            = kernel32.NewProc("Process32NextW")
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

// runningEditors walks the process list. The same question the other systems
// answer with a shell command, asked here through the API that avoids
// spawning tasklist on every tick.
func runningEditors() map[string]bool {
	const th32csSnapProcess = 0x00000002
	const maxPathShort = 260

	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snapshot == 0 || snapshot == ^uintptr(0) {
		return nil
	}
	defer procCloseHandle.Call(snapshot)

	type processEntry struct {
		Size              uint32
		Usage             uint32
		ProcessID         uint32
		DefaultHeapID     uintptr
		ModuleID          uint32
		Threads           uint32
		ParentProcessID   uint32
		PriorityClassBase int32
		Flags             uint32
		ExeFile           [maxPathShort]uint16
	}

	var entry processEntry
	entry.Size = uint32(unsafe.Sizeof(entry))

	found := map[string]bool{}
	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	for ret != 0 {
		name := syscall.UTF16ToString(entry.ExeFile[:])
		if shown, ok := editorFromProcess(name); ok {
			found[shown] = true
		}
		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	}
	return found
}
