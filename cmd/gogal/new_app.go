package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var appNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

type appManifest struct {
	Name        string              `json:"name"`
	Slug        string              `json:"slug"`
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	Route       string              `json:"route"`
	Version     string              `json:"version"`
	CreatedAt   string              `json:"created_at"`
	Backend     appBackendManifest  `json:"backend"`
	Frontend    appFrontendManifest `json:"frontend"`
	Modules     []appModuleManifest `json:"modules"`
}

type appBackendManifest struct {
	Controllers string `json:"controllers"`
	Hooks       string `json:"hooks"`
	Services    string `json:"services"`
}

type appFrontendManifest struct {
	Entry      string `json:"entry"`
	PagesDir   string `json:"pages_dir"`
	WidgetsDir string `json:"widgets_dir"`
}

type appModuleManifest struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Path        string `json:"path"`
	DocTypesDir string `json:"doctypes_dir"`
}

type newAppOptions struct {
	BenchPath   string
	Title       string
	Route       string
	Description string
	NoInput     bool
}

type newAppResult struct {
	BenchRoot    string
	AppName      string
	AppPath      string
	ManifestPath string
	ModuleName   string
	Route        string
}

func newNewAppCommand() *cobra.Command {
	options := &newAppOptions{}

	cmd := &cobra.Command{
		Use:   "new-app [app-name]",
		Short: "Create a new Gogal app scaffold inside a bench",
		Long: strings.TrimSpace(`Create a new installable Gogal app scaffold with module metadata, backend extension points, and frontend entrypoints.

The command is idempotent: rerunning it preserves existing source files and only fills in any missing scaffold pieces.`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := createApp(cmd, args[0], options)
			if err != nil {
				return err
			}

			cmd.Printf("Created app %s in bench %s\n", result.AppName, result.BenchRoot)
			cmd.Printf("App path: %s\n", result.AppPath)
			cmd.Printf("Manifest: %s\n", result.ManifestPath)
			cmd.Printf("Primary module: %s\n", result.ModuleName)
			cmd.Printf("Suggested route: %s\n", result.Route)
			cmd.Println("Scaffold includes backend stubs, frontend entrypoints, module metadata, fixtures, and migrations.")

			return nil
		},
	}

	cmd.Flags().StringVar(&options.BenchPath, "bench", ".", "Bench root directory")
	cmd.Flags().StringVar(&options.Title, "title", "", "Human-friendly app title (defaults to a title-cased version of the app name)")
	cmd.Flags().StringVar(&options.Route, "route", "", "Primary app route (defaults to /apps/<app-name>)")
	cmd.Flags().StringVar(&options.Description, "description", "", "Short app description")
	cmd.Flags().BoolVar(&options.NoInput, "no-input", false, "Disable interactive prompts and use defaults for missing values")

	return cmd
}

func createApp(cmd *cobra.Command, rawAppName string, options *newAppOptions) (*newAppResult, error) {
	appName := normalizeAppName(rawAppName)
	if !appNamePattern.MatchString(appName) {
		return nil, fmt.Errorf("invalid app name %q: use lowercase letters, numbers, dashes, or underscores, and start with a letter", rawAppName)
	}

	benchRoot, err := filepath.Abs(filepath.Clean(options.BenchPath))
	if err != nil {
		return nil, fmt.Errorf("resolve bench path: %w", err)
	}

	if _, err := initializeBench(benchRoot); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(cmd.InOrStdin())
	defaultTitle := humanizeSlug(appName)
	defaultRoute := "/apps/" + strings.ReplaceAll(appName, "_", "-")

	title := firstNonEmpty(options.Title, defaultTitle)
	if title, err = promptValue(cmd, reader, "App title", title, options.NoInput); err != nil {
		return nil, err
	}
	if strings.TrimSpace(title) == "" {
		title = defaultTitle
	}

	route := firstNonEmpty(options.Route, defaultRoute)
	if route, err = promptValue(cmd, reader, "Primary route", route, options.NoInput); err != nil {
		return nil, err
	}
	route = normalizeRoute(route, defaultRoute)

	description := strings.TrimSpace(options.Description)
	if description, err = promptValue(cmd, reader, "Description", description, options.NoInput); err != nil {
		return nil, err
	}

	appPath := filepath.Join(benchRoot, "apps", appName)
	moduleSlug := strings.ReplaceAll(appName, "-", "_")
	moduleLabel := title
	manifestPath := filepath.Join(appPath, "app.json")
	modulePath := filepath.Join(appPath, "modules", moduleSlug)
	docTypesDir := filepath.Join(modulePath, "doctypes")

	directories := []string{
		appPath,
		filepath.Join(appPath, "backend"),
		filepath.Join(appPath, "backend", "controllers"),
		filepath.Join(appPath, "backend", "hooks"),
		filepath.Join(appPath, "backend", "services"),
		filepath.Join(appPath, "frontend"),
		filepath.Join(appPath, "frontend", "src"),
		filepath.Join(appPath, "frontend", "src", "components"),
		filepath.Join(appPath, "frontend", "src", "pages"),
		filepath.Join(appPath, "frontend", "src", "lib"),
		filepath.Join(appPath, "fixtures"),
		filepath.Join(appPath, "migrations"),
		filepath.Join(appPath, "public"),
		filepath.Join(appPath, "scripts"),
		modulePath,
		docTypesDir,
	}

	for _, dir := range directories {
		if err := ensureDirectory(dir); err != nil {
			return nil, err
		}
	}

	manifest := defaultAppManifest(appName, title, description, route, moduleSlug, moduleLabel)
	existingManifest, err := readAppManifest(manifestPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if existingManifest != nil {
		mergeAppManifest(manifest, existingManifest)
	}
	manifest.Title = firstNonEmpty(strings.TrimSpace(title), manifest.Title)
	manifest.Description = firstNonEmpty(strings.TrimSpace(description), manifest.Description)
	manifest.Route = firstNonEmpty(strings.TrimSpace(route), manifest.Route)
	manifest.Modules = ensureManifestModules(manifest.Modules, moduleSlug, moduleLabel)

	if err := writeJSONFile(manifestPath, manifest); err != nil {
		return nil, err
	}

	files := map[string]string{
		filepath.Join(appPath, "README.md"):                                    appREADMEContents(manifest),
		filepath.Join(appPath, "backend", "controllers", "doc.go"):             backendDocContents("controllers", title),
		filepath.Join(appPath, "backend", "hooks", "doc.go"):                   backendDocContents("hooks", title),
		filepath.Join(appPath, "backend", "services", "doc.go"):                backendDocContents("services", title),
		filepath.Join(appPath, "backend", "hooks", "app_hooks.go"):             appHooksContents(title),
		filepath.Join(appPath, "frontend", "src", "index.js"):                  frontendEntryContents(manifest),
		filepath.Join(appPath, "frontend", "src", "pages", "HomePage.jsx"):     homePageContents(title, description),
		filepath.Join(appPath, "frontend", "src", "components", "AppCard.jsx"): appCardContents(),
		filepath.Join(appPath, "frontend", "src", "lib", "routes.js"):          frontendRoutesContents(manifest),
		filepath.Join(appPath, "modules", moduleSlug, "module.json"):           moduleManifestContents(moduleSlug, moduleLabel),
	}

	for path, content := range files {
		if err := writeFileIfMissing(path, content, 0o644); err != nil {
			return nil, err
		}
	}

	for _, keepPath := range []string{
		filepath.Join(docTypesDir, ".gitkeep"),
		filepath.Join(appPath, "fixtures", ".gitkeep"),
		filepath.Join(appPath, "migrations", ".gitkeep"),
		filepath.Join(appPath, "public", ".gitkeep"),
		filepath.Join(appPath, "scripts", ".gitkeep"),
	} {
		if err := writeFileIfMissing(keepPath, "", 0o644); err != nil {
			return nil, err
		}
	}

	return &newAppResult{
		BenchRoot:    benchRoot,
		AppName:      appName,
		AppPath:      appPath,
		ManifestPath: manifestPath,
		ModuleName:   moduleLabel,
		Route:        manifest.Route,
	}, nil
}

func defaultAppManifest(appName, title, description, route, moduleSlug, moduleLabel string) *appManifest {
	return &appManifest{
		Name:        appName,
		Slug:        appName,
		Title:       title,
		Description: description,
		Route:       route,
		Version:     "0.1.0",
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		Backend: appBackendManifest{
			Controllers: "backend/controllers",
			Hooks:       "backend/hooks",
			Services:    "backend/services",
		},
		Frontend: appFrontendManifest{
			Entry:      "frontend/src/index.js",
			PagesDir:   "frontend/src/pages",
			WidgetsDir: "frontend/src/components",
		},
		Modules: []appModuleManifest{
			{
				Name:        moduleSlug,
				Label:       moduleLabel,
				Path:        filepath.ToSlash(filepath.Join("modules", moduleSlug)),
				DocTypesDir: filepath.ToSlash(filepath.Join("modules", moduleSlug, "doctypes")),
			},
		},
	}
}

func readAppManifest(path string) (*appManifest, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest appManifest
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &manifest, nil
}

func mergeAppManifest(target, existing *appManifest) {
	if strings.TrimSpace(existing.Name) != "" {
		target.Name = existing.Name
	}
	if strings.TrimSpace(existing.Slug) != "" {
		target.Slug = existing.Slug
	}
	if strings.TrimSpace(existing.Title) != "" {
		target.Title = existing.Title
	}
	if strings.TrimSpace(existing.Description) != "" {
		target.Description = existing.Description
	}
	if strings.TrimSpace(existing.Route) != "" {
		target.Route = existing.Route
	}
	if strings.TrimSpace(existing.Version) != "" {
		target.Version = existing.Version
	}
	if strings.TrimSpace(existing.CreatedAt) != "" {
		target.CreatedAt = existing.CreatedAt
	}
	if strings.TrimSpace(existing.Backend.Controllers) != "" {
		target.Backend.Controllers = existing.Backend.Controllers
	}
	if strings.TrimSpace(existing.Backend.Hooks) != "" {
		target.Backend.Hooks = existing.Backend.Hooks
	}
	if strings.TrimSpace(existing.Backend.Services) != "" {
		target.Backend.Services = existing.Backend.Services
	}
	if strings.TrimSpace(existing.Frontend.Entry) != "" {
		target.Frontend.Entry = existing.Frontend.Entry
	}
	if strings.TrimSpace(existing.Frontend.PagesDir) != "" {
		target.Frontend.PagesDir = existing.Frontend.PagesDir
	}
	if strings.TrimSpace(existing.Frontend.WidgetsDir) != "" {
		target.Frontend.WidgetsDir = existing.Frontend.WidgetsDir
	}
	if len(existing.Modules) > 0 {
		target.Modules = existing.Modules
	}
}

func ensureManifestModules(modules []appModuleManifest, moduleSlug, moduleLabel string) []appModuleManifest {
	if len(modules) == 0 {
		return defaultAppManifest(moduleSlug, moduleLabel, "", "/apps/"+strings.ReplaceAll(moduleSlug, "_", "-"), moduleSlug, moduleLabel).Modules
	}

	for index := range modules {
		if strings.TrimSpace(modules[index].Name) == "" {
			modules[index].Name = moduleSlug
		}
		if strings.TrimSpace(modules[index].Label) == "" {
			modules[index].Label = moduleLabel
		}
		if strings.TrimSpace(modules[index].Path) == "" {
			modules[index].Path = filepath.ToSlash(filepath.Join("modules", modules[index].Name))
		}
		if strings.TrimSpace(modules[index].DocTypesDir) == "" {
			modules[index].DocTypesDir = filepath.ToSlash(filepath.Join(modules[index].Path, "doctypes"))
		}
	}

	return modules
}

func writeFileIfMissing(path string, content string, mode os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	if err := ensureDirectory(filepath.Dir(path)); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}

	return nil
}

func normalizeAppName(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	trimmed = strings.ReplaceAll(trimmed, " ", "-")
	return trimmed
}

func normalizeRoute(route string, fallback string) string {
	trimmed := strings.TrimSpace(route)
	if trimmed == "" {
		return fallback
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return trimmed
}

func humanizeSlug(value string) string {
	replacer := strings.NewReplacer("-", " ", "_", " ")
	parts := strings.Fields(replacer.Replace(strings.TrimSpace(value)))
	for index := range parts {
		parts[index] = strings.ToUpper(parts[index][:1]) + strings.ToLower(parts[index][1:])
	}
	if len(parts) == 0 {
		return "App"
	}
	return strings.Join(parts, " ")
}

func appREADMEContents(manifest *appManifest) string {
	var builder strings.Builder
	builder.WriteString("# ")
	builder.WriteString(manifest.Title)
	builder.WriteString("\n\n")
	builder.WriteString("This app was scaffolded by `gogal new-app` and is intended to be installed inside a Gogal bench.\n\n")
	if strings.TrimSpace(manifest.Description) != "" {
		builder.WriteString(manifest.Description)
		builder.WriteString("\n\n")
	}
	builder.WriteString("## Layout\n\n")
	builder.WriteString("- `app.json` – app manifest and integration metadata\n")
	builder.WriteString("- `backend/` – Go hooks, controllers, and services\n")
	builder.WriteString("- `frontend/` – React entrypoints, pages, and widgets\n")
	builder.WriteString("- `modules/` – module metadata and owned DocType JSON definitions\n")
	builder.WriteString("- `fixtures/` – seed data and exported records\n")
	builder.WriteString("- `migrations/` – future schema/data migration scripts\n\n")
	builder.WriteString("## Suggested next steps\n\n")
	builder.WriteString("1. Add DocType JSON files under `modules/")
	builder.WriteString(manifest.Modules[0].Name)
	builder.WriteString("/doctypes/`.\n")
	builder.WriteString("2. Register backend hooks and service logic in `backend/hooks`.\n")
	builder.WriteString("3. Build the runtime screens from `frontend/src/pages`.\n")
	return builder.String()
}

func backendDocContents(packageName string, title string) string {
	return fmt.Sprintf("package %s\n\n// %s package holds %s-specific extension points for the %s app.\n", packageName, packageName, packageName, title)
}

func appHooksContents(title string) string {
	return fmt.Sprintf(`package hooks

// Registry describes the backend extension points exposed by the %s app.
// Future CLI steps can wire this into app installation and lifecycle hooks.
type Registry struct {
	BeforeInstall []string
	AfterInstall  []string
	BeforeMigrate []string
	AfterMigrate  []string
}

// DefaultRegistry returns an empty hook registry ready for customization.
func DefaultRegistry() Registry {
	return Registry{}
}
`, title)
}

func frontendEntryContents(manifest *appManifest) string {
	return fmt.Sprintf(`import HomePage from "./pages/HomePage.jsx";
import { appRoutes } from "./lib/routes";

export const appMeta = {
  name: %q,
  title: %q,
  route: %q,
};

export { HomePage, appRoutes };
`, manifest.Name, manifest.Title, manifest.Route)
}

func homePageContents(title string, description string) string {
	if strings.TrimSpace(description) == "" {
		description = "Start building your module screens, dashboards, and runtime experiences here."
	}

	return fmt.Sprintf(`export default function HomePage() {
  return (
    <section className="rounded-3xl border border-slate-800 bg-slate-900/70 p-8 text-slate-100 shadow-2xl shadow-slate-950/30">
      <p className="text-sm font-semibold uppercase tracking-[0.3em] text-cyan-300">Gogal App</p>
      <h1 className="mt-3 text-3xl font-semibold text-white">%s</h1>
      <p className="mt-4 max-w-2xl text-sm leading-7 text-slate-300">%s</p>
    </section>
  );
}
`, title, description)
}

func appCardContents() string {
	return `export default function AppCard({ title, children }) {
  return (
    <div className="rounded-2xl border border-slate-800 bg-slate-950/60 p-5 text-slate-100">
      <h2 className="text-lg font-semibold text-white">{title}</h2>
      <div className="mt-3 text-sm text-slate-300">{children}</div>
    </div>
  );
}
`
}

func frontendRoutesContents(manifest *appManifest) string {
	return fmt.Sprintf(`import HomePage from "../pages/HomePage.jsx";

export const appRoutes = [
  {
    path: %q,
    label: %q,
    component: HomePage,
  },
];
`, manifest.Route, manifest.Title)
}

func moduleManifestContents(moduleSlug string, moduleLabel string) string {
	return fmt.Sprintf(`{
  "name": %q,
  "label": %q,
  "doctype_path": "doctypes"
}
`, moduleSlug, moduleLabel)
}
