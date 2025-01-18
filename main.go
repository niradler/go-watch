package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	IgnoreDirs   []string `json:"ignore_dirs" yaml:"ignore_dirs"`
	DebounceTime string   `json:"debounce_time" yaml:"debounce_time"`
	Rules        []Rule   `json:"rules" yaml:"rules"`
}

// Rule represents a pattern and associated commands.
type Rule struct {
	Patterns []string  `json:"patterns" yaml:"patterns"`
	Commands []Command `json:"commands" yaml:"commands"`
}

// Command represents a single command to be executed.
type Command struct {
	Cmd      string `json:"cmd" yaml:"cmd"`
	Parallel bool   `json:"parallel,omitempty" yaml:"parallel,omitempty"`
}

var (
	configFile   = flag.String("config", "", "Path to the configuration file")
	ignoreDirs   = flag.String("ignore-dirs", "", "Comma-separated list of directories to ignore")
	debounceTime = flag.String("debounce-time", "500ms", "Debounce time for file changes")
	rules        = flag.String("rules", "", "Comma-separated list of rules in the format pattern:command")
	logger       = log.New(os.Stdout, "[go-watch] ", log.LstdFlags|log.Lshortfile)
	watcher      *fsnotify.Watcher
)

func init() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.Fatalf("Failed to initialize file watcher: %v", err)
	}
}

func main() {
	flag.Parse()
	_ = godotenv.Load()

	config, err := loadConfig(*configFile)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	if *configFile == "" {
		if *ignoreDirs != "" {
			config.IgnoreDirs = strings.Split(*ignoreDirs, ",")
		}
		config.DebounceTime = *debounceTime
		if *rules != "" {
			config.Rules = parseRules(*rules)
		}
	}

	debounceDuration, err := time.ParseDuration(config.DebounceTime)
	if err != nil {
		logger.Fatalf("Invalid debounce time: %v", err)
	}

	logger.Println("Starting watcher...")
	addPatternsToWatcher(config)

	defer watcher.Close()

	var lastChange time.Time
	eventQueue := make(chan string)

	go func() {
		for path := range eventQueue {
			executeRules(path, config)
		}
	}()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if time.Since(lastChange) > debounceDuration {
				lastChange = time.Now()
				eventQueue <- event.Name
				logger.Printf("Change detected: %s", event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Printf("Watcher error: %v", err)
		}
	}
}

func parseRules(rules string) []Rule {
	var parsedRules []Rule
	rulePairs := strings.Split(rules, ",")
	for _, pair := range rulePairs {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			parsedRules = append(parsedRules, Rule{
				Patterns: []string{parts[0]},
				Commands: []Command{{Cmd: parts[1], Parallel: false}},
			})
		}
	}
	return parsedRules
}

func loadConfig(path string) (Config, error) {
	var config Config

	if path == "" {
		// Try default configuration files
		defaultFiles := []string{"go-watch.config.yaml", "go-watch.config.json"}
		for _, file := range defaultFiles {
			if _, err := os.Stat(file); err == nil {
				path = file
				break
			}
		}
	}

	if path == "" {
		logger.Println("No configuration file supplied and no default configuration file found.")
		return config, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return config, err
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return config, err
		}
	default:
		return config, fmt.Errorf("unsupported configuration file format: %s", path)
	}

	return config, nil
}

func addPatternsToWatcher(config Config) {
	for _, rule := range config.Rules {
		for _, pattern := range rule.Patterns {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				logger.Printf("Failed to resolve pattern %s: %v", pattern, err)
				continue
			}
			for _, match := range matches {
				if isIgnoredDir(match, config.IgnoreDirs) {
					continue
				}
				err := watcher.Add(match)
				if err != nil {
					logger.Printf("Failed to watch file %s: %v", match, err)
				} else {
					logger.Printf("Watching file: %s", match)
				}
			}
		}
	}
}

func isIgnoredDir(path string, ignoreDirs []string) bool {
	for _, ignore := range ignoreDirs {
		if strings.Contains(path, ignore) {
			return true
		}
	}
	return false
}

func executeRules(filePath string, config Config) {
	for _, rule := range config.Rules {
		for _, pattern := range rule.Patterns {
			// Use gobwas/glob to match the file path with the pattern
			g := glob.MustCompile(pattern)
			if g.Match(filePath) {
				for _, cmd := range rule.Commands {
					logger.Printf("Executing command: %s", cmd.Cmd)
					if !executeCommand(cmd) {
						// Stop executing further commands if one fails in non-parallel mode
						if !cmd.Parallel {
							logger.Printf("Stopping execution due to failure of command: %s", cmd.Cmd)
							break
						}
					}
				}
			}
		}
	}
}

func executeCommand(cmd Command) bool {
	command := exec.Command("sh", "-c", cmd.Cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Env = os.Environ()

	if cmd.Parallel {
		go func() {
			if err := command.Run(); err != nil {
				logger.Printf("Command failed: %s, Error: %v", cmd.Cmd, err)
			}
		}()
		return true
	} else {
		if err := command.Run(); err != nil {
			logger.Printf("Command failed: %s, Error: %v", cmd.Cmd, err)
			return false
		}
	}
	return true
}
