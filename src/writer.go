package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteIndexOutput writes all four index files atomically to .xojo_index directory
func WriteIndexOutput(indexPath string, output *IndexOutput) error {
	// Create index directory if it doesn't exist
	if err := os.MkdirAll(indexPath, 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	// Write all four files with atomic rename
	if err := writeJSON(filepath.Join(indexPath, "codetree.json"), output.CodeTree); err != nil {
		return fmt.Errorf("failed to write codetree.json: %w", err)
	}
	if err := writeJSON(filepath.Join(indexPath, "manifest.json"), output.Manifest); err != nil {
		return fmt.Errorf("failed to write manifest.json: %w", err)
	}
	if err := writeJSON(filepath.Join(indexPath, "dependencies.json"), output.Dependencies); err != nil {
		return fmt.Errorf("failed to write dependencies.json: %w", err)
	}
	if err := writeJSON(filepath.Join(indexPath, "meta.json"), output.Meta); err != nil {
		return fmt.Errorf("failed to write meta.json: %w", err)
	}

	return nil
}

// writeJSON writes data as indented JSON to a file atomically (write to temp, then rename)
func writeJSON(filePath string, data interface{}) error {
	// Marshal to JSON with indentation
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to temporary file
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// WriteCodeTree writes codetree.json
func WriteCodeTree(filePath string, tree CodeTree) error {
	return writeJSON(filePath, tree)
}

// WriteManifest writes manifest.json
func WriteManifest(filePath string, manifest Manifest) error {
	return writeJSON(filePath, manifest)
}

// WriteDependencies writes dependencies.json
func WriteDependencies(filePath string, deps DependencyGraph) error {
	return writeJSON(filePath, deps)
}

// WriteMeta writes meta.json
func WriteMeta(filePath string, meta MetaFile) error {
	return writeJSON(filePath, meta)
}
