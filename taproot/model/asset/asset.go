package asset

import (
	"crypto/sha256"
	"encoding/json"
	"reflect"

	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
)

var (
	ZeroPrevID PrevID
)

type Asset struct {
	Genesis
	Amount              int32
	ScriptPubkey        SerializedKey
	SplitCommitmentRoot *mssmt.ComputedNode
	PrevWitnesses       []Witness
}

func NewAsset(
	genesis Genesis, amount int32,
	scriptPubkey SerializedKey,
	splitCommitmentRoot *mssmt.ComputedNode,
) *Asset {

	return &Asset{
		Genesis:             genesis,
		Amount:              amount,
		ScriptPubkey:        scriptPubkey,
		SplitCommitmentRoot: splitCommitmentRoot,
		PrevWitnesses: []Witness{
			{
				PrevID:          &ZeroPrevID,
				SplitCommitment: nil,
			},
		},
	}
}

func (a *Asset) AssetCommitmentKey() [32]byte {
	h := sha256.New()

	genesisID := a.Genesis.ID()
	h.Write(genesisID[:])
	//h.Write([]byte(strconv.Itoa(int(a.Amount))))
	h.Write(a.ScriptPubkey.SchnorrSerialized())

	return [32]byte(h.Sum(nil))
}

func (a *Asset) Leaf() (*mssmt.LeafNode, error) {
	assetBytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	return mssmt.NewLeafNode(assetBytes, uint64(a.Amount)), nil
}

func (a *Asset) TapCommitmentKey() [32]byte {
	return a.ID()
}

func (a *Asset) Copy() *Asset {
	assetCopy := *a

	if len(a.PrevWitnesses) > 0 {
		assetCopy.PrevWitnesses = make([]Witness, len(a.PrevWitnesses))
		for idx := range a.PrevWitnesses {
			witness := a.PrevWitnesses[idx]

			var witnessCopy Witness
			if witness.PrevID != nil {
				witnessCopy.PrevID = &PrevID{
					OutPoint:  witness.PrevID.OutPoint,
					ID:        witness.PrevID.ID,
					ScriptKey: witness.PrevID.ScriptKey,
				}
			}

			if witness.SplitCommitment != nil {
				witnessCopy.SplitCommitment = &SplitCommitment{
					Proof:     *witness.SplitCommitment.Proof.Copy(),
					RootAsset: *witness.SplitCommitment.RootAsset.Copy(),
				}
			}
			assetCopy.PrevWitnesses[idx] = witnessCopy
		}
	}

	if a.SplitCommitmentRoot != nil {
		assetCopy.SplitCommitmentRoot = mssmt.NewComputedNode(
			a.SplitCommitmentRoot.NodeHash(),
			a.SplitCommitmentRoot.NodeSum(),
		)
	}
	return &assetCopy
}

// HasSplitCommitmentWitness returns true if an asset has a split commitment
// witness.
func (a *Asset) HasSplitCommitmentWitness() bool {
	if len(a.PrevWitnesses) != 1 {
		return false
	}

	return IsSplitCommitWitness(a.PrevWitnesses[0])
}

// IsGenesisAsset returns true if an asset is a genesis asset.
func (a *Asset) IsGenesisAsset() bool {
	return a.HasGenesisWitness()
}

// HasGenesisWitness determines whether an asset has a valid genesis witness,
// which should only have one input with a zero PrevID and empty witness and
// split commitment proof.
func (a *Asset) HasGenesisWitness() bool {
	if len(a.PrevWitnesses) != 1 {
		return false
	}

	witness := a.PrevWitnesses[0]
	if witness.PrevID == nil || witness.SplitCommitment != nil {
		return false
	}

	return *witness.PrevID == ZeroPrevID
}

// DeepEqual returns true if this asset is equal with the given asset.
func (a *Asset) DeepEqual(o *Asset) bool {

	// The ID commits to everything in the Genesis, including the type.
	if a.ID() != o.ID() {
		return false
	}

	if a.Amount != o.Amount {
		return false
	}

	if !mssmt.IsEqualNode(a.SplitCommitmentRoot, o.SplitCommitmentRoot) {
		return false
	}

	if !reflect.DeepEqual(a.ScriptPubkey, o.ScriptPubkey) {
		return false
	}

	if len(a.PrevWitnesses) != len(o.PrevWitnesses) {
		return false
	}

	for i := range a.PrevWitnesses {
		if !a.PrevWitnesses[i].DeepEqual(&o.PrevWitnesses[i]) {
			return false
		}
	}

	return true
}

func New(
	firstPrevOut wire.OutPoint,
	name string, outputIndex uint32,
	amount int32, scriptPubkey SerializedKey,
	splitCommitmentRoot *mssmt.ComputedNode,
) *Asset {

	genesis := NewGenesis(firstPrevOut, name, outputIndex)
	asset := NewAsset(genesis, amount, scriptPubkey, splitCommitmentRoot)

	return asset
}
