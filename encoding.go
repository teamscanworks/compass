package compass

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
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
