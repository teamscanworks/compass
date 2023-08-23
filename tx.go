package compass

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/types"
	ctypes "github.com/cosmos/cosmos-sdk/types"
)

// Returns an array of unconfirmed transactions present in the mempool
func (c *Client) UnconfirmedTransactions(ctx context.Context, limit *int) ([]types.Tx, error) {
	txns, err := c.RPC.UnconfirmedTxs(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unconfirmed transactions %+v", err)
	}
	return txns.Txs, nil
}

func (c *Client) DeserializeTransactions(txs []types.Tx) ([]ctypes.Tx, error) {
	var out = make([]ctypes.Tx, 0, len(txs))
	for _, tx := range txs {
		decodedTx, err := c.Codec.TxConfig.TxDecoder()(tx)
		if err != nil {
			// remove after testing
			fmt.Printf("failed to decode tx %+v\n", err)
			continue
		}
		out = append(out, decodedTx)
	}
	return out, nil
}
