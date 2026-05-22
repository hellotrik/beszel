package trayconfig

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigNormalize(t *testing.T) {
	c := Config{HubURL: " https://h.example ", Token: " t ", Key: " k ", Port: ""}
	c.Normalize()
	assert.Equal(t, "https://h.example", c.HubURL)
	assert.Equal(t, DefaultPort, c.Port)
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DATA_DIR", dir)

	cfg := Config{
		HubURL: "https://hub.example",
		Token:  "tok",
		Key:    "ssh-ed25519 AAA",
		Port:   "45876",
	}
	require.NoError(t, Save(cfg))

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.HubURL, loaded.HubURL)
	assert.Equal(t, cfg.Token, loaded.Token)
	assert.Equal(t, cfg.Key, loaded.Key)

	path := filepath.Join(dir, configFileName)
	assert.FileExists(t, path)
}

func TestMigrateFromEnv(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("HUB_URL", "https://hub.test")
	t.Setenv("TOKEN", "abc")
	t.Setenv("KEY", "ssh-ed25519 x")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "https://hub.test", cfg.HubURL)
	assert.Equal(t, "abc", cfg.Token)
}

func TestValidateRequiresHubURL(t *testing.T) {
	cfg := Config{Token: "t", Key: "k"}
	err := cfg.Validate()
	assert.Error(t, err)
}
