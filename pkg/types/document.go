package types

// ChromaDocument represents a chunk to be stored
type ChromaDocument struct {
	ID       string                 `json:"id"`
	Document string                 `json:"document"`
	Metadata map[string]interface{} `json:"metadata"`
} 