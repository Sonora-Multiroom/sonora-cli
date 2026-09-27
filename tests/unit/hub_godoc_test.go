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

// specArtifactRefPattern flags a comment that cites this repo's spec
// artifacts — research.md, data-model.md, spec.md, requirement/success-
// criteria IDs, section markers, constitution principles, or a NNN-feature
// directory name — which would leave hub/ unreadable to a consumer without
// this repo's specs (constitution Principle VII).
var specArtifactRefPattern = regexp.MustCompile(`research\.md|data-model\.md|spec\.md|\bFR-\d|\bSC-\d|§|[Pp]rinciple [IVX]+|[Cc]onstitution|\b0\d\d-[a-z]`)

// docExemptMethods implement standard interfaces (error, fmt.Stringer) or
// follow Go convention (Unwrap) and are exempt from the doc-comment
// requirement below.
var docExemptMethods = map[string]bool{"Error": true, "Unwrap": true, "String": true}

// TestHubGodoc parses every non-test .go file in hub/ and enforces
// constitution Principle VII's godoc rules: a package doc comment, a doc
// comment on every exported identifier (a doc on the enclosing const/var
// group counts for its members), no comment anywhere in the package citing
// this repo's spec artifacts, and standard-library-only imports.
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
		docPkg := doc.New(pkg, "./", doc.AllDecls|doc.PreserveAST)

		if strings.TrimSpace(docPkg.Doc) == "" {
			report(fset.Position(pkg.Pos()), "package %s has no package doc comment", pkgName)
		}

		checkDoc := func(name, doc string, pos token.Pos) {
			if strings.TrimSpace(doc) == "" {
				report(fset.Position(pos), "exported %s has no doc comment", name)
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
			for _, name := range v.Names {
				if ast.IsExported(name) && strings.TrimSpace(v.Doc) == "" {
					report(fset.Position(v.Decl.Pos()), "exported const %s has no doc comment (on itself or its group)", name)
				}
			}
		}
		for _, v := range docPkg.Vars {
			for _, name := range v.Names {
				if ast.IsExported(name) && strings.TrimSpace(v.Doc) == "" {
					report(fset.Position(v.Decl.Pos()), "exported var %s has no doc comment (on itself or its group)", name)
				}
			}
		}

		for _, f := range pkg.Files {
			// Every comment in the file — doc comments, inline comments, on
			// exported or unexported code alike — must read sensibly to a
			// consumer with no access to this repo's specs.
			for _, cg := range f.Comments {
				if specArtifactRefPattern.MatchString(cg.Text()) {
					report(fset.Position(cg.Pos()), "comment cites a spec artifact")
				}
			}
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
