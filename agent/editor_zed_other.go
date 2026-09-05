//go:build !darwin && !linux

package main

// Zed on Windows keeps the same database, but the agent reads window titles
// there and Zed's window does carry one — so this stays empty until someone
// reports otherwise.
func zedState() (project, file string) { return "", "" }
