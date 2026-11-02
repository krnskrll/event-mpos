package archtest

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const internalPrefix = "github.com/krnskrll/event-mpos/internal/"

// exempt packages wire concrete types together.
var exempt = map[string]bool{"app": true, "seed": true, "testutil": true, "archtest": true}

// layerRank orders a module's packages; a package may only import lower-ranked layers of its module.
var layerRank = map[string]int{"iface": 0, "repository": 1, "service": 2, "handler": 3}

// TestModuleBoundaries: modules talk to each other only through iface packages, and layers inside
// a module only depend downwards.
func TestModuleBoundaries(t *testing.T) {
	err := filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		from := filepath.ToSlash(filepath.Dir(strings.TrimPrefix(path, "../")))
		if exempt[top(from)] {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if to, ok := strings.CutPrefix(p, internalPrefix); ok && !allowed(from, to) {
				t.Errorf("%s must not import %s", from, to)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func allowed(from, to string) bool {
	if top(to) == "platform" {
		return true
	}
	if top(from) != top(to) {
		// platform stays independent of modules, except realtime's token check.
		return isIface(to) && (top(from) != "platform" || from == "platform/realtime")
	}
	f, fok := layerRank[last(from)]
	g, gok := layerRank[last(to)]
	return !fok || !gok || g < f
}

func top(p string) string { return strings.SplitN(p, "/", 2)[0] }

func last(p string) string { return p[strings.LastIndex(p, "/")+1:] }

func isIface(p string) bool { return strings.HasSuffix(p, "/iface") }

func TestAllowed(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{"order/service", "stock/iface", true},
		{"order/service", "stock/service", false},
		{"order/service", "order/repository", true},
		{"order/repository", "order/service", false},
		{"order/iface", "order/repository", false},
		{"order/handler", "order/service", true},
		{"integration/erp", "product/iface", true},
		{"integration/erp", "product/service", false},
		{"platform/common", "user/iface", false},
		{"platform/realtime", "auth/iface", true},
	}
	for _, c := range cases {
		if got := allowed(c.from, c.to); got != c.want {
			t.Errorf("allowed(%s, %s) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}
