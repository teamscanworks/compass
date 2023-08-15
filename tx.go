package compass

import (
	"fmt"

	"cosmossdk.io/x/tx/decode"
	"github.com/cometbft/cometbft/types"
)

// Returns an array of unconfirmed transactions present in the mempool
func (c *Client) UnconfirmedTransactions(limit *int) ([]types.Tx, error) {
	txns, err := c.RPC.UnconfirmedTxs(c.cctx.CmdContext, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unconfirmed transactions %+v", err)
	}
	return txns.Txs, nil
}

func DeserializeTransactions(txs []types.Tx) ([]decode.DecodedTx, error) {
	decoder, err := decode.NewDecoder(decode.Options{})
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
