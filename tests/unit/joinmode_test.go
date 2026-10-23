package unit

import (
	"slices"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/joinmode"
)

func TestJoinModeParse(t *testing.T) {
	tests := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"REPLACE", "REPLACE", true},
		{"mix", "MIX", true},
		{"Duck_Others", "DUCK_OTHERS", true},
		{"duck_others", "DUCK_OTHERS", true},
		{"", "", false},
		{"fade", "", false},
		{"duck-others", "", false},
		{"DUCK OTHERS", "", false},
		{" MIX", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, ok := joinmode.Parse(tt.in)
			if ok != tt.wantOK {
				t.Fatalf("Parse(%q) ok = %v, want %v", tt.in, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("Parse(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestJoinModeValues(t *testing.T) {
	want := []string{hub.JoinModeReplace, hub.JoinModeMix, hub.JoinModeDuckOthers}
	if !slices.Equal(joinmode.Values, want) {
		t.Errorf("Values = %v, want %v", joinmode.Values, want)
	}
}
