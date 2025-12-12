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
# Clone and run
git clone <your-repo>
cd go-pkgspy
go run main.go
```

Open http://localhost:8080 in your browser.

## Configuration

Edit the `packages` array in `main.go` to track your desired npm packages:

```go
packages = []string{
    "react/latest",
    "@angular/core/16.0.0", 
    "lodash/4.17.21",
}
```

Supported formats:
- `package-name/latest` - latest version
- `package-name/1.2.3` - specific version  
- `@scope/package/latest` - scoped packages

## API Endpoints

- `GET /` - Web dashboard
- `GET /refresh` - Force cache refresh

## Requirements

- Go 1.19+
- Internet connection for npm registry access

## License

MIT