#!/bin/bash
# Install script for macOS and Linux (Bash/Zsh)

# GitHub repository details
repoOwner="niradler"
repoName="go-watch"
installDir="$HOME/bin"  # Directory to install the binary

# Detect OS and architecture
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

# Map architecture to GitHub release format
case $arch in
    x86_64)
        arch="amd64"
        ;;
    arm64|aarch64)
        arch="arm64"
        ;;
    *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
esac

# Determine the binary name based on OS
case $os in
    darwin)
        binaryName="go-watch-darwin-$arch"
        ;;
    linux)
        binaryName="go-watch-linux-$arch"
        ;;
    *)
        echo "Unsupported OS: $os"
        exit 1
        ;;
esac

# Create the install directory if it doesn't exist
mkdir -p "$installDir"

# Download the latest release binary
releaseUrl="https://github.com/$repoOwner/$repoName/releases/latest/download/$binaryName"
outputPath="$installDir/go-watch"
curl -L "$releaseUrl" -o "$outputPath"

# Make the binary executable
chmod +x "$outputPath"

# Add the install directory to PATH if it's not already there
if [[ ":$PATH:" != *":$installDir:"* ]]; then
    echo "export PATH=\"$installDir:\$PATH\"" >> ~/.bashrc
    echo "export PATH=\"$installDir:\$PATH\"" >> ~/.zshrc
    echo "Added $installDir to PATH."
fi

echo "Installation complete! You can now run 'go-watch' from anywhere."