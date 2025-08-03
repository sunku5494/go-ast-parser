package analyzer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"sort"
	"strings"
)

// GetTypeString analyzes and returns type information from an AST expression.
// This function prioritizes using types.Info for accurate type names.
func GetTypeString(expr ast.Expr, info *types.Info) string {
	if tv := info.TypeOf(expr); tv != nil {
		return tv.String()
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + GetTypeString(t.X, info)
	case *ast.ArrayType:
		return "[]" + GetTypeString(t.Elt, info)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", GetTypeString(t.Key, info), GetTypeString(t.Value, info))
	case *ast.SelectorExpr:
		if ident, isIdent := t.X.(*ast.Ident); isIdent {
			if obj := info.Uses[ident]; obj != nil {
				if pkgName, isPkgName := obj.(*types.PkgName); isPkgName {
					return pkgName.Imported().Path() + "." + t.Sel.Name
				}
			}
		}
		return fmt.Sprintf("%s.%s", GetTypeString(t.X, info), t.Sel.Name)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		dir := ""
		switch t.Dir {
		case ast.SEND:
			dir = "chan<- "
		case ast.RECV:
			dir = "<-chan "
		default:
			dir = "chan "
		}
		return dir + GetTypeString(t.Value, info)
	case *ast.Ellipsis:
		return "..." + GetTypeString(t.Elt, info)
	case *ast.FuncType:
		return "func" + GetSignature(t, info)
	default:
		tmpFset := token.NewFileSet()
		var b bytes.Buffer
		if err := printer.Fprint(&b, tmpFset, expr); err == nil {
			return b.String()
		}
		return fmt.Sprintf("%T", expr)
	}
}

// GetSignature extracts function signature information from a function type.
func GetSignature(ft *ast.FuncType, info *types.Info) string {
	var params []string
	if ft.Params != nil {
		for _, field := range ft.Params.List {
			typeStr := GetTypeString(field.Type, info)
			if len(field.Names) == 0 {
				params = append(params, typeStr)
			} else {
				for _, name := range field.Names {
					params = append(params, name.Name+" "+typeStr)
				}
			}
		}
	}
	paramStr := "(" + strings.Join(params, ", ") + ")"

	var results []string
	if ft.Results != nil {
		for _, field := range ft.Results.List {
			typeStr := GetTypeString(field.Type, info)
			if len(field.Names) == 0 {
				results = append(results, typeStr)
			} else {
				for _, name := range field.Names {
					results = append(results, name.Name+" "+typeStr)
				}
			}
		}
	}
	resultStr := ""
	if len(results) > 0 {
		if len(results) == 1 && ft.Results.List[0].Names == nil {
			resultStr = " " + results[0]
		} else {
			resultStr = " (" + strings.Join(results, ", ") + ")"
		}
	}

	return paramStr + resultStr
}

// ExtractAccessedSymbols inspects a given AST node's subtree and collects all
// fully qualified symbol paths for imported packages.
func ExtractAccessedSymbols(node ast.Node, info *types.Info) []string {
	if node == nil || info == nil {
		return nil
	}
	
	// Using a map to automatically handle duplicate symbols
	accessed := make(map[string]bool)

	ast.Inspect(node, func(innerNode ast.Node) bool {
		if selExpr, ok := innerNode.(*ast.SelectorExpr); ok {
			// A SelectorExpr is of the form `X.Sel`
			// We check if `X` is an identifier (a package alias)
			if ident, isIdent := selExpr.X.(*ast.Ident); isIdent {
				// We use the TypesInfo to get the object that this identifier refers to.
				obj := info.Uses[ident]
				if obj == nil {
					return true // Continue inspection
				}
				
				// If the object is a package name (an import alias)
				if pkgName, isPkgName := obj.(*types.PkgName); isPkgName {
					// Get the full import path and the symbol name
					fullImportPath := pkgName.Imported().Path()
					symbolName := selExpr.Sel.Name
					
					// Construct the fully qualified symbol path
					fullyQualifiedSymbol := fullImportPath + "." + symbolName
					accessed[fullyQualifiedSymbol] = true
				}
			}
		}
		return true // Continue inspecting the child nodes
	})

	// Convert the map keys to a sorted slice for consistent output
	var result []string
	for symbol := range accessed {
		result = append(result, symbol)
	}
	sort.Strings(result)

	return result
} 