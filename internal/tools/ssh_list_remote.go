package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/pkg/sftp"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHListRemoteSchema returns the JSON Schema for the ssh_list_remote tool.
func SSHListRemoteSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "SSH session ID returned by ssh_connect",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Remote directory path to list",
			},
		},
		"required": []any{"session_id", "path"},
	})
	return s
}

// remoteFileInfo holds metadata about a remote file for JSON serialization.
type remoteFileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

// SSHListRemote returns a tool handler that lists files on a remote server via SFTP.
func SSHListRemote(mgr *ssh.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "session_id", "path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		sessionID := helpers.GetString(req.Arguments, "session_id")
		path := helpers.GetString(req.Arguments, "path")

		sess, err := mgr.Get(sessionID)
		if err != nil {
			return helpers.ErrorResult("ssh_session_error", err.Error()), nil
		}

		sftpClient, err := sftp.NewClient(sess.Client())
		if err != nil {
			return helpers.ErrorResult("sftp_error", fmt.Sprintf("create SFTP client: %s", err.Error())), nil
		}
		defer sftpClient.Close()

		entries, err := sftpClient.ReadDir(path)
		if err != nil {
			return helpers.ErrorResult("sftp_error", fmt.Sprintf("list directory %s: %s", path, err.Error())), nil
		}

		files := make([]remoteFileInfo, 0, len(entries))
		for _, entry := range entries {
			files = append(files, remoteFileInfo{
				Name:    entry.Name(),
				Size:    entry.Size(),
				Mode:    entry.Mode().String(),
				IsDir:   entry.IsDir(),
				ModTime: entry.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		resp, err := helpers.JSONResult(files)
		if err != nil {
			return helpers.ErrorResult("ssh_list_error", err.Error()), nil
		}
		return resp, nil
	}
}
