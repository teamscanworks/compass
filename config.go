package compass

import (
	"path"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

var (
	// Provides a default set of modules
	ModuleBasics = []module.AppModuleBasic{
		auth.AppModuleBasic{},
		authz.AppModuleBasic{},
		bank.AppModuleBasic{},
		// TODO: add osmosis governance proposal types here
		// TODO: add other proposal types here
		gov.NewAppModuleBasic([]client.ProposalHandler{
			paramsclient.ProposalHandler,
		}),
		crisis.AppModuleBasic{},
		distribution.AppModuleBasic{},
		mint.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		staking.AppModuleBasic{},
	}
)

// Allows configuration of the compass client
type ClientConfig struct {
	Key            string                  `json:"key" yaml:"key"`
	ChainID        string                  `json:"chain-id" yaml:"chain-id"`
	RPCAddr        string                  `json:"rpc-addr" yaml:"rpc-addr"`
	GRPCAddr       string                  `json:"grpc-addr" yaml:"grpc-addr"`
	AccountPrefix  string                  `json:"account-prefix" yaml:"account-prefix"`
	KeyringBackend string                  `json:"keyring-backend" yaml:"keyring-backend"`
	GasAdjustment  float64                 `json:"gas-adjustment" yaml:"gas-adjustment"`
	GasPrices      string                  `json:"gas-prices" yaml:"gas-prices"`
	MinGasAmount   uint64                  `json:"min-gas-amount" yaml:"min-gas-amount"`
	KeyDirectory   string                  `json:"key-directory" yaml:"key-directory"`
	Debug          bool                    `json:"debug" yaml:"debug"`
	Timeout        string                  `json:"timeout" yaml:"timeout"`
	BlockTimeout   string                  `json:"block-timeout" yaml:"block-timeout"`
	OutputFormat   string                  `json:"output-format" yaml:"output-format"`
	SignModeStr    string                  `json:"sign-mode" yaml:"sign-mode"`
	ExtraCodecs    []string                `json:"extra-codecs" yaml:"extra-codecs"`
	Modules        []module.AppModuleBasic `json:"-" yaml:"-"`
	Slip44         int                     `json:"slip44" yaml:"slip44"`
}

// Validates the client configuration
func (ccc *ClientConfig) Validate() error {
	if _, err := time.ParseDuration(ccc.Timeout); err != nil {
		return err
	}
	if ccc.BlockTimeout != "" {
		if _, err := time.ParseDuration(ccc.BlockTimeout); err != nil {
			return err
		}
	}
	return nil
}

// Formats the given directory for use as a keyring directory
func (ccc *ClientConfig) FormatKeysDir(home string) string {
	return path.Join(home, "keys", ccc.ChainID)
}

// Formats, and sets the home directory as the keyring direcotry
func (ccc *ClientConfig) SetKeysDir(home string) {
	ccc.KeyDirectory = ccc.FormatKeysDir(home)
}

// Returns a configuration suitable for the cosmoshub chain
func GetCosmosHubConfig(keyHome string, debug bool) *ClientConfig {
	cfg := &ClientConfig{
		Key:            "default",
		ChainID:        "cosmoshub-4",
		RPCAddr:        "https://cosmoshub-4.technofractal.com:443",
		GRPCAddr:       "https://gprc.cosmoshub-4.technofractal.com:443",
		AccountPrefix:  "cosmos",
		KeyringBackend: "test",
		GasAdjustment:  1.2,
		GasPrices:      "0.01uatom",
		MinGasAmount:   0,
		KeyDirectory:   keyHome,
		Debug:          debug,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
	}
	cfg.SetKeysDir(keyHome)
	return cfg
}

// Returns a configuration suitable for the osmosis blockchain
func GetOsmosisConfig(keyHome string, debug bool) *ClientConfig {
	cfg := &ClientConfig{
		Key:            "default",
		ChainID:        "osmosis-1",
		RPCAddr:        "https://osmosis-1.technofractal.com:443",
		GRPCAddr:       "https://gprc.osmosis-1.technofractal.com:443",
		AccountPrefix:  "osmo",
		KeyringBackend: "test",
		GasAdjustment:  1.2,
		GasPrices:      "0.01uosmo",
		MinGasAmount:   0,
		KeyDirectory:   keyHome,
		Debug:          debug,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
	}
	cfg.SetKeysDir(keyHome)
	return cfg
}

// Returns a configuration suitable for usage in simd environments
func GetSimdConfig() *ClientConfig {
	cfg := &ClientConfig{
		Key:            "default",
		ChainID:        "cosmoshub-4",
		RPCAddr:        "tcp://127.0.0.1:26657",
		GRPCAddr:       "127.0.0.1:9090",
		AccountPrefix:  "cosmos",
		KeyringBackend: "test",
		GasAdjustment:  1.2,
		GasPrices:      "0.01uatom",
		MinGasAmount:   0,
		KeyDirectory:   "keyring-test",
		Debug:          true,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
	}
	cfg.SetKeysDir("keyring-test")
	return cfg
}
