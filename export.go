package devtoolsssh

import (
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Register adds all SSH tools to the builder.
func Register(builder *plugin.PluginBuilder) {
	tp := &internal.ToolsPlugin{}
	tp.RegisterTools(builder)
}
