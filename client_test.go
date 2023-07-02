package compass_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamscanworks/compass"
	"go.uber.org/zap"
)

func TestClient(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	cfg := compass.GetSimdConfig()
	require.NotNil(t, cfg)
	client, err := compass.NewClient(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, client.GRPC)
	require.NotNil(t, client.RPC)
	require.NotNil(t, client.Keyring)
	require.NotNil(t, client.Codec)
	abcInfo, err := client.RPC.ABCIInfo(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, abcInfo.Response.LastBlockHeight, int64(1))
}
