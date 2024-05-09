package mint

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/server/internal/domain/asset"
	assetoutpoint "github.com/quocky/taproot-asset/server/internal/domain/asset_outpoint"
	chaintx "github.com/quocky/taproot-asset/server/internal/domain/chain_tx"
	"github.com/quocky/taproot-asset/server/internal/domain/genesis"
	manageutxo "github.com/quocky/taproot-asset/server/internal/domain/manage_utxo"
	assetsdk "github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/proof"
)

// InsertMintTxParams input parameter of genesis transaction
type InsertMintTxParams struct {
	Asset             *assetsdk.Asset        `json:"asset"`
	OutputIdx         int32                  `json:"output_idx"`
	AnchorTx          *wire.MsgTx            `json:"anchor_tx"`
	AmountSats        int32                  `json:"amount_sats"`
	AddressInfoPubkey assetsdk.SerializedKey `json:"address_info_pubkey"`
	TapScriptRootHash *chainhash.Hash        `json:"tap_script_root_hash"`
	ProofLocator      [32]byte               `json:"proof_locator"`
	MintProof         *proof.Proof           `json:"mint_proof"`
}

type InsertMintTxResult struct {
	GenesisAsset  asset.GenesisAsset          `json:"genesis_asset"`
	GenesisPoint  genesis.GenesisPoint        `json:"genesis"`
	AnchorTx      chaintx.ChainTx             `json:"anchor_tx"`
	ManagedUTXO   manageutxo.ManagedUtxo      `json:"managed_utxo"`
	AssetOutpoint assetoutpoint.AssetOutpoint `json:"asset_outpoint"`
}
