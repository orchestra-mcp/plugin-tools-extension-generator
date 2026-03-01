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

// ExtValidateSchema returns the JSON Schema for the ext_validate tool.
func ExtValidateSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Path to the extension directory containing manifest.json",
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// ExtValidate returns a handler that validates an extension's manifest.json.
func ExtValidate() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		directory := helpers.GetString(req.Arguments, "directory")
		manifestPath := filepath.Join(directory, "manifest.json")

		raw, err := os.ReadFile(manifestPath)
		if err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to read manifest.json: %v", err)), nil
		}

		var manifest map[string]any
		if err := json.Unmarshal(raw, &manifest); err != nil {
			return helpers.ErrorResult("parse_error", fmt.Sprintf("Invalid JSON in manifest.json: %v", err)), nil
		}

		var issues []string

		// Check required fields.
		for _, field := range []string{"name", "manifest_version", "version"} {
			if _, ok := manifest[field]; !ok {
				issues = append(issues, fmt.Sprintf("missing required field: %q", field))
			}
		}

		// Check manifest_version == 3.
		if mv, ok := manifest["manifest_version"]; ok {
			switch v := mv.(type) {
			case float64:
				if int(v) != 3 {
					issues = append(issues, fmt.Sprintf("manifest_version must be 3, got %d", int(v)))
				}
			default:
				issues = append(issues, "manifest_version must be a number")
			}
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Validation: %s\n\n", directory)
		if len(issues) == 0 {
			fmt.Fprintf(&b, "Valid MV3 extension manifest.\n")
		} else {
			fmt.Fprintf(&b, "Found %d issue(s):\n\n", len(issues))
			for _, issue := range issues {
				fmt.Fprintf(&b, "- %s\n", issue)
			}
		}

		return helpers.TextResult(b.String()), nil
	}
}
