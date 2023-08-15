package decode

import "cosmossdk.io/errors"

const (
	txCodespace = "tx"
)

var (
	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode     = errors.Register(txCodespace, 111, "tx parse error")
	ErrUnknownField = errors.Register(txCodespace, 222, "unknown protobuf field")
)
