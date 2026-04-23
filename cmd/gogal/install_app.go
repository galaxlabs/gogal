package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type installAppOptions struct {
	BenchPath string
	SiteName  string
}

type installAppResult struct {
	BenchRoot        string
	SiteName         string
	AppName          string
	SiteConfigPath   string
	RegistryPath     string
	InstalledApps    []string
	ManifestPath     string
	PrimaryRoute     string
	ModulesInstalled []string
}

type installedAppRegistry struct {
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Route        string   `json:"route"`
	Version      string   `json:"version"`
	AppPath      string   `json:"app_path"`
	ManifestPath string   `json:"manifest_path"`
	Modules      []string `json:"modules,omitempty"`
	InstalledAt  string   `json:"installed_at"`
	UpdatedAt    string   `json:"updated_at"`
}

func newInstallAppCommand() *cobra.Command {
	options := &installAppOptions{}

	cmd := &cobra.Command{
		Use:   "install-app [app-name]",
		Short: "Install a bench app onto a site",
		Long: strings.TrimSpace(`Install an existing Gogal app into an existing site.

The command is idempotent: it updates the site's installed app list, writes a per-site app registry record, and safely reuses existing installation metadata on reruns.`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := installApp(cmd, args[0], options)
			if err != nil {
				return err
			}

			cmd.Printf("Installed app %s on site %s\n", result.AppName, result.SiteName)
			cmd.Printf("Bench: %s\n", result.BenchRoot)
			cmd.Printf("Site config: %s\n", result.SiteConfigPath)
			cmd.Printf("Registry entry: %s\n", result.RegistryPath)
			cmd.Printf("Manifest: %s\n", result.ManifestPath)
			cmd.Printf("Primary route: %s\n", result.PrimaryRoute)
			if len(result.ModulesInstalled) > 0 {
				cmd.Printf("Modules: %s\n", strings.Join(result.ModulesInstalled, ", "))
			}
			cmd.Printf("Installed apps for %s: %s\n", result.SiteName, strings.Join(result.InstalledApps, ", "))

			return nil
		},
	}

	cmd.Flags().StringVar(&options.BenchPath, "bench", ".", "Bench root directory")
	cmd.Flags().StringVar(&options.SiteName, "site", "", "Site name to install the app into")
	_ = cmd.MarkFlagRequired("site")

	return cmd
}

func installApp(_ *cobra.Command, rawAppName string, options *installAppOptions) (*installAppResult, error) {
	appName := normalizeAppName(rawAppName)
	if !appNamePattern.MatchString(appName) {
		return nil, fmt.Errorf("invalid app name %q: use lowercase letters, numbers, dashes, or underscores, and start with a letter", rawAppName)
	}

	siteName := strings.TrimSpace(options.SiteName)
	if !siteNamePattern.MatchString(siteName) {
		return nil, fmt.Errorf("invalid site name %q: use letters, numbers, dots, dashes, or underscores", options.SiteName)
	}

	benchRoot, err := filepath.Abs(filepath.Clean(options.BenchPath))
	if err != nil {
		return nil, fmt.Errorf("resolve bench path: %w", err)
	}

	if _, err := initializeBench(benchRoot); err != nil {
		return nil, err
	}

	appPath := filepath.Join(benchRoot, "apps", appName)
	manifestPath := filepath.Join(appPath, "app.json")
	manifest, err := readAppManifest(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("app %q was not found at %s: run `gogal new-app %s --bench %s` first", appName, appPath, appName, benchRoot)
		}
		return nil, err
	}

	sitePath := filepath.Join(benchRoot, "sites", siteName)
	siteConfigPath := filepath.Join(sitePath, "site_config.json")
	siteConfig, err := readSiteConfig(siteConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("site %q was not found at %s: run `gogal new-site %s --bench %s` first", siteName, sitePath, siteName, benchRoot)
		}
		return nil, err
	}

	siteConfig.InstalledApps = appendUniqueNormalized(siteConfig.InstalledApps, manifest.Name)
	if err := writeJSONFile(siteConfigPath, siteConfig); err != nil {
		return nil, err
	}

	registryDir := filepath.Join(sitePath, "apps")
	if err := ensureDirectory(registryDir); err != nil {
		return nil, err
	}

	registryPath := filepath.Join(registryDir, manifest.Name+".json")
	existingRegistry, err := readInstalledAppRegistry(registryPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	registry := &installedAppRegistry{
		Name:         manifest.Name,
		Title:        manifest.Title,
		Route:        manifest.Route,
		Version:      manifest.Version,
		AppPath:      appPath,
		ManifestPath: manifestPath,
		Modules:      manifestModuleNames(manifest),
		InstalledAt:  now,
		UpdatedAt:    now,
	}
	if existingRegistry != nil {
		mergeInstalledAppRegistry(registry, existingRegistry)
		registry.Name = firstNonEmpty(manifest.Name, registry.Name)
		registry.Title = firstNonEmpty(manifest.Title, registry.Title)
		registry.Route = firstNonEmpty(manifest.Route, registry.Route)
		registry.Version = firstNonEmpty(manifest.Version, registry.Version)
		registry.AppPath = firstNonEmpty(appPath, registry.AppPath)
		registry.ManifestPath = firstNonEmpty(manifestPath, registry.ManifestPath)
		registry.Modules = normalizeUniqueStrings(manifestModuleNames(manifest))
		registry.UpdatedAt = now
	}
	if err := writeJSONFile(registryPath, registry); err != nil {
		return nil, err
	}

	return &installAppResult{
		BenchRoot:        benchRoot,
		SiteName:         siteName,
		AppName:          manifest.Name,
		SiteConfigPath:   siteConfigPath,
		RegistryPath:     registryPath,
		InstalledApps:    normalizeUniqueStrings(siteConfig.InstalledApps),
		ManifestPath:     manifestPath,
		PrimaryRoute:     manifest.Route,
		ModulesInstalled: manifestModuleNames(manifest),
	}, nil
}

func readInstalledAppRegistry(path string) (*installedAppRegistry, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var registry installedAppRegistry
	if err := json.Unmarshal(bytes, &registry); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &registry, nil
}

func mergeInstalledAppRegistry(target, existing *installedAppRegistry) {
	if strings.TrimSpace(existing.Name) != "" {
		target.Name = existing.Name
	}
	if strings.TrimSpace(existing.Title) != "" {
		target.Title = existing.Title
	}
	if strings.TrimSpace(existing.Route) != "" {
		target.Route = existing.Route
	}
	if strings.TrimSpace(existing.Version) != "" {
		target.Version = existing.Version
	}
	if strings.TrimSpace(existing.AppPath) != "" {
		target.AppPath = existing.AppPath
	}
	if strings.TrimSpace(existing.ManifestPath) != "" {
		target.ManifestPath = existing.ManifestPath
	}
	if len(existing.Modules) > 0 {
		target.Modules = normalizeUniqueStrings(existing.Modules)
	}
	if strings.TrimSpace(existing.InstalledAt) != "" {
		target.InstalledAt = existing.InstalledAt
	}
	if strings.TrimSpace(existing.UpdatedAt) != "" {
		target.UpdatedAt = existing.UpdatedAt
	}
}

func manifestModuleNames(manifest *appManifest) []string {
	modules := make([]string, 0, len(manifest.Modules))
	for _, module := range manifest.Modules {
		name := firstNonEmpty(module.Name, module.Label)
		if name == "" {
			continue
		}
		modules = append(modules, name)
	}
	return normalizeUniqueStrings(modules)
}

func appendUniqueNormalized(items []string, item string) []string {
	normalized := normalizeUniqueStrings(items)
	trimmed := strings.TrimSpace(item)
	if trimmed == "" {
		return normalized
	}
	for _, existing := range normalized {
		if strings.EqualFold(existing, trimmed) {
			return normalized
		}
	}
	return append(normalized, trimmed)
}

func normalizeUniqueStrings(values []string) []string {
	seen := make(map[string]string, len(values))
	ordered := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = trimmed
		ordered = append(ordered, trimmed)
	}
	return ordered
}
