package assetoutpointmodel

type UnspentOutpoint struct {
	ID                       string   `json:"id"`
	GenesisID                string   `json:"genesis_id"`
	ScriptKey                []byte   `json:"script_key"`
	Amount                   int32    `json:"amount"`
	SplitCommitmentRootHash  []byte   `json:"split_commitment_root_hash"`
	SplitCommitmentRootValue int32    `json:"split_commitment_root_value"`
	AnchorUtxoID             string   `json:"anchor_utxo_id"`
	ProofLocator             []byte   `json:"proof_locator"`
	Proof                    []byte   `json:"proof"`
	Spent                    bool     `json:"spent"`
	Outpoint                 string   `json:"outpoint"`
	AmtSats                  int32    `json:"amt_sats"`
	InternalKey              []byte   `json:"internal_key"`
	TaprootAssetRoot         []byte   `json:"taproot_asset_root"`
	ScriptOutput             []byte   `json:"script_output"`
	TxID                     string   `json:"tx_id"`
	RelatedAnchorAssets      [][]byte `json:"related_anchor_asset"`
	RelatedAnchorAssetProofs [][]byte `json:"related_anchor_asset_proof"`
}
