package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-ssh/internal/ssh"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSHInteractiveStreamSchema returns the JSON Schema for the ssh_interactive_stream streaming tool.
func SSHInteractiveStreamSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"interactive_id": map[string]any{
				"type":        "string",
				"description": "Interactive session ID returned by ssh_interactive_open",
			},
		},
		"required": []any{"interactive_id"},
	})
	return s
}

// SSHInteractiveStream returns a streaming tool handler that subscribes to an
// interactive SSH session's output and pushes each chunk through the streaming
// channel. The stream runs until the context is cancelled or the session closes.
func SSHInteractiveStream(iMgr *ssh.InteractiveManager) func(ctx context.Context, req *pluginv1.StreamStart, chunks chan<- []byte) error {
	return func(ctx context.Context, req *pluginv1.StreamStart, chunks chan<- []byte) error {
		interactiveID := ""
		if req.Arguments != nil {
			if v, ok := req.Arguments.GetFields()["interactive_id"]; ok {
				interactiveID = v.GetStringValue()
			}
		}
		if interactiveID == "" {
			return fmt.Errorf("missing required parameter: interactive_id")
		}

		ch, unsub, err := iMgr.Subscribe(interactiveID)
		if err != nil {
			return err
		}
		defer unsub()

		for {
			select {
			case data, ok := <-ch:
				if !ok {
					return nil
				}
				select {
				case chunks <- data:
				case <-ctx.Done():
					return ctx.Err()
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
