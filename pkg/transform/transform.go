package transform

import (
	"go/ast"
	"go/types"
	"sort"
	"strconv"
	"strings"
)

// ApplyQualifierReplacements inspects the given node's subtree for SelectorExprs
// and replaces package qualifiers with their full import paths in the chunkCode string.
// It uses a two-pass replacement strategy with unique placeholders to prevent cascading
// replacements where a full import path might contain another package alias.
func ApplyQualifierReplacements(chunkCode string, node ast.Node, info *types.Info) string {
	if node == nil || info == nil {
		return chunkCode
	}

	replacements := make(map[string]string)

	ast.Inspect(node, func(innerNode ast.Node) bool {
		if selExpr, ok := innerNode.(*ast.SelectorExpr); ok {
			if ident, isIdent := selExpr.X.(*ast.Ident); isIdent {
				obj := info.Uses[ident]
				if obj == nil {
					return true
				}
				if pkgName, isPkgName := obj.(*types.PkgName); isPkgName {
					fullImportPath := pkgName.Imported().Path()
					if ident.Name != fullImportPath {
						replacements[ident.Name] = fullImportPath
					}
				}
			}
		}
		return true
	})

	if len(replacements) == 0 {
		return chunkCode
	}

	tempMap := make(map[string]string)
	finalMap := make(map[string]string)

	placeholderPrefix := "__GO_QUALIFIER_TEMP_"
	i := 0
	for oldQualifier, fullPath := range replacements {
		placeholder := placeholderPrefix + strconv.Itoa(i) + "__"
		tempMap[oldQualifier] = placeholder
		finalMap[placeholder] = fullPath
		i++
	}

	var sortedOldQualifiers []string
	for q := range tempMap {
		sortedOldQualifiers = append(sortedOldQualifiers, q)
	}
	sort.Slice(sortedOldQualifiers, func(i, j int) bool {
		return len(sortedOldQualifiers[i]) > len(sortedOldQualifiers[j])
	})

	for _, oldQualifier := range sortedOldQualifiers {
		placeholder := tempMap[oldQualifier]
		chunkCode = strings.ReplaceAll(chunkCode, oldQualifier+".", placeholder+".")
	}

	var sortedPlaceholders []string
	for p := range finalMap {
		sortedPlaceholders = append(sortedPlaceholders, p)
	}
	sort.Slice(sortedPlaceholders, func(i, j int) bool {
		return len(sortedPlaceholders[i]) > len(sortedPlaceholders[j])
	})

	for _, placeholder := range sortedPlaceholders {
		fullPath := finalMap[placeholder]
		chunkCode = strings.ReplaceAll(chunkCode, placeholder+".", fullPath+".")
	}

	return chunkCode
} 