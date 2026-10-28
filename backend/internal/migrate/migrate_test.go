package migrate

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedMigrationsPresent(t *testing.T) {
	entries, err := fs.ReadDir(files, "migrations")
	if err != nil {
		t.Fatalf("migrations are not embedded: %v", err)
	}

	count := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			count++
		}
	}

	if count == 0 {
		t.Fatal("no .sql migrations embedded")
	}
}
