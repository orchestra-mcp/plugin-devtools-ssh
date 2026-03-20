package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHInteractiveSendSchema returns the JSON Schema for the ssh_interactive_send tool.
func SSHInteractiveSendSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"interactive_id": map[string]any{
				"type":        "string",
				"description": "Interactive session ID returned by ssh_interactive_open",
			},
			"input": map[string]any{
				"type":        "string",
				"description": "Input text to send to the interactive session (include \\n for Enter)",
			},
		},
		"required": []any{"interactive_id", "input"},
	})
	return s
}

// SSHInteractiveSend returns a tool handler that writes input to an interactive
// SSH session's stdin.
func SSHInteractiveSend(iMgr *ssh.InteractiveManager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "interactive_id", "input"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		id := helpers.GetString(req.Arguments, "interactive_id")
		input := helpers.GetString(req.Arguments, "input")

		if err := iMgr.Send(id, input); err != nil {
			return helpers.ErrorResult("ssh_interactive_send_error", err.Error()), nil
		}

		return helpers.TextResult("ok"), nil
	}
}
