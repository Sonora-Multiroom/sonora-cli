package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sonora-cli/internal/config"
)

// setHome points both HOME (Unix) and USERPROFILE (Windows) at dir so
// os.UserHomeDir() resolves to it regardless of platform.
func setHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func writeConfigFile(t *testing.T, home, contents string) string {
	t.Helper()
	dir := filepath.Join(home, ".config", "sonora")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestResolveHubURL_FlagTakesPrecedence(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	writeConfigFile(t, home, `{"hubUrl": "http://config.example:8080"}`)
	t.Setenv("MULTIROOM_URL", "http://env.example:8080")

	got, err := config.ResolveHubURL("http://flag.example:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://flag.example:8080" {
		t.Errorf("got %q, want flag value", got)
	}
}

func TestResolveHubURL_EnvUsedWhenNoFlag(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	writeConfigFile(t, home, `{"hubUrl": "http://config.example:8080"}`)
	t.Setenv("MULTIROOM_URL", "http://env.example:8080")

	got, err := config.ResolveHubURL("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://env.example:8080" {
		t.Errorf("got %q, want env value", got)
	}
}

func TestResolveHubURL_ConfigFileUsedWhenNoFlagOrEnv(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	writeConfigFile(t, home, `{"hubUrl": "http://config.example:8080"}`)

	got, err := config.ResolveHubURL("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://config.example:8080" {
		t.Errorf("got %q, want config file value", got)
	}
}

func TestResolveHubURL_DefaultWhenNothingSet(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	got, err := config.ResolveHubURL("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://localhost:8080" {
		t.Errorf("got %q, want built-in default", got)
	}
}

func TestResolveHubURL_MissingConfigFileIsNotAnError(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	// No config.json written at all.

	got, err := config.ResolveHubURL("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "http://localhost:8080" {
		t.Errorf("got %q, want built-in default", got)
	}
}

func TestResolveHubURL_MalformedConfigFileIsUsageError(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	path := writeConfigFile(t, home, `{not valid json`)

	_, err := config.ResolveHubURL("")
	if err == nil {
		t.Fatal("expected an error for malformed config file, got nil")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("expected error to name the config file path %q, got: %v", path, err)
	}
}

func TestResolveHubURL_NonStringHubURLIsUsageError(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)
	path := writeConfigFile(t, home, `{"hubUrl": 12345}`)

	_, err := config.ResolveHubURL("")
	if err == nil {
		t.Fatal("expected an error for non-string hubUrl, got nil")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("expected error to name the config file path %q, got: %v", path, err)
	}
}
