# vault-plugin-secrets-secp256k1sign

**UNDER CONSTRUCTION!**

Blockchain-agnostic HashiCorp Vault plugin that supports secp256k1 based signing, with an API interface that turns the vault into a software-based HSM device.
Based on [EthSign by Kaleido.io](https://github.com/kaleido-io/vault-plugin-secrets-ethsign) 

![Overview](/resources/overview.png)

The plugin only exposes the following endpoints to enable the client to generate signing keys for the secp256k1 curve suitable for signing any data (including blockchain transactions), list existing signing keys by their names and addresses, and `/signRaw` endpoint for each account. The generated private keys are saved in the vault as a secret. It never gives out the private keys.

## Build
These dependencies are needed:

* go 1.16

To build the binary:
```
make all
```

The output is `secpsign`

## Installing the Plugin on HashiCorp Vault server
The plugin must be registered and enabled on the vault server as a secret engine.

### Enabling on a dev mode server
The easiest way to try out the plugin is using a dev mode server to load it.

Download the binary: [https://www.vaultproject.io/downloads/](https://www.vaultproject.io/downloads/)

First copy the build output binary `secpsign` to the plugins folder, say `~/.vault.d/vault-plugins/`.
```
./vault server -dev -dev-plugin-dir=/Users/alice/.vault.d/vault_plugins/
```

After the dev server starts, the plugin should have already been registered in the system plugins catalog:
```
$ ./vault login <root token>
$ ./vault read sys/plugins/catalog
Key         Value
---         -----
auth        [alicloud app-id approle aws azure centrify cert cf gcp github jwt kubernetes ldap oci oidc okta pcf radius userpass]
database    [cassandra-database-plugin elasticsearch-database-plugin hana-database-plugin influxdb-database-plugin mongodb-database-plugin mssql-database-plugin mysql-aurora-database-plugin mysql-database-plugin mysql-legacy-database-plugin mysql-rds-database-plugin postgresql-database-plugin]
secret      [ad alicloud aws azure cassandra consul secpsign gcp gcpkms kv mongodb mssql mysql nomad pki postgresql rabbitmq ssh totp transit]
```

Note the `secpsign` entry in the secret section. Now it's ready to be enabled:
```
 ./vault secrets enable -path=secp -description="Secp265k1 Wallet" -plugin-name=secpsign plugin
```

To verify the new secret engine based on the plugin has been enabled:
```
$ ./vault secrets list
Path          Type         Accessor              Description
----          ----         --------              -----------
cubbyhole/    cubbyhole    cubbyhole_1f1e372d    per-token private secret storage
secp/         secpsign     secpsign_d9f104c7     Secp265k1 Wallet
identity/     identity     identity_382e2000     identity store
secret/       kv           kv_32f5a684           key/value secret storage
sys/          system       system_21e0c7c7       system endpoints used for control, policy and debugging
```

### Enabling on a non-dev mode server
Setting up a non-dev mode server is beyond the scope of this README, as this is a very sensitive IT operation. But a simple procedure can be found in [the wiki page](https://github.com/kaleido-io/vault-plugin-secrets-ethsign/wiki/Setting-Up-A-Local-HashiCorp-Vault-Server).

Before enabling the plugin on the server, it must first be registered.

First copy the binary to the plugin folder for the server (consult the configuration file for the plugin folder location). Then calculate a SHA256 hash for the binary.
```
shasum -a 256 ./secpsign
```

Use the hash to register the plugin with vault:
```
 ./vault write sys/plugins/catalog/eth-hsm sha_256=$SHA command="secpsign"
```
> If the target vault server is enabled for TLS, and is using a self-signed certificate or other non-verifiable TLS certificate, then the command value needs to contain the switch to turn off TLS verify: `command="secpsign -tls-skip-verify"`

Once registered, just like in dev mode, it's ready to be enabled as a secret engine:
```
 ./vault secrets enable -path=secp -description="Secp265k1 Wallet" -plugin-name=secpsign plugin
```

## Interacting with the secpsign Plugin
The plugin does not interact with the target blockchain. It has very simple responsibilities: sign transactions for submission to a blockchain.
There are 2 ways of dealing with singing:
1) Building and signing TX inside the plugin logic - there is legacy `/sign` API inherited form Kaleido.io ethsign plugin. This API works with Ethereum transactions only.
1) Building TX externally and signing it inside the plugin logic - there is new `/signRaw` API for it. This API can be used to produce any ECDSA Secp256k1 signatures that can be used with any other blockchain that use ECDSA Secp256k1 signatures (Bitcoin, for example). But, the TX building logic is not inculed in the plugin in this case, it just signs data provided externally, making it blockchain agnostic. 

### Creating A New Signing Account
Create a new account in the vault by POSTing to the `/accounts` endpoint.

Using the REST API:
```
$ curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"addressType":"P2PKH"} http://localhost:8200/v1/secp/accounts |jq

{
  "request_id": "a183425c-0998-0888-c768-8dda4ff60bef",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "address": "0xb579cbf259a8d36b22f2799eeeae5f3553b11eb7"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
$ vault write -force secp/accounts addressType=P2PKH

Key        Value
---        -----
address    1MBHQs5p9YxwEuAjsnshCQiawWQGUAMcoU
```

Optional `addressType` value in the request should contain the type of address that should be generated.
Supported types are:
*`P2PKH` - Bitcoin P2PKH (legacy) address
*`P2PKH-Testnet` - Bitcoin P2PKH (legacy) address for Testnet
*`P2SH` - not supported yet
*`P2WPKH` - not supported yet
*`P2TR` - not supported yet
*`ETH` - Ethereum account address (default value). 
if no value is specified, Ethereum account address will be generated

### Importing An Existing Private Key
You can also create a new signing account by importing from an existing private key. The private key is passed in as a hexidecimal string, without the '0x' prfix.

Using the REST API:
```
$ curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"privateKey":"ec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2"}' http://localhost:8200/v1/secp/accounts |jq

{
  "request_id": "a183425c-0998-0888-c768-8dda4ff60bef",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "address": "0xd5bcc62d9b1087a5cfec116c24d6187dd40fdf8a"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
$ vault write secp/accounts privateKey=ec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2

Key        Value
---        -----
address    0xd5bcc62d9b1087a5cfec116c24d6187dd40fdf8a
```

Optional `addressType` value in the request should contain the type of address that should be generated.
Supported types are:
*`P2PKH` - Bitcoin P2PKH (legacy) address
*`P2PKH-Testnet` - Bitcoin P2PKH (legacy) address for Testnet
*`P2SH` - not supported yet
*`P2WPKH` - not supported yet
*`P2TR` - not supported yet
*`ETH` - Ethereum account address (default value). 
if no value is specified, Ethereum account address will be generated

### List Existing Accounts
The list command only returns the addresses of the signing accounts. To return the private keys, use the `/export/accounts/:address` endpoint.

Using the REST API:
```
$  curl -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp/accounts?list=true |jq

{
  "request_id": "56c31ef5-9757-1ff4-354e-3b18ecd8ea77",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "keys": [
      "0xb579cbf259a8d36b22f2799eeeae5f3553b11eb7",
      "0x54edadf1696986c1884534bc6b633ff9a7fdb747"
    ]
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
g$ vault list secp/accounts

Keys
----
0x73b508a63af509a28fb034bf4742bb1a91fcbc4e
0xd5bcc62d9b1087a5cfec116c24d6187dd40fdf8a
```

### Reading Individual Accounts
Inspect the key using the address. Only the address of the signing account is returned. To return the private key, use the `/export/accounts/:address` endpoint.

Using the REST API:
```
$  curl -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp/accounts/0x54edadf1696986c1884534bc6b633ff9a7fdb747 |jq

{
  "request_id": "a183425c-0998-0888-c768-8dda4ff60bef",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "address": "0xb579cbf259a8d36b22f2799eeeae5f3553b11eb7",
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
$ vault read secp/accounts/12SkHVY1iGomTLit6aRafK3TtGakCBnWVu

Key        Value
---        -----
address    12SkHVY1iGomTLit6aRafK3TtGakCBnWVu
publicKey  a022743c2a6930a0bee3bdac72c84e2158e78498b91a8ecae7bb45a26804fe1697ebe5a397ba27695d5522b3e6550e200de8b9cb77129af1afd19e9545ec94aa
```

### Export An Account
You can also export the account by returning the private key. Since keys export is very sensitive operation its access rights should be configured properly and also the keys exporting is only possible as encrypted text, so the GPG public key should be provided as a rsaPublicKey parameter

Using the REST API (TODO: Add rsaPublicKey parameter example):
```
$  curl -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp/export/accounts/0x54edadf1696986c1884534bc6b633ff9a7fdb747 |jq 

{
  "request_id": "a183425c-0998-0888-c768-8dda4ff60bef",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "address": "0xb579cbf259a8d36b22f2799eeeae5f3553b11eb7",
    "privateKey": "ec85999367d32fbbe02dd600a2a44550b95274cc67d14375a9f0bce233f13ad2"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

Using the command line:
```
$ vault read secp/export/accounts/0xd5bcc62d9b1087a5cfec116c24d6187dd40fdf8a rsaPublicKey="$(<my-public.key)"

Key           Value
---           -----
address       0xd5bcc62d9b1087a5cfec116c24d6187dd40fdf8a
privateKey    wcBMA2balAeaF6fgAQgANYCQmk+wKCqm7vIyTFXc/kUSsO/WDmOyDnq1khdYIQzt3+tjvUNs8mDpNgYXC1aLIAta6Zd+EA97NSXIgD6CUzbyz8PQJ4+0smnsUMQY9Lyo6V8yia9XyNgv04jB89iEPQeCqZ+dZk9Mqitpq4vcqFKklv51TUHmxs8FPrvRLahbGaqa+ 
```

Returned privateKey value should be decoded from base64 and then decrypted using GPG utility

### Build and Sign Ethereum Transaction (legacy mode)
Use one of the accounts to sign a transaction.

Using the REST API:
```
$  curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp/accounts/0xc9389f98b1c5f5f9b6b61b5e3769471d550ad596/sign -d '{"data":"0x60fe47b10000000000000000000000000000000000000000000000000000000000000014","gas":30791,"gasPrice":0,"nonce":"0x0","to":"0xca0fe7354981aeb9d051e2f709055eb50b774087"}' |jq

{
  "request_id": "4b68c813-eda9-e3c7-4651-e9dbc526bf47",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "signed_transaction": "0xf888808083015f9094b401069f06a24155774bf8a0f6654ea299c8f68780a460fe47b10000000000000000000000000000000000000000000000000000000000000014840ea23e3fa088f4f5505f6f1da6c9a543863d5c7537e0dfc58618dbf34517c80875283d1e07a0583ecdc23ba3333a3f25611fffe0ec7fb585e9b9af93941f6e3ef8c8ef410698",
    "transaction_hash": "0x7ac47960a9398ae73b994c46fcb8834068195a2d3468c40a1eaad7ed4a15e68e"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

To sign a contract deploy, simply skip the `to` parameter in the JSON payload.

To use EIP155 signer, instead of Homestead signer, pass in `chainId` in the JSON payload.

The `signed_transaction` value in the response is already RLP encoded and can be submitted to an Ethereum blockchain directly.

### Sign a Transaction 
Use one of the accounts to sign a transaction.

Using the REST API:
```
$  curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/secp/accounts/1MBHQs5p9YxwEuAjsnshCQiawWQGUAMcoU/signRaw -d '"payload": "0x44fd2527dcebf3756a9cd61cf0b5313cb34e2d4de079810ed310b078e4616727"' |jq

{
  "request_id": "4b68c813-eda9-e3c7-4651-e9dbc526bf47",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "signature": "0x496c74441f3830feff4ef24df5a7ea5f100e1741e5bac85c206e1e0f51914d472815b8036e8ebfac06d88763deb3d68db214c46aa7cd12c8ebeaad109f98f9ed01",    
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```
The `payload` value in the request should contain hex encoded data to be signed (should start with 0x prefix).
The `signature` value in the response contains signature value (r,s) in hex encoded form (starts with 0x prefix).


## Access Policies
The plugin's endpoint paths are designed such that admin-level access policies vs. user-level access policies can be easily separated.

### Sample User Level Policy:
Use the following policy to assign to a regular user level access token, with the abilities to list keys, read individual keys and sign transactions.

```
/*
 * Ability to list existing keys ("list")
 */
path "secp/accounts" {
  capabilities = ["list", "update"]
}
/*
 * Ability to retrieve individual keys ("read"), sign transactions ("create")
 */
path "secp/accounts/*" {
  capabilities = ["create", "read"]
}
```

### Sample Admin Level Policy:
Use the following policy to assign to a admin level access token, with the full ability to create keys, import existing private keys, export private keys, read/delete individual keys, and sign transactions.

```
/*
 * Ability to create key ("update") and list existing keys ("list")
 */
path "secp/accounts" {
  capabilities = ["update", "list"]
}
/*
 * Ability to retrieve individual keys ("read"), sign transactions ("create") and delete keys ("delete")
 */
path "secp/accounts/*" {
  capabilities = ["create", "read", "delete"]
}
/*
 * Ability to export private keys ("read")
 */
path "secp/export/accounts/*" {
  capabilities = ["read"]
}
```
