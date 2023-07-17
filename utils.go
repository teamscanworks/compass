package compass

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	libclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
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

// Broadcasts a transaction, returning the transaction hash. This is is not thread-safe, as the factory can
// return a sequence number that is behind X sequences depending on the amount of transactions that have been sent
// and the time to confirm transactions.
//
// To be as safe as possible it's recommended the caller use `SendTransaction`
func (c *Client) BroadcastTx(ctx context.Context, msgs ...sdk.Msg) (string, error) {
	factory, err := c.factory.Prepare(c.cctx)
	if err != nil {
		return "", fmt.Errorf("failed to prepare transaction %s", err)
	}

	unsignedTx, err := factory.BuildUnsignedTx(msgs...)
	if err != nil {
		return "", fmt.Errorf("failed to build unsigned transaction %s", err)
	}

	if err := tx.Sign(c.cctx.CmdContext, factory, c.cctx.GetFromName(), unsignedTx, true); err != nil {
		return "", fmt.Errorf("failed to sign transaction %s", err)
	}

	txBytes, err := c.cctx.TxConfig.TxEncoder()(unsignedTx.GetTx())
	if err != nil {
		return "", fmt.Errorf("failed to get transaction encoder %s", err)
	}

	res, err := c.cctx.BroadcastTx(txBytes)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction %s", err)
	}

	txBytes, err = hex.DecodeString(res.TxHash)
	if err != nil {
		return "", fmt.Errorf("failed to decode string %s", err)
	}

	// allow up to 10 seconds for the transaction to be confirmed before bailing
	//
	// TODO: allow this to be configurable
	exitTicker := time.After(time.Second * 10)
	checkTicker := time.NewTicker(time.Second)
	defer checkTicker.Stop()
	for {
		select {
		case <-exitTicker:
			return "", fmt.Errorf("failed to confirm transaction")
		case <-checkTicker.C:
			if _, err := c.cctx.Client.Tx(ctx, txBytes, false); err != nil {
				continue
			}
			return res.TxHash, nil
		}
	}
}
