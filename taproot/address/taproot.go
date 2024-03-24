package address

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
)

// AddrResult is struct contain all information of address when it is created
type AddrResult struct {
	Address           *btcutil.AddressTaproot
	TapscriptRootHash *chainhash.Hash
	Pubkey            asset.SerializedKey
	TapCommitment     *commitment.TapCommitment
}

// CreateTapAddrByCommitment create taproot address with public key of owner
// with [32]byte data to tapscript branch. The purpose of this is insert data to
// onchain through tap address.
func (tap *TapAddr) CreateTapAddrByCommitment(
	serializedPubkey asset.SerializedKey,
	topCommitment *commitment.TapCommitment,
) (*AddrResult, error) {

	pubkey, err := serializedPubkey.ToPubKey()
	if err != nil {
		return nil, err
	}

	//					tapscriptrootHash
	//                   /       \
	//					/         \
	//				tapleaf		tapleaf
	tapleaf := topCommitment.TapLeaf()

	tapScriptTree := txscript.AssembleTaprootScriptTree(tapleaf)
	tapScriptRootHash := tapScriptTree.LeafMerkleProofs[0].RootNode.TapHash()

	//					outputkey
	//				 /			\
	//				/			 \
	//			pubkey 			tapScriptRootHash
	//
	// fomular:  taprootKey = internalKey + (h_tapTweak(internalKey || merkleRoot)*G)
	outputKey := txscript.ComputeTaprootOutputKey(pubkey, tapScriptRootHash[:])

	address, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), tap.NetWork)
	if err != nil {
		return nil, err
	}

	return &AddrResult{
		Address:           address,
		TapscriptRootHash: &tapScriptRootHash,
		Pubkey:            serializedPubkey,
		TapCommitment:     topCommitment,
	}, nil
}
