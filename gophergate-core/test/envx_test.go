package gophergatecore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Cyrof/GopherGate/gophergate-core/envx"
	"github.com/stretchr/testify/require"
)

func TestIsDev_FromEnvVar(t *testing.T) {
	t.Setenv("GOPHERGATE_ENV", "dev")
	require.True(t, envx.IsDev())

	t.Setenv("GOPHERGATE_ENV", "prod")
	require.False(t, envx.IsDev())

	require.NoError(t, os.Unsetenv("GOPHERGATE_ENV"))
	require.False(t, envx.IsDev())
}

func TestLoadDotenvIfPresent_SetsWhenUnset(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	require.NoError(t, os.WriteFile(filepath.Join(tmp, ".env"), []byte("GOPHERGATE_ENV=dev\n"), 0644))

	require.NoError(t, os.Unsetenv("GOPHERGATE_ENV"))

	envx.LoadDotenvIfPresent()

	require.Equal(t, "dev", os.Getenv("GOPHERGATE_ENV"))
	require.True(t, envx.IsDev())
}

func TestLoadDotenvIfPresent_DoesNotOverrideExisting(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	require.NoError(t, os.WriteFile(filepath.Join(tmp, ".env"), []byte("GOPHERGATE_ENV=dev\n"), 0644))

	t.Setenv("GOPHERGATE_ENV", "prod")

	envx.LoadDotenvIfPresent()

	require.Equal(t, "prod", os.Getenv("GOPHERGATE_ENV"))
	require.False(t, envx.IsDev())
}
