package benchmanager

import (
	"os"
	"path/filepath"
	"sort"
)

type Status struct {
	BenchRoot      string
	Sites          []string
	Apps           []string
	MigrationState string
	PostgresState  string
	PortState      string
}

func LoadStatus(root string) Status {
	abs, _ := filepath.Abs(root)
	return Status{
		BenchRoot:      abs,
		Sites:          listDirs(filepath.Join(abs, "sites")),
		Apps:           listDirs(filepath.Join(abs, "apps")),
		MigrationState: "Unknown (placeholder)",
		PostgresState:  "Connected status from doctor (placeholder)",
		PortState:      "Port 8080 check available in doctor",
	}
}

func listDirs(path string) []string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return []string{}
	}
	out := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}
