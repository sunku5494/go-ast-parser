package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sunku5494/go-ast-parser/pkg/loader"
	"github.com/sunku5494/go-ast-parser/pkg/output"
	"github.com/sunku5494/go-ast-parser/pkg/parser"
)

func main() {
	// Define command-line flag for project path
	projectPath := flag.String("path", "", "Absolute path to the Go module's root directory (must contain go.mod file)")
	flag.Parse()

	// Validate that project path is provided
	if *projectPath == "" {
		fmt.Fprintf(os.Stderr, "Error: Project path is required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -path /path/to/go/project\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Validate that the path exists and contains go.mod
	if _, err := os.Stat(*projectPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Project path does not exist: %s\n", *projectPath)
		os.Exit(1)
	}

	goModPath := filepath.Join(*projectPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: go.mod file not found in project path: %s\n", *projectPath)
		fmt.Fprintf(os.Stderr, "Make sure the path points to a Go module root directory\n")
		os.Exit(1)
	}

	fmt.Printf("Processing Go project at: %s\n", *projectPath)

	// Step 1: Load packages from project
	allPkgs, err := loader.LoadGoProject(*projectPath)
	if err != nil {
		log.Fatalf("Error loading Go project: %v", err)
	}

	// Step 2: Parse packages and extract code chunks
	chunks, err := parser.ParsePackages(allPkgs, *projectPath)
	if err != nil {
		log.Fatalf("Error parsing packages: %v", err)
	}

	// Step 3: Write chunks to JSON output
	outputFileName := "code_chunks.json"
	err = output.WriteChunksToJSON(chunks, outputFileName)
	if err != nil {
		log.Fatalf("Error writing output: %v", err)
	}
} 