package main

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme         string  `yaml:"theme"         json:"theme"`
	FontSize      int     `yaml:"fontSize"      json:"fontSize"`
	FontFamily    string  `yaml:"fontFamily"    json:"fontFamily"`
	FontLigatures bool    `yaml:"fontLigatures" json:"fontLigatures"`
	CursorStyle   string  `yaml:"cursorStyle"   json:"cursorStyle"`
	CursorBlink   bool    `yaml:"cursorBlink"   json:"cursorBlink"`
	LineHeight    float64 `yaml:"lineHeight"    json:"lineHeight"`
	Scrollback    int     `yaml:"scrollback"    json:"scrollback"`
	CopyOnSelect  bool    `yaml:"copyOnSelect"  json:"copyOnSelect"`
	Port          int     `yaml:"port"          json:"port"`
	Shell         string  `yaml:"shell"         json:"shell"`
	Layout        string  `yaml:"layout,omitempty" json:"layout,omitempty"`
}

var (
	configMu       sync.RWMutex
	configFilePath string // set by main.go via --config flag
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
		return ".zeno.yaml"
	}
	return filepath.Join(home, ".zeno.yaml")
}

func loadConfig() Config {
	cfg := defaultConfig()
	data, err := os.ReadFile(configPath())
	if err != nil {
		saveConfig(cfg)
		return cfg
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return cfg
	}

	if fileCfg.Theme != "" {
		cfg.Theme = fileCfg.Theme
	}
	if fileCfg.FontSize > 0 {
		cfg.FontSize = fileCfg.FontSize
	}
	if fileCfg.FontFamily != "" {
		cfg.FontFamily = fileCfg.FontFamily
	}
	cfg.FontLigatures = fileCfg.FontLigatures
	if fileCfg.CursorStyle != "" {
		cfg.CursorStyle = fileCfg.CursorStyle
	}
	cfg.CursorBlink = fileCfg.CursorBlink
	if fileCfg.LineHeight > 0 {
		cfg.LineHeight = fileCfg.LineHeight
	}
	if fileCfg.Scrollback > 0 {
		cfg.Scrollback = fileCfg.Scrollback
	}
	cfg.CopyOnSelect = fileCfg.CopyOnSelect
	if fileCfg.Port > 0 {
		cfg.Port = fileCfg.Port
	}
	if fileCfg.Shell != "" {
		cfg.Shell = fileCfg.Shell
	}
	if fileCfg.Layout != "" {
		cfg.Layout = fileCfg.Layout
	}

	return cfg
}

func saveConfig(cfg Config) error {
	configMu.Lock()
	defer configMu.Unlock()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}
