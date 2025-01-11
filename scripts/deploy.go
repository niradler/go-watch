package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	projectName = "go-watch" // Replace with your project name
	version     = "v0.0.6"   // Replace with your version
)

func main() {
	// Step 1: Run tests
	fmt.Println("Running tests...")
	if err := runTests(); err != nil {
		log.Fatalf("Tests failed: %v", err)
	}
	fmt.Println("Tests passed!")

	// Step 2: Build binaries for all platforms
	fmt.Println("Building binaries...")
	platforms := []struct {
		os   string
		arch string
	}{
		{"windows", "amd64"},
		{"darwin", "amd64"},
		{"linux", "amd64"},
	}

	for _, platform := range platforms {
		if err := buildBinary(platform.os, platform.arch); err != nil {
			log.Fatalf("Failed to build for %s/%s: %v", platform.os, platform.arch, err)
		}
	}
	fmt.Println("Binaries built successfully!")

	// Step 3: Create GitHub release and upload binaries
	fmt.Println("Creating GitHub release...")
	if err := createGitHubRelease(); err != nil {
		log.Fatalf("Failed to create GitHub release: %v", err)
	}
	fmt.Println("GitHub release created successfully!")
}

// runTests runs `go test` and checks for errors
func runTests() error {
	cmd := exec.Command("go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// buildBinary builds the binary for a specific platform
func buildBinary(goos, goarch string) error {
	// Generate a unique binary name
	binaryName := fmt.Sprintf("%s-%s-%s", projectName, goos, goarch)
	if goos == "windows" {
		binaryName += ".exe"
	}

	// Output path
	output := filepath.Join("bin", binaryName)

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll("bin", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", output, ".")
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOOS=%s", goos), fmt.Sprintf("GOARCH=%s", goarch))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// createGitHubRelease creates a GitHub release and uploads binaries
func createGitHubRelease() error {
	// Ensure GitHub CLI (gh) is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is not installed. Install it from https://cli.github.com")
	}

	// Create the release
	releaseCmd := exec.Command("gh", "release", "create", version, "--title", version, "--notes", "Release notes")
	releaseCmd.Stdout = os.Stdout
	releaseCmd.Stderr = os.Stderr
	if err := releaseCmd.Run(); err != nil {
		return fmt.Errorf("failed to create GitHub release: %v", err)
	}

	// Upload binaries
	binDir := "bin"
	files, err := os.ReadDir(binDir)
	if err != nil {
		return fmt.Errorf("failed to read bin directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			path := filepath.Join(binDir, file.Name())
			uploadCmd := exec.Command("gh", "release", "upload", version, path)
			uploadCmd.Stdout = os.Stdout
			uploadCmd.Stderr = os.Stderr
			if err := uploadCmd.Run(); err != nil {
				return fmt.Errorf("failed to upload %s: %v", path, err)
			}
			fmt.Printf("Uploaded %s\n", path)
		}
	}

	return nil
}
