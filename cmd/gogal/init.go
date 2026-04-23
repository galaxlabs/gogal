package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultDBHost        = "127.0.0.1"
	defaultDBPort        = 5432
	defaultBasePort      = 8000
	defaultRedisCache    = "redis://127.0.0.1:6379/0"
	defaultRedisQueue    = "redis://127.0.0.1:6379/1"
	defaultRedisSocketIO = "redis://127.0.0.1:6379/2"
)

type commonSiteConfig struct {
	DBHost        string `json:"db_host"`
	DBPort        int    `json:"db_port"`
	RedisCache    string `json:"redis_cache"`
	RedisQueue    string `json:"redis_queue"`
	RedisSocketIO string `json:"redis_socketio"`
	BasePort      int    `json:"base_port"`
}

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [bench-name]",
		Short: "Initialize a new Gogal multi-tenant bench",
		Long: strings.TrimSpace(`Create the foundational Gogal bench directory layout.

This command is idempotent. Running it again updates missing directories and merges any missing keys into sites/common_site_config.json without duplicating files.`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			benchRoot, err := filepath.Abs(filepath.Clean(args[0]))
			if err != nil {
				return fmt.Errorf("resolve bench path: %w", err)
			}

			result, err := initializeBench(benchRoot)
			if err != nil {
				return err
			}

			cmd.Printf("Initialized Gogal bench at %s\n", result.BenchRoot)
			cmd.Println()
			cmd.Println("Directories ready:")
			for _, dir := range result.Directories {
				cmd.Printf("  - %s\n", dir)
			}
			cmd.Println()
			cmd.Printf("Common site config: %s\n", result.CommonSiteConfigPath)
			cmd.Println("Next steps:")
			cmd.Println("  1. cd into the bench directory")
			cmd.Println("  2. run `gogal new-app <app-name>` to scaffold your first installable module package")
			cmd.Println("  3. run `gogal new-site <site-name>` to create your first tenant")
			cmd.Println("  4. run `gogal install-app <app-name> --site <site-name>` to attach the app to that tenant")

			return nil
		},
	}

	return cmd
}

type initResult struct {
	BenchRoot            string
	Directories          []string
	CommonSiteConfigPath string
}

func initializeBench(benchRoot string) (*initResult, error) {
	directories := []string{
		benchRoot,
		filepath.Join(benchRoot, "apps"),
		filepath.Join(benchRoot, "sites"),
		filepath.Join(benchRoot, "config"),
	}

	for _, dir := range directories {
		if err := ensureDirectory(dir); err != nil {
			return nil, err
		}
	}

	configPath := filepath.Join(benchRoot, "sites", "common_site_config.json")
	config, err := ensureCommonSiteConfig(configPath)
	if err != nil {
		return nil, err
	}

	writtenDirectories := make([]string, 0, len(directories))
	for _, dir := range directories[1:] {
		relativePath, relErr := filepath.Rel(benchRoot, dir)
		if relErr != nil {
			writtenDirectories = append(writtenDirectories, dir)
			continue
		}
		writtenDirectories = append(writtenDirectories, relativePath)
	}
	sort.Strings(writtenDirectories)

	if config == nil {
		return nil, fmt.Errorf("common site config was not created")
	}

	return &initResult{
		BenchRoot:            benchRoot,
		Directories:          writtenDirectories,
		CommonSiteConfigPath: configPath,
	}, nil
}

func ensureDirectory(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}
	return nil
}

func ensureCommonSiteConfig(configPath string) (*commonSiteConfig, error) {
	config := defaultCommonSiteConfig()

	existingConfig, err := readCommonSiteConfig(configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if existingConfig != nil {
		mergeCommonSiteConfig(config, existingConfig)
	}

	if err := writeJSONFile(configPath, config); err != nil {
		return nil, err
	}

	return config, nil
}

func defaultCommonSiteConfig() *commonSiteConfig {
	return &commonSiteConfig{
		DBHost:        defaultDBHost,
		DBPort:        defaultDBPort,
		RedisCache:    defaultRedisCache,
		RedisQueue:    defaultRedisQueue,
		RedisSocketIO: defaultRedisSocketIO,
		BasePort:      defaultBasePort,
	}
}

func readCommonSiteConfig(configPath string) (*commonSiteConfig, error) {
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config commonSiteConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("parse %s: %w", configPath, err)
	}

	return &config, nil
}

func mergeCommonSiteConfig(target, existing *commonSiteConfig) {
	if strings.TrimSpace(existing.DBHost) != "" {
		target.DBHost = existing.DBHost
	}
	if existing.DBPort != 0 {
		target.DBPort = existing.DBPort
	}
	if strings.TrimSpace(existing.RedisCache) != "" {
		target.RedisCache = existing.RedisCache
	}
	if strings.TrimSpace(existing.RedisQueue) != "" {
		target.RedisQueue = existing.RedisQueue
	}
	if strings.TrimSpace(existing.RedisSocketIO) != "" {
		target.RedisSocketIO = existing.RedisSocketIO
	}
	if existing.BasePort != 0 {
		target.BasePort = existing.BasePort
	}
}

func writeJSONFile(path string, value any) error {
	if err := ensureDirectory(filepath.Dir(path)); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	bytes = append(bytes, '\n')

	tempFilePath := path + ".tmp"
	if err := os.WriteFile(tempFilePath, bytes, 0o644); err != nil {
		return fmt.Errorf("write temp file %s: %w", tempFilePath, err)
	}
	if err := os.Rename(tempFilePath, path); err != nil {
		return fmt.Errorf("rename temp file for %s: %w", path, err)
	}

	return nil
}
