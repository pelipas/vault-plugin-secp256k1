package backend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
)

func TestHDWallets(t *testing.T) {
	assert := assert.New(t)

	b, _ := getBackend(t)

	// create hdw1
	req := logical.TestRequest(t, logical.UpdateOperation, "hdwallets")
	storage := req.Storage
	data := map[string]interface{}{
		"name":        "",
		"addressType": "P2PKH",
		"seedPhrase":  "",
	}
	req.Data = data

	res, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	name1 := res.Data["name"].(string)
	println(fmt.Sprintf("res.Data = %s", res.Data))

	// create key2
	req = logical.TestRequest(t, logical.UpdateOperation, "hdwallets")
	req.Storage = storage
	req.Data = data
	res, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	name2 := res.Data["name"].(string)

	req = logical.TestRequest(t, logical.ListOperation, "hdwallets")
	req.Storage = storage
	resp, err := b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	println(fmt.Sprintf("resp.Data = %s", resp.Data))

	expected1 := &logical.Response{
		Data: map[string]interface{}{
			"keys": []string{name1, name2},
		},
	}
	expected2 := &logical.Response{
		Data: map[string]interface{}{
			"keys": []string{name2, name1},
		},
	}

	if !reflect.DeepEqual(resp, expected1) && !reflect.DeepEqual(resp, expected2) {
		t.Fatalf("bad response.\n\nexpected: %#v\n\nGot: %#v", expected1, resp)
	}

	req = logical.TestRequest(t, logical.UpdateOperation, "hdwallets")
	req.Storage = storage
	data = map[string]interface{}{
		"name":        "Test_HDWallet",
		"addressType": "P2PKH",
		"seedPhrase":  "wink lava series fame scorpion friend eyebrow thank blanket blur network own direct cover assume swap science aunt supreme token anxiety roast pink where",
	}
	req.Data = data
	res, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	req = logical.TestRequest(t, logical.ReadOperation, "export/hdwallets/Test_HDWallet")
	req.Storage = storage
	res, err = b.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	b.Logger().Info("Master key data for HDWallet ", "masterKey", res.Data["masterKey"])
	assert.Equal("wink lava series fame scorpion friend eyebrow thank blanket blur network own direct cover assume swap science aunt supreme token anxiety roast pink where", res.Data["seedPhrase"])

	//panic("Bye!")
}
