# Go AST Parser

A modular Go tool that analyzes Go projects and extracts structured code chunks for semantic search and analysis.

## 🚀 Quick Start

```bash
# Build the tool
make build

# Analyze a Go project
./bin/go-ast-parser -path /path/to/your/go/project

# Output: code_chunks.json
```

## 📋 Features

- ✅ **Comprehensive Analysis** - Processes main module + vendor dependencies
- ✅ **Rich Metadata** - Types, symbols, functions, methods extraction  
- ✅ **JSON Output** - Structured data for semantic search systems
- ✅ **Modular Architecture** - Clean, testable, maintainable codebase
- ✅ **Type-Safe Analysis** - Uses official Go AST and type checking tools

## 🏗️ Architecture

- **`cmd/go-ast-parser`** - CLI entry point
- **`pkg/loader`** - Package loading (main + vendor)
- **`pkg/parser`** - AST parsing & chunk extraction
- **`pkg/analyzer`** - Type analysis & symbol extraction
- **`pkg/transform`** - Code transformations
- **`pkg/output`** - JSON serialization
- **`pkg/types`** - Core data structures

📖 **[Full Architecture Documentation](ARCHITECTURE.md)**

## 🔧 Build & Usage

```bash
# Available commands
make help

# Build binary
make build

# Clean artifacts
make clean
```

## 📊 Output Format

```json
{
  "id": "file_path:line_start-line_end-entity_name",
  "document": "actual_code_content", 
  "metadata": {
    "file_path": "/path/to/file.go",
    "package_name": "main",
    "entity_type": "function",
    "accessed_symbols": ["package.Symbol"]
  }
}
```

## 📝 Requirements

- Go 1.23+
- Target project must have `go.mod` file

---

*For detailed architecture documentation with diagrams, see [ARCHITECTURE.md](ARCHITECTURE.md)* 