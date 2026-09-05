package main

import "testing"

// Every editor writes its window title differently, and the title is all this
// has to work with. These are the real ones, copied off running windows.
func TestSplitTitle(t *testing.T) {
	cases := []struct {
		name    string
		title   string
		file    string
		project string
	}{
		{"zed", "chat_service.go — server", "chat_service.go", "server"},
		{"zed, unsaved", "● main.go — hub", "main.go", "hub"},
		{"vs code", "composer.tsx — hub — Visual Studio Code", "composer.tsx", "hub"},
		{"vs code, project only", "hub — Visual Studio Code", "", "hub"},
		{"sublime", "index.html - remakeizmir", "index.html", "remakeizmir"},
		{"a path sneaks in", "/Users/can/gizli/main.go — /Users/can/gizli", "main.go", "gizli"},
		{"no separator", "server", "", "server"},
		{"empty", "   ", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file, project := splitTitle(tc.title)
			if file != tc.file || project != tc.project {
				t.Errorf("got (%q, %q), want (%q, %q)", file, project, tc.file, tc.project)
			}
		})
	}
}

func TestLanguageOf(t *testing.T) {
	if got := languageOf("chat_service.go"); got != "go" {
		t.Errorf("go: %q", got)
	}
	if got := languageOf("page.tsx"); got != "typescript" {
		t.Errorf("tsx: %q", got)
	}
	if got := languageOf("nedir bu"); got != "" {
		t.Errorf("bilinmeyen: %q", got)
	}
}
