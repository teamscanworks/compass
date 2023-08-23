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

// if hardError == true, deserialization will fail at first error
func (c *Client) DeserializeTransactions(txs []types.Tx, hardError bool) ([]ctypes.Tx, error) {
	var out = make([]ctypes.Tx, 0, len(txs))
	for _, tx := range txs {
		/* this works for some stride transactions but not others however it always works fro cosmos
		var raw v1beta1.TxRaw
		err := proto.UnmarshalOptions{Merge: true, AllowPartial: true, DiscardUnknown: false}.Unmarshal(tx, &raw)
		if err != nil && !hardError {
			fmt.Println("raw unmarshal failed ", err)
		}*/
		//decodedTx, err := authtx.DefaultTxDecoder(c.cctx.Codec)(tx)
		decodedTx, err := c.Codec.TxConfig.TxDecoder()(tx)
		if err != nil && !hardError {
			// remove after testing
			fmt.Printf("failed to decode tx %+v\n", err)
			continue
		} else if err != nil && hardError {
			return nil, fmt.Errorf("decode encountered hard error %+v", err)
		}
		out = append(out, decodedTx)
	}
	return out, nil
}
