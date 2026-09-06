package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectClientEnvironment(t *testing.T) {
	t.Setenv("FRPPC_ID", "old-id")
	t.Setenv("FRPPC_SECRET", "old-secret")
	t.Setenv("FRPPC_RPC_URL", "grpc://old:9001")

	handled, err := injectClientEnvironment([]string{
		"-s", "e20e5234-d95e-42b4-ad16-05c05489e663",
		"-i", "xsilen.c.dev",
		"--rpc-url", "grpc://localhost:9001",
	})

	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "xsilen.c.dev", os.Getenv("FRPPC_ID"))
	assert.Equal(t, "e20e5234-d95e-42b4-ad16-05c05489e663", os.Getenv("FRPPC_SECRET"))
	assert.Equal(t, "grpc://localhost:9001", os.Getenv("FRPPC_RPC_URL"))
}

func TestInjectClientEnvironmentRequiresCredentials(t *testing.T) {
	handled, err := injectClientEnvironment([]string{"-i", "xsilen.c.dev"})

	assert.True(t, handled)
	assert.EqualError(t, err, "client: -s is required")
}

func TestInjectClientEnvironmentDelegatesArtisanCommands(t *testing.T) {
	handled, err := injectClientEnvironment([]string{"artisan", "about"})

	require.NoError(t, err)
	assert.False(t, handled)
}
