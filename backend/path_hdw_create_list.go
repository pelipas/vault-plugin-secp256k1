package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathHdwCreateAndList(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "hdwallets/?",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ListOperation:   b.listHDWallets,
			logical.UpdateOperation: b.createHDWallet,
		},
		HelpSynopsis: "List all the HD Wallets maintained by the plugin backend and create new accounts.",
		HelpDescription: `

    LIST - list all accounts
    POST - create a new account

    `,
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of a HD Wallet to be created. If not present, the request generates master key address",
				Default:     "",
			},
			"addressType": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Type of address to be generated (possible values are ETH, P2PK, P2PKH, P2SH, P2WPKH, P2WSH, P2TR). If not present, the request generates ETH address.",
				Default:     "",
			},
			"seedPhrase": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Seed phrase. If not present, the request generates new seed.",
				Default:     "",
			},
		},
	}
}
