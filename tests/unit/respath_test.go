package unit

import (
	"strings"
	"testing"

	"sonora-cli/internal/cli/respath"
)

func TestParse_CanonicalNamesOnly(t *testing.T) {
	cases := []struct {
		name string
		arg  string
		want respath.ResourceKind
	}{
		{"inputs", "inputs", respath.Inputs},
		{"outputs", "outputs", respath.Outputs},
		{"groups", "groups", respath.Groups},
		{"routes", "routes", respath.Routes},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := respath.Parse(c.arg)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v, want nil", c.arg, err)
			}
			if p.Kind != c.want {
				t.Errorf("Parse(%q).Kind = %v, want %v", c.arg, p.Kind, c.want)
			}
			if p.ID != "" {
				t.Errorf("Parse(%q).ID = %q, want empty (collection form)", c.arg, p.ID)
			}
		})
	}
}

func TestParse_WithID(t *testing.T) {
	p, err := respath.Parse("outputs/office-speaker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Kind != respath.Outputs {
		t.Errorf("Kind = %v, want Outputs", p.Kind)
	}
	if p.ID != "office-speaker" {
		t.Errorf("ID = %q, want office-speaker", p.ID)
	}
}

func TestParse_IDPattern_ValidCharacters(t *testing.T) {
	for _, id := range []string{"a", "A1_-z", "123", "abc-DEF_123"} {
		p, err := respath.Parse("outputs/" + id)
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", id, err)
		}
		if p.ID != id {
			t.Errorf("ID = %q, want %q", p.ID, id)
		}
	}
}

func TestParse_IDPattern_InvalidCharacters(t *testing.T) {
	for _, id := range []string{"has space", "has.dot", "has#hash", "has$dollar"} {
		if _, err := respath.Parse("outputs/" + id); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", id)
		}
	}
}

func TestParse_ExtraSlashIsMalformed(t *testing.T) {
	if _, err := respath.Parse("out/foo/bar"); err == nil {
		t.Error("expected error for id containing an extra '/', got nil")
	}
}

func TestParse_EmptyIDAfterSlashIsMalformed(t *testing.T) {
	if _, err := respath.Parse("outputs/"); err == nil {
		t.Error("expected error for empty id segment, got nil")
	}
}

func TestParse_IDTooLong(t *testing.T) {
	long := strings.Repeat("a", 256)
	if _, err := respath.Parse("outputs/" + long); err == nil {
		t.Error("expected error for id exceeding 255 characters, got nil")
	}
}

func TestParse_UnrecognizedResource(t *testing.T) {
	if _, err := respath.Parse("bogus"); err == nil {
		t.Error("expected error for unrecognized resource, got nil")
	}
	if _, err := respath.Parse("bogus/some-id"); err == nil {
		t.Error("expected error for unrecognized resource with id, got nil")
	}
}

func TestParse_AliasesResolveToSameKind(t *testing.T) {
	cases := []struct {
		alias     string
		canonical string
		wantKind  respath.ResourceKind
	}{
		{"in", "inputs", respath.Inputs},
		{"out", "outputs", respath.Outputs},
		{"gr", "groups", respath.Groups},
		{"rt", "routes", respath.Routes},
	}
	for _, c := range cases {
		t.Run(c.alias, func(t *testing.T) {
			aliasPath, err := respath.Parse(c.alias)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v, want nil", c.alias, err)
			}
			if aliasPath.Kind != c.wantKind {
				t.Errorf("Parse(%q).Kind = %v, want %v", c.alias, aliasPath.Kind, c.wantKind)
			}

			canonicalPath, err := respath.Parse(c.canonical)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v, want nil", c.canonical, err)
			}
			if aliasPath.Kind != canonicalPath.Kind {
				t.Errorf("Parse(%q).Kind = %v, Parse(%q).Kind = %v, want equal", c.alias, aliasPath.Kind, c.canonical, canonicalPath.Kind)
			}
		})
	}
}

func TestParse_AliasesWithID(t *testing.T) {
	cases := []struct {
		alias    string
		wantKind respath.ResourceKind
	}{
		{"in", respath.Inputs},
		{"out", respath.Outputs},
		{"gr", respath.Groups},
		{"rt", respath.Routes},
	}
	for _, c := range cases {
		t.Run(c.alias, func(t *testing.T) {
			p, err := respath.Parse(c.alias + "/some-id")
			if err != nil {
				t.Fatalf("Parse(%q) error = %v, want nil", c.alias+"/some-id", err)
			}
			if p.Kind != c.wantKind {
				t.Errorf("Kind = %v, want %v", p.Kind, c.wantKind)
			}
			if p.ID != "some-id" {
				t.Errorf("ID = %q, want some-id", p.ID)
			}
		})
	}
}

func TestNames_EnumeratesCanonicalResources(t *testing.T) {
	names := respath.Names()
	want := map[string]bool{"inputs": true, "outputs": true, "groups": true, "routes": true}
	if len(names) != len(want) {
		t.Fatalf("Names() = %v, want 4 entries", names)
	}
	for _, n := range names {
		if !want[n] {
			t.Errorf("Names() contains unexpected entry %q", n)
		}
	}
}
