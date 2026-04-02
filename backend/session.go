package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

type RingBuffer struct {
	mu   sync.Mutex
	buf  []byte
	size int
	w    int
	full bool
}

func newRingBuffer(size int) *RingBuffer {
	return &RingBuffer{buf: make([]byte, size), size: size}
}

func (r *RingBuffer) Write(p []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range p {
		r.buf[r.w] = b
		r.w = (r.w + 1) % r.size
		if r.w == 0 {
			r.full = true
		}
	}
}

func (r *RingBuffer) Bytes() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.full {
		return append([]byte(nil), r.buf[:r.w]...)
	}
	out := make([]byte, r.size)
	copy(out, r.buf[r.w:])
	copy(out[r.size-r.w:], r.buf[:r.w])
	return out
}

type Session struct {
	ID       string
	ptmx     *os.File
	cmd      *exec.Cmd
	ring     *RingBuffer
	clients  []*websocket.Conn
	clientMu sync.Mutex
	writeMus map[*websocket.Conn]*sync.Mutex
	lastSeen time.Time
	cols     uint16
	rows     uint16
	done     chan struct{}
	closed   bool
}

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*Session
	command  string
	args     []string
	timeout  time.Duration
}

func newSessionManager(command string, args []string, timeout time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		command:  command,
		args:     args,
		timeout:  timeout,
	}
	go sm.reaper()
	return sm
}

func (sm *SessionManager) reaper() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		sm.mu.Lock()
		for id, s := range sm.sessions {
			s.clientMu.Lock()
			nClients := len(s.clients)
			s.clientMu.Unlock()
			if nClients == 0 && !s.closed && time.Since(s.lastSeen) > sm.timeout {
				log.Printf("Session %s timed out, cleaning up", id[:8])
				s.close()
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate session ID: %v", err)
	}
	return fmt.Sprintf("%x", b)
}

func (sm *SessionManager) Create() (*Session, error) {
	cmd := exec.Command(sm.command, sm.args...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	s := &Session{
		ID:       generateSessionID(),
		ptmx:     ptmx,
		cmd:      cmd,
		ring:     newRingBuffer(256 * 1024), // 256KB
		writeMus: make(map[*websocket.Conn]*sync.Mutex),
		lastSeen: time.Now(),
		done:     make(chan struct{}),
	}

	// PTY reader — fans out to all clients + ring buffer
	go func() {
		buf := make([]byte, readBufSize)
		writeBuf := make([]byte, writeBufSize)
		writeBuf[0] = msgData

		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				close(s.done)
				return
			}

			// Write to ring buffer
			s.ring.Write(buf[:n])

			// Fan out to all clients
			copy(writeBuf[1:], buf[:n])
			msg := make([]byte, n+1)
			copy(msg, writeBuf[:n+1])

			s.clientMu.Lock()
			alive := make([]*websocket.Conn, 0, len(s.clients))
			for _, c := range s.clients {
				mu := s.writeMus[c]
				mu.Lock()
				err := c.WriteMessage(websocket.BinaryMessage, msg)
				mu.Unlock()
				if err == nil {
					alive = append(alive, c)
				}
			}
			s.clients = alive
			s.clientMu.Unlock()
		}
	}()

	// Ping ticker
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.clientMu.Lock()
				for _, c := range s.clients {
					mu := s.writeMus[c]
					mu.Lock()
					c.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second))
					mu.Unlock()
				}
				s.clientMu.Unlock()
			case <-s.done:
				return
			}
		}
	}()

	sm.mu.Lock()
	sm.sessions[s.ID] = s
	sm.mu.Unlock()

	return s, nil
}

func (sm *SessionManager) Get(id string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.sessions[id]
}

func (s *Session) AddClient(conn *websocket.Conn) {
	s.clientMu.Lock()
	s.clients = append(s.clients, conn)
	s.writeMus[conn] = &sync.Mutex{}
	s.clientMu.Unlock()
}

func (s *Session) RemoveClient(conn *websocket.Conn) {
	s.clientMu.Lock()
	for i, c := range s.clients {
		if c == conn {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			delete(s.writeMus, conn)
			break
		}
	}
	s.lastSeen = time.Now()
	s.clientMu.Unlock()
}

func (s *Session) ReplayTo(conn *websocket.Conn) {
	data := s.ring.Bytes()
	if len(data) == 0 {
		return
	}
	msg := make([]byte, len(data)+1)
	msg[0] = msgData
	copy(msg[1:], data)

	s.clientMu.Lock()
	mu := s.writeMus[conn]
	s.clientMu.Unlock()
	if mu != nil {
		mu.Lock()
		conn.WriteMessage(websocket.BinaryMessage, msg)
		mu.Unlock()
	}
}

func (s *Session) close() {
	s.closed = true
	s.ptmx.Close()
	if s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
}
