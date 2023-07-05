package compass_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamscanworks/compass"
)

func TestKeysDir(t *testing.T) {
	t.Run("Simd", func(t *testing.T) {
		cfg := compass.GetSimdConfig()
		require.Equal(t, cfg.KeyDirectory, "keyring-test/keys/cosmoshub-4")
	})
}
