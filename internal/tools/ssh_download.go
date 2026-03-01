package tools

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/pkg/sftp"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHDownloadSchema returns the JSON Schema for the ssh_download tool.
func SSHDownloadSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "SSH session ID returned by ssh_connect",
			},
			"remote_path": map[string]any{
				"type":        "string",
				"description": "Remote file path to download",
			},
			"local_path": map[string]any{
				"type":        "string",
				"description": "Local destination path",
			},
		},
		"required": []any{"session_id", "remote_path", "local_path"},
	})
	return s
}

// SSHDownload returns a tool handler that downloads a file from a remote server via SFTP.
func SSHDownload(mgr *ssh.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "session_id", "remote_path", "local_path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		sessionID := helpers.GetString(req.Arguments, "session_id")
		remotePath := helpers.GetString(req.Arguments, "remote_path")
		localPath := helpers.GetString(req.Arguments, "local_path")

		sess, err := mgr.Get(sessionID)
		if err != nil {
			return helpers.ErrorResult("ssh_session_error", err.Error()), nil
		}

		sftpClient, err := sftp.NewClient(sess.Client())
		if err != nil {
			return helpers.ErrorResult("sftp_error", fmt.Sprintf("create SFTP client: %s", err.Error())), nil
		}
		defer sftpClient.Close()

		remoteFile, err := sftpClient.Open(remotePath)
		if err != nil {
			return helpers.ErrorResult("sftp_error", fmt.Sprintf("open remote file: %s", err.Error())), nil
		}
		defer remoteFile.Close()

		// Create local directory if needed
		localDir := filepath.Dir(localPath)
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return helpers.ErrorResult("file_error", fmt.Sprintf("create local directory: %s", err.Error())), nil
		}

		localFile, err := os.Create(localPath)
		if err != nil {
			return helpers.ErrorResult("file_error", fmt.Sprintf("create local file: %s", err.Error())), nil
		}
		defer localFile.Close()

		written, err := io.Copy(localFile, remoteFile)
		if err != nil {
			return helpers.ErrorResult("sftp_error", fmt.Sprintf("download file: %s", err.Error())), nil
		}

		return helpers.TextResult(fmt.Sprintf("Downloaded %s to %s (%d bytes)", remotePath, localPath, written)), nil
	}
}
