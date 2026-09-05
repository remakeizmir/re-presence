//go:build linux

package main

import (
	"os/exec"
	"strings"
)

// Reading the focused window on Linux.
//
// X11 only, through xdotool. Wayland deliberately gives no unprivileged
// process the focused window's title, which is the whole mechanism here — on
// Wayland the agent reports nothing and says so at startup.

func focusedWindow() (string, string, error) {
	id, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return "", "", err
	}
	window := strings.TrimSpace(string(id))

	class, err := exec.Command("xdotool", "getwindowclassname", window).Output()
	if err != nil {
		return "", "", err
	}
	name, err := exec.Command("xdotool", "getwindowname", window).Output()
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(string(class)), strings.TrimSpace(string(name)), nil
}

// runningEditors lists editors that are running, focused or not. ps rather
// than xdotool: this one is about the process, not the window, and it works
// under Wayland too even though the title does not.
func runningEditors() map[string]bool {
	out, err := exec.Command("ps", "-e", "-o", "comm=").Output()
	if err != nil {
		return nil
	}
	return editorsAmong(strings.Split(string(out), "\n"))
}
