package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHExecSchema returns the JSON Schema for the ssh_exec tool.
func SSHExecSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "SSH session ID returned by ssh_connect",
			},
			"command": map[string]any{
				"type":        "string",
				"description": "Command to execute on the remote server",
			},
		},
		"required": []any{"session_id", "command"},
	})
	return s
}

// SSHExec returns a tool handler that executes a command on a remote server.
func SSHExec(mgr *ssh.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "session_id", "command"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		sessionID := helpers.GetString(req.Arguments, "session_id")
		command := helpers.GetString(req.Arguments, "command")

		output, err := mgr.Exec(sessionID, command)
		if err != nil {
			return helpers.ErrorResult("ssh_exec_error", err.Error()), nil
		}

		return helpers.TextResult(output), nil
	}
}
