package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	// Parse --config first (before loading config)
	cfgPath := flag.String("config", "", "Config file path (default: ~/.zeno.toml)")
	port := flag.Int("port", 0, "HTTP server port")
	shell := flag.String("shell", "", "Shell to run")
	noOpen := flag.Bool("no-open", false, "Don't auto-open browser")
	secret := flag.String("secret", "", "Secret token for access")
	useTLS := flag.Bool("tls", false, "Enable HTTPS with TLS")
	certFile := flag.String("cert", "", "TLS certificate file path")
	keyFile := flag.String("key", "", "TLS private key file path")
	flag.Parse()

	// Set config path if provided, then load
	if *cfgPath != "" {
		setConfigPath(*cfgPath)
	}
	cfg := loadConfig()

	// CLI flags override config (only if explicitly set)
	if *port == 0 {
		*port = cfg.Port
	}
	if *shell == "" {
		*shell = cfg.Shell
	}

	if *secret == "" {
		*secret = os.Getenv("ZENO_SECRET")
	}

	// Determine command to run
	command := *shell
	var args []string
	if posArgs := flag.Args(); len(posArgs) > 0 && posArgs[0] == "ssh" {
		if len(posArgs) < 2 {
			log.Fatal("Usage: zeno ssh user@host [-p port] [-i keyfile]")
		}
		command = "ssh"
		args = posArgs[1:]
	}

	if (*certFile != "" || *keyFile != "") && !*useTLS {
		*useTLS = true
	}
	if (*certFile != "") != (*keyFile != "") {
		log.Fatal("Both --cert and --key must be provided together")
	}

	scheme := "http"
	if *useTLS {
		scheme = "https"
	}

	addr := fmt.Sprintf(":%d", *port)
	srv := newServer(command, args, *secret, *useTLS, &cfg)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("Zeno running on %s://localhost:%d", scheme, *port)
		log.Printf("Config: %s", configPath())

		var err error
		if *useTLS {
			if *certFile != "" {
				err = httpServer.ListenAndServeTLS(*certFile, *keyFile)
			} else {
				cert, fingerprint, genErr := generateSelfSignedCert()
				if genErr != nil {
					log.Fatalf("Failed to generate TLS certificate: %v", genErr)
				}
				log.Printf("TLS certificate fingerprint: %s", fingerprint)
				httpServer.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
				err = httpServer.ListenAndServeTLS("", "")
			}
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	if !*noOpen {
		go openBrowser(fmt.Sprintf("%s://localhost:%d", scheme, *port))
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)
}

func defaultShell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/sh"
}

func openBrowser(url string) {
	time.Sleep(100 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	cmd.Start()
}
