package compass

import (
	"fmt"
	"os"
	"sync"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

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
	CCtx    client.Context
	Factory tx.Factory
	Keyring keyring.Keyring

	Codec Codec

	initFn  sync.Once
	closeFn sync.Once

	factoryLock sync.Mutex
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
		)
		if err != nil {
			initErr = fmt.Errorf("failed to dial grpc server node %s", err)
			return
		}
		c.GRPC = grpcConn
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

		c.CCtx = c.configClientContext(client.Context{}.WithTxConfig(txCfg))

		factory, err := tx.NewFactoryCLI(c.ClientContext(), pflag.NewFlagSet("", pflag.ExitOnError))
		if err != nil {
			initErr = fmt.Errorf("failed to initialize tx factory %s", err)
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
	return c.CCtx
}

func (c *Client) configTxFactory(input tx.Factory) tx.Factory {
	return input.
		WithAccountRetriever(c).
		WithChainID(c.cfg.ChainID).
		WithGasAdjustment(c.cfg.GasAdjustment).
		WithGasPrices(c.cfg.GasPrices).
		WithKeybase(c.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithSimulateAndExecute(true)
}

func (c *Client) configClientContext(cctx client.Context) client.Context {
	return cctx.
		//WithViper("breaker").
		WithAccountRetriever(c).
		WithChainID(c.cfg.ChainID).
		WithKeyring(c.Keyring).
		WithGRPCClient(c.GRPC).
		WithClient(c.RPC).WithSignModeStr(signing.SignMode_SIGN_MODE_DIRECT.String()).
		WithCodec(c.Codec.Marshaler).
		WithInterfaceRegistry(c.Codec.InterfaceRegistry) //.WithOutput(os.Stdout)
}

func (c *Client) PrepareClientContext(cctx client.Context) error {
	c.factoryLock.Lock()
	defer c.factoryLock.Unlock()
	factory, err := c.Factory.Prepare(cctx)
	if err != nil {
		c.log.Error("failed to prepare factory", zap.Error(err))
		return err
	}
	c.Factory = factory
	return nil
}

func (c *Client) SetFromAddress() error {
	activeKp, err := c.GetActiveKeypair()
	if err != nil {
		return err
	}
	if activeKp == nil {
		c.log.Warn("no keys found, you should create at least one")
	} else {
		c.log.Info("configured from address", zap.String("from.address", activeKp.String()))
		c.CCtx = c.CCtx.WithFromAddress(*activeKp)
	}
	return nil
}

// Returns the keypair actively in use for signing transactions (the first key in the keyring).
// If no address has been configured returns `nil, nil`
func (c *Client) GetActiveKeypair() (*sdktypes.AccAddress, error) {
	keys, err := c.Keyring.List()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	kp, err := keys[0].GetAddress()
	if err != nil {
		return nil, err
	}
	return &kp, nil
}
