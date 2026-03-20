package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHInteractiveCloseSchema returns the JSON Schema for the ssh_interactive_close tool.
func SSHInteractiveCloseSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"interactive_id": map[string]any{
				"type":        "string",
				"description": "Interactive session ID returned by ssh_interactive_open",
			},
		},
		"required": []any{"interactive_id"},
	})
	return s
}

// SSHInteractiveClose returns a tool handler that closes an interactive SSH session.
func SSHInteractiveClose(iMgr *ssh.InteractiveManager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "interactive_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		id := helpers.GetString(req.Arguments, "interactive_id")

		if err := iMgr.Close(id); err != nil {
			return helpers.ErrorResult("ssh_interactive_close_error", err.Error()), nil
		}

		return helpers.TextResult("interactive session closed"), nil
	}
}
