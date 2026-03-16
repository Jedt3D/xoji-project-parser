package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ReadMeta reads and parses meta.json
func readMeta(filePath string) (*MetaFile, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var meta MetaFile
	if err := json.Unmarshal(bytes, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse meta.json: %w", err)
	}

	return &meta, nil
}

// computeProjectHash computes SHA256 hash of .xojo_project file
func computeProjectHash(projectFilePath string) (string, error) {
	file, err := os.Open(projectFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
