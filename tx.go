package compass

import (
	"fmt"

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
