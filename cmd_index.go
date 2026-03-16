package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CmdIndex handles `xoji index [--file path] [--project path]`
func CmdIndex(projectPath, filePath string) error {
	// Find .xojo_project if not specified
	if projectPath == "" {
		var err error
		projectPath, err = FindXojoProject("")
		if err != nil {
			return err
		}
	} else {
		// Check if projectPath is a directory or a file
		fi, err := os.Stat(projectPath)
		if err != nil {
			return fmt.Errorf("invalid project path: %w", err)
		}

		// If it's a directory, search for .xojo_project in it
		if fi.IsDir() {
			var err error
			projectPath, err = FindXojoProject(projectPath)
			if err != nil {
				return err
			}
		}
	}

	// Parse the project
	project, err := ParseXojoProject(projectPath)
	if err != nil {
		return err
	}

	var output *IndexOutput

	if filePath != "" {
		// Incremental update: load existing indexes
		metaPath := filepath.Join(project.IndexPath, "meta.json")
		treePath := filepath.Join(project.IndexPath, "codetree.json")
		manifestPath := filepath.Join(project.IndexPath, "manifest.json")
		depsPath := filepath.Join(project.IndexPath, "dependencies.json")

		// Try to load existing indexes
		var existing *IndexOutput
		metaBytes, err := os.ReadFile(metaPath)
		if err == nil {
			var meta MetaFile
			if err := json.Unmarshal(metaBytes, &meta); err == nil {
				treeBytes, _ := os.ReadFile(treePath)
				manifestBytes, _ := os.ReadFile(manifestPath)
				depsBytes, _ := os.ReadFile(depsPath)

				var tree CodeTree
				var manifest Manifest
				var deps DependencyGraph

				json.Unmarshal(treeBytes, &tree)
				json.Unmarshal(manifestBytes, &manifest)
				json.Unmarshal(depsBytes, &deps)

				existing = &IndexOutput{
					CodeTree:     tree,
					Manifest:     manifest,
					Dependencies: deps,
					Meta:         meta,
				}
			}
		}

		// Build incremental index
		output, err = BuildIncrementalIndex(project, filePath, existing)
		if err != nil {
			return err
		}
	} else {
		// Full index build
		output, err = BuildFullIndex(project)
		if err != nil {
			return err
		}
	}

	// Write indexes to disk
	if err := WriteIndexOutput(project.IndexPath, output); err != nil {
		return err
	}

	fmt.Printf("Indexes written to %s\n", project.IndexPath)
	return nil
}
