package ssh

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Session represents an active SSH connection.
type Session struct {
	ID        string
	Host      string
	User      string
	Port      int
	client    *ssh.Client
	ConnectedAt time.Time
}

// SessionInfo is a read-only snapshot of a session for listing.
type SessionInfo struct {
	ID          string `json:"id"`
	Host        string `json:"host"`
	User        string `json:"user"`
	Port        int    `json:"port"`
	ConnectedAt string `json:"connected_at"`
}

// Manager manages active SSH sessions.
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewManager creates a new SSH session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// generateID creates a session ID in the format ssh-XXXXXX (6 random hex chars).
func generateID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return "ssh-" + hex.EncodeToString(b)
}

// Connect establishes an SSH connection and stores the session.
// keyPath is the path to a PEM-encoded private key file.
// If password is provided and keyPath is empty, password auth is used.
func Connect(mgr *Manager, host, user, keyPath, password string, port int) (string, error) {
	if port == 0 {
		port = 22
	}

	var authMethods []ssh.AuthMethod

	if keyPath != "" {
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return "", fmt.Errorf("read SSH key %s: %w", keyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return "", fmt.Errorf("parse SSH key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return "", fmt.Errorf("no authentication method provided: supply key_path or password")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("SSH dial %s: %w", addr, err)
	}

	id := generateID()
	sess := &Session{
		ID:          id,
		Host:        host,
		User:        user,
		Port:        port,
		client:      client,
		ConnectedAt: time.Now(),
	}

	mgr.mu.Lock()
	mgr.sessions[id] = sess
	mgr.mu.Unlock()

	return id, nil
}

// Disconnect closes and removes an SSH session.
func (m *Manager) Disconnect(id string) error {
	m.mu.Lock()
	sess, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("session %s not found", id)
	}
	delete(m.sessions, id)
	m.mu.Unlock()

	return sess.client.Close()
}

// Get retrieves a session by ID.
func (m *Manager) Get(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session %s not found", id)
	}
	return sess, nil
}

// Client returns the underlying *ssh.Client for this session.
func (s *Session) Client() *ssh.Client {
	return s.client
}

// Exec runs a command on the remote server via the given session.
func (m *Manager) Exec(id, command string) (string, error) {
	sess, err := m.Get(id)
	if err != nil {
		return "", err
	}

	session, err := sess.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("command failed: %s", stderr.String())
		}
		return "", fmt.Errorf("command failed: %w", err)
	}

	return stdout.String(), nil
}

// List returns info about all active sessions.
func (m *Manager) List() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]SessionInfo, 0, len(m.sessions))
	for _, s := range m.sessions {
		infos = append(infos, SessionInfo{
			ID:          s.ID,
			Host:        s.Host,
			User:        s.User,
			Port:        s.Port,
			ConnectedAt: s.ConnectedAt.Format(time.RFC3339),
		})
	}
	return infos
}
