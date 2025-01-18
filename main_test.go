package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gobwas/glob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCommand represents a mock implementation of the Command structure
type MockCommand struct {
	mock.Mock
}

func (m *MockCommand) Execute() error {
	args := m.Called()
	return args.Error(0)
}

// Test loading configuration
func TestLoadConfig(t *testing.T) {
	configPath := "test_config.yaml"
	configData := []byte(`
ignore_dirs:
  - "bin"
  - ".git"
debounce_time: "500ms"
rules:
  - patterns:
      - "**/*.go"
    commands:
      - cmd: "go test -v ./..."
        parallel: false
`)

	err := os.MkdirAll("tmp", 0755)
	assert.NoError(t, err)

	// Create test config file
	err = os.WriteFile(configPath, configData, 0644)
	assert.NoError(t, err)
	defer os.RemoveAll("tmp")
	// Load the config
	config, err := loadConfig(configPath)
	assert.NoError(t, err)

	// Assertions
	assert.Equal(t, 2, len(config.IgnoreDirs))
	assert.Equal(t, "500ms", config.DebounceTime)
	assert.Len(t, config.Rules, 1)
	assert.Equal(t, "**/*.go", config.Rules[0].Patterns[0])
	assert.Equal(t, "go test -v ./...", config.Rules[0].Commands[0].Cmd)
}

// Test pattern matching and rule execution
func TestPatternMatchingAndRuleExecution(t *testing.T) {
	// Mock the watcher to simulate file changes
	mockCmd := new(MockCommand)
	mockCmd.On("Execute").Return(nil)

	// Test rules with patterns
	filePath := "tmp/deploy.go"
	config := Config{
		Rules: []Rule{
			{
				Patterns: []string{"**/*.go"},
				Commands: []Command{
					{Cmd: "go test -v ./...", Parallel: false},
				},
			},
		},
	}

	// Simulate pattern match
	for _, rule := range config.Rules {
		for _, pattern := range rule.Patterns {
			g := glob.MustCompile(pattern)
			if g.Match(filePath) {
				// Simulate execution of commands
				for _, cmd := range rule.Commands {
					fmt.Println("Executing command:", cmd.Cmd)
					mockCmd.Execute()
				}
			}
		}
	}

	mockCmd.AssertExpectations(t)
}

// Test watcher for file events
func TestFileWatcher(t *testing.T) {
	// Simulate file changes
	fileChanges := []string{
		"tmp/deploy.go",
		"main.go",
	}

	// Set up configuration with pattern to match *.go files
	config := Config{
		Rules: []Rule{
			{
				Patterns: []string{"*.go"},
				Commands: []Command{
					{Cmd: "go test -v ./...", Parallel: false},
				},
			},
		},
	}

	// Watch for file changes
	for _, file := range fileChanges {
		time.Sleep(1 * time.Second) // Simulate debounce

		// Check if the file path matches any pattern
		for _, rule := range config.Rules {
			for _, pattern := range rule.Patterns {
				g := glob.MustCompile(pattern)
				if g.Match(file) {
					// Simulate command execution
					for _, cmd := range rule.Commands {
						fmt.Println("Executing command:", cmd.Cmd)
						// Directly call Execute here (mocked above)
						mockCmd := new(MockCommand)
						mockCmd.On("Execute").Return(nil)
						mockCmd.Execute()
						mockCmd.AssertExpectations(t)
					}
				}
			}
		}
	}
}

// Test adding new files dynamically
func TestAddingNewFiles(t *testing.T) {
	// Set up configuration with a rule that watches *.go files
	config := Config{
		Rules: []Rule{
			{
				Patterns: []string{"*.go"},
				Commands: []Command{
					{Cmd: "go test -v ./...", Parallel: false},
				},
			},
		},
	}

	err := os.MkdirAll("tmp", 0755)
	assert.NoError(t, err)
	// Simulate adding a new file to the watched directory
	newFile := "tmp/newfile.go"
	err = os.WriteFile(newFile, []byte("package main"), 0644)
	assert.NoError(t, err)
	defer os.RemoveAll("tmp")

	// Watch for new files
	// Here, you could extend the test by simulating that `newFile` is added dynamically
	// and your watcher should be able to catch this event (mocking can be used for this part).

	// Simulate that a change was detected and execute commands
	for _, rule := range config.Rules {
		for _, pattern := range rule.Patterns {
			g := glob.MustCompile(pattern)
			if g.Match(newFile) {
				for _, cmd := range rule.Commands {
					fmt.Println("Executing command:", cmd.Cmd)
					mockCmd := new(MockCommand)
					mockCmd.On("Execute").Return(nil)
					mockCmd.Execute()
					mockCmd.AssertExpectations(t)
				}
			}
		}
	}
}
