# go-pkgspy

A lightweight Go web server that tracks npm package versions with caching and one-click install commands.

## Features

- 🔍 Track multiple npm packages in a clean web dashboard
- ⚡ 24-hour caching for fast performance
- 📋 One-click copy for install commands
- 📦 Export all dependencies as JSON
- 🔄 Manual refresh capability
- 🎯 Support for scoped packages and specific versions

## Quick Start

```bash
# Set NPM token (required)
export NPM_TOKEN=your_npm_token_here

# Clone and run
git clone <your-repo>
cd go-pkgspy
go run main.go

# Or build for your OS
## Windows
GOOS=windows GOARCH=amd64 go build -o ./out/go-pkgspy.exe

## Linux
GOOS=linux GOARCH=amd64 go build -o ./out/go-pkgspy

## macOS
GOOS=darwin GOARCH=amd64 go build -o ./out/go-pkgspy

# Run the binary with systemd
[Service]
Environment=NPM_TOKEN=your_token
ExecStart=/path/to/go-pkgspy
...other systemd config
```

Open http://localhost:9090 in your browser.

## Configuration

Create a `packages.txt` file in the project root with one package per line:

```
react/latest
@angular/core/16.0.0
lodash/4.17.21
```

Supported formats:
- `package-name/latest` - latest version
- `package-name/1.2.3` - specific version  
- `@scope/package/latest` - scoped packages

## API Endpoints

- `GET /` - Web dashboard
- `GET /refresh` - Force cache refresh

## Requirements

- Go 1.23+
- NPM token (set as `NPM_TOKEN` environment variable)
- Internet connection for npm registry access
- `packages.txt` file with packages to track

