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

// enforce flips to true once every module sits behind its iface package (task T15).
const enforce = false

// exempt packages may wire concrete types together.
var exempt = map[string]bool{"app": true, "seed": true, "testutil": true, "archtest": true}

// TestModuleBoundaries: a module may import another module only through its iface package,
// and iface/platform packages may not import concrete module code.
func TestModuleBoundaries(t *testing.T) {
	var violations []string
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
				violations = append(violations, from+" -> "+to)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range violations {
		if enforce {
			t.Error(v)
		} else {
			t.Log(v)
		}
	}
}

func allowed(from, to string) bool {
	if top(to) == "platform" || isIface(to) {
		return true
	}
	if top(from) == "platform" || isIface(from) || top(from) == "integration" {
		return false
	}
	return top(from) == top(to)
}

func top(p string) string { return strings.SplitN(p, "/", 2)[0] }

func isIface(p string) bool { return strings.HasSuffix(p, "/iface") }
