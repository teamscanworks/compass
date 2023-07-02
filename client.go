package compass

import (
	"fmt"
	"sync"

	cclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"google.golang.org/grpc"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"go.uber.org/zap"
)

type Client struct {
	log  *zap.Logger
	cfg  *ClientConfig
	RPC  *rpchttp.HTTP
	GRPC *grpc.ClientConn

	initFn sync.Once
}

func NewClient(log *zap.Logger, cfg *ClientConfig) (*Client, error) {
	logger := log.Named("compass")
	rpc := &Client{
		log: logger,
		cfg: cfg,
	}
	return rpc, rpc.Initialize()
}

func (c *Client) Initialize() error {
	var initErr error = nil
	c.initFn.Do(func() {
		if c.log == nil {
			initErr = fmt.Errorf("invalid client object: no logger")
			return
		}
		if c.cfg == nil {
			initErr = fmt.Errorf("invalid client object: no config")
			return
		}
		rpc, err := cclient.NewClientFromNode(c.cfg.RPCAddr)
		if err != nil {
			initErr = fmt.Errorf("failed to construct client from node %s", err)
			return
		}

		grpcConn, err := grpc.Dial(
			c.cfg.GRPCAddr,      // your gRPC server address.
			grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
			// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registr
			// if the request/response types contain interface instead of 'nil' you should pass the application spe
			grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
		)
		if err != nil {
			initErr = fmt.Errorf("failed to dial grpc server node %s", err)
			return
		}

		c.RPC = rpc
		c.GRPC = grpcConn
	})

	return initErr
}
