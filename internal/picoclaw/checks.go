package picoclaw

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"gogal/internal/config"
	"gogal/internal/database"
	"gogal/internal/proxy"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CheckResult struct {
	Name    string
	OK      bool
	Details string
}

func RunDoctorChecks(root string, site string) ([]CheckResult, error) {
	results := []CheckResult{}

	if _, err := exec.LookPath("go"); err != nil {
		results = append(results, CheckResult{Name: "Go version", OK: false, Details: err.Error()})
	} else {
		out, _ := exec.Command("go", "version").CombinedOutput()
		results = append(results, CheckResult{Name: "Go version", OK: true, Details: string(out)})
	}

	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		results = append(results, CheckResult{Name: "Go module initialized", OK: false, Details: "missing go.mod"})
	} else {
		results = append(results, CheckResult{Name: "Go module initialized", OK: true, Details: "go.mod found"})
	}

	if _, err := os.Stat(filepath.Join(root, "sites", "common_site_config.json")); err != nil {
		results = append(results, CheckResult{Name: "common_site_config.json", OK: false, Details: err.Error()})
	} else {
		results = append(results, CheckResult{Name: "common_site_config.json", OK: true, Details: "found"})
	}

	cfg, err := config.LoadSiteConfig(root, site)
	if err != nil {
		results = append(results, CheckResult{Name: "site_config.json", OK: false, Details: err.Error()})
		return results, nil
	}
	results = append(results, CheckResult{Name: "site_config.json", OK: true, Details: "found"})

	dsn := database.DBConfig{Host: cfg.DBHost, Port: cfg.DBPort, Name: cfg.DBName, User: cfg.DBUser, Password: cfg.DBPassword}.DSN()
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		results = append(results, CheckResult{Name: "PostgreSQL connection", OK: false, Details: err.Error()})
	} else {
		sqlDB, err := gdb.DB()
		if err != nil {
			results = append(results, CheckResult{Name: "PostgreSQL connection", OK: false, Details: err.Error()})
			goto portCheck
		}
		defer sqlDB.Close()
		if pingErr := sqlDB.Ping(); pingErr != nil {
			results = append(results, CheckResult{Name: "PostgreSQL connection", OK: false, Details: pingErr.Error()})
		} else {
			results = append(results, CheckResult{Name: "PostgreSQL connection", OK: true, Details: "connected"})
			if privErr := database.CheckSchemaPrivileges(sqlDB, cfg.DBUser); privErr != nil {
				results = append(results, CheckResult{Name: "Schema privileges", OK: false, Details: privErr.Error()})
			} else {
				results = append(results, CheckResult{Name: "Schema privileges", OK: true, Details: "USAGE+CREATE on public"})
			}
			if createErr := database.CanCreateTable(sqlDB); createErr != nil {
				results = append(results, CheckResult{Name: "CREATE TABLE in public", OK: false, Details: createErr.Error()})
			} else {
				results = append(results, CheckResult{Name: "CREATE TABLE in public", OK: true, Details: "ok"})
			}
		}
	}

portCheck:
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		results = append(results, CheckResult{Name: "Port 8080 availability", OK: false, Details: err.Error()})
	} else {
		_ = ln.Close()
		results = append(results, CheckResult{Name: "Port 8080 availability", OK: true, Details: "available"})
	}

	if filepath.Base(root) == "" {
		results = append(results, CheckResult{Name: "Working directory", OK: false, Details: "invalid working directory"})
	} else {
		results = append(results, CheckResult{Name: "Working directory", OK: true, Details: fmt.Sprintf("%s", root)})
	}

	tr := proxy.CheckTraefikStatus()
	results = append(results, CheckResult{Name: "Traefik (optional)", OK: true, Details: tr.Message})

	return results, nil
}
