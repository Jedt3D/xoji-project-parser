package main

// ProjectItem represents a single item in a .xojo_project file
type ProjectItem struct {
	ItemType     string // Class, Module, Interface, DesktopWindow, etc.
	DisplayName  string
	RelativePath string
	ID           string
	ParentID     string
}

// CodeEntry represents a parsed .xojo_code file
type CodeEntry struct {
	Type        string            `json:"type"`
	Entity      string            `json:"entity"`
	Inherits    string            `json:"inherits,omitempty"`
	Implements  []string          `json:"implements,omitempty"`
	Methods     map[string]int    `json:"methods"`
	Properties  map[string]int    `json:"properties"`
	Events      map[string]int    `json:"events"`
	Constants   map[string]int    `json:"constants"`
	Enums       map[string]int    `json:"enums"`
	Hooks       map[string]int    `json:"hooks"`
}

// WindowControl represents a control in a window file
type WindowControl struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// WindowEntry represents a parsed .xojo_window file
type WindowEntry struct {
	Type       string            `json:"type"`
	Entity     string            `json:"entity"`
	Controls   []WindowControl   `json:"controls"`
	Methods    map[string]int    `json:"methods"`
	Events     map[string]int    `json:"events"`
	Properties map[string]int    `json:"properties"`
}

// CodeTree is the root structure for codetree.json
type CodeTree map[string]interface{}

// ManifestEntry represents a single file in manifest.json
type ManifestEntry struct {
	Path       string `json:"path"`
	Type       string `json:"type"`
	Entity     string `json:"entity"`
	Size       int64  `json:"size"`
	ModTime    string `json:"modTime"`
	SizeBytes  int64  `json:"sizeBytes"`
}

// Manifest is the root structure for manifest.json
type Manifest []ManifestEntry

// MetaFile represents meta.json
type MetaFile struct {
	Version     string            `json:"version"`
	IndexedAt   string            `json:"indexedAt"`
	ProjectHash string            `json:"projectHash"`
	Files       map[string]FileMeta `json:"files"`
}

// FileMeta stores mtime and size for a single file
type FileMeta struct {
	MTime string `json:"mtime"`
	Size  int64  `json:"size"`
}

// DependencyGraph represents the dependency relationships
type DependencyGraph struct {
	Classes map[string]ClassDeps `json:"classes"`
	Reverse map[string][]string  `json:"reverse"`
}

// ClassDeps represents dependencies for a single class
type ClassDeps struct {
	Inherits string   `json:"inherits,omitempty"`
	Implements []string `json:"implements,omitempty"`
}

// IndexOutput holds all four index structures
type IndexOutput struct {
	CodeTree      CodeTree
	Manifest      Manifest
	Dependencies  DependencyGraph
	Meta          MetaFile
}

// Project holds all project-level data needed during indexing
type Project struct {
	RootPath  string
	Items     []ProjectItem
	IndexPath string // .xojo_index directory path
}
