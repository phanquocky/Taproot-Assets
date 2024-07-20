package assetoutpoint

import (
	"github.com/quocky/taproot-asset/server/internal/domain/common"
	manageutxo "github.com/quocky/taproot-asset/server/internal/domain/manage_utxo"
	"github.com/quocky/taproot-asset/taproot/model/asset"
)

type AssetOutpoint struct {
	common.Entity            `json:",inline"`
	GenesisID                common.ID           `json:"genesis_id"`
	Genesis                  *asset.GenesisAsset `json:"genesis,omitempty"`
	ScriptKey                []byte              `json:"script_key"`
	Amount                   int32               `json:"amount"`
	SplitCommitmentRootHash  []byte              `json:"split_commitment_root_hash"`
	SplitCommitmentRootValue int32               `json:"split_commitment_root_value"`
	AnchorUtxoID             common.ID           `json:"anchor_utxo_id"`
	ProofLocator             []byte              `json:"proof_locator"`
	Spent                    bool                `json:"spent"`
}

type UnspentOutpoint struct {
	AssetOutpoint          `json:",inline"`
	manageutxo.ManagedUtxo `json:"res"`
	RelatedAssets          []*AssetOutpoint `json:"related_assets"`
}
