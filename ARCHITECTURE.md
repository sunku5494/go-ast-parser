# Go AST Parser - Architecture Documentation

## 📋 Executive Summary

The Go AST Parser is a modular command-line tool that analyzes Go projects and extracts structured code chunks for code indexing, search, and static analysis. It processes both main module and vendor dependencies, generating comprehensive metadata about functions, types, and symbols for exact code search and navigation.


**Processing Pipeline:** Load → Parse → Analyze → Transform → Output
**Languages:** Pure Go with go/ast and golang.org/x/tools

---

## 🏗️ System Architecture

### Core Components

| Package | Responsibility | Key Functions |
|---------|---------------|---------------|
| `cmd/go-ast-parser` | CLI Entry Point | Flag parsing, input validation, orchestration |
| `pkg/loader` | Package Loading | LoadGoProject(), vendor + main module loading |
| `pkg/parser` | AST Parsing | ParsePackages(), declaration processing |
| `pkg/analyzer` | Type Analysis | GetTypeString(), ExtractAccessedSymbols() |
| `pkg/transform` | Code Transformation | ApplyQualifierReplacements() |
| `pkg/output` | Output Handling | WriteChunksToJSON() |
| `pkg/types` | Data Structures | ChromaDocument struct |

### Architecture Diagram

```mermaid
graph TB
    %% Input/Output
    INPUT[["Go Project Path<br/>(CLI Input)"]]
    OUTPUT[["code_chunks.json<br/>"]]
    
    %% Main Entry Point
    MAIN["cmd/go-ast-parser/main.go<br/>🚀 CLI Entry Point<br/>• Flag parsing<br/>• Input validation<br/>• Orchestration"]
    
    %% Core Packages
    LOADER["pkg/loader<br/>📦 Package Loader<br/>• LoadGoProject()<br/>• Main module loading<br/>• Vendor directory loading<br/>• Package deduplication"]
    
    PARSER["pkg/parser<br/>🔍 AST Parser<br/>• ParsePackages()<br/>• AST traversal<br/>• Declaration processing<br/>• Chunk extraction"]
    
    ANALYZER["pkg/analyzer<br/>🧠 Type Analyzer<br/>• GetTypeString()<br/>• GetSignature()<br/>• ExtractAccessedSymbols()<br/>• Type inference"]
    
    TRANSFORM["pkg/transform<br/>🔄 Code Transformer<br/>• ApplyQualifierReplacements()<br/>• Package qualifier resolution<br/>• Import path expansion"]
    
    OUTPUT_PKG["pkg/output<br/>📤 Output Handler<br/>• WriteChunksToJSON()<br/>• JSON serialization<br/>• File writing"]
    
    TYPES["pkg/types<br/>📋 Data Structures<br/>• ChromaDocument<br/>• Metadata schema<br/>• JSON tags"]
    
    %% External Dependencies  
    GO_PACKAGES["golang.org/x/tools/go/packages<br/>🛠️ Go Tools<br/>• Package loading<br/>• AST parsing<br/>• Type checking"]
    
    %% Data Structures
    PKG_DATA[("Go Packages<br/>[]*packages.Package")]
    CHUNK_DATA[("Code Chunks<br/>[]ChromaDocument")]
    
    %% Flow Connections
    INPUT --> MAIN
    MAIN --> LOADER
    MAIN --> PARSER  
    MAIN --> OUTPUT_PKG
    
    LOADER --> PKG_DATA
    PKG_DATA --> PARSER
    PARSER --> CHUNK_DATA
    CHUNK_DATA --> OUTPUT_PKG
    OUTPUT_PKG --> OUTPUT
    
    %% Dependencies
    PARSER --> ANALYZER
    PARSER --> TRANSFORM
    PARSER --> TYPES
    OUTPUT_PKG --> TYPES
    
    LOADER --> GO_PACKAGES
    PARSER --> GO_PACKAGES
    ANALYZER --> GO_PACKAGES
    TRANSFORM --> GO_PACKAGES
    
    %% Processing Flow Numbers
    MAIN -.->|"1. Load"| LOADER
    MAIN -.->|"2. Parse"| PARSER  
    MAIN -.->|"3. Output"| OUTPUT_PKG
```

---

## 🔄 Processing Pipeline

### Data Flow Diagram

```mermaid
flowchart TD
    START([🚀 Start: go-ast-parser -path /project])
    
    %% Input Validation
    VALIDATE{{"🔍 Validate Input<br/>• Path exists?<br/>• go.mod exists?"}}
    ERROR_EXIT[("❌ Exit with Error")]
    
    %% Step 1: Package Loading
    LOAD_MAIN["📦 Load Main Module<br/>packages.Load('./...')<br/>from project root"]
    LOAD_VENDOR["📦 Load Vendor Packages<br/>packages.Load('./...')<br/>from vendor directory"]
    DEDUPE["🔄 Deduplicate Packages<br/>Remove duplicate IDs<br/>Merge package lists"]
    
    %% Step 2: AST Processing  
    PROCESS_PKG["🔍 Process Each Package<br/>• Validate TypesInfo<br/>• Check Syntax trees<br/>• Verify Fset"]
    
    PROCESS_FILE["📄 Process Each File<br/>• Read source code<br/>• Check if vendored<br/>• Extract file metadata"]
    
    PROCESS_DECL["📋 Process Each Declaration<br/>• Function declarations<br/>• Type declarations<br/>• Value declarations"]
    
    %% Analysis & Transformation
    ANALYZE["🧠 Analyze Declaration<br/>• Extract type information<br/>• Get function signatures<br/>• Find accessed symbols"]
    
    TRANSFORM_CODE["🔄 Transform Code<br/>• Replace package qualifiers<br/>• Expand import paths<br/>• Apply transformations"]
    
    CREATE_CHUNK["📝 Create Chunk<br/>• Generate unique ID<br/>• Package metadata<br/>• Create ChromaDocument"]
    
    %% Step 3: Output
    COLLECT["📊 Collect All Chunks<br/>Aggregate from all<br/>packages and files"]
    
    SERIALIZE["📤 Serialize to JSON<br/>json.MarshalIndent<br/>Pretty formatting"]
    
    WRITE_FILE["💾 Write to File<br/>code_chunks.json<br/>"]
    
    SUCCESS([✅ Success: Chunks Extracted])
    
    %% Connections
    START --> VALIDATE
    VALIDATE -->|Valid| LOAD_MAIN
    VALIDATE -->|Invalid| ERROR_EXIT
    
    LOAD_MAIN --> LOAD_VENDOR
    LOAD_VENDOR --> DEDUPE
    
    DEDUPE --> PROCESS_PKG
    PROCESS_PKG --> PROCESS_FILE
    PROCESS_FILE --> PROCESS_DECL
    
    PROCESS_DECL --> ANALYZE
    ANALYZE --> TRANSFORM_CODE
    TRANSFORM_CODE --> CREATE_CHUNK
    
    CREATE_CHUNK --> COLLECT
    COLLECT --> SERIALIZE
    SERIALIZE --> WRITE_FILE
    WRITE_FILE --> SUCCESS
```

---

## 📊 Data Structures

### ChromaDocument Schema
```json
{
  "id": "file_path:line_start-line_end-entity_name",
  "document": "actual_code_content",
  "metadata": {
    "file_path": "/path/to/file.go",
    "package_name": "main",
    "is_vendored": false,
    "accessed_symbols": ["package.Symbol"],
    "entity_type": "function|method|struct|interface|alias_or_basic|const|var",
    "entity_name": "EntityName",
    "receiver_type": "ReceiverType" // for methods only
  }
}
```

---

## 📝 Implementation Notes

### Key Design Decisions:
- **Vendor Inclusion** - Processes both main and vendor code for completeness
- **Metadata Richness** - Comprehensive symbol and type information
- **Unique IDs** - File path + line range + entity name for chunk identification
- **JSON Output** - Human-readable format for easy integration
- **Modular Architecture** - Package-based organization for maintainability

### Dependencies:
- **golang.org/x/tools/go/packages** - Official Go package loading
- **Standard Library** - go/ast, go/token, go/types for AST processing

---

*Module: github.com/sunku5494/go-ast-parser*
*Version: 1.0.0* 
