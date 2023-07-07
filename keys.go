package compass

import (
	ckeys "github.com/cosmos/cosmos-sdk/client/keys"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// KeyOutput contains mnemonic and address of key
type KeyOutput struct {
	Mnemonic string `json:"mnemonic" yaml:"mnemonic"`
	Address  string `json:"address" yaml:"address"`
}

func (cc *Client) AddKey(name string, coinType uint32) (output *KeyOutput, err error) {
	ko, err := cc.KeyAddOrRestore(name, coinType)
	if err != nil {
		return nil, err
	}
	return ko, nil
}

func (cc *Client) RestoreKey(name, mnemonic string, coinType uint32) (address string, err error) {
	ko, err := cc.KeyAddOrRestore(name, coinType, mnemonic)
	if err != nil {
		return "", err
	}
	return ko.Address, nil
}

func (cc *Client) ShowAddress(name string) (address string, err error) {
	info, err := cc.Keyring.Key(name)
	if err != nil {
		return "", err
	}
	acc, err := info.GetAddress()
	if err != nil {
		return "", nil
	}
	out, err := cc.EncodeBech32AccAddr(acc)
	if err != nil {
		return "", err
	}
	return out, nil
}

func (cc *Client) ListAddresses() (map[string]string, error) {
	out := map[string]string{}
	info, err := cc.Keyring.List()
	if err != nil {
		return nil, err
	}
	for _, k := range info {
		acc, err := k.GetAddress()
		if err != nil {
			return nil, err
		}
		addr, err := cc.EncodeBech32AccAddr(acc)
		if err != nil {
			return nil, err
		}
		out[k.Name] = addr
	}
	return out, nil
}

func (cc *Client) DeleteKey(name string) error {
	if err := cc.Keyring.Delete(name); err != nil {
		return err
	}
	return nil
}

func (cc *Client) KeyExists(name string) bool {
	k, err := cc.Keyring.Key(name)
	if err != nil {
		return false
	}

	return k.Name == name

}

func (cc *Client) ExportPrivKeyArmor(keyName string) (armor string, err error) {
	return cc.Keyring.ExportPrivKeyArmor(keyName, ckeys.DefaultKeyPass)
}

func (cc *Client) KeyAddOrRestore(keyName string, coinType uint32, mnemonic ...string) (*KeyOutput, error) {
	var mnemonicStr string
	var err error
	algo := keyring.SignatureAlgo(hd.Secp256k1)

	if len(mnemonic) > 0 {
		mnemonicStr = mnemonic[0]
	} else {
		mnemonicStr, err = CreateMnemonic()
		if err != nil {
			return nil, err
		}
	}

	info, err := cc.Keyring.NewAccount(keyName, mnemonicStr, "", hd.CreateHDPath(coinType, 0, 0).String(), algo)
	if err != nil {
		return nil, err
	}

	acc, err := info.GetAddress()
	if err != nil {
		return nil, err
	}

	out, err := cc.EncodeBech32AccAddr(acc)
	if err != nil {
		return nil, err
	}
	return &KeyOutput{Mnemonic: mnemonicStr, Address: out}, nil
}
