package internal

import (
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// ToolsPlugin registers all SSH tools.
type ToolsPlugin struct{}

// RegisterTools registers all 11 SSH tools (10 regular + 1 streaming) with the plugin builder.
func (tp *ToolsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	mgr := ssh.NewManager()
	iMgr := ssh.NewInteractiveManager()

	builder.RegisterTool("ssh_connect",
		"Connect to an SSH server",
		tools.SSHConnectSchema(), tools.SSHConnect(mgr))

	builder.RegisterTool("ssh_exec",
		"Execute command on remote server",
		tools.SSHExecSchema(), tools.SSHExec(mgr))

	builder.RegisterTool("ssh_disconnect",
		"Disconnect SSH session",
		tools.SSHDisconnectSchema(), tools.SSHDisconnect(mgr))

	builder.RegisterTool("ssh_list_sessions",
		"List active SSH sessions",
		tools.SSHListSessionsSchema(), tools.SSHListSessions(mgr))

	builder.RegisterTool("ssh_upload",
		"Upload file to remote server via SFTP",
		tools.SSHUploadSchema(), tools.SSHUpload(mgr))

	builder.RegisterTool("ssh_download",
		"Download file from remote server via SFTP",
		tools.SSHDownloadSchema(), tools.SSHDownload(mgr))

	builder.RegisterTool("ssh_list_remote",
		"List files on remote server via SFTP",
		tools.SSHListRemoteSchema(), tools.SSHListRemote(mgr))

	// Interactive SSH session tools.
	builder.RegisterTool("ssh_interactive_open",
		"Open an interactive SSH shell with PTY allocation",
		tools.SSHInteractiveOpenSchema(), tools.SSHInteractiveOpen(mgr, iMgr))

	builder.RegisterTool("ssh_interactive_send",
		"Send input to an interactive SSH session",
		tools.SSHInteractiveSendSchema(), tools.SSHInteractiveSend(iMgr))

	builder.RegisterTool("ssh_interactive_close",
		"Close an interactive SSH session",
		tools.SSHInteractiveCloseSchema(), tools.SSHInteractiveClose(iMgr))

	builder.RegisterStreamingTool("ssh_interactive_stream",
		"Stream real-time output from an interactive SSH session",
		tools.SSHInteractiveStreamSchema(), tools.SSHInteractiveStream(iMgr))
}
