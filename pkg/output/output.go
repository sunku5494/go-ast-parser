package output

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/sunku5494/go-ast-parser/pkg/types"
)

// WriteChunksToJSON writes the code chunks to a JSON file with pretty formatting.
func WriteChunksToJSON(chunks []types.ChromaDocument, filename string) error {
	jsonData, err := json.MarshalIndent(chunks, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling chunks to JSON: %w", err)
	}

	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON to file: %w", err)
	}

	fmt.Printf("Successfully extracted %d code chunks to %s\n", len(chunks), filename)
	return nil
} 