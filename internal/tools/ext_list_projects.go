package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ExtListProjectsSchema returns the JSON Schema for the ext_list_projects tool.
func ExtListProjectsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Directory to scan for extension projects (defaults to current directory)",
			},
		},
	})
	return s
}

// ExtListProjects returns a handler that scans a directory for Chrome extension projects.
func ExtListProjects() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		scanDir := helpers.GetString(req.Arguments, "directory")
		if scanDir == "" {
			var err error
			scanDir, err = os.Getwd()
			if err != nil {
				return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to get working directory: %v", err)), nil
			}
		}

		entries, err := os.ReadDir(scanDir)
		if err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to read directory: %v", err)), nil
		}

		var projects []string
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			manifestPath := filepath.Join(scanDir, entry.Name(), "manifest.json")
			if _, err := os.Stat(manifestPath); err == nil {
				projects = append(projects, entry.Name())
			}
		}

		if len(projects) == 0 {
			return helpers.TextResult(fmt.Sprintf("No extension projects found in %s", scanDir)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Extension Projects in %s (%d)\n\n", scanDir, len(projects))
		for _, p := range projects {
			fmt.Fprintf(&b, "- %s\n", p)
		}

		return helpers.TextResult(b.String()), nil
	}
}
