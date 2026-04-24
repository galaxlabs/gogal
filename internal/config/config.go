package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type RuntimeConfig struct {
	BenchRoot string
	SiteName  string
	HTTPPort  int
	Common    CommonSiteConfig
	Site      SiteConfig
}

func LoadRuntimeConfig() (RuntimeConfig, error) {
	root, err := os.Getwd()
	if err != nil {
		return RuntimeConfig{}, err
	}
	return LoadRuntimeConfigFromRoot(root)
}

func LoadRuntimeConfigFromRoot(root string) (RuntimeConfig, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return RuntimeConfig{}, err
	}
	site := strings.TrimSpace(os.Getenv("GOGAL_SITE"))
	if site == "" {
		site = "example.local"
	}
	common, err := LoadCommonSiteConfig(absRoot)
	if err != nil {
		common = DefaultCommonSiteConfig()
	}
	siteCfg, err := LoadSiteConfig(absRoot, site)
	if err != nil {
		siteCfg = DefaultSiteConfig(site)
	}
	port := common.BasePort
	if port == 0 {
		port = 8080
	}
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		parsed, parseErr := strconv.Atoi(p)
		if parseErr != nil {
			return RuntimeConfig{}, fmt.Errorf("invalid PORT: %w", parseErr)
		}
		port = parsed
	}
	return RuntimeConfig{
		BenchRoot: absRoot,
		SiteName:  site,
		HTTPPort:  port,
		Common:    common,
		Site:      siteCfg,
	}, nil
}
