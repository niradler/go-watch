# Install script for Windows (PowerShell)

# GitHub repository details
$repoOwner = "niradler"
$repoName = "go-watch"
$binaryName = "go-watch-windows-amd64.exe"
$installDir = "$env:USERPROFILE\bin"  # Directory to install the binary
$pathEnv = [Environment]::GetEnvironmentVariable("Path", "User")

# Create the install directory if it doesn't exist
if (-Not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

# Download the latest release binary
$releaseUrl = "https://github.com/$repoOwner/$repoName/releases/latest/download/$binaryName"
$outputPath = "$installDir\go-watch.exe"
Invoke-WebRequest -Uri $releaseUrl -OutFile $outputPath

# Add the install directory to PATH if it's not already there
if (-Not ($pathEnv -split ";" -contains $installDir)) {
    [Environment]::SetEnvironmentVariable("Path", "$pathEnv;$installDir", "User")
    Write-Host "Added $installDir to PATH."
}

Write-Host "Installation complete! You can now run 'go-watch' from anywhere."