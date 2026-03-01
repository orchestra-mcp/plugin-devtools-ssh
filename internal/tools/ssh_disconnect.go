package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHDisconnectSchema returns the JSON Schema for the ssh_disconnect tool.
func SSHDisconnectSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "SSH session ID to disconnect",
			},
		},
		"required": []any{"session_id"},
	})
	return s
}

// SSHDisconnect returns a tool handler that disconnects an SSH session.
func SSHDisconnect(mgr *ssh.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "session_id"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		sessionID := helpers.GetString(req.Arguments, "session_id")

		if err := mgr.Disconnect(sessionID); err != nil {
			return helpers.ErrorResult("ssh_disconnect_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Disconnected session %s", sessionID)), nil
	}
}
