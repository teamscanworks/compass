package compass

import (
	"fmt"
	"os"
	"sync"

	"github.com/cosmos/cosmos-sdk/client"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	cclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
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

	// sequence number setting can potentially lead to race conditions, so lock access
	// to the ability to send transactions to a single tx at a time
	//
	// NOTE: this isn't very performant, so should probably move to a goroutine that
	// runs in a goroutine, accepting messages to send through a channel
	txLock sync.Mutex
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

// Triggers keyring migration, ensuring that the factory, and client context are updated
func (c *Client) MigrateKeyring() error {
	_, err := c.Keyring.MigrateAll()
	if err != nil {
		return err
	}
	c.Factory = c.Factory.WithKeybase(c.Keyring)
	c.CCtx = c.CCtx.WithKeyring(c.Keyring)
	return nil
}

// Returns previously initialized transaction factory
func (c *Client) TxFactory() tx.Factory {
	return c.Factory
}

func (c *Client) UpdateFromName(name string) {
	c.CCtx = c.CCtx.WithFromName(name)
}

// Returns an instance of client.Context, used widely throughout cosmos-sdk
func (c *Client) ClientContext() client.Context {
	return c.CCtx
}
func (c *Client) SendTransaction(msg sdktypes.Msg) error {
	c.txLock.Lock()
	defer c.txLock.Unlock()
	if err := c.prepare(); err != nil {
		return fmt.Errorf("transaction preparation failed %v", err)
	}
	c.log.Info("sending transaction")
	if err := tx.GenerateOrBroadcastTxWithFactory(c.CCtx, c.Factory, msg); err != nil {
		return fmt.Errorf("failed to send transaction %v", err)
	}
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
		c.CCtx = c.CCtx.WithFromAddress(*activeKp).WithFromName(c.cfg.Key)
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

// Returns the keyring record located at the given index, returning an error
// if there are less than `idx` keys in the keyring
func (c *Client) KeyringRecordAt(idx int) (*keyring.Record, error) {
	keys, err := c.Keyring.List()
	if err != nil {
		return nil, err
	}
	if len(keys) < idx-1 {
		return nil, fmt.Errorf("key length %v less than index %v", len(keys), idx)
	}
	return keys[idx], nil
}

// prepartion is susceptible to race conditions so its an private function
func (c *Client) prepare() error {
	kp, err := c.GetActiveKeypair()
	if err != nil {
		return err
	}
	factory, err := c.Factory.Prepare(c.CCtx)
	if err != nil {
		c.log.Error("failed to prepare factory", zap.Error(err))
		return err
	}
	c.Factory = factory
	_, seq, err := c.Factory.AccountRetriever().GetAccountNumberSequence(c.CCtx, *kp)
	if err != nil {
		return err
	}
	c.Factory = c.Factory.WithSequence(seq)
	return nil
}

func (c *Client) configTxFactory(input tx.Factory) tx.Factory {
	return input.
		WithAccountRetriever(c).
		WithChainID(c.cfg.ChainID).
		WithGasAdjustment(c.cfg.GasAdjustment).
		WithGasPrices(c.cfg.GasPrices).
		WithKeybase(c.Keyring).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		// prevents some runtime panics  due to misconfigured clients causing the error messages to be logged
		WithSimulateAndExecute(true)
}

func (c *Client) configClientContext(cctx client.Context) client.Context {
	return cctx.
		//WithViper("breaker").
		WithAccountRetriever(c).
		WithChainID(c.cfg.ChainID).
		WithKeyring(c.Keyring).
		WithGRPCClient(c.GRPC).
		WithClient(c.RPC).
		WithSignModeStr(signing.SignMode_SIGN_MODE_DIRECT.String()).
		WithCodec(c.Codec.Marshaler).
		WithInterfaceRegistry(c.Codec.InterfaceRegistry).
		WithBroadcastMode("sync").
		// this is important to set otherwise it is not possible to programmatically sign
		// transactions with cosmos-sdk as it will expect the user to provide input
		WithSkipConfirmation(true)
}
