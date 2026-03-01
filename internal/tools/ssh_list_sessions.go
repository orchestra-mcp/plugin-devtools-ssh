package tools

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHListSessionsSchema returns the JSON Schema for the ssh_list_sessions tool.
func SSHListSessionsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

// SSHListSessions returns a tool handler that lists active SSH sessions.
func SSHListSessions(mgr *ssh.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		sessions := mgr.List()

		if len(sessions) == 0 {
			return helpers.TextResult("No active SSH sessions"), nil
		}

		resp, err := helpers.JSONResult(sessions)
		if err != nil {
			return helpers.ErrorResult("ssh_list_error", err.Error()), nil
		}
		return resp, nil
	}
}
