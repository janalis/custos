package config

import (
	"encoding/json"
	"errors"
	"os"

	"custos/internal/platform/safeio"
)

// LoadConfined loads settings at exactly root, without ancestor discovery or
// following links outside the root. The caller owns the root handle.
func LoadConfined(root *os.Root) (*Config, error) {
	read := func(name string) ([]byte, error) { return safeio.ReadRootFile(root, name, MaxFileSize) }

	var settings File
	b, err := read(FileName)
	if err == nil {
		if err := json.Unmarshal(b, &settings); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	// Validate Composer access even when the chosen PHP version is explicit.
	if _, err := read("composer.json"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return resolve(root.Name(), settings, read)
}

// ResolveCaptured resolves settings from a captured input manifest.
func ResolveCaptured(root string, sources map[string][]byte) (*Config, error) {
	var settings File
	if b := sources[FileName]; b != nil {
		if err := json.Unmarshal(b, &settings); err != nil {
			return nil, err
		}
	}
	return resolve(root, settings, func(name string) ([]byte, error) {
		b, ok := sources[name]
		if !ok {
			return nil, os.ErrNotExist
		}
		return b, nil
	})
}
