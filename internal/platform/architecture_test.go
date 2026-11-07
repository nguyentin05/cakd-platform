package platform_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePrefix = "github.com/nguyentin05/cakd-platform/"

func TestHexagonalArchitectureBoundaries(t *testing.T) {
	repoRoot := findRepoRoot(t)
	internalDir := filepath.Join(repoRoot, "internal")

	fset := token.NewFileSet()
	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(internalDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		layer := strings.Split(rel, "/")[0]

		// Only enforce on hexagonal layers: domain, usecase, adapter, infrastructure, composition, platform
		switch layer {
		case "domain", "usecase", "adapter", "infrastructure", "composition", "platform":
		default:
			return nil
		}

		fileAst, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", rel, err)
		}

		for _, imp := range fileAst.Imports {
			impPath := strings.Trim(imp.Path.Value, `"`)
			checkImportBoundary(t, rel, layer, impPath)
		}

		if layer == "domain" || layer == "usecase" || layer == "infrastructure" {
			checkForbiddenCalls(t, fset, rel, layer, fileAst)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal dir failed: %v", err)
	}
}

func checkImportBoundary(t *testing.T, relFile, layer, impPath string) {
	t.Helper()

	switch layer {
	case "domain":
		if strings.HasPrefix(impPath, modulePrefix) {
			sub := strings.TrimPrefix(impPath, modulePrefix)
			if !strings.HasPrefix(sub, "internal/domain") {
				t.Errorf("[Hexagonal Violation] %s (domain) must not import %s", relFile, impPath)
			}
		} else if strings.Contains(impPath, ".") {
			t.Errorf("[Hexagonal Violation] %s (domain) must only import Go stdlib or internal/domain/*, found external import %s", relFile, impPath)
		}

	case "usecase":
		if strings.HasPrefix(impPath, modulePrefix) {
			sub := strings.TrimPrefix(impPath, modulePrefix)
			allowed := strings.HasPrefix(sub, "internal/domain") ||
				strings.HasPrefix(sub, "internal/usecase") ||
				strings.HasPrefix(sub, "internal/platform")
			if !allowed {
				t.Errorf("[Hexagonal Violation] %s (usecase) must not import %s", relFile, impPath)
			}
		}

	case "adapter":
		if strings.HasPrefix(impPath, modulePrefix) {
			sub := strings.TrimPrefix(impPath, modulePrefix)
			if strings.HasPrefix(sub, "internal/infrastructure") || strings.HasPrefix(sub, "cmd/") {
				t.Errorf("[Hexagonal Violation] %s (driving adapter) must not import driven infrastructure or cmd (%s)", relFile, impPath)
			}
		}

	case "infrastructure":
		if strings.HasPrefix(impPath, modulePrefix) {
			sub := strings.TrimPrefix(impPath, modulePrefix)
			if strings.HasPrefix(sub, "internal/adapter") ||
				strings.HasPrefix(sub, "internal/composition") ||
				strings.HasPrefix(sub, "cmd/") ||
				(strings.HasPrefix(sub, "internal/usecase") && sub != "internal/usecase/ports") {
				t.Errorf("[Hexagonal Violation] %s (driven infrastructure) must only depend on domain, usecase/ports, or platform; found %s", relFile, impPath)
			}
		}
	}
}

func checkForbiddenCalls(t *testing.T, fset *token.FileSet, relFile, layer string, fileAst *ast.File) {
	t.Helper()

	ast.Inspect(fileAst, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.SelectorExpr:
			if ident, ok := node.X.(*ast.Ident); ok {
				sel := ident.Name + "." + node.Sel.Name
				switch sel {
				case "os.Stdin", "term.ReadPassword":
					pos := fset.Position(node.Pos())
					t.Errorf("[I/O Boundary Violation] %s:%d (%s) must not read interactive stdin (%s)", relFile, pos.Line, layer, sel)
				case "os.Stdout":
					pos := fset.Position(node.Pos())
					t.Errorf("[I/O Boundary Violation] %s:%d (%s) must not reference os.Stdout directly; use progress.Sink", relFile, pos.Line, layer)
				}
			}
		case *ast.CallExpr:
			if selExpr, ok := node.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selExpr.X.(*ast.Ident); ok {
					callName := ident.Name + "." + selExpr.Sel.Name
					if callName == "os.Setenv" {
						pos := fset.Position(node.Pos())
						t.Errorf("[State Boundary Violation] %s:%d (%s) must not mutate process-wide environment via os.Setenv", relFile, pos.Line, layer)
					}
					if (layer == "domain" || layer == "usecase") && strings.HasPrefix(callName, "fmt.Print") {
						pos := fset.Position(node.Pos())
						t.Errorf("[I/O Boundary Violation] %s:%d (%s) must not print directly to stdout via %s", relFile, pos.Line, layer, callName)
					}
				}
			}
		}
		return true
	})
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate go.mod")
		}
		dir = parent
	}
}
