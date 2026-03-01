// Command tools-extension-generator is the entry point for the
// tools.extension-generator plugin binary. It provides 8 MCP tools
// for scaffolding and managing Chrome extensions.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra-mcp/plugin-tools-extension-generator/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

func main() {
	builder := plugin.New("tools.extension-generator").
		Version("0.1.0").
		Description("Chrome extension scaffolding and management tools").
		Author("Orchestra").
		Binary("tools-extension-generator")

	tp := &internal.ExtensionPlugin{}
	tp.RegisterTools(builder)

	p := builder.BuildWithTools()
	p.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := p.Run(ctx); err != nil {
		log.Fatalf("tools.extension-generator: %v", err)
	}
}
