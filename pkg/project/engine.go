// Package project stores the HTTP engine selected when creating a project.
package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Engine identifies a generated project's HTTP integration.
type Engine string

const (
	Gin        Engine = "gin"
	Stdlib     Engine = "stdlib"
	Fiber      Engine = "fiber"
	ConfigFile        = ".gobase.json"
)

// Settings are used by code generation, not application runtime configuration.
type Settings struct {
	Engine Engine `json:"engine"`
}

// Validate rejects engines without a supported integration.
func (e Engine) Validate() error {
	switch e {
	case Gin, Stdlib, Fiber:
		return nil
	default:
		return fmt.Errorf("unsupported HTTP engine %q; choose gin, stdlib, or fiber", e)
	}
}

// Read preserves Fiber generation for existing projects without settings.
func Read(root string) (Settings, error) {
	data, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if errors.Is(err, os.ErrNotExist) {
		return Settings{Engine: Fiber}, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("read project settings: %w", err)
	}
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, fmt.Errorf("parse project settings: %w", err)
	}
	return settings, settings.Engine.Validate()
}
