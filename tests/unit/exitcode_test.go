package unit

import (
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode"
)

func TestExitcodeUsage(t *testing.T) {
	if exitcode.Usage != 2 {
		t.Errorf("exitcode.Usage = %d, want 2", exitcode.Usage)
	}
}

func TestExitcodeFor_Table(t *testing.T) {
	cases := map[hub.ErrorClass]int{
		hub.ClassNone:               0,
		hub.ClassHub:                3,
		hub.ClassNetwork:            4,
		hub.ClassNotFound:           5,
		hub.ClassValidation:         6,
		hub.ClassRouteFailed:        8,
		hub.ClassSourceUnreachable:  9,
		hub.ClassServiceUnavailable: 10,
		hub.ClassInputNotFound:      11,
		hub.ClassTargetNotFound:     12,
		hub.ClassTTSUnavailable:     13,
	}
	for class, want := range cases {
		if got := exitcode.For(class); got != want {
			t.Errorf("exitcode.For(%v) = %d, want %d", class, got, want)
		}
	}
}

func TestExitcodeFor_Distinct(t *testing.T) {
	classes := []hub.ErrorClass{
		hub.ClassHub, hub.ClassNetwork, hub.ClassNotFound, hub.ClassValidation,
		hub.ClassRouteFailed, hub.ClassSourceUnreachable, hub.ClassServiceUnavailable,
		hub.ClassInputNotFound, hub.ClassTargetNotFound, hub.ClassTTSUnavailable,
	}
	seen := map[int]bool{exitcode.Usage: true}
	for _, c := range classes {
		code := exitcode.For(c)
		if seen[code] {
			t.Errorf("exit code %d reused (class %v)", code, c)
		}
		seen[code] = true
		if code == 7 {
			t.Errorf("class %v reuses retired exit code 7", c)
		}
	}
}
