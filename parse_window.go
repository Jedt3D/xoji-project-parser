package main

import (
	"bufio"
	"os"
	"strings"
)

// parseWindowFile parses a single .xojo_window file and returns a WindowEntry
func parseWindowFile(filePath string) (*WindowEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entry := &WindowEntry{
		Controls:   []WindowControl{},
		Methods:    make(map[string]int),
		Properties: make(map[string]int),
		Events:     make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// State machine variables
	state := "INIT"
	pendingName := false
	pendingLineNum := 0
	currentControl := ""
	depth := 0
	controlMap := make(map[string]bool)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		lowered := strings.ToLower(trimmed)

		// Handle different states
		switch state {
		case "INIT":
			// Look for #tag DesktopWindow or #tag Window
			if strings.HasPrefix(lowered, "#tag desktopwindow") || strings.HasPrefix(lowered, "#tag window") {
				state = "IN_WINDOW_DEF"
			}

		case "IN_WINDOW_DEF":
			// Look for Begin DesktopWindow/Window with window name
			if (strings.HasPrefix(lowered, "begin desktopwindow ") || strings.HasPrefix(lowered, "begin window ")) && depth == 0 {
				// Extract window name
				parts := strings.Fields(trimmed)
				if len(parts) >= 3 {
					entry.Type = parts[1] // "DesktopWindow" or "Window"
					entry.Entity = parts[2]
				}
				depth = 1
			} else if (strings.HasPrefix(lowered, "begin ") || strings.HasPrefix(lowered, "begin:")) && depth > 0 {
				// Look for nested control Begin blocks (depth >= 1)
				// Extract control type and name from "Begin Type Name" or similar
				parts := strings.Fields(trimmed)
				if len(parts) >= 3 {
					controlType := parts[1]
					controlName := parts[2]
					// Only add controls at depth 1 (direct children)
					if depth == 1 {
						// Dedup by name
						if !controlMap[controlName] {
							entry.Controls = append(entry.Controls, WindowControl{
								Name: controlName,
								Type: controlType,
							})
							controlMap[controlName] = true
						}
					}
				}
				depth++
			}

			// Check for bare "End" to pop the stack
			if trimmed == "End" {
				depth--
				if depth == 0 {
					// End of window definition
					state = "AFTER_DEF"
				}
			}

			// Check for end of window tag
			if strings.HasPrefix(lowered, "#tag enddesktopwindow") || strings.HasPrefix(lowered, "#tag endwindow") {
				state = "AFTER_DEF"
			}

		case "AFTER_DEF":
			// Look for #tag WindowCode
			if strings.HasPrefix(lowered, "#tag windowcode") {
				state = "IN_WC"
			}
			// Look for #tag Events
			if strings.HasPrefix(lowered, "#tag events ") {
				parts := strings.Fields(trimmed)
				if len(parts) >= 3 {
					currentControl = parts[2]
				}
				state = "IN_CE"
			}
			// Look for #tag ViewBehavior to skip
			if strings.HasPrefix(lowered, "#tag viewbehavior") {
				state = "SKIP_VIEW"
			}

		case "IN_WC":
			// Inside WindowCode, handle method/event/property declarations same as code files
			if strings.HasPrefix(lowered, "#tag endwindowcode") {
				state = "AFTER_WC"
			} else if strings.HasPrefix(lowered, "#tag method") || strings.HasPrefix(lowered, "#tag menuhandler") {
				state = "IN_WC_METHOD"
				pendingName = true
				pendingLineNum = lineNum
			} else if strings.HasPrefix(lowered, "#tag property") || strings.HasPrefix(lowered, "#tag computedproperty") {
				state = "IN_WC_PROPERTY"
				pendingName = true
				pendingLineNum = lineNum
			} else if strings.HasPrefix(lowered, "#tag event") {
				state = "IN_WC_EVENT"
				pendingName = true
				pendingLineNum = lineNum
			}

		case "IN_WC_METHOD":
			if pendingName && trimmed != "" && !strings.HasPrefix(lowered, "#tag") {
				name := extractSubFunctionName(trimmed)
				if name != "" {
					entry.Methods[name] = pendingLineNum
				}
				pendingName = false
			}
			if strings.HasPrefix(lowered, "#tag endmethod") || strings.HasPrefix(lowered, "#tag endmenuhandler") {
				state = "IN_WC"
			}

		case "IN_WC_PROPERTY":
			if pendingName && trimmed != "" && !strings.HasPrefix(lowered, "#tag") {
				name := extractPropertyName(trimmed)
				if name != "" {
					entry.Properties[name] = pendingLineNum
				}
				pendingName = false
			}
			if strings.HasPrefix(lowered, "#tag endproperty") || strings.HasPrefix(lowered, "#tag endcomputedproperty") {
				state = "IN_WC"
			}

		case "IN_WC_EVENT":
			if pendingName && trimmed != "" && !strings.HasPrefix(lowered, "#tag") {
				name := extractEventName(trimmed)
				if name != "" {
					entry.Events[name] = pendingLineNum
				}
				pendingName = false
			}
			if strings.HasPrefix(lowered, "#tag endevent") {
				state = "IN_WC"
			}

		case "AFTER_WC":
			// Look for #tag Events ControlName
			if strings.HasPrefix(lowered, "#tag events ") {
				parts := strings.Fields(trimmed)
				if len(parts) >= 3 {
					currentControl = parts[2]
				}
				state = "IN_CE"
			}
			// Look for #tag ViewBehavior to skip
			if strings.HasPrefix(lowered, "#tag viewbehavior") {
				state = "SKIP_VIEW"
			}

		case "IN_CE":
			// Inside Events for a control
			if strings.HasPrefix(lowered, "#tag endevents") {
				state = "AFTER_WC"
			} else if strings.HasPrefix(lowered, "#tag event") {
				state = "IN_CE_EVENT"
				pendingName = true
				pendingLineNum = lineNum
			}

		case "IN_CE_EVENT":
			if pendingName && trimmed != "" && !strings.HasPrefix(lowered, "#tag") {
				name := extractEventName(trimmed)
				if name != "" {
					// Store as "ControlName.EventName"
					key := currentControl + "." + name
					entry.Events[key] = pendingLineNum
				}
				pendingName = false
			}
			if strings.HasPrefix(lowered, "#tag endevent") {
				state = "IN_CE"
			}

		case "SKIP_VIEW":
			if strings.HasPrefix(lowered, "#tag endviewbehavior") {
				state = "AFTER_WC"
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entry, nil
}
