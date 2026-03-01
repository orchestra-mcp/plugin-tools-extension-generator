package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ExtAddContentScriptSchema returns the JSON Schema for the ext_add_content_script tool.
func ExtAddContentScriptSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Path to the extension directory containing manifest.json",
			},
			"matches": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "URL match patterns for the content script (e.g. [\"https://*.example.com/*\"])",
			},
			"run_at": map[string]any{
				"type":        "string",
				"description": "When to inject the script: document_start, document_end, or document_idle (optional)",
			},
		},
		"required": []any{"directory", "matches"},
	})
	return s
}

// ExtAddContentScript returns a handler that appends a content_scripts entry to manifest.json.
func ExtAddContentScript() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		if len(helpers.GetStringSlice(req.Arguments, "matches")) == 0 {
			return helpers.ErrorResult("validation_error", "missing required fields: matches"), nil
		}

		directory := helpers.GetString(req.Arguments, "directory")
		matches := helpers.GetStringSlice(req.Arguments, "matches")
		runAt := helpers.GetString(req.Arguments, "run_at")

		manifestPath := filepath.Join(directory, "manifest.json")
		raw, err := os.ReadFile(manifestPath)
		if err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to read manifest.json: %v", err)), nil
		}

		var manifest map[string]any
		if err := json.Unmarshal(raw, &manifest); err != nil {
			return helpers.ErrorResult("parse_error", fmt.Sprintf("Failed to parse manifest.json: %v", err)), nil
		}

		// Build the new content_scripts entry.
		matchesAny := make([]any, len(matches))
		for i, m := range matches {
			matchesAny[i] = m
		}
		entry := map[string]any{
			"matches": matchesAny,
			"js":      []any{"content.js"},
		}
		if runAt != "" {
			entry["run_at"] = runAt
		}

		// Append to existing content_scripts list or create it.
		existing, ok := manifest["content_scripts"].([]any)
		if !ok {
			existing = []any{}
		}
		manifest["content_scripts"] = append(existing, entry)

		updated, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to marshal manifest: %v", err)), nil
		}
		if err := os.WriteFile(manifestPath, updated, 0644); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write manifest.json: %v", err)), nil
		}

		return helpers.TextResult(fmt.Sprintf("Content script added for: %s", strings.Join(matches, ", "))), nil
	}
}
