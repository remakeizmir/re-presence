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

// Windows names a running program by its executable, and the executable
// rarely matches what is written on the window: JetBrains ships idea64.exe,
// Android Studio ships studio64.exe, and Visual Studio — a different program
// from VS Code — is devenv.exe.
func TestEditorFromProcess(t *testing.T) {
	known := map[string]string{
		"Zed.exe":            "Zed",
		"Code.exe":           "VS Code",
		"cursor.exe":         "Cursor",
		"idea64.exe":         "IntelliJ",
		"studio64.exe":       "Android Studio",
		"devenv.exe":         "Visual Studio",
		"sublime_text":       "Sublime Text",
		"Zed":                "Zed", // macOS reports it without the suffix
		"Visual Studio Code": "",    // the window's name, not the process
	}

	for input, want := range known {
		got, ok := editorFromProcess(input)
		if want == "" {
			if ok {
				t.Errorf("%q editör sayıldı (%q)", input, got)
			}
			continue
		}
		if !ok || got != want {
			t.Errorf("%q → %q (%v), beklenen %q", input, got, ok, want)
		}
	}
}

func TestEditorFromProcessIgnoresEverythingElse(t *testing.T) {
	for _, name := range []string{"chrome.exe", "Slack", "explorer.exe", "", "  "} {
		if _, ok := editorFromProcess(name); ok {
			t.Errorf("%q editör sayıldı", name)
		}
	}
}

// The card follows the editor being open rather than being in front, so the
// process list has to be read the same way on every system: names, sometimes
// paths, sometimes with an .exe on the end.
func TestEditorsAmong(t *testing.T) {
	found := editorsAmong([]string{
		"Finder", "stable", "OrbStack", "Dia", // Warp calls itself "stable"
		"zed",
		"/usr/share/code/code",
		"Code.exe",
		"  ",
		"chrome",
	})

	for _, want := range []string{"Zed", "VS Code"} {
		if !found[want] {
			t.Errorf("%s bulunamadı: %v", want, found)
		}
	}
	if len(found) != 2 {
		t.Errorf("fazladan bir şey editör sayıldı: %v", found)
	}
}
