package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// CmdCheck handles `xoji check` - compares file mtimes against meta.json
// Returns true if fresh (exit 0), false if stale (exit 1)
func CmdCheck(projectPath string) (bool, error) {
	// Find .xojo_project if not specified
	if projectPath == "" {
		var err error
		projectPath, err = FindXojoProject("")
		if err != nil {
			return false, err
		}
	}

	// Parse the project
	project, err := ParseXojoProject(projectPath)
	if err != nil {
		return false, err
	}

	// Try to read meta.json
	metaPath := filepath.Join(project.IndexPath, "meta.json")
	meta, err := readMeta(metaPath)
	if err != nil {
		// No meta.json means stale
		return false, nil
	}

	// Check if any source file has changed since last index
	for relPath, fileMeta := range meta.Files {
		fullPath := filepath.Join(project.RootPath, relPath)

		fi, err := os.Stat(fullPath)
		if err != nil {
			// File deleted or inaccessible = stale
			return false, nil
		}

		// Compare mtime
		currentTime := fi.ModTime().UTC().Format("2006-01-02T15:04:05Z")
		if currentTime != fileMeta.MTime {
			// File has changed = stale
			return false, nil
		}

		// Compare size as additional check
		if fi.Size() != fileMeta.Size {
			return false, nil
		}
	}

	// Check if new files have been added to the project
	if err := checkNewFiles(project, meta); err != nil {
		return false, err
	}

	// All files match = fresh
	return true, nil
}

// checkNewFiles checks if any new files have been added since last index
func checkNewFiles(project *Project, meta *MetaFile) error {
	indexedFiles := make(map[string]bool)
	for relPath := range meta.Files {
		indexedFiles[relPath] = true
	}

	// Check if any project items are missing from meta
	for _, item := range project.Items {
		if !indexedFiles[item.RelativePath] {
			// New file found
			return fmt.Errorf("new file found: %s", item.RelativePath)
		}
	}

	return nil
}
