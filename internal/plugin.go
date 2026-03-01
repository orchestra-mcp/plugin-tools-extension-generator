package internal

import (
	"github.com/orchestra-mcp/sdk-go/plugin"
	"github.com/orchestra-mcp/plugin-tools-extension-generator/internal/tools"
)

// ExtensionPlugin registers all Chrome extension tools with the plugin builder.
type ExtensionPlugin struct{}

// RegisterTools registers all 8 extension generator tools with the plugin builder.
func (ep *ExtensionPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	builder.RegisterTool("ext_scaffold",
		"Scaffold a new Chrome MV3 extension with boilerplate files",
		tools.ExtScaffoldSchema(), tools.ExtScaffold())

	builder.RegisterTool("ext_add_content_script",
		"Add a content script entry to an extension's manifest.json",
		tools.ExtAddContentScriptSchema(), tools.ExtAddContentScript())

	builder.RegisterTool("ext_validate",
		"Validate an extension directory's manifest.json for required fields",
		tools.ExtValidateSchema(), tools.ExtValidate())

	builder.RegisterTool("ext_build",
		"Package an extension directory into a distributable zip archive",
		tools.ExtBuildSchema(), tools.ExtBuild())

	builder.RegisterTool("ext_list_projects",
		"List Chrome extension projects found in a directory",
		tools.ExtListProjectsSchema(), tools.ExtListProjects())

	builder.RegisterTool("ext_add_feature",
		"Add a feature handler, popup button, or side panel to an extension",
		tools.ExtAddFeatureSchema(), tools.ExtAddFeature())

	builder.RegisterTool("ext_connect_orchestra",
		"Inject Orchestra SDK connector code into a Chrome extension",
		tools.ExtConnectOrchestraSchema(), tools.ExtConnectOrchestra())

	builder.RegisterTool("ext_publish",
		"Validate extension and show store submission instructions",
		tools.ExtPublishSchema(), tools.ExtPublish())
}
