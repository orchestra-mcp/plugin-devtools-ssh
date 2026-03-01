package tools

import (
	"context"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	internalssh "github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"google.golang.org/protobuf/types/known/structpb"
)

// ---------- Helpers ----------

func newMgr() *internalssh.Manager {
	return internalssh.NewManager()
}

func callTool(
	t *testing.T,
	handler func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error),
	args map[string]any,
) *pluginv1.ToolResponse {
	t.Helper()
	var s *structpb.Struct
	if args != nil {
		var err error
		s, err = structpb.NewStruct(args)
		if err != nil {
			t.Fatalf("callTool: build args: %v", err)
		}
	}
	resp, err := handler(context.Background(), &pluginv1.ToolRequest{Arguments: s})
	if err != nil {
		t.Fatalf("callTool: unexpected error: %v", err)
	}
	return resp
}

func isError(resp *pluginv1.ToolResponse) bool {
	return resp != nil && !resp.Success
}

func errorCode(resp *pluginv1.ToolResponse) string {
	return resp.GetErrorCode()
}

// ---------- ssh_connect ----------

func TestSSHConnect_MissingHost(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHConnect(mgr), map[string]any{
		"user": "testuser",
	})
	if !isError(resp) {
		t.Fatal("expected error when host is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHConnect_MissingUser(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHConnect(mgr), map[string]any{
		"host": "127.0.0.1",
	})
	if !isError(resp) {
		t.Fatal("expected error when user is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHConnect_ConnectionFail(t *testing.T) {
	mgr := newMgr()
	// Port 2 is never listening; connection must fail quickly.
	resp := callTool(t, SSHConnect(mgr), map[string]any{
		"host":     "127.0.0.1",
		"user":     "testuser",
		"password": "wrong",
		"port":     float64(2),
	})
	if !isError(resp) {
		t.Fatal("expected error for connection to non-listening port")
	}
}

// ---------- ssh_exec ----------

func TestSSHExec_MissingSessionID(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHExec(mgr), map[string]any{
		"command": "ls",
	})
	if !isError(resp) {
		t.Fatal("expected error when session_id is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHExec_UnknownSession(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHExec(mgr), map[string]any{
		"session_id": "ssh-unknown",
		"command":    "ls",
	})
	if !isError(resp) {
		t.Fatal("expected error for unknown session_id")
	}
}

// ---------- ssh_disconnect ----------

func TestSSHDisconnect_MissingSessionID(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHDisconnect(mgr), map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error when session_id is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHDisconnect_UnknownSession(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHDisconnect(mgr), map[string]any{
		"session_id": "ssh-unknown",
	})
	if !isError(resp) {
		t.Fatal("expected error for unknown session_id")
	}
}

// ---------- ssh_list_sessions ----------

func TestSSHListSessions_Empty(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHListSessions(mgr), map[string]any{})
	if isError(resp) {
		t.Fatalf("expected success for empty session list, got error: %s — %s", errorCode(resp), resp.ErrorMessage)
	}
	if resp.Result == nil {
		t.Fatal("expected non-nil result")
	}
	textVal, ok := resp.Result.Fields["text"]
	if !ok {
		t.Fatal("expected 'text' field in result")
	}
	if textVal.GetStringValue() == "" {
		t.Fatal("expected non-empty text for empty session list")
	}
}

// ---------- ssh_upload ----------

func TestSSHUpload_MissingArgs(t *testing.T) {
	mgr := newMgr()
	// Missing session_id entirely.
	resp := callTool(t, SSHUpload(mgr), map[string]any{
		"local_path":  "/tmp/test",
		"remote_path": "/tmp/test",
	})
	if !isError(resp) {
		t.Fatal("expected error when session_id is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHUpload_UnknownSession(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHUpload(mgr), map[string]any{
		"session_id":  "ssh-unknown",
		"local_path":  "/tmp/test",
		"remote_path": "/tmp/test",
	})
	if !isError(resp) {
		t.Fatal("expected error for unknown session_id")
	}
}

// ---------- ssh_download ----------

func TestSSHDownload_MissingArgs(t *testing.T) {
	mgr := newMgr()
	// Missing session_id entirely.
	resp := callTool(t, SSHDownload(mgr), map[string]any{
		"remote_path": "/tmp/remote",
		"local_path":  "/tmp/local",
	})
	if !isError(resp) {
		t.Fatal("expected error when session_id is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHDownload_UnknownSession(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHDownload(mgr), map[string]any{
		"session_id":  "ssh-unknown",
		"remote_path": "/tmp/remote",
		"local_path":  "/tmp/local",
	})
	if !isError(resp) {
		t.Fatal("expected error for unknown session_id")
	}
}

// ---------- ssh_list_remote ----------

func TestSSHListRemote_MissingSessionID(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHListRemote(mgr), map[string]any{
		"path": "/tmp",
	})
	if !isError(resp) {
		t.Fatal("expected error when session_id is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHListRemote_UnknownSession(t *testing.T) {
	mgr := newMgr()
	resp := callTool(t, SSHListRemote(mgr), map[string]any{
		"session_id": "ssh-unknown",
		"path":       "/tmp",
	})
	if !isError(resp) {
		t.Fatal("expected error for unknown session_id")
	}
}
