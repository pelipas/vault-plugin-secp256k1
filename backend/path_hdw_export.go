package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathHdwExport(b *backend) *framework.Path {
	return &framework.Path{
		Pattern:      "export/hdwallets/" + framework.GenericNameRegex("name"),
		HelpSynopsis: "Export an Ethereum account",
		HelpDescription: `

    GET - return the account by the name with the private key

    `,
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{Type: framework.TypeString},
		},
		ExistenceCheck: b.pathExistenceCheck,
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.exportHDWallet,
		},
	}
}
