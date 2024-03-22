package commitment

import (
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
)

type CommittedAssets map[[32]byte]*asset.Asset

type AssetCommitment struct {
	TapKey [32]byte
	Root   *mssmt.BranchNode
	tree   mssmt.Tree
	assets CommittedAssets
}
