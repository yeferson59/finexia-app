package app

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// modulePath is the import prefix every internal package shares.
const modulePath = "github.com/yeferson59/finexia-app/internal/"

// domainSegments are the first path segments that identify a domain module or
// the composition root — anything that carries business logic or wiring, as
// opposed to the technical shared kernel (platform) and the shared type leaf
// (identity).
var domainSegments = map[string]bool{
	"app": true, "auth": true, "user": true, "portfolio": true,
	"market": true, "marketing": true, "notification": true,
	"scheduler": true, "health": true, "routes": true,
	"handlers": true, "services": true, "repositories": true,
	"entities": true, "dtos": true, "middlewares": true,
}

// internalImports parses every non-test .go file under dir and returns the set
// of internal (finexia-app/internal/...) packages it imports, keyed by the
// path segment right after internal/.
func internalImports(t *testing.T, dir string) map[string][]string {
	t.Helper()
	// Tests run with the package dir (internal/app) as the working directory,
	// so internal/ is one level up.
	root := filepath.Join("..", dir)
	byImporter := map[string][]string{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		for _, spec := range f.Imports {
			imp, uerr := strconv.Unquote(spec.Path.Value)
			if uerr != nil {
				return uerr
			}
			if strings.HasPrefix(imp, modulePath) {
				byImporter[path] = append(byImporter[path], strings.TrimPrefix(imp, modulePath))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return byImporter
}

func firstSegment(pkg string) string {
	if i := strings.IndexByte(pkg, '/'); i >= 0 {
		return pkg[:i]
	}
	return pkg
}

// TestPlatformStaysAKernel asserts platform/* never reaches into a domain
// module, identity or the composition root: the shared kernel must not know
// the business.
func TestPlatformStaysAKernel(t *testing.T) {
	for file, imports := range internalImports(t, "platform") {
		for _, imp := range imports {
			seg := firstSegment(imp)
			if domainSegments[seg] || seg == "identity" {
				t.Errorf("%s imports internal/%s: platform must not depend on domain code", file, imp)
			}
		}
	}
}

// TestIdentityStaysALeaf asserts the shared identity types import nothing else
// internal: it is the leaf every module may safely depend on.
func TestIdentityStaysALeaf(t *testing.T) {
	for file, imports := range internalImports(t, "identity") {
		for _, imp := range imports {
			t.Errorf("%s imports internal/%s: identity must stay a dependency-free leaf", file, imp)
		}
	}
}

// TestNothingImportsCompositionRoot asserts no module reaches back into
// internal/app: wiring flows one way, from app down into the modules.
func TestNothingImportsCompositionRoot(t *testing.T) {
	for _, dir := range []string{
		"auth", "user", "portfolio", "market", "marketing",
		"notification", "scheduler", "health", "platform", "identity",
	} {
		for file, imports := range internalImports(t, dir) {
			for _, imp := range imports {
				if firstSegment(imp) == "app" {
					t.Errorf("%s imports internal/app: modules must not depend on the composition root", file)
				}
			}
		}
	}
}
