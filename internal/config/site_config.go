package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CommonSiteConfig struct {
	DBHost        string `json:"db_host"`
	DBPort        int    `json:"db_port"`
	BasePort      int    `json:"base_port"`
	DeveloperMode bool   `json:"developer_mode"`
}

type SiteConfig struct {
	SiteName   string `json:"site_name"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
}

func DefaultCommonSiteConfig() CommonSiteConfig {
	return CommonSiteConfig{
		DBHost:        "127.0.0.1",
		DBPort:        5432,
		BasePort:      8080,
		DeveloperMode: true,
	}
}

func DefaultSiteConfig(siteName string) SiteConfig {
	return SiteConfig{
		SiteName:   siteName,
		DBName:     "gogaldb",
		DBUser:     "gogaluser",
		DBPassword: "gogal123",
		DBHost:     "127.0.0.1",
		DBPort:     5432,
	}
}

func LoadCommonSiteConfig(root string) (CommonSiteConfig, error) {
	path := filepath.Join(root, "sites", "common_site_config.json")
	cfg := DefaultCommonSiteConfig()
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read common_site_config.json: %w", err)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("parse common_site_config.json: %w", err)
	}
	if cfg.DBHost == "" {
		cfg.DBHost = "127.0.0.1"
	}
	if cfg.DBPort == 0 {
		cfg.DBPort = 5432
	}
	if cfg.BasePort == 0 {
		cfg.BasePort = 8080
	}
	return cfg, nil
}

func LoadSiteConfig(root string, site string) (SiteConfig, error) {
	path := filepath.Join(root, "sites", site, "site_config.json")
	cfg := DefaultSiteConfig(site)
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read site_config.json: %w", err)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("parse site_config.json: %w", err)
	}
	if cfg.SiteName == "" {
		cfg.SiteName = site
	}
	if cfg.DBHost == "" {
		cfg.DBHost = "127.0.0.1"
	}
	if cfg.DBPort == 0 {
		cfg.DBPort = 5432
	}
	return cfg, nil
}

func SaveJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}
