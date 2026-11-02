package app_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

// TestContractMatchesRoutes keeps api/serve/*.yml and the router in step. The router side is
// routes.golden, which TestRoutesUnchanged pins to the real router.
func TestContractMatchesRoutes(t *testing.T) {
	golden, err := os.ReadFile(filepath.Join("testdata", "routes.golden"))
	if err != nil {
		t.Fatal(err)
	}
	routes := map[string]bool{}
	for _, r := range strings.Split(strings.TrimSpace(string(golden)), "\n") {
		routes[r] = true
	}

	spec, err := specRoutes(filepath.Join("..", "..", "api", "serve"))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range sortedKeys(routes) {
		if !spec[r] {
			t.Errorf("route missing from api/serve: %s", r)
		}
	}
	for _, r := range sortedKeys(spec) {
		if !routes[r] {
			t.Errorf("spec documents a route the router lacks: %s", r)
		}
	}
}

var pathParam = regexp.MustCompile(`\{(\w+)\}`)

func specRoutes(dir string) (map[string]bool, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var doc struct {
			Paths map[string]map[string]any `yaml:"paths"`
		}
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		for p, ops := range doc.Paths {
			for m := range ops {
				switch m {
				case "get", "post", "put", "patch", "delete":
					out[strings.ToUpper(m)+" "+pathParam.ReplaceAllString(p, ":$1")] = true
				}
			}
		}
	}
	return out, nil
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
