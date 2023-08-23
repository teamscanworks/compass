package compass_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/require"
	"github.com/teamscanworks/compass"
	"go.uber.org/zap"
)

type AutoGenerated struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		NTxs       string   `json:"n_txs"`
		Total      string   `json:"total"`
		TotalBytes string   `json:"total_bytes"`
		Txs        []string `json:"txs"`
	} `json:"result"`
}

func TestDeserializeUnconfirmedTxCosmosHub(t *testing.T) {
	data, err := ioutil.ReadFile("unconfirmed_txs.json")
	require.NoError(t, err)
	var jOut AutoGenerated
	require.NoError(t, json.Unmarshal(data, &jOut))

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	cfg := compass.GetCosmosHubConfig("keys", true)
	require.NotNil(t, cfg)
	client, err := compass.NewClient(logger, cfg, []keyring.Option{compass.DefaultSignatureOptions()})
	require.NoError(t, err)
	txs, err := client.UnconfirmedTransactions(context.Background(), nil)
	require.NoError(t, err)

	decodedTxs, err := client.DeserializeTransactions(txs, true)
	require.NoError(t, err)
	for _, tx := range decodedTxs {
		t.Log("tx ", tx)
	}
}

func TestDeserializeUnconfirmedTxStride(t *testing.T) {
	data, err := ioutil.ReadFile("unconfirmed_txs.json")
	require.NoError(t, err)
	var jOut AutoGenerated
	require.NoError(t, json.Unmarshal(data, &jOut))

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	cfg := compass.GetStrideConfig()
	require.NotNil(t, cfg)
	client, err := compass.NewClient(logger, cfg, []keyring.Option{compass.DefaultSignatureOptions()})
	require.NoError(t, err)
	txs, err := client.UnconfirmedTransactions(context.Background(), nil)
	require.NoError(t, err)
	// if no txs found, read from test file
	if len(txs) == 0 {
		t.Log("padding")
		for _, tx := range jOut.Result.Txs {
			txs = append(txs, []byte(tx))
		}
	}

	decodedTxs, err := client.DeserializeTransactions(txs, true)
	require.NoError(t, err)
	for _, tx := range decodedTxs {
		t.Log("tx ", tx)
	}
	panic("k")
}
