# go-watch

A lightweight, fast, and configurable file watcher for Go, inspired by `nodemon`. It automatically watches for changes in source files and restarts processes or executes commands when changes are detected. It also supports live reload for frontend workflows.

## Features

- **High Performance**: Uses `fsnotify` for efficient file system event handling.
- **Cross-Platform**: Works on Windows, macOS, and Linux.
- **Advanced Watch Capabilities**: Supports multiple extensions and ignores specified directories.
- **Live Reload**: WebSocket-based live reload for frontend workflows.
- **Configuration File**: Supports JSON or YAML configuration files.
- **Debounce Mechanism**: Prevents rapid restarts.

## Installation

```bash
go install github.com/niradler/go-watch@latest
```

without go, you can use the binaries

```sh
curl -fsSL https://raw.githubusercontent.com/niradler/go-watch/master/scripts/install.sh -o install.sh && bash install.sh
```

or for windows:

```ps1
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/niradler/go-watch/master/scripts/install.ps1" -OutFile install.ps1; .\install.ps1
```

## Usage

### Basic Usage (CLI Arguments)

Watch `.go` and `.js` files, ignore `node_modules` and `.git` directories, and run `go run main.go` on file changes:

```bash
go-watch --ext go,js --ignore node_modules,.git --cmd "go run main.go"
```

### Live Reload

Enable live reload for frontend workflows on port `35729`:

```bash
go-watch --live-reload --live-reload-port 35729
```

### Configuration File

Create a `go-watch.config.json` or `go-watch.config.yaml` file for more advanced configurations.

#### JSON Example (`go-watch.config.json`)

```json
{
  "watch_dirs": [".", "src"],
  "ignore_dirs": ["node_modules", ".git"],
  "rules": [
    {
      "extensions": ["go"],
      "command": "go run main.go"
    },
    {
      "extensions": ["js"],
      "command": "node index.js"
    }
  ],
  "debounce_time": "500ms",
  "live_reload": true,
  "live_reload_port": 35729
}
```

Run with the config file:

```bash
go-watch --config go-watch.config.json
```

## Use Cases

### 1. Watching a Go Project

Watch `.go` files and restart the Go application on changes:

```bash
go-watch --ext go --cmd "go run main.go"
```

### 2. Watching a Node.js Project

Watch `.js` files and restart the Node.js application on changes:

```bash
go-watch --ext js --cmd "node index.js"
```

### 3. Watching a Frontend Project

Watch `.html`, `.css`, and `.js` files, and enable live reload:

```bash
go-watch --ext html,css,js --live-reload --live-reload-port 35729
```

### 4. Ignoring Specific Directories

Watch `.go` files but ignore the `vendor` and `tmp` directories:

```bash
go-watch --ext go --ignore vendor,tmp --cmd "go run main.go"
```

### 5. Using a Configuration File

Create a `go-watch.config.json` file:

```json
{
  "watch_dirs": [".", "src"],
  "ignore_dirs": ["node_modules", ".git"],
  "extensions": ["go", "js"],
  "command": "go run main.go",
  "debounce_time": "1s",
  "live_reload": false
}
```

Run with the config file:

```bash
go-watch --config go-watch.config.json
```

## Configuration Options

| Option            | Description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| `--config`        | Path to a JSON or YAML configuration file.                                  |
| `--ext`           | Comma-separated list of file extensions to watch (e.g., `go,js`).           |
| `--ignore`        | Comma-separated list of directories to ignore (e.g., `node_modules,.git`).  |
| `--cmd`           | Command to execute when changes are detected (e.g., `go run main.go`).      |
| `--debounce`      | Debounce time for file changes (e.g., `500ms`, `1s`).                       |
| `--live-reload`   | Enable live reload for frontend workflows.                                  |
| `--live-reload-port` | Port for the live reload server (default: `35729`).                     |

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT
