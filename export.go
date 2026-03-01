package toolsextensiongenerator

import (
	"github.com/orchestra-mcp/plugin-tools-extension-generator/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all extension generator tools to the builder.
func Register(builder *plugin.PluginBuilder) {
	ep := &internal.ExtensionPlugin{}
	ep.RegisterTools(builder)
}
