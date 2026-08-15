package backend_test

import (
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/yukihito-jokyu/topic2html/backend/"

func TestNamedLayersExistAndLegacyInternalDirectoryIsAbsent(t *testing.T) {
	t.Parallel()
	root := backendRoot(t)
	tests := []struct {
		name      string
		directory string
		legacy    bool
	}{
		{
			name:      "cmd",
			directory: "cmd",
		},
		{
			name:      "handler",
			directory: "handler",
		},
		{
			name:      "usecase",
			directory: "usecase",
		},
		{
			name:      "repository",
			directory: "repository",
		},
		{
			name:      "domain",
			directory: "domain",
		},
		{
			name:      "legacy internal",
			directory: "internal",
			legacy:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := os.Stat(filepath.Join(root, tt.directory))
			if tt.legacy {
				if !errors.Is(err, os.ErrNotExist) {
					t.Errorf("legacy internal directory remains: %v", err)
				}

				return
			}
			if err != nil || !info.IsDir() {
				t.Errorf("required layer directory %q is missing", tt.directory)
			}
		})
	}
}

func TestNamedLayerDependencies(t *testing.T) {
	t.Parallel()
	root := backendRoot(t)
	tests := []struct {
		name      string
		layer     string
		forbidden []string
	}{
		{
			name:      "domain",
			layer:     "domain",
			forbidden: []string{modulePath, "github.com/gin-gonic/gin", "github.com/jackc/pgx", "net/http", "os"},
		},
		{
			name:      "usecase",
			layer:     "usecase",
			forbidden: []string{modulePath + "handler/", modulePath + "repository/", modulePath + "cmd/", "github.com/gin-gonic/gin", "github.com/jackc/pgx", "net/http", "os"},
		},
		{
			name:      "repository",
			layer:     "repository",
			forbidden: []string{modulePath + "handler/", modulePath + "cmd/", "github.com/gin-gonic/gin", "os"},
		},
		{
			name:      "handler",
			layer:     "handler",
			forbidden: []string{modulePath + "repository/", modulePath + "cmd/", "os"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertLayerHasNoForbiddenImports(t, filepath.Join(root, tt.layer), tt.forbidden)
		})
	}
}

func backendRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Dir(file)
}

func assertLayerHasNoForbiddenImports(t *testing.T, directory string, forbidden []string) {
	t.Helper()
	err := filepath.WalkDir(directory, func(sourceFile string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(sourceFile) != ".go" {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), sourceFile, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range parsed.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			for _, prohibited := range forbidden {
				if path == prohibited || strings.HasPrefix(path, prohibited) {
					t.Errorf("%s imports forbidden dependency %q", sourceFile, path)
				}
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", directory, err)
	}
}
