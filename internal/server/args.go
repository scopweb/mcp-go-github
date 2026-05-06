// Package server: argument parsing helpers for JSON-RPC tool calls.
//
// JSON-RPC numeric arguments arrive as float64 (per encoding/json default).
// These helpers convert safely without panicking on missing or wrong-typed args.
package server

import (
	"fmt"
	"strconv"
)

// getIntArg extracts an int from arguments map. Accepts float64 (JSON default),
// int, int64, json.Number, or numeric string. Returns error if missing or invalid.
func getIntArg(args map[string]interface{}, key string) (int, error) {
	raw, exists := args[key]
	if !exists || raw == nil {
		return 0, fmt.Errorf("required parameter '%s' is missing", key)
	}
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case int32:
		return int(v), nil
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("parameter '%s' must be an integer, got string %q", key, v)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("parameter '%s' must be a number, got %T", key, raw)
	}
}

// getInt64Arg is the int64 variant for IDs that exceed int32 range (e.g. workflow run_id).
func getInt64Arg(args map[string]interface{}, key string) (int64, error) {
	raw, exists := args[key]
	if !exists || raw == nil {
		return 0, fmt.Errorf("required parameter '%s' is missing", key)
	}
	switch v := raw.(type) {
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case string:
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parameter '%s' must be an integer, got string %q", key, v)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("parameter '%s' must be a number, got %T", key, raw)
	}
}

// getStringArg extracts a string. Returns error if missing or non-string.
// Use this when the parameter is required; for optional strings keep the
// existing `s, _ := args[key].(string)` pattern.
func getStringArg(args map[string]interface{}, key string) (string, error) {
	raw, exists := args[key]
	if !exists || raw == nil {
		return "", fmt.Errorf("required parameter '%s' is missing", key)
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("parameter '%s' must be a string, got %T", key, raw)
	}
	if s == "" {
		return "", fmt.Errorf("parameter '%s' must not be empty", key)
	}
	return s, nil
}
