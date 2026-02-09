package server

// Annotation helpers for MCP tools
// Spec: https://modelcontextprotocol.io/specification/2025-11-25/server/tools#tool-annotations

// ReadOnly annotation for tools that don't modify state
func ReadOnlyAnnotation() map[string]interface{} {
	return map[string]interface{}{
		"readOnlyHint":    true,
		"destructiveHint": false,
		"idempotentHint":  true,
	}
}

// Idempotent annotation for tools that can be called multiple times safely
func IdempotentAnnotation() map[string]interface{} {
	return map[string]interface{}{
		"readOnlyHint":    false,
		"destructiveHint": false,
		"idempotentHint":  true,
	}
}

// Destructive annotation for tools that cause irreversible changes
func DestructiveAnnotation() map[string]interface{} {
	return map[string]interface{}{
		"readOnlyHint":    false,
		"destructiveHint": true,
		"idempotentHint":  false,
	}
}

// Modifying annotation for tools that modify state but are reversible
func ModifyingAnnotation() map[string]interface{} {
	return map[string]interface{}{
		"readOnlyHint":    false,
		"destructiveHint": false,
		"idempotentHint":  false,
	}
}

// OpenWorld annotation for tools that interact with external entities
func OpenWorldAnnotation() map[string]interface{} {
	return map[string]interface{}{
		"openWorldHint": true,
	}
}

// CombineAnnotations merges multiple annotation sets
func CombineAnnotations(annotations ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, ann := range annotations {
		for k, v := range ann {
			result[k] = v
		}
	}
	return result
}
