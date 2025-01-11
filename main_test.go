package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// TestIsIgnored tests the isIgnored function
func TestIsIgnored(t *testing.T) {
	ignoreDirs := []string{"node_modules", ".git"}
	if !isIgnored("node_modules", ignoreDirs) {
		t.Error("Expected node_modules to be ignored")
	}
	if isIgnored("src", ignoreDirs) {
		t.Error("Expected src to not be ignored")
	}
}

// TestIsWatchedFile tests the isWatchedFile function
func TestIsWatchedFile(t *testing.T) {
	extensions := []string{"go", "js"}
	if !isWatchedFile("main.go", extensions) {
		t.Error("Expected main.go to be watched")
	}
	if isWatchedFile("README.md", extensions) {
		t.Error("Expected README.md to not be watched")
	}
}

// TestLoadConfig tests the loadConfig function
func TestLoadConfig(t *testing.T) {
	// Create a temporary JSON config file
	configContent := `{
		"watch_dirs": [".", "src"],
		"ignore_dirs": ["node_modules", ".git"],
		"extensions": ["go", "js"],
		"command": "go run main.go",
		"debounce_time": "500ms",
		"live_reload": true,
		"live_reload_port": 35729
	}`
	tmpFile, err := os.CreateTemp("", "config*.json") // Ensure the file has a .json extension
	if err != nil {
		t.Fatal("Failed to create temp config file:", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatal("Failed to write to temp config file:", err)
	}

	// Load the config
	config, err := loadConfig(tmpFile.Name(), log.New(os.Stdout, "[test] ", log.LstdFlags))
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Command != "go run main.go" {
		t.Error("Expected command to be 'go run main.go'")
	}
	if !config.LiveReload {
		t.Error("Expected live reload to be enabled")
	}
}

// TestLiveReload tests the live reload functionality
func TestLiveReload(t *testing.T) {
	if !isFrontendFile("index.html") {
		t.Error("Expected index.html to be a frontend file")
	}
	if isFrontendFile("main.go") {
		t.Error("Expected main.go to not be a frontend file")
	}
}

// TestCommandExecution tests the command execution functionality
func TestCommandExecution(t *testing.T) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "echo Hello, World!")
	} else {
		cmd = exec.Command("echo", "Hello, World!")
	}

	output, err := cmd.Output()
	if err != nil {
		t.Error("Command execution failed:", err)
	}

	// Normalize line endings for cross-platform compatibility
	expected := "Hello, World!\n"
	if runtime.GOOS == "windows" {
		expected = "Hello, World!\r\n"
	}

	if strings.TrimSpace(string(output)) != strings.TrimSpace(expected) {
		t.Errorf("Unexpected command output: %q (expected: %q)", string(output), expected)
	}
}

// TestInvalidConfig tests handling of invalid config files
func TestInvalidConfig(t *testing.T) {
	// Create a temporary invalid config file
	configContent := `invalid json`
	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal("Failed to create temp config file:", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatal("Failed to write to temp config file:", err)
	}

	// Attempt to load the invalid config
	_, err = loadConfig(tmpFile.Name(), log.New(os.Stdout, "[test] ", log.LstdFlags))
	if err == nil {
		t.Error("Expected error for invalid config file, got nil")
	}
}

// TestIsFrontendFile tests the isFrontendFile function
func TestIsFrontendFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"index.html", true},
		{"styles.css", true},
		{"script.js", true},
		{"main.go", false},
		{"README.md", false},
	}

	for _, test := range tests {
		result := isFrontendFile(test.path)
		if result != test.expected {
			t.Errorf("isFrontendFile(%q) = %v, expected %v", test.path, result, test.expected)
		}
	}
}

// TestIsWatchedFileEdgeCases tests edge cases for isWatchedFile
func TestIsWatchedFileEdgeCases(t *testing.T) {
	tests := []struct {
		path       string
		extensions []string
		expected   bool
	}{
		{"main.go", []string{"go"}, true},
		{"main.go", []string{"js"}, false},
		{"script.js", []string{"js"}, true},
		{"script.js", []string{"go"}, false},
		{"README.md", []string{"go", "js"}, false},
		{"", []string{"go"}, false}, // Empty path
	}

	for _, test := range tests {
		result := isWatchedFile(test.path, test.extensions)
		if result != test.expected {
			t.Errorf("isWatchedFile(%q, %v) = %v, expected %v", test.path, test.extensions, result, test.expected)
		}
	}
}
