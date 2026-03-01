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

// ExtPublishSchema returns the JSON Schema for the ext_publish tool.
func ExtPublishSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Path to the extension directory",
			},
			"store": map[string]any{
				"type":        "string",
				"description": "Target store: chrome, firefox, or edge (optional, default: chrome)",
				"enum":        []any{"chrome", "firefox", "edge"},
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// ExtPublish returns a handler that validates an extension and returns store submission instructions.
func ExtPublish() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		directory := helpers.GetString(req.Arguments, "directory")
		store := helpers.GetString(req.Arguments, "store")
		if store == "" {
			store = "chrome"
		}

		manifestPath := filepath.Join(directory, "manifest.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			return helpers.ErrorResult("validation_error", fmt.Sprintf("no manifest.json found in %s", directory)), nil
		}

		var instructions string
		switch store {
		case "firefox":
			instructions = "Visit https://addons.mozilla.org/developers/, click 'Submit a New Add-on', upload your zip."
		case "edge":
			instructions = "Visit https://partner.microsoft.com/en-us/dashboard/microsoftedge/overview, click 'Create new extension', upload your zip."
		default: // chrome
			instructions = "Visit https://chrome.google.com/webstore/developer/dashboard, click 'New Item', upload your zip from ext_build."
		}

		return helpers.TextResult(fmt.Sprintf("## Publish to %s\n\n%s", store, instructions)), nil
	}
}
