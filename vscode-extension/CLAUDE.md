# VS Code Extension

This directory contains the Visual Studio Code extension for the Cow programming language, providing syntax highlighting, language server integration, and editor features.

## Directory Structure
```
vscode-extension/
├── src/
│   ├── extension.ts   # Extension entry point
│   └── client.ts      # Language server client
├── syntaxes/
│   └── cow-lang.tmGrammar.json  # Syntax highlighting rules
└── package.json       # Extension manifest
```

## Current Implementation
- Basic extension structure with TypeScript placeholders
- Syntax highlighting grammar for Cow language
- Language server client setup (placeholder)

## Development Commands
```bash
cd vscode-extension
npm install
npm run compile
npm run watch
```

## Installation
The extension can be packaged and installed in VS Code using the `vsce` tool once development is complete.