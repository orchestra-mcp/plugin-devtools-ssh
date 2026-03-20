package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHInteractiveOpenSchema returns the JSON Schema for the ssh_interactive_open tool.
func SSHInteractiveOpenSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "SSH session ID returned by ssh_connect",
			},
			"cols": map[string]any{
				"type":        "number",
				"description": "Terminal width in columns (default: 80)",
			},
			"rows": map[string]any{
				"type":        "number",
				"description": "Terminal height in rows (default: 24)",
			},
		},
		"required": []any{"session_id"},
	})
	return s
}

// SSHInteractiveOpen returns a tool handler that opens an interactive SSH shell
// with PTY allocation on an existing SSH connection.
func SSHInteractiveOpen(sshMgr *ssh.Manager, iMgr *ssh.InteractiveManager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "session_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		sessionID := helpers.GetString(req.Arguments, "session_id")
		cols := helpers.GetInt(req.Arguments, "cols")
		rows := helpers.GetInt(req.Arguments, "rows")

		id, err := iMgr.Open(sshMgr, sessionID, cols, rows)
		if err != nil {
			return helpers.ErrorResult("ssh_interactive_open_error", err.Error()), nil
		}

		return helpers.JSONResult(map[string]any{
			"interactive_id": id,
			"ssh_session_id": sessionID,
			"cols":           cols,
			"rows":           rows,
		})
	}
}
