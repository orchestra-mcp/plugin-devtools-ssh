package ssh

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	gossh "golang.org/x/crypto/ssh"
)

// InteractiveSessionInfo holds metadata about an interactive SSH session.
type InteractiveSessionInfo struct {
	ID            string `json:"id"`
	SSHSessionID  string `json:"ssh_session_id"`
	Cols          int    `json:"cols"`
	Rows          int    `json:"rows"`
}

// InteractiveSession wraps an SSH session with PTY allocation for interactive use.
type InteractiveSession struct {
	ID           string
	SSHSessionID string
	session      *gossh.Session
	stdin        io.WriteCloser
	cols         int
	rows         int
	mu           sync.Mutex

	// Subscriber fan-out for streaming output.
	subscribersMu sync.Mutex
	subscribers   map[int]chan []byte
	nextSubID     int
}

// InteractiveManager manages interactive SSH sessions.
type InteractiveManager struct {
	sessions map[string]*InteractiveSession
	mu       sync.RWMutex
}

// NewInteractiveManager creates a new interactive session manager.
func NewInteractiveManager() *InteractiveManager {
	return &InteractiveManager{
		sessions: make(map[string]*InteractiveSession),
	}
}

// generateInteractiveID creates an ID in the format "issh-" + 6 random hex chars.
func generateInteractiveID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return "issh-" + hex.EncodeToString(b)
}

// Open starts an interactive SSH shell session with PTY allocation on an existing
// SSH connection managed by the given Manager.
func (im *InteractiveManager) Open(sshMgr *Manager, sshSessionID string, cols, rows int) (string, error) {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	sess, err := sshMgr.Get(sshSessionID)
	if err != nil {
		return "", err
	}

	sshSession, err := sess.Client().NewSession()
	if err != nil {
		return "", fmt.Errorf("create SSH session: %w", err)
	}

	// Request PTY.
	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	if err := sshSession.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		sshSession.Close()
		return "", fmt.Errorf("request PTY: %w", err)
	}

	stdin, err := sshSession.StdinPipe()
	if err != nil {
		sshSession.Close()
		return "", fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		sshSession.Close()
		return "", fmt.Errorf("stdout pipe: %w", err)
	}

	if err := sshSession.Shell(); err != nil {
		sshSession.Close()
		return "", fmt.Errorf("start shell: %w", err)
	}

	id := generateInteractiveID()
	isess := &InteractiveSession{
		ID:           id,
		SSHSessionID: sshSessionID,
		session:      sshSession,
		stdin:        stdin,
		cols:         cols,
		rows:         rows,
		subscribers:  make(map[int]chan []byte),
	}

	// Background goroutine reads stdout and fans out to subscribers.
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])

				isess.subscribersMu.Lock()
				for _, ch := range isess.subscribers {
					select {
					case ch <- data:
					default:
					}
				}
				isess.subscribersMu.Unlock()
			}
			if err != nil {
				// Close all subscriber channels.
				isess.subscribersMu.Lock()
				for subID, ch := range isess.subscribers {
					close(ch)
					delete(isess.subscribers, subID)
				}
				isess.subscribersMu.Unlock()
				return
			}
		}
	}()

	im.mu.Lock()
	im.sessions[id] = isess
	im.mu.Unlock()

	return id, nil
}

// Send writes input to the interactive session's stdin.
func (im *InteractiveManager) Send(id, input string) error {
	im.mu.RLock()
	isess, ok := im.sessions[id]
	im.mu.RUnlock()
	if !ok {
		return fmt.Errorf("interactive session %q not found", id)
	}

	_, err := isess.stdin.Write([]byte(input))
	if err != nil {
		return fmt.Errorf("write to stdin: %w", err)
	}
	return nil
}

// Subscribe creates a channel that receives a copy of all output from the
// interactive session. Returns the read channel and an unsubscribe function.
func (im *InteractiveManager) Subscribe(id string) (<-chan []byte, func(), error) {
	im.mu.RLock()
	isess, ok := im.sessions[id]
	im.mu.RUnlock()
	if !ok {
		return nil, nil, fmt.Errorf("interactive session %q not found", id)
	}

	ch := make(chan []byte, 64)

	isess.subscribersMu.Lock()
	subID := isess.nextSubID
	isess.nextSubID++
	isess.subscribers[subID] = ch
	isess.subscribersMu.Unlock()

	unsub := func() {
		isess.subscribersMu.Lock()
		if _, exists := isess.subscribers[subID]; exists {
			delete(isess.subscribers, subID)
			close(ch)
		}
		isess.subscribersMu.Unlock()
	}

	return ch, unsub, nil
}

// Resize changes the terminal dimensions for an interactive session.
func (im *InteractiveManager) Resize(id string, cols, rows int) error {
	im.mu.RLock()
	isess, ok := im.sessions[id]
	im.mu.RUnlock()
	if !ok {
		return fmt.Errorf("interactive session %q not found", id)
	}

	if err := isess.session.WindowChange(rows, cols); err != nil {
		return fmt.Errorf("window change: %w", err)
	}

	isess.mu.Lock()
	isess.cols = cols
	isess.rows = rows
	isess.mu.Unlock()

	return nil
}

// List returns information about all active interactive sessions.
func (im *InteractiveManager) List() []InteractiveSessionInfo {
	im.mu.RLock()
	defer im.mu.RUnlock()

	infos := make([]InteractiveSessionInfo, 0, len(im.sessions))
	for _, isess := range im.sessions {
		isess.mu.Lock()
		infos = append(infos, InteractiveSessionInfo{
			ID:           isess.ID,
			SSHSessionID: isess.SSHSessionID,
			Cols:         isess.cols,
			Rows:         isess.rows,
		})
		isess.mu.Unlock()
	}
	return infos
}

// Close terminates an interactive session.
func (im *InteractiveManager) Close(id string) error {
	im.mu.Lock()
	isess, ok := im.sessions[id]
	if !ok {
		im.mu.Unlock()
		return fmt.Errorf("interactive session %q not found", id)
	}
	delete(im.sessions, id)
	im.mu.Unlock()

	_ = isess.stdin.Close()
	_ = isess.session.Close()

	return nil
}
