package compass

import (
	"context"
	"fmt"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	libclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"go.uber.org/zap"
)

// Returns a Cosmos JSON-RPC websocket client
func NewRPCClient(addr string, timeout time.Duration) (*rpchttp.HTTP, error) {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}
	httpClient.Timeout = timeout
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}
	return rpcClient, nil
}

// returns a keyring.Option that specifies a list of default algorithms
func DefaultSignatureOptions() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = keyring.SigningAlgoList{hd.Secp256k1}
		options.SupportedAlgosLedger = keyring.SigningAlgoList{hd.Secp256k1}
	}
}

// Returns a BIP39 mnemonic using the english language
func CreateMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}

// Broadcasts a transaction, returning the transaction hash
func (c *Client) BroadcastTx(ctx context.Context, msgs ...sdk.Msg) (string, error) {
	factory, err := c.Factory.Prepare(c.CCtx)
	if err != nil {
		return "", fmt.Errorf("failed to prepare transaction %s", err)
	}
	c.log.Debug("building unsigned tx")
	unsignedTx, err := factory.BuildUnsignedTx(msgs...)
	if err != nil {
		c.log.Debug("building tx failed", zap.Error(err))
		return "", fmt.Errorf("failed to build unsigned transaction %s", err)
	}
	c.log.Debug("signing transaction")
	if err := tx.Sign(c.CCtx.CmdContext, factory, c.CCtx.GetFromName(), unsignedTx, true); err != nil {
		c.log.Debug("signing failed", zap.Error(err))
		return "", fmt.Errorf("failed to sign transaction %s", err)
	}
	c.log.Debug("encoding transaction")
	txBytes, err := c.CCtx.TxConfig.TxEncoder()(unsignedTx.GetTx())
	if err != nil {
		c.log.Debug("encoding failed", zap.Error(err))
		return "", fmt.Errorf("failed to get transaction encoder %s", err)
	}
	c.log.Debug("broadcasting transaction")
	res, err := c.CCtx.BroadcastTx(txBytes)
	if err != nil {
		c.log.Debug("broadcast failed", zap.Error(err))
		return "", fmt.Errorf("failed to broadcast transaction %s", err)
	}
	c.log.Debug("confirming transaction", zap.String("tx.hash", res.TxHash))
	exitTicker := time.After(time.Second * 10)
	checkTicker := time.NewTicker(time.Second)
	defer checkTicker.Stop()
	for {
		select {
		case <-exitTicker:
			return "", fmt.Errorf("failed to confirm transaction")
		case <-checkTicker.C:
			txRes, err := c.CCtx.Client.Tx(ctx, []byte(res.TxHash), false)
			if err != nil {
				c.log.Debug("status check failed", zap.Error(err), zap.String("tx.hash", res.TxHash))
				continue
			}
			c.log.Info("confirmed transaction", zap.String("tx.hash", res.TxHash), zap.Any("response", txRes))
			return res.TxHash, nil
		}
	}
}
