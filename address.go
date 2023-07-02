package compass

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (cc *Client) EncodeBech32AccAddr(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(cc.cfg.AccountPrefix, addr)
}
func (cc *Client) EncodeBech32AccPub(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "pub"), addr)
}
func (cc *Client) EncodeBech32ValAddr(addr sdk.ValAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "valoper"), addr)
}
func (cc *Client) EncodeBech32ValPub(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "valoperpub"), addr)
}
func (cc *Client) EncodeBech32ConsAddr(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "valcons"), addr)
}
func (cc *Client) EncodeBech32ConsPub(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "valconspub"), addr)
}

func (cc *Client) DecodeBech32AccAddr(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, cc.cfg.AccountPrefix)
}
func (cc *Client) DecodeBech32AccPub(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "pub"))
}
func (cc *Client) DecodeBech32ValAddr(addr string) (sdk.ValAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "valoper"))
}
func (cc *Client) DecodeBech32ValPub(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "valoperpub"))
}
func (cc *Client) DecodeBech32ConsAddr(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "valcons"))
}
func (cc *Client) DecodeBech32ConsPub(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.cfg.AccountPrefix, "valconspub"))
}
