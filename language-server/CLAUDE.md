# Language Server

This directory contains the Language Server Protocol (LSP) implementation for the Cow programming language, providing editor support for features like autocomplete, diagnostics, and navigation.

## Directory Structure
```
language-server/
├── cmd/
│   └── server/        # LSP server executable
├── internal/
│   ├── server/        # Core server implementation
│   ├── handlers/      # LSP message handlers
│   └── analysis/      # Code analysis integration
└── go.mod
```

## Current Implementation
- Basic server structure with placeholder implementations

## Usage
The language server is designed to be launched by editor extensions and communicate via the Language Server Protocol over JSON-RPC.

## Development Commands
```bash
cd language-server
go run cmd/server/main.go
go build cmd/server/main.go
go test ./...
```