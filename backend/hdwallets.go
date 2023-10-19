package backend

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/tyler-smith/go-bip39"
)

type HDWallet struct {
	Name        string `json:"name"`
	AddressType string `json:"address_type"`
	SeedPhrase  string `json:"seed_phrase"`
	MasterKey   string `json:"master_key"`
}

func hdw_paths(b *backend) []*framework.Path {
	return []*framework.Path{
		pathHdwCreateAndList(b),
		pathHdwExport(b),
	}
}

func (b *backend) listHDWallets(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vals, err := req.Storage.List(ctx, "hdwallets/")
	if err != nil {
		b.Logger().Error("Failed to retrieve the list of HD Wallets", "error", err)
		return nil, err
	}

	return logical.ListResponse(vals), nil
}

func (b *backend) createHDWallet(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	seedPhrase := data.Get("seedPhrase").(string)
	addrType := data.Get("addressType").(string)
	name := data.Get("name").(string)

	var err error

	var seed []byte

	if seedPhrase == "" {
		//generate new seed
		entropy, err := bip39.NewEntropy(256) //TODO: parameterize
		if err != nil {
			b.Logger().Error("Failed to generate entropy", "error", err)
			return nil, err
		}

		seedPhrase, err = bip39.NewMnemonic(entropy)
		if err != nil {
			b.Logger().Error("Failed to generate mnemonic", "error", err)
			return nil, err
		}
	}

	seed = bip39.NewSeed(seedPhrase, "Secret Passphrase") //TODO: find out whether it makes sence to use this password

	var keyString string
	var keyAddress string

	switch addrType {
	case "P2PKH":
		params := &chaincfg.MainNetParams
		key, err := hdkeychain.NewMaster(seed, params)
		if err != nil { //TODO: check error and retry
			b.Logger().Error("Failed to init master node", "error", err)
			return nil, err
		}
		defer key.Zero()
		keyString = key.String()
		addr, err := key.Address(params)
		if err != nil { //TODO: check error and retry
			b.Logger().Error("Failed get master node address", "error", err)
			return nil, err
		}
		keyAddress = addr.String()

	case "P2PKH-Testnet":
		params := &chaincfg.TestNet3Params
		// Generate a new master node using the seed.
		key, err := hdkeychain.NewMaster(seed, params)
		if err != nil { //TODO: check error and retry
			b.Logger().Error("Failed to init master node", "error", err)
			return nil, err
		}
		defer key.Zero()
		keyString = key.String()
		addr, err := key.Address(params)
		if err != nil { //TODO: check error and retry
			b.Logger().Error("Failed get master node address", "error", err)
			return nil, err
		}
		keyAddress = addr.String()

		//default //ETH

	}

	if name == "" {
		name = keyAddress
	}

	hdwPath := fmt.Sprintf("hdwallets/%s", name)

	hdwJSON := &HDWallet{
		Name:        name,
		MasterKey:   keyString,
		SeedPhrase:  seedPhrase,
		AddressType: addrType,
	}

	entry, _ := logical.StorageEntryJSON(hdwPath, hdwJSON)
	err = req.Storage.Put(ctx, entry)
	if err != nil {
		b.Logger().Error("Failed to save the new hdwallet to storage", "error", err)
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"name": hdwJSON.Name,
		},
	}, nil
}

func (b *backend) retrieveHDWallet(ctx context.Context, req *logical.Request, name string) (*HDWallet, error) {

	path := fmt.Sprintf("hdwallets/%s", name)
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		b.Logger().Error("Failed to retrieve HDWallet by name", "path", path, "error", err)
		return nil, err
	}
	if entry == nil {
		// could not find the corresponding key for the address
		return nil, nil
	}
	var hdwallet HDWallet
	_ = entry.DecodeJSON(&hdwallet)
	return &hdwallet, nil
}

func (b *backend) exportHDWallet(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)

	b.Logger().Info("Retrieving master key data for HDWallet ", "name", name)
	hdwallet, err := b.retrieveHDWallet(ctx, req, name)
	if err != nil {
		return nil, err
	}
	if hdwallet == nil {
		return nil, fmt.Errorf("HDWallet does not exist")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"name":        hdwallet.Name,
			"addressType": hdwallet.AddressType,
			"seedPhrase":  hdwallet.SeedPhrase,
			"masterKey":   hdwallet.MasterKey,
		},
	}, nil
}
