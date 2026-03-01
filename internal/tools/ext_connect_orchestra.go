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

// ExtConnectOrchestraSchema returns the JSON Schema for the ext_connect_orchestra tool.
func ExtConnectOrchestraSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Path to the extension directory",
			},
			"host": map[string]any{
				"type":        "string",
				"description": "Orchestra host (optional, default: localhost)",
			},
			"port": map[string]any{
				"type":        "number",
				"description": "Orchestra port (optional, default: 4444)",
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// ExtConnectOrchestra returns a handler that injects Orchestra SDK connector code into an extension.
func ExtConnectOrchestra() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		directory := helpers.GetString(req.Arguments, "directory")
		host := helpers.GetString(req.Arguments, "host")
		if host == "" {
			host = "localhost"
		}
		port := helpers.GetInt(req.Arguments, "port")
		if port == 0 {
			port = 4444
		}

		connectorPath := filepath.Join(directory, "orchestra-connector.js")
		content := fmt.Sprintf(`// Orchestra SDK connector — auto-generated
const ORCHESTRA_HOST = '%s';
const ORCHESTRA_PORT = %d;

async function orchestraCall(tool, args) {
  const resp = await fetch(`+"`"+`http://${ORCHESTRA_HOST}:${ORCHESTRA_PORT}/mcp/tools/${tool}`+"`"+`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(args),
  });
  return resp.json();
}

window.__orchestra = { call: orchestraCall };
`, host, port)

		if err := os.WriteFile(connectorPath, []byte(content), 0644); err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to write orchestra-connector.js: %v", err)), nil
		}

		return helpers.TextResult(fmt.Sprintf("Orchestra connector written to %s", connectorPath)), nil
	}
}
