package tools

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ExtBuildSchema returns the JSON Schema for the ext_build tool.
func ExtBuildSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Path to the extension directory to package",
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// ExtBuild returns a handler that packages an extension directory into a zip archive.
func ExtBuild() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		directory := helpers.GetString(req.Arguments, "directory")

		// Resolve to absolute path so the zip output is predictable.
		absDir, err := filepath.Abs(directory)
		if err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to resolve directory: %v", err)), nil
		}

		zipPath := absDir + ".zip"
		zipFile, err := os.Create(zipPath)
		if err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to create zip file: %v", err)), nil
		}
		defer zipFile.Close()

		zw := zip.NewWriter(zipFile)
		defer zw.Close()

		err = filepath.Walk(absDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() {
				return nil
			}

			// Compute the archive-relative path.
			rel, err := filepath.Rel(absDir, path)
			if err != nil {
				return err
			}

			w, err := zw.Create(rel)
			if err != nil {
				return err
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(w, f)
			return err
		})
		if err != nil {
			return helpers.ErrorResult("io_error", fmt.Sprintf("Failed to build zip: %v", err)), nil
		}

		return helpers.TextResult(fmt.Sprintf("Extension built: %s.zip", directory)), nil
	}
}
