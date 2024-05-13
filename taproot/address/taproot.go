package address

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
)

type TapAddress struct {
	Address           *btcutil.AddressTaproot
	TapScriptRootHash *chainhash.Hash
	PubKey            asset.SerializedKey
	tapCommitment     *commitment.TapCommitment
}

func (tap *TapAddress) GetTapCommitment() *commitment.TapCommitment {
	return tap.tapCommitment
}

// CreateTapAddr create taproot address with public key of owner
// with [32]byte data to tapscript branch. The purpose of this is insert data to
// onchain through tap address.
func (tap *TapAddr) CreateTapAddr(
	userPubKey asset.SerializedKey,
	tapCommitment *commitment.TapCommitment,
) (*TapAddress, error) {

	pubkey, err := userPubKey.ToPubKey()
	if err != nil {
		return nil, err
	}

	//					tapscriptrootHash
	//                   /       \
	//					/         \
	//				tapleaf		tapleaf
	tapleaf := tapCommitment.TapLeaf()

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

	return &TapAddress{
		Address:           address,
		TapScriptRootHash: &tapScriptRootHash,
		PubKey:            userPubKey,
		tapCommitment:     tapCommitment,
	}, nil
}
