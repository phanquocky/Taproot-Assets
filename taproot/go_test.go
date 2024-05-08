package taproot

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	asset2 "github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestAbc(t *testing.T) {
	key := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}
	value := []byte("abc")
	leafNode := mssmt.LeafNode{
		Value: value,
	}
	tree := mssmt.NewCompactedTree(mssmt.NewDefaultStore())
	tree.Insert(nil, key, &leafNode)
	fmt.Println("abcdef ")
	b, err := json.Marshal(tree)
	if err != nil {
		fmt.Println("Marshal error: ", err)
		t.Error(err)
	}

	fmt.Println("abcbyte: ", string(b))
	t.Log("Test abc")
}

type TapCommitmentTest struct {
	TreeRoot *mssmt.BranchNode `json:"tree_root"`

	AssetCommitments []*commitment.AssetCommitment `json:"asset_commitments"`
}

func TestMarshalAssetCommitment(t *testing.T) {
	a := [33]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33}

	outpoint := wire.OutPoint{Hash: chainhash.Hash{}, Index: 1}

	mintAssets := asset2.NewGenesis(outpoint, "name", 0)

	asset := asset2.NewAsset(mintAssets, 30, a, nil)

	fmt.Println("asset: ", asset)

	ac, err := commitment.NewAssetCommitment(context.Background(), asset)
	require.NoError(t, err)

	tapcommitment, err := commitment.NewTapCommitment(ac)

	assetCommitments := make([]*commitment.AssetCommitment, 0)
	for _, assetCommitment := range tapcommitment.AssetCommitments {
		assetCommitments = append(assetCommitments, assetCommitment)
	}

	tapTest := TapCommitmentTest{
		TreeRoot:         tapcommitment.TreeRoot,
		AssetCommitments: assetCommitments,
	}
	fmt.Println("tapcommitment: ", tapTest)

	b, err := json.Marshal(tapTest)
	require.NoError(t, err)

	fmt.Println("tapcommitment: ", string(b))
}
