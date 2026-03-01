package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHConnectSchema returns the JSON Schema for the ssh_connect tool.
func SSHConnectSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"host": map[string]any{
				"type":        "string",
				"description": "SSH server hostname or IP address",
			},
			"user": map[string]any{
				"type":        "string",
				"description": "SSH username",
			},
			"key_path": map[string]any{
				"type":        "string",
				"description": "Path to PEM-encoded SSH private key file",
			},
			"port": map[string]any{
				"type":        "number",
				"description": "SSH port (default: 22)",
			},
			"password": map[string]any{
				"type":        "string",
				"description": "SSH password (used if key_path not provided)",
			},
		},
		"required": []any{"host", "user"},
	})
	return s
}

// SSHConnect returns a tool handler that connects to an SSH server.
func SSHConnect(mgr *ssh.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "host", "user"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		host := helpers.GetString(req.Arguments, "host")
		user := helpers.GetString(req.Arguments, "user")
		keyPath := helpers.GetString(req.Arguments, "key_path")
		password := helpers.GetString(req.Arguments, "password")
		port := helpers.GetInt(req.Arguments, "port")

		id, err := ssh.Connect(mgr, host, user, keyPath, password, port)
		if err != nil {
			return helpers.ErrorResult("ssh_connect_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Connected to %s@%s (session: %s)", user, host, id)), nil
	}
}
