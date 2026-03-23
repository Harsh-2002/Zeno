package main

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

//go:embed static
var frontendFS embed.FS

const (
	msgData   byte = 0x00
	msgResize byte = 0x01

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
	secret string
	config   *Config
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
		secret:    secret,
		config:     config,
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
		// Preserve server-only fields
		updated.Port = s.config.Port
		updated.Shell = s.config.Shell
		*s.config = updated
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
	rand.Read(tokenBytes)
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
<title>Zeno — Login</title>
<style>
@font-face{font-family:'JetBrains Mono';src:url('/fonts/jetbrains-mono-regular.woff2') format('woff2');font-weight:400;font-display:swap}
*{margin:0;padding:0;box-sizing:border-box}
html,body{height:100%%;background:#1e1e1e;font-family:'JetBrains Mono',monospace;display:flex;align-items:center;justify-content:center;-webkit-font-smoothing:antialiased}
.card{background:#252526;border:1px solid #3c3c3c;border-radius:12px;padding:40px;width:340px;box-shadow:0 8px 32px rgba(0,0,0,0.4)}
h1{color:#ccc;font-size:18px;font-weight:400;margin-bottom:8px;text-align:center}
.subtitle{color:#666;font-size:12px;text-align:center;margin-bottom:28px}
input{width:100%%;padding:10px 14px;background:#3c3c3c;border:1px solid #555;border-radius:6px;color:#ccc;font-size:14px;font-family:inherit;outline:none;margin-bottom:16px}
input:focus{border-color:#007acc}
button{width:100%%;padding:10px;background:#007acc;border:none;border-radius:6px;color:#fff;font-size:14px;font-family:inherit;cursor:pointer;transition:background 0.15s}
button:hover{background:#0098ff}
.error{color:#f14c4c;font-size:12px;text-align:center;margin-bottom:12px}
</style>
</head>
<body>
<div class="card">
<h1>Zeno</h1>
<p class="subtitle">Enter password to connect</p>
<div class="error">%s</div>
<form method="POST" action="/auth">
<input type="password" name="password" placeholder="Password" autofocus autocomplete="current-password" />
<button type="submit">Connect</button>
</form>
</div>
</body>
</html>`

// ─── WebSocket Handler ─────────────────────────────────────

func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	cmd := exec.Command(s.command, s.args...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("PTY start error: %v", err)
		conn.Close()
		return
	}

	var closeOnce sync.Once
	cleanup := func() {
		closeOnce.Do(func() {
			conn.Close()
			ptmx.Close()
			if cmd.Process != nil {
				cmd.Process.Kill()
				cmd.Wait()
			}
		})
	}
	defer cleanup()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var writeMu sync.Mutex

	// PTY -> WebSocket
	go func() {
		defer cleanup()

		buf := make([]byte, readBufSize)
		writeBuf := make([]byte, writeBufSize)
		writeBuf[0] = msgData

		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				writeMu.Lock()
				err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second))
				writeMu.Unlock()
				if err != nil {
					return
				}
			}
		}()

		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				return
			}
			copy(writeBuf[1:], buf[:n])
			writeMu.Lock()
			err = conn.WriteMessage(websocket.BinaryMessage, writeBuf[:n+1])
			writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}()

	// WebSocket -> PTY
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
			if _, err := ptmx.Write(message[1:]); err != nil {
				return
			}
		case msgResize:
			var size resizeMsg
			if err := json.Unmarshal(message[1:], &size); err != nil {
				log.Printf("Resize parse error: %v", err)
				continue
			}
			pty.Setsize(ptmx, &pty.Winsize{
				Rows: size.Rows,
				Cols: size.Cols,
			})
		}
	}
}
