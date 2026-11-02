package app_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Set UPDATE_GOLDEN=1 to rewrite the golden file after an intended route change.
func TestRoutesUnchanged(t *testing.T) {
	a := testutil.NewApp(t)
	var routes []string
	for _, r := range a.Router.Routes() {
		routes = append(routes, r.Method+" "+r.Path)
	}
	sort.Strings(routes)
	got := strings.Join(routes, "\n") + "\n"

	golden := filepath.Join("testdata", "routes.golden")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("routes changed:\n--- want\n%s--- got\n%s", want, got)
	}
}
