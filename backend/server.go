package main

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

//go:embed static
var frontendFS embed.FS

const (
	msgData    byte = 0x00
	msgResize  byte = 0x01
	msgSession byte = 0x02

	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second

	readBufSize  = 32768
	writeBufSize = readBufSize + 1
)

type resizeMsg struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

type server struct {
	command  string
	args     []string
	secret   string
	config   *Config
	sessions *SessionManager
	mux      *http.ServeMux

	authMu     sync.Mutex
	authTokens map[string]time.Time
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func newServer(command string, args []string, secret string, config *Config) *server {
	s := &server{
		command:    command,
		args:       args,
		secret:     secret,
		config:     config,
		sessions:   newSessionManager(command, args, 60*time.Second),
		authTokens: make(map[string]time.Time),
	}
	s.mux = http.NewServeMux()

	frontendSub, err := fs.Sub(frontendFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/auth", s.handleAuth)
	s.mux.HandleFunc("/api/config", s.authWrapFunc(s.handleConfig))
	s.mux.HandleFunc("/api/files", s.authWrapFunc(s.handleFiles))
	s.mux.HandleFunc("/api/upload", s.authWrapFunc(s.handleUpload))
	s.mux.HandleFunc("/api/download", s.authWrapFunc(s.handleDownload))
	s.mux.HandleFunc("/api/rename", s.authWrapFunc(s.handleRename))
	s.mux.HandleFunc("/api/delete", s.authWrapFunc(s.handleDelete))
	s.mux.Handle("/", s.authWrap(http.FileServer(http.FS(frontendSub))))
	s.mux.HandleFunc("/ws", s.authWrapFunc(s.handleWS))

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ─── Config API ────────────────────────────────────────────

func (s *server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		configMu.RLock()
		defer configMu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.config)

	case http.MethodPut:
		var updated Config
		if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		configMu.Lock()
		updated.Port = s.config.Port
		updated.Shell = s.config.Shell
		*s.config = updated
		configMu.Unlock()
		if err := saveConfig(updated); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.config)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ─── Auth ──────────────────────────────────────────────────

func (s *server) validToken(token string) bool {
	if token == "" {
		return false
	}
	s.authMu.Lock()
	defer s.authMu.Unlock()
	exp, ok := s.authTokens[token]
	if !ok || time.Now().After(exp) {
		delete(s.authTokens, token)
		return false
	}
	return true
}

func (s *server) authWrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.secret == "" {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie("zeno-token")
		if err != nil || !s.validToken(cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) authWrapFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.secret == "" {
			next(w, r)
			return
		}
		cookie, err := r.Cookie("zeno-token")
		if err != nil || !s.validToken(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if s.secret == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	errMsg := ""
	if r.URL.Query().Get("error") == "1" {
		errMsg = "Invalid password"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, loginHTML, errMsg)
}

func (s *server) handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	pw := r.FormValue("password")

	if subtle.ConstantTimeCompare([]byte(pw), []byte(s.secret)) != 1 {
		http.Redirect(w, r, "/login?error=1", http.StatusFound)
		return
	}

	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(tokenBytes)

	s.authMu.Lock()
	s.authTokens[token] = time.Now().Add(24 * time.Hour)
	s.authMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "zeno-token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Zeno</title>
<link rel="icon" type="image/svg+xml" href="/favicon.svg">
<style>
@font-face{font-family:'JetBrains Mono';src:url('/fonts/jetbrains-mono-regular.woff2') format('woff2');font-weight:400;font-display:swap}
*{margin:0;padding:0;box-sizing:border-box}
html,body{height:100%%;background:#1a1a1a;font-family:'JetBrains Mono',monospace;display:flex;align-items:center;justify-content:center;-webkit-font-smoothing:antialiased;overflow:hidden}
.frame{width:420px;max-width:90vw;animation:appear .4s ease}
@keyframes appear{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:translateY(0)}}
.terminal{background:#1e1e1e;border:1.5px solid #333;border-radius:10px;overflow:hidden}
.titlebar{height:36px;background:#2d2d2d;display:flex;align-items:center;padding:0 14px;border-bottom:1px solid #333;gap:7px}
.dot{width:11px;height:11px;border-radius:50%%;border:1px solid rgba(255,255,255,0.06)}
.dot.r{background:#ff5f57}.dot.y{background:#febc2e}.dot.g{background:#28c840}
.titlebar-text{flex:1;text-align:center;color:#666;font-size:11px;margin-right:40px}
.body{padding:28px 24px 24px}
.logo{display:flex;align-items:center;gap:10px;margin-bottom:24px}
.logo svg{flex-shrink:0}
.logo-text{color:#888;font-size:11px;letter-spacing:0.5px}
.prompt{display:flex;align-items:center;gap:0;margin-bottom:6px}
.prompt-symbol{color:#666;font-size:13px;padding:8px 0 8px 2px;flex-shrink:0}
input{flex:1;background:transparent;border:none;color:#e0e0e0;font-size:13px;font-family:inherit;outline:none;padding:8px 8px;caret-color:#e0e0e0;letter-spacing:0.3px}
input::placeholder{color:#444}
.prompt-line{height:1px;background:#333;margin-bottom:16px}
.actions{display:flex;align-items:center;justify-content:space-between}
.hint{color:#444;font-size:10px}
.hint kbd{color:#555;background:#2a2a2a;padding:1px 5px;border-radius:3px;border:1px solid #333;font-family:inherit;font-size:10px}
button{background:transparent;border:1px solid #444;border-radius:6px;color:#888;font-size:11px;font-family:inherit;cursor:pointer;padding:6px 16px;transition:all .15s;letter-spacing:0.3px}
button:hover{border-color:#666;color:#ccc}
.error{color:#bf616a;font-size:11px;margin-bottom:12px;min-height:16px}
</style>
</head>
<body>
<div class="frame">
<div class="terminal">
<div class="titlebar">
<span class="dot r"></span><span class="dot y"></span><span class="dot g"></span>
<span class="titlebar-text">zeno</span>
</div>
<div class="body">
<div class="logo">
<svg width="28" height="28" viewBox="0 0 32 32"><rect width="32" height="32" rx="7" fill="#2d2d2d"/><path d="M8 9h16v2.5L11.5 23H24v2.5H8V23l12.5-11.5H8z" fill="#888"/></svg>
<span class="logo-text">TERMINAL ACCESS</span>
</div>
<div class="error">%s</div>
<form method="POST" action="/auth">
<div class="prompt">
<span class="prompt-symbol">&#x276F;</span>
<input type="password" name="password" placeholder="enter secret" autofocus autocomplete="current-password" />
</div>
<div class="prompt-line"></div>
<div class="actions">
<span class="hint">press <kbd>enter</kbd> to connect</span>
<button type="submit">connect</button>
</div>
</form>
</div>
</div>
</div>
</body>
</html>`

// ─── WebSocket Handler ─────────────────────────────────────

// ─── WebSocket Handler (Session-based) ─────────────────────

type sessionMsg struct {
	Action    string `json:"action"`
	SessionID string `json:"sessionID"`
}

func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Read first message — expect session connect
	_, firstMsg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	var session *Session

	if len(firstMsg) > 0 && firstMsg[0] == msgSession {
		var sm sessionMsg
		if err := json.Unmarshal(firstMsg[1:], &sm); err == nil && sm.SessionID != "" {
			session = s.sessions.Get(sm.SessionID)
		}
	}

	// Create new session if not reconnecting
	if session == nil {
		session, err = s.sessions.Create()
		if err != nil {
			log.Printf("Session create error: %v", err)
			conn.Close()
			return
		}
	}

	// Add this client
	session.AddClient(conn)

	// Replay ring buffer for reconnecting clients
	session.ReplayTo(conn)

	// Send session ID back
	resp, err := json.Marshal(sessionMsg{Action: "session", SessionID: session.ID})
	if err != nil {
		log.Printf("Failed to marshal session response: %v", err)
		return
	}
	respMsg := make([]byte, len(resp)+1)
	respMsg[0] = msgSession
	copy(respMsg[1:], resp)
	session.clientMu.Lock()
	mu := session.writeMus[conn]
	session.clientMu.Unlock()
	if mu != nil {
		mu.Lock()
		conn.WriteMessage(websocket.BinaryMessage, respMsg)
		mu.Unlock()
	}

	// WebSocket → PTY
	defer func() {
		session.RemoveClient(conn)
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if len(message) == 0 {
			continue
		}
		switch message[0] {
		case msgData:
			if _, err := session.ptmx.Write(message[1:]); err != nil {
				return
			}
		case msgResize:
			var size resizeMsg
			if err := json.Unmarshal(message[1:], &size); err != nil {
				log.Printf("Resize parse error: %v", err)
				continue
			}
			session.cols = size.Cols
			session.rows = size.Rows
			pty.Setsize(session.ptmx, &pty.Winsize{Rows: size.Rows, Cols: size.Cols})
		}
	}
}

// ─── File Browser / Upload / Download ──────────────────────

type fileEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

type fileListing struct {
	Path    string      `json:"path"`
	Entries []fileEntry `json:"entries"`
}

func (s *server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	session := s.sessions.Get(sessionID)
	if session == nil {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	cwd, err := getProcessCwd(session.cmd.Process.Pid)
	if err != nil {
		cwd = "."
	}

	reqPath := r.URL.Query().Get("path")
	if reqPath == "" {
		reqPath = "."
	}
	if strings.Contains(reqPath, "..") || filepath.IsAbs(reqPath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(cwd, reqPath)
	dirEntries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, "Cannot read directory", http.StatusInternalServerError)
		return
	}

	entries := make([]fileEntry, 0, len(dirEntries))
	// Folders first
	for _, e := range dirEntries {
		if e.IsDir() {
			entries = append(entries, fileEntry{Name: e.Name(), IsDir: true})
		}
	}
	// Then files
	for _, e := range dirEntries {
		if !e.IsDir() {
			info, _ := e.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			entries = append(entries, fileEntry{Name: e.Name(), IsDir: false, Size: size})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileListing{Path: fullPath, Entries: entries})
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	session := s.sessions.Get(sessionID)
	if session == nil {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB limit
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Bad upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	cwd, err := getProcessCwd(session.cmd.Process.Pid)
	if err != nil {
		cwd = "."
	}

	// Optional subpath for uploading into subdirectories
	subPath := r.URL.Query().Get("path")
	if subPath != "" && !strings.Contains(subPath, "..") && !filepath.IsAbs(subPath) {
		cwd = filepath.Join(cwd, subPath)
	}

	filename := filepath.Base(header.Filename)
	if filename == ".." || filename == "." {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	destPath := filepath.Join(cwd, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name": filename,
		"size": written,
		"path": destPath,
	})
}

func (s *server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	session := s.sessions.Get(sessionID)
	if session == nil {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" || strings.Contains(filePath, "..") || filepath.IsAbs(filePath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	cwd, err := getProcessCwd(session.cmd.Process.Pid)
	if err != nil {
		cwd = "."
	}

	fullPath := filepath.Join(cwd, filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(fullPath)))
	http.ServeFile(w, r, fullPath)
}

func (s *server) handleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	session := s.sessions.Get(sessionID)
	if session == nil {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	var body struct {
		Path    string `json:"path"`
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if body.Path == "" || body.NewName == "" ||
		strings.Contains(body.Path, "..") || strings.Contains(body.NewName, "..") ||
		filepath.IsAbs(body.Path) || filepath.IsAbs(body.NewName) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	cwd, err := getProcessCwd(session.cmd.Process.Pid)
	if err != nil {
		cwd = "."
	}

	oldPath := filepath.Join(cwd, body.Path)
	newPath := filepath.Join(filepath.Dir(oldPath), filepath.Base(body.NewName))

	if err := os.Rename(oldPath, newPath); err != nil {
		http.Error(w, "Rename failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	session := s.sessions.Get(sessionID)
	if session == nil {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" || strings.Contains(filePath, "..") || filepath.IsAbs(filePath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	cwd, err := getProcessCwd(session.cmd.Process.Pid)
	if err != nil {
		cwd = "."
	}

	fullPath := filepath.Join(cwd, filePath)
	if err := os.RemoveAll(fullPath); err != nil {
		http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
