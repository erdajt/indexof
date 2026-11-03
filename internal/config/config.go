// SPDX-FileCopyrightText: 2026 erdajt <erdajtsopjani.tech@gmail.com>
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Editor             string `yaml:"editor"`
	Retries            int    `yaml:"retries"`
	TimeoutSec         int    `yaml:"timeout_sec"`
	UserAgent          string `yaml:"user_agent"`
	ShowHidden         bool   `yaml:"show_hidden"`
	DownloadDir        string `yaml:"download_dir"`
	DeleteOnExit       bool   `yaml:"delete_on_exit"`
	MoveToCwd          bool   `yaml:"move_to_cwd"`
	BigFileThresholdMB int    `yaml:"big_file_threshold_mb"`
	Theme              string `yaml:"theme"`
}

func Default() Config {
	return Config{
		Editor:             "nvim",
		Retries:            3,
		TimeoutSec:         15,
		UserAgent:          "github.com/erdajt/indexof/1.0",
		ShowHidden:         true,
		DownloadDir:        filepath.Join(os.TempDir(), "indexof"),
		DeleteOnExit:       true,
		MoveToCwd:          false,
		BigFileThresholdMB: 50,
		Theme:              "default",
	}
}

// indexof config path
func path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "indexof", "config.yaml")
}

func Load() (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path())
	if err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(filepath.Dir(path()), 0o755)
			_ = Save(cfg)
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	cfg.sanitize()
	return cfg, nil
}

func Save(cfg Config) error {
	cfg.sanitize()
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path(), data, 0o644)
}

func (c *Config) sanitize() {
	if c.Editor == "" {
		c.Editor = "nvim"
	}
	if c.Retries < 1 {
		c.Retries = 3
	}
	if c.TimeoutSec < 5 {
		c.TimeoutSec = 15
	}
	if c.DownloadDir == "" {
		c.DownloadDir = filepath.Join(os.TempDir(), "indexof")
	}
	if c.BigFileThresholdMB < 1 {
		c.BigFileThresholdMB = 50
	}
	if c.Theme == "" {
		c.Theme = "default"
	}
}
