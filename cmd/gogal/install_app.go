package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type installAppOptions struct {
	BenchPath string
	SiteName  string
}

type installAppResult struct {
	BenchRoot      string
	SiteName       string
	AppName        string
	AppPath        string
	SiteConfigPath string
	AppsTxtPath    string
	InstalledApps  []string
}

func newInstallAppCommand() *cobra.Command {
	options := &installAppOptions{}

	cmd := &cobra.Command{
		Use:   "install-app [app-name]",
		Short: "Install a bench app onto a site",
		Long: strings.TrimSpace(`Install an app that exists under apps/<app-name> into a specific Gogal site.

This foundational command validates the app scaffold, registers it in the site's site_config.json, and maintains an apps.txt registry file. Running it again is safe and will not duplicate entries.`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := installApp(cmd, args[0], options)
			if err != nil {
				return err
			}

			cmd.Printf("Installed app %s on site %s\n", result.AppName, result.SiteName)
			cmd.Printf("Bench: %s\n", result.BenchRoot)
			cmd.Printf("App path: %s\n", result.AppPath)
			cmd.Printf("Site config: %s\n", result.SiteConfigPath)
			cmd.Printf("App registry: %s\n", result.AppsTxtPath)
			cmd.Printf("Installed apps: %s\n", strings.Join(result.InstalledApps, ", "))

			return nil
		},
	}

	cmd.Flags().StringVar(&options.BenchPath, "bench", ".", "Bench root directory")
	cmd.Flags().StringVar(&options.SiteName, "site", "", "Target site name inside the bench")
	_ = cmd.MarkFlagRequired("site")

	return cmd
}

func installApp(cmd *cobra.Command, rawAppName string, options *installAppOptions) (*installAppResult, error) {
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
	appManifestPath := filepath.Join(appPath, "app.json")
	manifest, err := readAppManifest(appManifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("app %q was not found in bench %s (expected %s)", appName, benchRoot, appManifestPath)
		}
		return nil, err
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return nil, fmt.Errorf("app manifest %s is missing a name", appManifestPath)
	}

	sitePath := filepath.Join(benchRoot, "sites", siteName)
	if info, err := os.Stat(sitePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("site %q does not exist in bench %s", siteName, benchRoot)
		}
		return nil, fmt.Errorf("stat site path %s: %w", sitePath, err)
	} else if !info.IsDir() {
		return nil, fmt.Errorf("site path %s is not a directory", sitePath)
	}

	siteConfigPath := filepath.Join(sitePath, "site_config.json")
	siteConfig, err := readSiteConfig(siteConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("site config %s was not found; create the site first with gogal new-site", siteConfigPath)
		}
		return nil, err
	}

	siteConfig.InstalledApps = appendUniquePreserveOrder(siteConfig.InstalledApps, manifest.Name)
	if err := writeJSONFile(siteConfigPath, siteConfig); err != nil {
		return nil, err
	}

	appsTxtPath := filepath.Join(sitePath, "apps.txt")
	if err := writeAppsRegistry(appsTxtPath, siteConfig.InstalledApps); err != nil {
		return nil, err
	}

	return &installAppResult{
		BenchRoot:      benchRoot,
		SiteName:       siteName,
		AppName:        manifest.Name,
		AppPath:        appPath,
		SiteConfigPath: siteConfigPath,
		AppsTxtPath:    appsTxtPath,
		InstalledApps:  append([]string(nil), siteConfig.InstalledApps...),
	}, nil
}

func writeAppsRegistry(path string, installedApps []string) error {
	if err := ensureDirectory(filepath.Dir(path)); err != nil {
		return err
	}

	cleanApps := appendUniquePreserveOrder(nil, installedApps...)
	content := strings.Join(cleanApps, "\n")
	if content != "" {
		content += "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write app registry %s: %w", path, err)
	}

	return nil
}

func appendUniquePreserveOrder(existing []string, values ...string) []string {
	seen := make(map[string]struct{}, len(existing)+len(values))
	result := make([]string, 0, len(existing)+len(values))

	appendValue := func(value string) {
		trimmed := normalizeAppName(value)
		if trimmed == "" {
			return
		}
		if _, ok := seen[trimmed]; ok {
			return
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}

	for _, value := range existing {
		appendValue(value)
	}
	for _, value := range values {
		appendValue(value)
	}

	return result
}
