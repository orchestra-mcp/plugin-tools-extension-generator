package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// callTool invokes a tool handler with the given argument map.
func callTool(t *testing.T, handler func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error), args map[string]any) *pluginv1.ToolResponse {
	t.Helper()
	s, err := structpb.NewStruct(args)
	if err != nil {
		t.Fatalf("callTool: build args: %v", err)
	}
	resp, err := handler(context.Background(), &pluginv1.ToolRequest{Arguments: s})
	if err != nil {
		t.Fatalf("callTool: unexpected error: %v", err)
	}
	return resp
}

// isError returns true when the response is not successful.
func isError(resp *pluginv1.ToolResponse) bool {
	return !resp.Success
}

// errorCode returns the error code from a response.
func errorCode(resp *pluginv1.ToolResponse) string {
	return resp.GetErrorCode()
}

// makeExtDir creates a temporary directory with a minimal manifest.json.
func makeExtDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	manifest := `{"name":"Test","manifest_version":3,"version":"1.0.0","permissions":[]}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0644); err != nil {
		t.Fatalf("makeExtDir: write manifest.json: %v", err)
	}
	return dir
}

// ---- ext_scaffold ----

func TestExtScaffold_MissingName(t *testing.T) {
	resp := callTool(t, ExtScaffold(), map[string]any{
		"directory": t.TempDir(),
	})
	if !isError(resp) {
		t.Fatal("expected error, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
}

func TestExtScaffold_Valid(t *testing.T) {
	parent := t.TempDir()
	resp := callTool(t, ExtScaffold(), map[string]any{
		"name":      "my-ext",
		"directory": parent,
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
	text := resp.Result.Fields["text"].GetStringValue()
	if !strings.Contains(text, "my-ext") {
		t.Fatalf("expected response to contain extension path, got: %s", text)
	}
	if _, err := os.Stat(filepath.Join(parent, "my-ext", "manifest.json")); err != nil {
		t.Fatalf("manifest.json not created: %v", err)
	}
}

// ---- ext_add_content_script ----

func TestExtAddContentScript_MissingArgs(t *testing.T) {
	resp := callTool(t, ExtAddContentScript(), map[string]any{
		"directory": t.TempDir(),
		// missing matches
	})
	if !isError(resp) {
		t.Fatal("expected error, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
}

func TestExtAddContentScript_Valid(t *testing.T) {
	dir := makeExtDir(t)
	resp := callTool(t, ExtAddContentScript(), map[string]any{
		"directory": dir,
		"matches":   []any{"https://*.example.com/*"},
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
}

// ---- ext_validate ----

func TestExtValidate_MissingDirectory(t *testing.T) {
	resp := callTool(t, ExtValidate(), map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
}

func TestExtValidate_Valid(t *testing.T) {
	dir := makeExtDir(t)
	resp := callTool(t, ExtValidate(), map[string]any{
		"directory": dir,
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
}

// ---- ext_build ----

func TestExtBuild_MissingDirectory(t *testing.T) {
	resp := callTool(t, ExtBuild(), map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
}

func TestExtBuild_Valid(t *testing.T) {
	dir := makeExtDir(t)
	resp := callTool(t, ExtBuild(), map[string]any{
		"directory": dir,
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
	zipPath := dir + ".zip"
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("zip file not created at %s: %v", zipPath, err)
	}
}

// ---- ext_list_projects ----

func TestExtListProjects_MissingDirectory(t *testing.T) {
	resp := callTool(t, ExtListProjects(), map[string]any{
		"directory": "/nonexistent/path/that/does/not/exist",
	})
	if !isError(resp) {
		t.Fatal("expected error for nonexistent directory, got success")
	}
}

func TestExtListProjects_Empty(t *testing.T) {
	dir := t.TempDir()
	resp := callTool(t, ExtListProjects(), map[string]any{
		"directory": dir,
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
}

// ---- ext_add_feature ----

func TestExtAddFeature_MissingArgs(t *testing.T) {
	resp := callTool(t, ExtAddFeature(), map[string]any{
		// missing directory
		"feature": "background-handler",
	})
	if !isError(resp) {
		t.Fatal("expected error, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
}

func TestExtAddFeature_InvalidFeature(t *testing.T) {
	dir := makeExtDir(t)
	resp := callTool(t, ExtAddFeature(), map[string]any{
		"directory": dir,
		"feature":   "unknown",
	})
	if !isError(resp) {
		t.Fatal("expected error, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
}

func TestExtAddFeature_BackgroundHandler(t *testing.T) {
	dir := makeExtDir(t)
	resp := callTool(t, ExtAddFeature(), map[string]any{
		"directory": dir,
		"feature":   "background-handler",
		"name":      "TestHandler",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
	content, err := os.ReadFile(filepath.Join(dir, "background.js"))
	if err != nil {
		t.Fatalf("background.js not found: %v", err)
	}
	if !strings.Contains(string(content), "TestHandler") {
		t.Fatalf("expected background.js to contain TestHandler, got: %s", string(content))
	}
}

// ---- ext_connect_orchestra ----

func TestExtConnectOrchestra_MissingDirectory(t *testing.T) {
	resp := callTool(t, ExtConnectOrchestra(), map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
}

func TestExtConnectOrchestra_Valid(t *testing.T) {
	dir := makeExtDir(t)
	resp := callTool(t, ExtConnectOrchestra(), map[string]any{
		"directory": dir,
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
	connectorPath := filepath.Join(dir, "orchestra-connector.js")
	if _, err := os.Stat(connectorPath); err != nil {
		t.Fatalf("orchestra-connector.js not created: %v", err)
	}
}

// ---- ext_publish ----

func TestExtPublish_MissingDirectory(t *testing.T) {
	resp := callTool(t, ExtPublish(), map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
}

func TestExtPublish_NoManifest(t *testing.T) {
	dir := t.TempDir() // empty dir, no manifest.json
	resp := callTool(t, ExtPublish(), map[string]any{
		"directory": dir,
	})
	if !isError(resp) {
		t.Fatal("expected error for missing manifest, got success")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %s", errorCode(resp))
	}
	if !strings.Contains(resp.ErrorMessage, "no manifest.json") {
		t.Fatalf("expected error message to contain 'no manifest.json', got: %s", resp.ErrorMessage)
	}
}

func TestExtPublish_Chrome(t *testing.T) {
	dir := makeExtDir(t)
	resp := callTool(t, ExtPublish(), map[string]any{
		"directory": dir,
		"store":     "chrome",
	})
	if isError(resp) {
		t.Fatalf("expected success, got error: %s — %s", resp.ErrorCode, resp.ErrorMessage)
	}
	text := resp.Result.Fields["text"].GetStringValue()
	if !strings.Contains(text, "chrome.google.com") {
		t.Fatalf("expected response to contain chrome.google.com, got: %s", text)
	}
}
