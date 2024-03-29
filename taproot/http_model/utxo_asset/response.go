package utxoasset

type UnspentAssetResp struct {
	GenesisAsset     GenesisAsset
	UnspentOutpoints []UnspentOutpoint
	GenesisPoint     GenesisPoint
	InputFilesBytes  [][]byte
}

type GenesisAsset struct {
	// use hex.Encoder to convert to []byte
	AssetID        string `json:"asset_id"`
	AssetName      string `json:"asset_name"`
	Supply         int32  `json:"supply"`
	OutputIndex    int32  `json:"output_index"`
	GenesisPointID string `json:"genesis_point_id"`
}

type GenesisPoint struct {
	PrevOut    string `json:"prev_out"`
	AnchorTxID string `json:"anchor_tx_id"`
}

type UnspentOutpoint struct {
	GenesisID                string `json:"genesis_id"`
	ScriptKey                []byte `json:"script_key"`
	Amount                   int32  `json:"amount"`
	SplitCommitmentRootHash  []byte `json:"split_commitment_root_hash"`
	SplitCommitmentRootValue int32  `json:"split_commitment_root_value"`
	AnchorUtxoID             string `json:"anchor_utxo_id"`
	ProofLocator             []byte `json:"proof_locator"`
	Spent                    bool   `json:"spent"`
	Outpoint                 string `json:"outpoint"`
	AmtSats                  int32  `json:"amt_sats"`
	InternalKey              []byte `json:"internal_key"`
	TaprootAssetRoot         []byte `json:"taproot_asset_root"`
	ScriptOutput             []byte `json:"script_output"`
	TxID                     string `json:"tx_id"`
}
