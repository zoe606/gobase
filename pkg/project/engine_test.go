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
	for _, engine := range []project.Engine{project.Gin, project.Stdlib, project.Fiber} {
		require.NoError(t, os.WriteFile(filepath.Join(root, project.ConfigFile), []byte(`{"engine":"`+string(engine)+`"}`), 0o600))
		settings, err := project.Read(root)
		require.NoError(t, err)
		require.Equal(t, engine, settings.Engine)
	}
	for _, data := range []string{`{`, `{"engine":"unknown"}`, `{}`} {
		require.NoError(t, os.WriteFile(filepath.Join(root, project.ConfigFile), []byte(data), 0o600))
		_, err := project.Read(root)
		require.Error(t, err)
	}
	require.NoError(t, os.Remove(filepath.Join(root, project.ConfigFile)))
	require.NoError(t, os.Mkdir(filepath.Join(root, project.ConfigFile), 0o700))
	_, err = project.Read(root)
	require.Error(t, err)
}
