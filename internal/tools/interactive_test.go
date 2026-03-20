package tools

import (
	"context"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	internalssh "github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"google.golang.org/protobuf/types/known/structpb"
)

func newInteractiveMgr() *internalssh.InteractiveManager {
	return internalssh.NewInteractiveManager()
}

// ---------- ssh_interactive_open ----------

func TestSSHInteractiveOpen_MissingSessionID(t *testing.T) {
	mgr := newMgr()
	iMgr := newInteractiveMgr()
	resp := callTool(t, SSHInteractiveOpen(mgr, iMgr), map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error when session_id is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHInteractiveOpen_UnknownSession(t *testing.T) {
	mgr := newMgr()
	iMgr := newInteractiveMgr()
	resp := callTool(t, SSHInteractiveOpen(mgr, iMgr), map[string]any{
		"session_id": "ssh-unknown",
	})
	if !isError(resp) {
		t.Fatal("expected error for unknown SSH session")
	}
	if errorCode(resp) != "ssh_interactive_open_error" {
		t.Fatalf("expected ssh_interactive_open_error, got %q", errorCode(resp))
	}
}

// ---------- ssh_interactive_send ----------

func TestSSHInteractiveSend_MissingArgs(t *testing.T) {
	iMgr := newInteractiveMgr()
	resp := callTool(t, SSHInteractiveSend(iMgr), map[string]any{
		"input": "ls\n",
	})
	if !isError(resp) {
		t.Fatal("expected error when interactive_id is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHInteractiveSend_MissingInput(t *testing.T) {
	iMgr := newInteractiveMgr()
	resp := callTool(t, SSHInteractiveSend(iMgr), map[string]any{
		"interactive_id": "issh-abc123",
	})
	if !isError(resp) {
		t.Fatal("expected error when input is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHInteractiveSend_UnknownSession(t *testing.T) {
	iMgr := newInteractiveMgr()
	resp := callTool(t, SSHInteractiveSend(iMgr), map[string]any{
		"interactive_id": "issh-unknown",
		"input":          "ls\n",
	})
	if !isError(resp) {
		t.Fatal("expected error for unknown interactive session")
	}
	if errorCode(resp) != "ssh_interactive_send_error" {
		t.Fatalf("expected ssh_interactive_send_error, got %q", errorCode(resp))
	}
}

// ---------- ssh_interactive_close ----------

func TestSSHInteractiveClose_MissingID(t *testing.T) {
	iMgr := newInteractiveMgr()
	resp := callTool(t, SSHInteractiveClose(iMgr), map[string]any{})
	if !isError(resp) {
		t.Fatal("expected error when interactive_id is missing")
	}
	if errorCode(resp) != "validation_error" {
		t.Fatalf("expected validation_error, got %q", errorCode(resp))
	}
}

func TestSSHInteractiveClose_UnknownSession(t *testing.T) {
	iMgr := newInteractiveMgr()
	resp := callTool(t, SSHInteractiveClose(iMgr), map[string]any{
		"interactive_id": "issh-unknown",
	})
	if !isError(resp) {
		t.Fatal("expected error for unknown interactive session")
	}
	if errorCode(resp) != "ssh_interactive_close_error" {
		t.Fatalf("expected ssh_interactive_close_error, got %q", errorCode(resp))
	}
}

// ---------- ssh_interactive_stream ----------

func TestSSHInteractiveStream_MissingID(t *testing.T) {
	iMgr := newInteractiveMgr()
	handler := SSHInteractiveStream(iMgr)
	chunks := make(chan []byte, 16)
	err := handler(context.Background(), &pluginv1.StreamStart{}, chunks)
	if err == nil {
		t.Fatal("expected error when interactive_id is missing")
	}
	if err.Error() != "missing required parameter: interactive_id" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHInteractiveStream_UnknownSession(t *testing.T) {
	iMgr := newInteractiveMgr()
	handler := SSHInteractiveStream(iMgr)
	args, _ := structpb.NewStruct(map[string]any{
		"interactive_id": "issh-unknown",
	})
	chunks := make(chan []byte, 16)
	err := handler(context.Background(), &pluginv1.StreamStart{Arguments: args}, chunks)
	if err == nil {
		t.Fatal("expected error for unknown interactive session")
	}
}

// ---------- InteractiveManager unit tests ----------

func TestInteractiveManager_ListEmpty(t *testing.T) {
	iMgr := newInteractiveMgr()
	infos := iMgr.List()
	if len(infos) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(infos))
	}
}

func TestInteractiveManager_SubscribeNotFound(t *testing.T) {
	iMgr := newInteractiveMgr()
	_, _, err := iMgr.Subscribe("issh-nonexistent")
	if err == nil {
		t.Fatal("expected error subscribing to non-existent session")
	}
}

func TestInteractiveManager_ResizeNotFound(t *testing.T) {
	iMgr := newInteractiveMgr()
	err := iMgr.Resize("issh-nonexistent", 120, 40)
	if err == nil {
		t.Fatal("expected error resizing non-existent session")
	}
}

func TestInteractiveManager_SendNotFound(t *testing.T) {
	iMgr := newInteractiveMgr()
	err := iMgr.Send("issh-nonexistent", "hello\n")
	if err == nil {
		t.Fatal("expected error sending to non-existent session")
	}
}

func TestInteractiveManager_CloseNotFound(t *testing.T) {
	iMgr := newInteractiveMgr()
	err := iMgr.Close("issh-nonexistent")
	if err == nil {
		t.Fatal("expected error closing non-existent session")
	}
}
