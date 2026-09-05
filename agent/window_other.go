//go:build !darwin && !linux && !windows

package main

import "fmt"

// Everything else. The agent still runs — it will simply never find an editor
// — because failing to start is a worse answer than saying so once.
func focusedWindow() (app, title string, err error) {
	return "", "", fmt.Errorf("bu işletim sisteminde pencere başlığı okunamıyor")
}

func runningEditors() map[string]bool { return nil }
