package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// parseCodeFile parses a single .xojo_code file and returns a CodeEntry
func parseCodeFile(filePath string) (*CodeEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entry := &CodeEntry{
		Methods:    make(map[string]int),
		Properties: make(map[string]int),
		Events:     make(map[string]int),
		Constants:  make(map[string]int),
		Enums:      make(map[string]int),
		Hooks:      make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// State machine variables
	state := "INIT"
	pendingName := false
	pendingLineNum := 0
	implements := []string{}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Case-insensitive #tag matching
		lowered := strings.ToLower(trimmed)

		// Handle different states
		switch state {
		case "INIT":
			// Look for #tag Class/Module/Interface
			if strings.HasPrefix(lowered, "#tag class") {
				entry.Type = "Class"
				state = "IN_ENTITY"
				pendingName = true
				pendingLineNum = lineNum
			} else if strings.HasPrefix(lowered, "#tag module") {
				entry.Type = "Module"
				state = "IN_ENTITY"
				pendingName = true
				pendingLineNum = lineNum
			} else if strings.HasPrefix(lowered, "#tag interface") {
				entry.Type = "Interface"
				state = "IN_ENTITY"
				pendingName = true
				pendingLineNum = lineNum
			}

		case "IN_ENTITY":
			// Extract entity name from declaration line
			if pendingName && trimmed != "" {
				entry.Entity = extractEntityName(trimmed)
				state = "IN_CLASS"
				pendingName = false
			}

		case "IN_CLASS":
			if strings.HasPrefix(lowered, "#tag endclass") || strings.HasPrefix(lowered, "#tag endmodule") || strings.HasPrefix(lowered, "#tag endinterface") {
				state = "DONE"
				break
			}

			// Skip ViewBehavior, Note, CompatibilityFlags
			if strings.HasPrefix(lowered, "#tag viewbehavior") {
				state = "SKIP_VIEW"
			} else if strings.HasPrefix(lowered, "#tag compatibilityflags") {
				// Single line, just skip
			} else if strings.HasPrefix(lowered, "#tag instance") && strings.Contains(lowered, " instance") {
				// Handle uppercase T variant: #tag Instance — single line, skip
			} else if strings.HasPrefix(lowered, "#tag note") {
				state = "SKIP_NOTE"
			} else if strings.HasPrefix(lowered, "#tag method") || strings.HasPrefix(lowered, "#tag menuhandler") {
				state = "IN_METHOD"
				pendingName = true
				pendingLineNum = lineNum
			} else if strings.HasPrefix(lowered, "#tag property") || strings.HasPrefix(lowered, "#tag computedproperty") {
				state = "IN_PROPERTY"
				pendingName = true
				pendingLineNum = lineNum
			} else if strings.HasPrefix(lowered, "#tag event") {
				state = "IN_EVENT"
				pendingName = true
				pendingLineNum = lineNum
			} else if strings.HasPrefix(lowered, "#tag constant") {
				// Extract Name= from same line
				if name := extractAttrValue(trimmed, "Name"); name != "" {
					entry.Constants[name] = lineNum
				}
				state = "IN_CONSTANT"
			} else if strings.HasPrefix(lowered, "#tag enum") {
				// Extract Name= from same line
				if name := extractAttrValue(trimmed, "Name"); name != "" {
					entry.Enums[name] = lineNum
				}
				state = "IN_ENUM"
			} else if strings.HasPrefix(lowered, "#tag hook") {
				state = "IN_HOOK"
				pendingName = true
				pendingLineNum = lineNum
			} else if strings.HasPrefix(lowered, "inherits ") {
				entry.Inherits = extractLastWord(trimmed)
			} else if strings.HasPrefix(lowered, "implements ") {
				impl := extractLastWord(trimmed)
				if impl != "" {
					implements = append(implements, impl)
					entry.Implements = implements
				}
			}

		case "IN_METHOD":
			if pendingName && trimmed != "" && !strings.HasPrefix(lowered, "#tag") {
				name := extractSubFunctionName(trimmed)
				if name != "" {
					entry.Methods[name] = pendingLineNum
				}
				pendingName = false
			}
			if strings.HasPrefix(lowered, "#tag endmethod") || strings.HasPrefix(lowered, "#tag endmenuhandler") {
				state = "IN_CLASS"
			}

		case "IN_PROPERTY":
			if pendingName && trimmed != "" && !strings.HasPrefix(lowered, "#tag") {
				name := extractPropertyName(trimmed)
				if name != "" {
					entry.Properties[name] = pendingLineNum
				}
				pendingName = false
			}
			if strings.HasPrefix(lowered, "#tag endproperty") || strings.HasPrefix(lowered, "#tag endcomputedproperty") {
				state = "IN_CLASS"
			}

		case "IN_EVENT":
			if pendingName && trimmed != "" && !strings.HasPrefix(lowered, "#tag") {
				name := extractEventName(trimmed)
				if name != "" {
					entry.Events[name] = pendingLineNum
				}
				pendingName = false
			}
			if strings.HasPrefix(lowered, "#tag endevent") {
				state = "IN_CLASS"
			}

		case "IN_CONSTANT":
			if strings.HasPrefix(lowered, "#tag endconstant") {
				state = "IN_CLASS"
			}

		case "IN_ENUM":
			if strings.HasPrefix(lowered, "#tag endenum") {
				state = "IN_CLASS"
			}

		case "IN_HOOK":
			if pendingName && trimmed != "" && !strings.HasPrefix(lowered, "#tag") {
				// Strip "Event " prefix if present
				name := strings.TrimPrefix(trimmed, "Event ")
				name = strings.TrimPrefix(name, "event ")
				name = extractEventName(name)
				if name != "" {
					entry.Hooks[name] = pendingLineNum
				}
				pendingName = false
			}
			if strings.HasPrefix(lowered, "#tag endhook") {
				state = "IN_CLASS"
			}

		case "SKIP_VIEW":
			if strings.HasPrefix(lowered, "#tag endviewbehavior") {
				state = "IN_CLASS"
			}

		case "SKIP_NOTE":
			if strings.HasPrefix(lowered, "#tag endnote") {
				state = "IN_CLASS"
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entry, nil
}

// Helper functions for name extraction

func extractEntityName(line string) string {
	// Line format: "Protected Class Foo" or "Class Bar"
	// Extract last word (the class/module name)
	words := strings.Fields(line)
	if len(words) > 0 {
		return words[len(words)-1]
	}
	return ""
}

func extractSubFunctionName(line string) string {
	// Line format: "Sub Foo(...)" or "Function Bar(...)" or with modifiers
	line = strings.TrimSpace(line)

	// Skip comments
	if strings.HasPrefix(line, "//") {
		return ""
	}

	// Remove modifiers
	for _, mod := range []string{"Private ", "Protected ", "Static ", "Override "} {
		line = strings.TrimPrefix(line, mod)
	}

	// Find Sub or Function keyword
	subIdx := strings.Index(strings.ToLower(line), "sub ")
	funcIdx := strings.Index(strings.ToLower(line), "function ")

	var start int
	if subIdx >= 0 {
		start = subIdx + 4
	} else if funcIdx >= 0 {
		start = funcIdx + 9
	} else {
		return ""
	}

	// Extract from keyword to opening paren or space
	remainder := line[start:]
	endIdx := strings.IndexAny(remainder, "( ")
	if endIdx >= 0 {
		return strings.TrimSpace(remainder[:endIdx])
	}
	return strings.TrimSpace(remainder)
}

func extractPropertyName(line string) string {
	// First word is the property name
	words := strings.Fields(line)
	if len(words) > 0 {
		return words[0]
	}
	return ""
}

func extractEventName(line string) string {
	// Similar to Sub/Function extraction
	return extractSubFunctionName(line)
}

func extractAttrValue(line, attrName string) string {
	// Find "Name=VALUE" or "Name = VALUE"
	pattern := regexp.MustCompile(`(?i)` + attrName + `\s*=\s*([^,\s]+)`)
	matches := pattern.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractLastWord(line string) string {
	words := strings.Fields(line)
	if len(words) > 0 {
		return words[len(words)-1]
	}
	return ""
}
