package compass

import (
	"fmt"
	"os"
	"sync"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	cclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Client provides a lightweight RPC/gRPC client for interacting with the cosmos blockchain, and is a fork of https://github.com/strangelove-ventures/lens
type Client struct {
	log     *zap.Logger
	cfg     *ClientConfig
	RPC     *rpchttp.HTTP
	GRPC    *grpc.ClientConn
	Factory tx.Factory
	Keyring keyring.Keyring

	Codec Codec

	initFn  sync.Once
	closeFn sync.Once
}

// Returns a new compass client used to interact with the cosmos blockchain
func NewClient(log *zap.Logger, cfg *ClientConfig, keyringOptions []keyring.Option) (*Client, error) {
	logger := log.Named("compass")
	rpc := &Client{
		log:   logger,
		cfg:   cfg,
		Codec: MakeCodec(cfg.Modules, []string{}),
	}
	return rpc, rpc.Initialize(keyringOptions)
}

// Closes internal clients used by compass
func (c *Client) Close() error {
	var closeErr error = fmt.Errorf("client already close")
	c.closeFn.Do(func() {
		if closeErr = c.GRPC.Close(); closeErr != nil {
			return
		}
		closeErr = nil

	})
	return closeErr
}

// Initializes the compass client, and should be called immediately after instantiation
func (c *Client) Initialize(keyringOptions []keyring.Option) error {
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
		keyInfo, err := keyring.New(c.cfg.ChainID, c.cfg.KeyringBackend, c.cfg.KeyDirectory, os.Stdin, c.Codec.Marshaler, keyringOptions...)
		if err != nil {
			initErr = fmt.Errorf("failed to initialize keyring %s", err)
			return
		}
		c.Keyring = keyInfo
		rpc, err := cclient.NewClientFromNode(c.cfg.RPCAddr)
		if err != nil {
			initErr = fmt.Errorf("failed to construct client from node %s", err)
			return
		}
		c.RPC = rpc
		grpcConn, err := grpc.Dial(
			c.cfg.GRPCAddr,      // your gRPC server address.
			grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism
			//grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
		)
		if err != nil {
			initErr = fmt.Errorf("failed to dial grpc server node %s", err)
			return
		}
		c.GRPC = grpcConn
		factory, err := tx.NewFactoryCLI(c.ClientContext(), pflag.NewFlagSet("", pflag.ExitOnError))
		if err != nil {
			initErr = fmt.Errorf("failed to initialize tx factory %s", err)
			return
		}
		signOpts, err := authtx.NewDefaultSigningOptions()
		if err != nil {
			initErr = fmt.Errorf("failed to get tx opts %s", err)
			return
		}
		txCfg, err := authtx.NewTxConfigWithOptions(c.Codec.Marshaler, authtx.ConfigOptions{
			SigningOptions: signOpts,
		})
		if err != nil {
			initErr = fmt.Errorf("failed to initialize tx config %s", err)
			return
		}
		c.Factory = c.configTxFactory(factory.WithTxConfig(txCfg))

		c.log.Info("initialized client")
	})

	return initErr
}

// Returns previously initialized transaction factory
func (c *Client) TxFactory() tx.Factory {
	return c.Factory
}

// Returns an instance of client.Context, used widely throughout cosmos-sdk
func (c *Client) ClientContext() client.Context {
	return client.Context{}.
		WithViper("breaker").
		WithAccountRetriever(c).
		WithChainID(c.cfg.ChainID).
		WithKeyring(c.Keyring).
		WithGRPCClient(c.GRPC).
		WithClient(c.RPC).
		WithCodec(c.Codec.Marshaler).
		WithInterfaceRegistry(c.Codec.InterfaceRegistry)
}

func (c *Client) configTxFactory(input tx.Factory) tx.Factory {
	return input.
		WithAccountRetriever(c).
		WithChainID(c.cfg.ChainID).
		WithGasAdjustment(c.cfg.GasAdjustment).
		WithGasPrices(c.cfg.GasPrices).
		WithKeybase(c.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
}
