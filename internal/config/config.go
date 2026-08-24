// Package config resolves CLI configuration, including the hub base URL.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultHubURL = "http://localhost:8080"

// ResolveHubURL resolves the hub base URL using, in order of precedence:
// the --hub-url flag value (flagVal), the MULTIROOM_URL environment
// variable, the hubUrl field in ~/.config/sonora/config.json, and finally
// the built-in default. The config file is read lazily and its absence is
// not an error; a malformed file or a non-string hubUrl field is a usage
// error naming the file.
func ResolveHubURL(flagVal string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if envVal := os.Getenv("MULTIROOM_URL"); envVal != "" {
		return envVal, nil
	}

	path, err := configFilePath()
	if err != nil {
		return defaultHubURL, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultHubURL, nil
		}
		return "", fmt.Errorf("reading config file %s: %w", path, err)
	}

	var raw struct {
		HubURL json.RawMessage `json:"hubUrl"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("config file %s is not valid JSON: %w", path, err)
	}
	if raw.HubURL == nil {
		return defaultHubURL, nil
	}

	var hubURL string
	if err := json.Unmarshal(raw.HubURL, &hubURL); err != nil {
		return "", fmt.Errorf("config file %s: hubUrl must be a string", path)
	}
	if hubURL == "" {
		return defaultHubURL, nil
	}
	return hubURL, nil
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "sonora", "config.json"), nil
}
