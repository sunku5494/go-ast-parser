package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/sunku5494/go-ast-parser/pkg/analyzer"
	"github.com/sunku5494/go-ast-parser/pkg/transform"
	"github.com/sunku5494/go-ast-parser/pkg/types"
)

// ParsePackages extracts code chunks from loaded Go packages.
// It processes each package's AST to create documented chunks with metadata.
func ParsePackages(allPkgs []*packages.Package, projectPath string) ([]types.ChromaDocument, error) {
	var allChunks []types.ChromaDocument

	// Resolve the absolute path of the vendor directory once for `is_vendored` check
	vendorDirPath := filepath.Join(projectPath, "vendor")
	absVendorPath, err := filepath.Abs(vendorDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for vendor directory: %w", err)
	}

	// Process all unique packages
	for _, pkg := range allPkgs {
		if pkg.TypesInfo == nil || pkg.Syntax == nil || pkg.Fset == nil {
			log.Printf("Skipping package %s due to missing type information, syntax trees, or fileset.", pkg.ID)
			continue
		}

		chunks, err := processPackage(pkg, absVendorPath)
		if err != nil {
			log.Printf("Error processing package %s: %v", pkg.ID, err)
			continue
		}

		allChunks = append(allChunks, chunks...)
	}

	return allChunks, nil
}

// processPackage processes a single package and extracts all code chunks from it.
func processPackage(pkg *packages.Package, absVendorPath string) ([]types.ChromaDocument, error) {
	var chunks []types.ChromaDocument

	for _, file := range pkg.Syntax {
		filePath := pkg.Fset.File(file.Pos()).Name()
		originalFileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading file %s: %v", filePath, err)
			continue
		}

		packageName := pkg.Name
		originalFileContentString := string(originalFileBytes)

		// Determine if the file is from the vendor directory using a robust check
		isVendored := strings.HasPrefix(filePath, absVendorPath+string(filepath.Separator))

		fileChunks := processFileDeclarations(file, pkg, filePath, packageName, isVendored, originalFileContentString)
		chunks = append(chunks, fileChunks...)
	}

	return chunks, nil
}

// processFileDeclarations processes all declarations in a single file.
func processFileDeclarations(file *ast.File, pkg *packages.Package, filePath, packageName string, isVendored bool, originalFileContentString string) []types.ChromaDocument {
	var chunks []types.ChromaDocument

	for _, decl := range file.Decls {
		metadata := map[string]interface{}{
			"file_path":    filePath,
			"package_name": packageName,
			"is_vendored":  isVendored,
		}

		startPos := pkg.Fset.Position(decl.Pos())
		endPos := pkg.Fset.Position(decl.End())

		startOffset := startPos.Offset
		endOffset := endPos.Offset

		if startOffset < 0 || endOffset > len(originalFileContentString) || startOffset > endOffset {
			log.Printf("Warning: Invalid offsets for declaration in %s (line %d): start=%d, end=%d, file_len=%d. Skipping declaration.",
				filePath, startPos.Line, startOffset, endOffset, len(originalFileContentString))
			continue
		}
		declChunkCode := originalFileContentString[startOffset:endOffset]

		// Extract all accessed symbols for this declaration
		accessedSymbols := analyzer.ExtractAccessedSymbols(decl, pkg.TypesInfo)
		metadata["accessed_symbols"] = accessedSymbols

		declChunks := processDeclaration(decl, pkg, declChunkCode, metadata, filePath, startPos, endPos)
		if declChunks != nil {
			chunks = append(chunks, declChunks...)
		}
	}

	return chunks
}

// processDeclaration processes a single AST declaration and returns ChromaDocuments.
func processDeclaration(decl ast.Decl, pkg *packages.Package, declChunkCode string, metadata map[string]interface{}, filePath string, startPos, endPos token.Position) []types.ChromaDocument {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		chunk := processFunctionDeclaration(d, pkg, declChunkCode, metadata, filePath, startPos, endPos)
		if chunk != nil {
			return []types.ChromaDocument{*chunk}
		}
		return nil
	case *ast.GenDecl:
		return processGeneralDeclaration(d, pkg, declChunkCode, metadata, filePath, startPos, endPos)
	default:
		return nil
	}
}

// processFunctionDeclaration processes function and method declarations.
func processFunctionDeclaration(funcDecl *ast.FuncDecl, pkg *packages.Package, declChunkCode string, metadata map[string]interface{}, filePath string, startPos, endPos token.Position) *types.ChromaDocument {
	metadata["entity_type"] = "function"
	metadata["entity_name"] = funcDecl.Name.Name

	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		metadata["entity_type"] = "method"
		receiverType := analyzer.GetTypeString(funcDecl.Recv.List[0].Type, pkg.TypesInfo)
		metadata["receiver_type"] = receiverType
		metadata["entity_name"] = receiverType + "." + funcDecl.Name.Name
	}

	finalChunkCode := transform.ApplyQualifierReplacements(declChunkCode, funcDecl, pkg.TypesInfo)

	return &types.ChromaDocument{
		ID:       fmt.Sprintf("%s:%d-%d-%s", filePath, startPos.Line, endPos.Line, funcDecl.Name.Name),
		Document: finalChunkCode,
		Metadata: metadata,
	}
}

// processGeneralDeclaration processes type, const, and var declarations and returns all chunks.
func processGeneralDeclaration(genDecl *ast.GenDecl, pkg *packages.Package, declChunkCode string, metadata map[string]interface{}, filePath string, startPos, endPos token.Position) []types.ChromaDocument {
	if genDecl.Tok == token.IMPORT {
		return nil // Skip import declarations
	}

	var chunks []types.ChromaDocument

	// Process each specification in the general declaration
	for _, spec := range genDecl.Specs {
		specStartPos := pkg.Fset.Position(spec.Pos())
		specEndPos := pkg.Fset.Position(spec.End())
		
		chunk := processSpecification(spec, genDecl, pkg, metadata, filePath, specStartPos, specEndPos)
		if chunk != nil {
			chunks = append(chunks, *chunk)
		}
	}

	return chunks
}

// processSpecification processes individual specifications (type, const, var).
func processSpecification(spec ast.Spec, genDecl *ast.GenDecl, pkg *packages.Package, baseMetadata map[string]interface{}, filePath string, specStartPos, specEndPos token.Position) *types.ChromaDocument {
	// Create a copy of the base metadata for this specification
	specMetadata := make(map[string]interface{})
	for k, v := range baseMetadata {
		specMetadata[k] = v
	}

	switch s := spec.(type) {
	case *ast.TypeSpec:
		return processTypeSpecification(s, pkg, specMetadata, filePath, specStartPos, specEndPos)
	case *ast.ValueSpec:
		return processValueSpecification(s, pkg, specMetadata, filePath, specStartPos, specEndPos)
	default:
		return nil
	}
}

// processTypeSpecification processes type declarations (struct, interface, etc.).
func processTypeSpecification(typeSpec *ast.TypeSpec, pkg *packages.Package, specMetadata map[string]interface{}, filePath string, specStartPos, specEndPos token.Position) *types.ChromaDocument {
	entityName := typeSpec.Name.Name
	specMetadata["entity_name"] = entityName

	if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
		specMetadata["type_category"] = "struct"
	} else if _, isInterface := typeSpec.Type.(*ast.InterfaceType); isInterface {
		specMetadata["type_category"] = "interface"
	} else {
		specMetadata["type_category"] = "alias_or_basic"
	}

	// Read the original code for this specification
	originalFileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		return nil
	}
	originalFileContentString := string(originalFileBytes)
	
	specStartOffset := specStartPos.Offset
	specEndOffset := specEndPos.Offset
	
	if specStartOffset < 0 || specEndOffset > len(originalFileContentString) || specStartOffset > specEndOffset {
		log.Printf("Warning: Invalid offsets for spec in %s (line %d): start=%d, end=%d, file_len=%d. Skipping spec.",
			filePath, specStartPos.Line, specStartOffset, specEndOffset, len(originalFileContentString))
		return nil
	}
	specChunkCode := originalFileContentString[specStartOffset:specEndOffset]

	finalChunkCode := transform.ApplyQualifierReplacements(specChunkCode, typeSpec, pkg.TypesInfo)

	return &types.ChromaDocument{
		ID:       fmt.Sprintf("%s:%d-%d-%s", filePath, specStartPos.Line, specEndPos.Line, entityName),
		Document: finalChunkCode,
		Metadata: specMetadata,
	}
}

// processValueSpecification processes const and var declarations.
func processValueSpecification(valueSpec *ast.ValueSpec, pkg *packages.Package, specMetadata map[string]interface{}, filePath string, specStartPos, specEndPos token.Position) *types.ChromaDocument {
	var names []string
	for _, name := range valueSpec.Names {
		names = append(names, name.Name)
	}
	entityName := strings.Join(names, ", ")
	specMetadata["entity_name"] = entityName

	if valueSpec.Type != nil {
		specMetadata["type"] = analyzer.GetTypeString(valueSpec.Type, pkg.TypesInfo)
	} else if len(valueSpec.Values) > 0 {
		if tv := pkg.TypesInfo.TypeOf(valueSpec.Values[0]); tv != nil {
			specMetadata["type"] = tv.String()
		}
	}

	// Read the original code for this specification
	originalFileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file %s: %v", filePath, err)
		return nil
	}
	originalFileContentString := string(originalFileBytes)
	
	specStartOffset := specStartPos.Offset
	specEndOffset := specEndPos.Offset
	
	if specStartOffset < 0 || specEndOffset > len(originalFileContentString) || specStartOffset > specEndOffset {
		log.Printf("Warning: Invalid offsets for spec in %s (line %d): start=%d, end=%d, file_len=%d. Skipping spec.",
			filePath, specStartPos.Line, specStartOffset, specEndOffset, len(originalFileContentString))
		return nil
	}
	specChunkCode := originalFileContentString[specStartOffset:specEndOffset]

	finalChunkCode := transform.ApplyQualifierReplacements(specChunkCode, valueSpec, pkg.TypesInfo)

	return &types.ChromaDocument{
		ID:       fmt.Sprintf("%s:%d-%d-%s", filePath, specStartPos.Line, specEndPos.Line, entityName),
		Document: finalChunkCode,
		Metadata: specMetadata,
	}
} 