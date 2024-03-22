package commitment

import "github.com/quocky/taproot-asset/taproot/model/mssmt"

type AssetCommitments map[[32]byte]*AssetCommitment

type TapCommitment struct {
	TreeRoot         *mssmt.BranchNode
	tree             mssmt.Tree
	assetCommitments AssetCommitments
}
