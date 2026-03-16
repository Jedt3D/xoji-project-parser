package main

// buildDependencies extracts class relationships from parsed code entries
func buildDependencies(codeTree CodeTree) DependencyGraph {
	graph := DependencyGraph{
		Classes: make(map[string]ClassDeps),
		Reverse: make(map[string][]string),
	}

	// First pass: collect all classes and their inheritance/implementation
	for _, entry := range codeTree {
		var codeEntry *CodeEntry
		switch e := entry.(type) {
		case *CodeEntry:
			codeEntry = e
		default:
			continue
		}

		if codeEntry.Entity == "" {
			continue
		}

		classDeps := ClassDeps{
			Inherits:   codeEntry.Inherits,
			Implements: codeEntry.Implements,
		}

		graph.Classes[codeEntry.Entity] = classDeps

		// Build reverse map: if Foo inherits from BaseClass, add Foo to reverse[BaseClass]
		if codeEntry.Inherits != "" {
			graph.Reverse[codeEntry.Inherits] = append(graph.Reverse[codeEntry.Inherits], codeEntry.Entity)
		}
	}

	return graph
}
