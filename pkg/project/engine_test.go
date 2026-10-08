package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"go-boilerplate/pkg/project"

	"github.com/stretchr/testify/require"
)

func TestSettings(t *testing.T) {
	root := t.TempDir()
	settings, err := project.Read(root)
	require.NoError(t, err)
	require.Equal(t, project.Fiber, settings.Engine)
	require.Equal(t, project.Full, settings.Profile)
	for _, engine := range []project.Engine{project.Gin, project.Stdlib, project.Fiber} {
		require.NoError(t, os.WriteFile(filepath.Join(root, project.ConfigFile), []byte(`{"engine":"`+string(engine)+`"}`), 0o600))
		settings, err := project.Read(root)
		require.NoError(t, err)
		require.Equal(t, engine, settings.Engine)
		require.Equal(t, project.Full, settings.Profile)
	}
	for _, profile := range []project.Profile{project.Full, project.Minimal} {
		require.NoError(t, os.WriteFile(filepath.Join(root, project.ConfigFile), []byte(`{"engine":"gin","profile":"`+string(profile)+`"}`), 0o600))
		settings, err := project.Read(root)
		require.NoError(t, err)
		require.Equal(t, profile, settings.Profile)
	}
	for _, data := range []string{`{`, `{"engine":"unknown"}`, `{}`, `{"engine":"gin","profile":"unknown"}`} {
		require.NoError(t, os.WriteFile(filepath.Join(root, project.ConfigFile), []byte(data), 0o600))
		_, err := project.Read(root)
		require.Error(t, err)
	}
	require.NoError(t, os.Remove(filepath.Join(root, project.ConfigFile)))
	require.NoError(t, os.Mkdir(filepath.Join(root, project.ConfigFile), 0o700))
	_, err = project.Read(root)
	require.Error(t, err)
}
