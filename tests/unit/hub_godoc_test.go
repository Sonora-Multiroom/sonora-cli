package unit

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// specArtifactRefPattern flags godoc that cites this repo's spec artifacts —
// research.md, data-model.md, spec.md, requirement/success-criteria IDs,
// section markers, constitution principles, or a NNN-feature directory name
// — which would leave hub/ godoc unreadable to a consumer without this
// repo's specs (constitution Principle VII).
var specArtifactRefPattern = regexp.MustCompile(`research\.md|data-model\.md|spec\.md|\bFR-\d|\bSC-\d|§|[Pp]rinciple [IVX]+|[Cc]onstitution|\b0\d\d-[a-z]`)

// docExemptMethods implement standard interfaces (error, fmt.Stringer) or
// follow Go convention (Unwrap) and are exempt from the doc-comment
// requirement below.
var docExemptMethods = map[string]bool{"Error": true, "Unwrap": true, "String": true}

// TestHubGodoc parses every non-test .go file in hub/ and enforces
// constitution Principle VII's godoc rules: a package doc comment, a doc
// comment on every exported identifier (a doc on the enclosing const/var
// group counts for its members), no citation of this repo's spec artifacts,
// and standard-library-only imports.
func TestHubGodoc(t *testing.T) {
	const hubDir = "../../hub"

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, hubDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing %s: %v", hubDir, err)
	}

	var violations []string
	report := func(pos token.Position, format string, args ...any) {
		violations = append(violations, fmt.Sprintf("%s:%d: %s", filepath.Base(pos.Filename), pos.Line, fmt.Sprintf(format, args...)))
	}

	for pkgName, pkg := range pkgs {
		docPkg := doc.New(pkg, "./", doc.AllDecls)

		if strings.TrimSpace(docPkg.Doc) == "" {
			report(fset.Position(pkg.Pos()), "package %s has no package doc comment", pkgName)
		} else if specArtifactRefPattern.MatchString(docPkg.Doc) {
			report(fset.Position(pkg.Pos()), "package doc cites a spec artifact")
		}

		checkDoc := func(name, doc string, pos token.Pos) {
			if strings.TrimSpace(doc) == "" {
				report(fset.Position(pos), "exported %s has no doc comment", name)
				return
			}
			if specArtifactRefPattern.MatchString(doc) {
				report(fset.Position(pos), "doc comment for %s cites a spec artifact", name)
			}
		}

		for _, f := range docPkg.Funcs {
			if ast.IsExported(f.Name) {
				checkDoc(f.Name, f.Doc, f.Decl.Pos())
			}
		}
		for _, tp := range docPkg.Types {
			if ast.IsExported(tp.Name) {
				checkDoc(tp.Name, tp.Doc, tp.Decl.Pos())
			}
			for _, m := range tp.Methods {
				if ast.IsExported(m.Name) && !docExemptMethods[m.Name] {
					checkDoc(tp.Name+"."+m.Name, m.Doc, m.Decl.Pos())
				}
			}
			for _, f := range tp.Funcs {
				if ast.IsExported(f.Name) {
					checkDoc(f.Name, f.Doc, f.Decl.Pos())
				}
			}
		}
		for _, v := range docPkg.Consts {
			if strings.TrimSpace(v.Doc) != "" && specArtifactRefPattern.MatchString(v.Doc) {
				report(fset.Position(v.Decl.Pos()), "doc comment for const group cites a spec artifact")
			}
			for _, name := range v.Names {
				if ast.IsExported(name) && strings.TrimSpace(v.Doc) == "" {
					report(fset.Position(v.Decl.Pos()), "exported const %s has no doc comment (on itself or its group)", name)
				}
			}
		}
		for _, v := range docPkg.Vars {
			if strings.TrimSpace(v.Doc) != "" && specArtifactRefPattern.MatchString(v.Doc) {
				report(fset.Position(v.Decl.Pos()), "doc comment for var group cites a spec artifact")
			}
			for _, name := range v.Names {
				if ast.IsExported(name) && strings.TrimSpace(v.Doc) == "" {
					report(fset.Position(v.Decl.Pos()), "exported var %s has no doc comment (on itself or its group)", name)
				}
			}
		}

		for _, f := range pkg.Files {
			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				first, _, _ := strings.Cut(path, "/")
				if strings.Contains(first, ".") {
					report(fset.Position(imp.Pos()), "non-stdlib import %q", path)
				}
			}
		}
	}

	sort.Strings(violations)
	for _, v := range violations {
		t.Error(v)
	}
}
