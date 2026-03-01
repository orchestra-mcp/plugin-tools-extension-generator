package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ExtScaffoldSchema returns the JSON Schema for the ext_scaffold tool.
func ExtScaffoldSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Extension name",
			},
			"directory": map[string]any{
				"type":        "string",
				"description": "Parent directory where the extension folder will be created",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Extension description (optional)",
			},
			"permissions": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "List of Chrome permissions to request (optional)",
			},
		},
		"required": []any{"name", "directory"},
	})
	return s
}

// ExtScaffold returns a handler that creates a Chrome MV3 extension scaffold.
func ExtScaffold() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name", "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")
		directory := helpers.GetString(req.Arguments, "directory")
		description := helpers.GetString(req.Arguments, "description")
		permissions := helpers.GetStringSlice(req.Arguments, "permissions")

		extDir := filepath.Join(directory, name)

		// Create the extension directory.
		if err := os.MkdirAll(extDir, 0755); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to create directory: %v", err)), nil
		}

		// Create icons/ subdirectory.
		if err := os.MkdirAll(filepath.Join(extDir, "icons"), 0755); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to create icons directory: %v", err)), nil
		}

		// Build manifest.json.
		permsAny := make([]any, len(permissions))
		for i, p := range permissions {
			permsAny[i] = p
		}
		manifest := map[string]any{
			"name":             name,
			"description":      description,
			"version":          "1.0.0",
			"manifest_version": 3,
			"permissions":      permsAny,
		}
		manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to marshal manifest: %v", err)), nil
		}
		if err := os.WriteFile(filepath.Join(extDir, "manifest.json"), manifestBytes, 0644); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write manifest.json: %v", err)), nil
		}

		// background.js
		bgJS := "chrome.runtime.onInstalled.addListener(() => { console.log('Extension installed'); });\n"
		if err := os.WriteFile(filepath.Join(extDir, "background.js"), []byte(bgJS), 0644); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write background.js: %v", err)), nil
		}

		// popup.html
		popupHTML := fmt.Sprintf("<!DOCTYPE html>\n<html>\n<head><meta charset=\"utf-8\"><title>%s</title></head>\n<body>\n  <h1>%s</h1>\n  <script src=\"popup.js\"></script>\n</body>\n</html>\n", name, name)
		if err := os.WriteFile(filepath.Join(extDir, "popup.html"), []byte(popupHTML), 0644); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write popup.html: %v", err)), nil
		}

		// popup.js
		popupJS := "document.addEventListener('DOMContentLoaded', () => { console.log('Popup ready'); });\n"
		if err := os.WriteFile(filepath.Join(extDir, "popup.js"), []byte(popupJS), 0644); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write popup.js: %v", err)), nil
		}

		// content.js
		contentJS := "console.log('Content script loaded');\n"
		if err := os.WriteFile(filepath.Join(extDir, "content.js"), []byte(contentJS), 0644); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write content.js: %v", err)), nil
		}

		return helpers.TextResult(fmt.Sprintf("Extension scaffolded at %s/%s", directory, name)), nil
	}
}
