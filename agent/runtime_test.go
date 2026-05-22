package agent

import (
	"testing"

	"github.com/henrygd/beszel/agent/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyRuntimeConfig(t *testing.T) {
	t.Setenv("HUB_URL", "")
	t.Setenv("TOKEN", "")
	t.Setenv("KEY", "")

	err := ApplyRuntimeConfig(RuntimeConfig{
		HubURL: "https://hub.example",
		Token:  "tok",
		Key:    "ssh-ed25519 AAA",
	})
	require.NoError(t, err)

	hub, _ := utils.GetEnv("HUB_URL")
	assert.Equal(t, "https://hub.example", hub)
}

func TestApplyRuntimeConfigInvalidURL(t *testing.T) {
	err := ApplyRuntimeConfig(RuntimeConfig{
		HubURL: "not-a-url",
		Token:  "t",
		Key:    "k",
	})
	assert.Error(t, err)
}
