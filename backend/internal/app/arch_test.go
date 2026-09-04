package app

import (
	"go/ast"
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
	"mcp": true, "scheduler": true, "health": true, "routes": true,
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

			if after, found := strings.CutPrefix(imp, modulePath); found {
				byImporter[path] = append(byImporter[path], after)
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
	if before, _, found := strings.Cut(pkg, "/"); found {
		return before
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

// TestModulesOwnTheirConfig asserts no domain module imports platform/config:
// reading the environment is the composition root's job, and each module
// declares its own small Config struct that app populates. internal/migrator is
// exempt — it is a second entrypoint, not a module.
func TestModulesOwnTheirConfig(t *testing.T) {
	for _, dir := range []string{
		"auth", "user", "portfolio", "market", "marketing",
		"notification", "mcp", "scheduler", "health",
	} {
		for file, imports := range internalImports(t, dir) {
			for _, imp := range imports {
				if imp == "platform/config" {
					t.Errorf("%s imports internal/platform/config: modules declare their own Config, populated by internal/app", file)
				}
			}
		}
	}
}

// TestServiceFirstModulesStayIndependent asserts that user and marketing
// import no other domain module. That is what lets the composition root build
// their services first and hand them to auth as ordinary constructor
// arguments: auth reads users/roles through user.Service and advances the
// waitlist through marketing.Service, so if either grew a dependency back on
// auth the graph would be cyclic again and the wiring would need a setter.
//
// The routes those two modules serve still need auth's guards — they take them
// through the authMiddleware interface, in the second construction step, which
// costs no import.
func TestServiceFirstModulesStayIndependent(t *testing.T) {
	for _, dir := range []string{"user", "marketing"} {
		for file, imports := range internalImports(t, dir) {
			for _, imp := range imports {
				seg := firstSegment(imp)
				if seg != dir && domainSegments[seg] {
					t.Errorf("%s imports internal/%s: user and marketing must depend on no other domain module, so their services can be built before auth", file, imp)
				}
			}
		}
	}
}

// TestNothingImportsCompositionRoot asserts no module reaches back into
// internal/app: wiring flows one way, from app down into the modules.
func TestNothingImportsCompositionRoot(t *testing.T) {
	for _, dir := range []string{
		"auth", "user", "portfolio", "market", "marketing",
		"notification", "mcp", "scheduler", "health", "platform", "identity",
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

// serviceAccessorCalls parses every non-test .go file under dir and returns the
// zero-argument `x.Service()` calls it contains, keyed by file.
func serviceAccessorCalls(t *testing.T, dir string) map[string][]string {
	t.Helper()
	root := filepath.Join("..", dir)
	byFile := map[string][]string{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 0 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Service" {
				return true
			}
			byFile[path] = append(byFile[path], fset.Position(call.Pos()).String())

			return true
		})

		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}

	return byFile
}

// TestOnlyAppCallsServiceAccessors asserts no domain module calls a Module's
// Service() accessor. That accessor hands out the concrete *service, which is
// the one way to step around the consumer-defined interfaces every cross-module
// dependency goes through: portfolio already imports market for the Asset type,
// so nothing but this test stops it from holding a *market.Module and reaching
// past its own AssetReader to the whole service.
//
// Only the composition root may call it, because handing those services to the
// interfaces that consume them is precisely its job.
func TestOnlyAppCallsServiceAccessors(t *testing.T) {
	for _, dir := range []string{
		"auth", "user", "portfolio", "market", "marketing",
		"notification", "mcp", "scheduler", "health", "platform", "identity",
	} {
		for file, calls := range serviceAccessorCalls(t, dir) {
			for _, pos := range calls {
				t.Errorf("%s calls Service() at %s: modules consume each other through their own interfaces, never through the concrete service; only internal/app may reach for it", file, pos)
			}
		}
	}
}
