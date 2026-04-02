package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Theme         string  `toml:"theme"         json:"theme"`
	FontSize      int     `toml:"fontSize"      json:"fontSize"`
	FontFamily    string  `toml:"fontFamily"    json:"fontFamily"`
	FontLigatures bool    `toml:"fontLigatures" json:"fontLigatures"`
	CursorStyle   string  `toml:"cursorStyle"   json:"cursorStyle"`
	CursorBlink   bool    `toml:"cursorBlink"   json:"cursorBlink"`
	LineHeight    float64 `toml:"lineHeight"    json:"lineHeight"`
	Scrollback    int     `toml:"scrollback"    json:"scrollback"`
	CopyOnSelect  bool    `toml:"copyOnSelect"  json:"copyOnSelect"`
	StartCommand  string  `toml:"startCommand"  json:"startCommand"`
	Port          int     `toml:"port"          json:"port"`
	Shell         string  `toml:"shell"         json:"shell"`
	PersistSessions bool   `toml:"persistSessions" json:"persistSessions"`
	Workspace       string `toml:"workspace,omitempty" json:"workspace,omitempty"`
}

var (
	configMu       sync.RWMutex
	configFilePath string
)

func defaultConfig() Config {
	return Config{
		Theme:       "dark",
		FontSize:    14,
		FontFamily:  "JetBrains Mono",
		CursorStyle: "block",
		CursorBlink: true,
		LineHeight:  1.1,
		Scrollback:  100000,
		Port:        8080,
		Shell:       defaultShell(),
	}
}

func setConfigPath(path string) {
	configFilePath = path
}

func configPath() string {
	if configFilePath != "" {
		return configFilePath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".zeno.toml"
	}
	return filepath.Join(home, ".zeno.toml")
}

func loadConfig() Config {
	cfg := defaultConfig()

	if _, err := toml.DecodeFile(configPath(), &cfg); err != nil {
		// File doesn't exist or is invalid — create with defaults
		saveConfig(cfg)
	}

	return cfg
}

func saveConfig(cfg Config) error {
	configMu.Lock()
	defer configMu.Unlock()
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}
	return os.WriteFile(configPath(), buf.Bytes(), 0600)
}
