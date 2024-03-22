package asset

import "github.com/quocky/taproot-asset/taproot/model/mssmt"

type SplitCommitment struct {
	Proof     mssmt.Proof
	RootAsset Asset
}
