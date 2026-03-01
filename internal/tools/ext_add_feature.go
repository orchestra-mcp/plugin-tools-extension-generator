package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ExtAddFeatureSchema returns the JSON Schema for the ext_add_feature tool.
func ExtAddFeatureSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Path to the extension directory",
			},
			"feature": map[string]any{
				"type":        "string",
				"description": "Feature type to add: background-handler, popup-button, or side-panel",
				"enum":        []any{"background-handler", "popup-button", "side-panel"},
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Feature name for generated code (optional, default: MyFeature)",
			},
		},
		"required": []any{"directory", "feature"},
	})
	return s
}

// ExtAddFeature returns a handler that adds a feature to an existing Chrome extension.
func ExtAddFeature() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory", "feature"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		directory := helpers.GetString(req.Arguments, "directory")
		feature := helpers.GetString(req.Arguments, "feature")
		name := helpers.GetString(req.Arguments, "name")
		if name == "" {
			name = "MyFeature"
		}

		switch feature {
		case "background-handler":
			bgPath := filepath.Join(directory, "background.js")
			content := fmt.Sprintf("\n// %s handler\nchrome.runtime.onMessage.addListener((msg) => { if (msg.type === '%s') { /* TODO */ } });\n", name, name)
			f, err := os.OpenFile(bgPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to open background.js: %v", err)), nil
			}
			defer f.Close()
			if _, err := f.WriteString(content); err != nil {
				return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write background.js: %v", err)), nil
			}
			return helpers.TextResult(fmt.Sprintf("Added background handler to %s", bgPath)), nil

		case "popup-button":
			jsPath := filepath.Join(directory, name+".js")
			content := fmt.Sprintf("document.getElementById('%sBtn')?.addEventListener('click', () => { chrome.runtime.sendMessage({ type: '%s' }); });\n", name, name)
			if err := os.WriteFile(jsPath, []byte(content), 0644); err != nil {
				return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write %s.js: %v", name, err)), nil
			}
			return helpers.TextResult(fmt.Sprintf("Created popup button script at %s", jsPath)), nil

		case "side-panel":
			htmlPath := filepath.Join(directory, name+"-panel.html")
			content := fmt.Sprintf("<!DOCTYPE html>\n<html>\n<head><meta charset=\"utf-8\"><title>%s Panel</title></head>\n<body>\n  <h1>%s</h1>\n  <!-- TODO: side panel content -->\n</body>\n</html>\n", name, name)
			if err := os.WriteFile(htmlPath, []byte(content), 0644); err != nil {
				return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write %s-panel.html: %v", name, err)), nil
			}
			return helpers.TextResult(fmt.Sprintf("Created side panel at %s", htmlPath)), nil

		default:
			return helpers.ErrorResult("validation_error", fmt.Sprintf("invalid feature %q: must be background-handler, popup-button, or side-panel", feature)), nil
		}
	}
}
