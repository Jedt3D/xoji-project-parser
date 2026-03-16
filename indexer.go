package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BuildFullIndex orchestrates building all three indexes (codetree, manifest, dependencies)
func BuildFullIndex(project *Project) (*IndexOutput, error) {
	codeTree := make(CodeTree)
	manifest := Manifest{}
	meta := &MetaFile{
		Version:   "0.1.0",
		IndexedAt: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Files:     make(map[string]FileMeta),
	}

	// Compute project hash
	projectPath := filepath.Join(project.RootPath, filepath.Base(project.RootPath)) + ".xojo_project"
	hash, err := computeProjectHash(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute project hash: %w", err)
	}
	meta.ProjectHash = hash

	// Parse each project item
	for _, item := range project.Items {
		filePath := filepath.Join(project.RootPath, item.RelativePath)

		// Stat the file
		fi, err := os.Stat(filePath)
		if err != nil {
			continue // Skip missing files
		}

		// Store file metadata for freshness checking
		meta.Files[item.RelativePath] = FileMeta{
			MTime: fi.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
			Size:  fi.Size(),
		}

		// Parse the file based on its type
		var entry interface{}
		switch {
		case item.ItemType == "Class" || item.ItemType == "Module" || item.ItemType == "Interface":
			if filepath.Ext(filePath) == ".xojo_code" {
				codeEntry, err := parseCodeFile(filePath)
				if err != nil {
					continue
				}
				entry = codeEntry
			}
		case item.ItemType == "DesktopWindow", item.ItemType == "MenuBar", item.ItemType == "DesktopToolbar":
			if filepath.Ext(filePath) == ".xojo_window" {
				windowEntry, err := parseWindowFile(filePath)
				if err != nil {
					continue
				}
				entry = windowEntry
			}
		}

		// Add to codetree if we got an entry
		if entry != nil {
			codeTree[item.RelativePath] = entry

			// Add to manifest
			manifest = append(manifest, ManifestEntry{
				Path:      item.RelativePath,
				Type:      item.ItemType,
				Entity:    getEntityName(entry),
				Size:      fi.Size(),
				ModTime:   fi.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
				SizeBytes: fi.Size(),
			})
		}
	}

	// Build dependency graph
	deps := buildDependencies(codeTree)

	return &IndexOutput{
		CodeTree:     codeTree,
		Manifest:     manifest,
		Dependencies: deps,
		Meta:         *meta,
	}, nil
}

// BuildIncrementalIndex re-indexes a single file and updates existing indexes
func BuildIncrementalIndex(project *Project, filePath string, existing *IndexOutput) (*IndexOutput, error) {
	// Re-read all existing indexes into memory (they're small)
	output := existing
	if output == nil {
		return nil, fmt.Errorf("no existing index to update")
	}

	// Find the project item matching this file
	var item *ProjectItem
	for i := range project.Items {
		if project.Items[i].RelativePath == filePath {
			item = &project.Items[i]
			break
		}
	}

	if item == nil {
		return nil, fmt.Errorf("file not in project: %s", filePath)
	}

	fullPath := filepath.Join(project.RootPath, filePath)
	fi, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	// Parse the updated file
	var entry interface{}
	switch {
	case item.ItemType == "Class" || item.ItemType == "Module" || item.ItemType == "Interface":
		if filepath.Ext(fullPath) == ".xojo_code" {
			codeEntry, err := parseCodeFile(fullPath)
			if err != nil {
				return nil, err
			}
			entry = codeEntry
		}
	case item.ItemType == "DesktopWindow", item.ItemType == "MenuBar", item.ItemType == "DesktopToolbar":
		if filepath.Ext(fullPath) == ".xojo_window" {
			windowEntry, err := parseWindowFile(fullPath)
			if err != nil {
				return nil, err
			}
			entry = windowEntry
		}
	}

	if entry != nil {
		// Update codetree
		output.CodeTree[filePath] = entry

		// Update manifest
		found := false
		for i, m := range output.Manifest {
			if m.Path == filePath {
				output.Manifest[i] = ManifestEntry{
					Path:      filePath,
					Type:      item.ItemType,
					Entity:    getEntityName(entry),
					Size:      fi.Size(),
					ModTime:   fi.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
					SizeBytes: fi.Size(),
				}
				found = true
				break
			}
		}
		if !found {
			output.Manifest = append(output.Manifest, ManifestEntry{
				Path:      filePath,
				Type:      item.ItemType,
				Entity:    getEntityName(entry),
				Size:      fi.Size(),
				ModTime:   fi.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
				SizeBytes: fi.Size(),
			})
		}

		// Update metadata
		output.Meta.IndexedAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		output.Meta.Files[filePath] = FileMeta{
			MTime: fi.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
			Size:  fi.Size(),
		}

		// Rebuild dependency graph
		output.Dependencies = buildDependencies(output.CodeTree)
	}

	return output, nil
}

// Helper function to extract entity name from parsed entry
func getEntityName(entry interface{}) string {
	switch e := entry.(type) {
	case *CodeEntry:
		return e.Entity
	case *WindowEntry:
		return e.Entity
	}
	return ""
}
