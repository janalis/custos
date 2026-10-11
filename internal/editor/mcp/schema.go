package mcp

func toolSchema(name string) map[string]any {
	fields := map[string]any{
		"workspaceId":  map[string]any{"type": "string", "description": "Authorized workspace ID; may be omitted only when a default exists."},
		"checkpointId": map[string]any{"type": "string"}, "planId": map[string]any{"type": "string"},
		"paths":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": 100},
		"path":   map[string]any{"type": "string", "description": "Workspace-relative path."},
		"buffer": map[string]any{"type": "string", "description": "Supplied PHP source; never persisted and cannot authorize disk writes."},
		"query":  map[string]any{"type": "string"}, "rule": map[string]any{"type": "string"}, "severity": map[string]any{"type": "string", "enum": []string{"error", "warning", "info"}},
		"cursor": map[string]any{"type": "string"}, "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 20},
		"offset":    map[string]any{"type": "integer", "minimum": 0, "description": "Zero-based byte offset, not a character offset."},
		"action":    map[string]any{"type": "string", "enum": []string{"list", "resume", "delete"}, "default": "list"},
		"iterative": map[string]any{"type": "boolean", "default": false},
		"choices":   map[string]any{"type": "array", "maxItems": 100, "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"findingId", "fix"}, "properties": map[string]any{"findingId": map[string]any{"type": "string"}, "fix": map[string]any{"type": "integer", "minimum": 0, "description": "Zero-based index in the finding's fix choices."}}}},
	}
	selections := map[string][]string{
		"workspace": {"workspaceId", "cursor", "limit"}, "analyze": {"workspaceId", "paths", "path", "buffer"},
		"findings": {"workspaceId", "checkpointId", "rule", "severity", "path", "cursor", "limit"},
		"rules":    {"workspaceId", "query", "rule", "cursor", "limit"}, "symbols": {"workspaceId", "checkpointId", "query", "path", "cursor", "limit"},
		"type": {"workspaceId", "checkpointId", "path", "offset"}, "prepare_fixes": {"workspaceId", "checkpointId", "planId", "choices", "iterative", "path", "rule", "cursor", "limit"},
		"apply_fixes": {"workspaceId", "planId"}, "checkpoints": {"workspaceId", "checkpointId", "action", "cursor", "limit"},
	}
	required := map[string][]string{"findings": {"checkpointId"}, "symbols": {"checkpointId"}, "type": {"checkpointId", "path", "offset"}, "prepare_fixes": {"checkpointId"}, "apply_fixes": {"planId"}}
	props := map[string]any{}
	for _, field := range selections[name] {
		props[field] = fields[field]
	}
	schema := map[string]any{"type": "object", "additionalProperties": false, "properties": props}
	if fields := required[name]; len(fields) > 0 {
		schema["required"] = fields
	}
	return schema
}
