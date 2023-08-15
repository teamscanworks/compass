package compass

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/types"
	"github.com/teamscanworks/compass/decode"
)

// Returns an array of unconfirmed transactions present in the mempool
func (c *Client) UnconfirmedTransactions(ctx context.Context, limit *int) ([]types.Tx, error) {
	txns, err := c.RPC.UnconfirmedTxs(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unconfirmed transactions %+v", err)
	}
	return txns.Txs, nil
}

func (c *Client) DeserializeTransactions(txs []types.Tx) ([]decode.DecodedTx, error) {
	decoder, err := decode.NewDecoder(decode.Options{
		SigningContext: c.Codec.InterfaceRegistry.SigningContext(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to construct decoder %+v", err)
	}
	var out = make([]decode.DecodedTx, len(txs))
	for _, tx := range txs {
		decodedTx, err := decoder.Decode(tx)
		if err != nil {
			// remove after testing
			fmt.Printf("failed to decode tx %+v\n", err)
			continue
		}
		out = append(out, *decodedTx)
	}
	return out, nil
}
