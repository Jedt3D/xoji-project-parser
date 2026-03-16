package main

import (
	"os"
	"testing"
)

func TestComputeProjectHash(t *testing.T) {
	// Create a temporary file with known content
	tmpFile, err := os.CreateTemp("", "test_project_*.xojo_project")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testContent := `Type=Desktop
RBProjectVersion=2025.031
Class=App;App.xojo_code;&h0000000001;&h0000000000;false`

	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tmpFile.Close()

	// Compute hash
	hash1, err := computeProjectHash(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}

	if hash1 == "" {
		t.Error("Hash should not be empty")
	}

	// Verify it's a valid SHA256 hash (64 hex characters)
	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}

	// Compute hash again to verify consistency
	hash2, err := computeProjectHash(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to compute hash second time: %v", err)
	}

	if hash1 != hash2 {
		t.Error("Hash should be consistent for same file")
	}
}

func TestComputeProjectHashDifferentContent(t *testing.T) {
	// Create first temporary file
	tmpFile1, err := os.CreateTemp("", "test_project1_*.xojo_project")
	if err != nil {
		t.Fatalf("Failed to create temp file 1: %v", err)
	}
	defer os.Remove(tmpFile1.Name())

	tmpFile1.WriteString("Content 1")
	tmpFile1.Close()

	// Create second temporary file with different content
	tmpFile2, err := os.CreateTemp("", "test_project2_*.xojo_project")
	if err != nil {
		t.Fatalf("Failed to create temp file 2: %v", err)
	}
	defer os.Remove(tmpFile2.Name())

	tmpFile2.WriteString("Content 2")
	tmpFile2.Close()

	// Compute hashes
	hash1, err := computeProjectHash(tmpFile1.Name())
	if err != nil {
		t.Fatalf("Failed to compute hash 1: %v", err)
	}

	hash2, err := computeProjectHash(tmpFile2.Name())
	if err != nil {
		t.Fatalf("Failed to compute hash 2: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Different files should produce different hashes")
	}
}

func TestComputeProjectHashNonexistent(t *testing.T) {
	_, err := computeProjectHash("/nonexistent/path/to/project.xojo_project")
	if err == nil {
		t.Error("Should return error for nonexistent file")
	}
}

func TestReadMetaValid(t *testing.T) {
	// Create a temporary JSON meta file
	tmpFile, err := os.CreateTemp("", "meta_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	metaContent := `{
  "version": "1.0.0",
  "indexedAt": "2026-03-17T12:00:00Z",
  "projectHash": "abc123def456",
  "files": {
    "App.xojo_code": {
      "mtime": "2026-03-17T10:00:00Z",
      "size": 1234
    },
    "MainWindow.xojo_window": {
      "mtime": "2026-03-17T11:00:00Z",
      "size": 5678
    }
  }
}`

	tmpFile.WriteString(metaContent)
	tmpFile.Close()

	// Read meta
	meta, err := readMeta(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read meta: %v", err)
	}

	if meta.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", meta.Version)
	}

	if meta.ProjectHash != "abc123def456" {
		t.Errorf("Expected projectHash abc123def456, got %s", meta.ProjectHash)
	}

	if len(meta.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(meta.Files))
	}

	appFile, ok := meta.Files["App.xojo_code"]
	if !ok {
		t.Error("Expected App.xojo_code in files")
	}

	if appFile.Size != 1234 {
		t.Errorf("Expected size 1234 for App.xojo_code, got %d", appFile.Size)
	}
}

func TestReadMetaInvalid(t *testing.T) {
	// Create a temporary file with invalid JSON
	tmpFile, err := os.CreateTemp("", "meta_invalid_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("{ invalid json")
	tmpFile.Close()

	_, err = readMeta(tmpFile.Name())
	if err == nil {
		t.Error("Should return error for invalid JSON")
	}
}

func TestReadMetaNonexistent(t *testing.T) {
	_, err := readMeta("/nonexistent/path/to/meta.json")
	if err == nil {
		t.Error("Should return error for nonexistent file")
	}
}

func TestMetaFileStructure(t *testing.T) {
	meta := MetaFile{
		Version:     "1.0.0",
		IndexedAt:   "2026-03-17T12:00:00Z",
		ProjectHash: "test_hash",
		Files: map[string]FileMeta{
			"file1.xojo_code": {
				MTime: "2026-03-17T10:00:00Z",
				Size:  1000,
			},
		},
	}

	if meta.Version != "1.0.0" {
		t.Error("Meta version should be 1.0.0")
	}

	if len(meta.Files) != 1 {
		t.Error("Meta should have 1 file entry")
	}

	fileMeta := meta.Files["file1.xojo_code"]
	if fileMeta.Size != 1000 {
		t.Error("File size should be 1000")
	}
}
