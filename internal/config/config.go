package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Regions map[string]RegionConfig `yaml:"regions" json:"regions"`
	Proxy   ProxyConfig             `yaml:"proxy" json:"proxy"`
}

type RegionConfig struct {
	Cookie string `yaml:"cookie" json:"cookie"`
}

type ProxyConfig struct {
	HTTP  string `yaml:"http" json:"http"`
	HTTPS string `yaml:"https" json:"https"`
}

func ConfigDir() string {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "byte-cli")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "byte-cli")
}

func ConfigFile() string { return filepath.Join(ConfigDir(), "config.yaml") }

func Load() (AppConfig, error) {
	cfg := AppConfig{Regions: map[string]RegionConfig{}}
	data, err := os.ReadFile(ConfigFile())
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Regions == nil {
		cfg.Regions = map[string]RegionConfig{}
	}
	return cfg, nil
}

func Save(cfg AppConfig) error {
	if cfg.Regions == nil {
		cfg.Regions = map[string]RegionConfig{}
	}
	if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile(), data, 0o600)
}

func GetCookie(cfg AppConfig, region string) string {
	return cfg.Regions[region].Cookie
}

func SetCookie(cfg *AppConfig, region, cookie string) {
	if cfg.Regions == nil {
		cfg.Regions = map[string]RegionConfig{}
	}
	cfg.Regions[region] = RegionConfig{Cookie: cookie}
}

func MaskCookie(cookie string) string {
	if len(cookie) <= 8 {
		return strings.Repeat("*", len(cookie))
	}
	return cookie[:4] + strings.Repeat("*", len(cookie)-8) + cookie[len(cookie)-4:]
}
