package loader

import (
	"go/token"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// LoadGoProject loads packages from both the main module and vendor directory.
// It returns a slice of unique packages and handles deduplication.
func LoadGoProject(projectPath string) ([]*packages.Package, error) {
	fset := token.NewFileSet()

	// Check if vendor directory exists
	vendorDirPath := filepath.Join(projectPath, "vendor")
	if _, err := os.Stat(vendorDirPath); os.IsNotExist(err) {
		log.Printf("Warning: Vendor directory does NOT exist at %s. Please run 'go mod vendor' in your project root.", vendorDirPath)
	} else if err != nil {
		log.Printf("Error checking vendor directory at %s: %v", vendorDirPath, err)
	} else {
		log.Printf("Vendor directory EXISTS at %s.", vendorDirPath)
	}

	// List to hold all packages loaded from both main module and vendor
	var allPkgs []*packages.Package
	loadedPkgIDs := make(map[string]bool) // To deduplicate packages by ID

	// Step 1: Load packages from the main module
	log.Printf("Loading packages from main module (%s)...", projectPath)
	mainModuleCfg := CreatePackageConfig(projectPath, fset)
	mainPkgs, err := packages.Load(mainModuleCfg, "./...")
	if err != nil {
		log.Printf("Warning: packages.Load for main module returned an error: %v. Attempting to process available packages.", err)
	}
	log.Printf("Finished loading %d packages from main module.", len(mainPkgs))

	for _, pkg := range mainPkgs {
		if _, ok := loadedPkgIDs[pkg.ID]; !ok {
			allPkgs = append(allPkgs, pkg)
			loadedPkgIDs[pkg.ID] = true
		}
	}

	// Step 2: Load packages directly from the vendor directory
	// This ensures all vendored packages are included, even if not directly
	// referenced by the main module's go.mod (e.g., if it's a transitive dependency
	// that packages.Load didn't fully resolve in the first pass).
	log.Printf("Loading packages directly from vendor directory (%s)...", vendorDirPath)
	vendorCfg := CreatePackageConfig(vendorDirPath, fset)
	vendorPkgs, err := packages.Load(vendorCfg, "./...")
	if err != nil {
		log.Printf("Warning: packages.Load for vendor directory returned an error: %v. Attempting to process available packages.", err)
	}
	log.Printf("Finished loading %d packages from vendor directory.", len(vendorPkgs))

	for _, pkg := range vendorPkgs {
		if _, ok := loadedPkgIDs[pkg.ID]; !ok {
			allPkgs = append(allPkgs, pkg)
			loadedPkgIDs[pkg.ID] = true
		}
	}

	log.Printf("Total unique packages loaded: %d", len(allPkgs))

	// Diagnostic logging of loaded packages
	logLoadedPackages(allPkgs)

	return allPkgs, nil
}

// CreatePackageConfig creates a standardized package configuration for loading.
func CreatePackageConfig(workDir string, fset *token.FileSet) *packages.Config {
	return &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Fset:  fset,
		Dir:   workDir,
		Tests: false,
		// Logf:  log.Printf, // Uncomment for verbose go/packages logging
	}
}

// logLoadedPackages provides diagnostic output about loaded packages.
func logLoadedPackages(allPkgs []*packages.Package) {
	log.Println("--- Listing ALL Loaded Packages (Main Module + Vendor) ---")
	for _, pkg := range allPkgs {
		log.Printf("Package ID: %s", pkg.ID)
		if len(pkg.GoFiles) > 0 {
			numFilesToPrint := 3 // Print up to 3 files for brevity
			if len(pkg.GoFiles) < numFilesToPrint {
				numFilesToPrint = len(pkg.GoFiles)
			}
			for i := 0; i < numFilesToPrint; i++ {
				log.Printf("  File: %s", pkg.GoFiles[i])
			}
			if len(pkg.GoFiles) > numFilesToPrint {
				log.Printf("  ...and %d more files.", len(pkg.GoFiles)-numFilesToPrint)
			}
		} else {
			log.Println("  No Go files found for this package.")
		}
	}
	log.Println("-----------------------------------------------------")
} 