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

// SSHUploadSchema returns the JSON Schema for the ssh_upload tool.
func SSHUploadSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"session_id": map[string]any{
				"type":        "string",
				"description": "SSH session ID returned by ssh_connect",
			},
			"local_path": map[string]any{
				"type":        "string",
				"description": "Local file path to upload",
			},
			"remote_path": map[string]any{
				"type":        "string",
				"description": "Remote destination path",
			},
		},
		"required": []any{"session_id", "local_path", "remote_path"},
	})
	return s
}

// SSHUpload returns a tool handler that uploads a file to a remote server via SFTP.
func SSHUpload(mgr *ssh.Manager) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "session_id", "local_path", "remote_path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		sessionID := helpers.GetString(req.Arguments, "session_id")
		localPath := helpers.GetString(req.Arguments, "local_path")
		remotePath := helpers.GetString(req.Arguments, "remote_path")

		sess, err := mgr.Get(sessionID)
		if err != nil {
			return helpers.ErrorResult("ssh_session_error", err.Error()), nil
		}

		sftpClient, err := sftp.NewClient(sess.Client())
		if err != nil {
			return helpers.ErrorResult("sftp_error", fmt.Sprintf("create SFTP client: %s", err.Error())), nil
		}
		defer sftpClient.Close()

		localFile, err := os.Open(localPath)
		if err != nil {
			return helpers.ErrorResult("file_error", fmt.Sprintf("open local file: %s", err.Error())), nil
		}
		defer localFile.Close()

		// Create remote directory if needed
		remoteDir := filepath.Dir(remotePath)
		_ = sftpClient.MkdirAll(remoteDir)

		remoteFile, err := sftpClient.Create(remotePath)
		if err != nil {
			return helpers.ErrorResult("sftp_error", fmt.Sprintf("create remote file: %s", err.Error())), nil
		}
		defer remoteFile.Close()

		written, err := io.Copy(remoteFile, localFile)
		if err != nil {
			return helpers.ErrorResult("sftp_error", fmt.Sprintf("upload file: %s", err.Error())), nil
		}

		return helpers.TextResult(fmt.Sprintf("Uploaded %s to %s (%d bytes)", localPath, remotePath, written)), nil
	}
}
