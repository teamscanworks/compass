package compass

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Codec struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

func MakeCodec(moduleBasics []module.AppModuleBasic, extraCodecs []string) Codec {
	modBasic := module.NewBasicManager(moduleBasics...)
	encodingConfig := MakeCodecConfig(modBasic)

	return encodingConfig
}

func MakeCodecConfig(modBase module.BasicManager) Codec {
	interfaceRegistry := types.NewInterfaceRegistry()
	anyRegistry := NewAnyInterfaceRegistry(interfaceRegistry)
	amino := codec.NewLegacyAmino()
	std.RegisterInterfaces(anyRegistry)
	std.RegisterLegacyAminoCodec(amino)
	modBase.RegisterInterfaces(anyRegistry)
	modBase.RegisterLegacyAminoCodec(amino)

	authtypes.RegisterInterfaces(anyRegistry)
	banktypes.RegisterInterfaces(anyRegistry)
	crisistypes.RegisterInterfaces(anyRegistry)
	distributiontypes.RegisterInterfaces(anyRegistry)
	proposal.RegisterInterfaces(anyRegistry)
	slashingtypes.RegisterInterfaces(anyRegistry)
	stakingtypes.RegisterInterfaces(anyRegistry)
	vestingtypes.RegisterInterfaces(anyRegistry)
	authztypes.RegisterInterfaces(anyRegistry)

	marshaler := codec.NewProtoCodec(anyRegistry)
	/*
		defaultOpts, err := tx.NewDefaultSigningOptions()
		if err != nil {
			panic(err)
		}
		txConfig, err := tx.NewTxConfigWithOptions(marshaler, tx.ConfigOptions{
			SigningOptions:   defaultOpts,
			EnabledSignModes: tx.DefaultSignModes,
		})
		if err != nil {
			panic(err)
		}
	*/
	return Codec{
		InterfaceRegistry: anyRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             amino,
	}
}
