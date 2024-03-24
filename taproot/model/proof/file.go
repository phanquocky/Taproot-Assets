package proof

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/utils"
	"log"
	"os"
)

const LocatorFilePath = "./locator/%x"

var (
	ErrNoProofAvailable = errors.New("no proof available")
)

type HashedProof struct {
	ProofBytes []byte
	Hash       [sha256.Size]byte
}

type File struct {
	Proofs []*HashedProof
}

// NewFile returns a new proof file given a version and a series of state
// transition Proofs.
func NewFile(proofs ...Proof) (*File, error) {
	var (
		prevHash     [sha256.Size]byte
		linkedProofs = make([]*HashedProof, len(proofs))
	)

	// We start out with the zero Hash as the previous Hash and then create
	// the checksum of SHA256(prev_hash || proof) as our incremental
	// checksum for each of the Proofs, basically building a proof chain
	// similar to Bitcoin's time chain.
	for idx := range proofs {
		proof := proofs[idx]

		proofBytes, err := json.Marshal(proof)
		if err != nil {
			return nil, err
		}

		linkedProofs[idx] = &HashedProof{
			ProofBytes: proofBytes,
			Hash:       hashProof(proofBytes, prevHash),
		}
		prevHash = linkedProofs[idx].Hash
	}

	return &File{
		Proofs: linkedProofs,
	}, nil
}

// hashProof hashes a proof's content together with the previous Hash and
// re-uses the given buffer:
//
//	SHA256(prev_hash || proof_bytes).
func hashProof(proofBytes []byte, prevHash [32]byte) [32]byte {
	h := sha256.New()
	_, _ = h.Write(prevHash[:])
	_, _ = h.Write(proofBytes)
	return *(*[32]byte)(h.Sum(nil))
}

// IsEmpty returns true if the file does not contain any Proofs.
func (f *File) IsEmpty() bool {
	return len(f.Proofs) == 0
}

// IsValid combines multiple sanity checks for proof file validity.
func (f *File) IsValid() error {
	if f.IsEmpty() {
		return ErrNoProofAvailable
	}

	return nil
}

// ProofAt returns the proof at the given index. If the file is empty, this
// returns ErrNoProofAvailable.
func (f *File) ProofAt(index uint32) (*Proof, error) {
	if err := f.IsValid(); err != nil {
		return nil, err
	}

	if index > uint32(len(f.Proofs))-1 {
		return nil, fmt.Errorf("invalid index %d", index)
	}

	var (
		proof      = &Proof{}
		proofBytes = f.Proofs[index].ProofBytes
	)
	if err := json.Unmarshal(proofBytes, &proof); err != nil {
		return nil, fmt.Errorf("error decoding proof: %v", err)
	}

	return proof, nil
}

// LastProof returns the last proof in the chain of Proofs. If the file is
// empty, this return nil.
func (f *File) LastProof() (*Proof, error) {
	if err := f.IsValid(); err != nil {
		return nil, err
	}

	return f.ProofAt(uint32(len(f.Proofs)) - 1)
}

// AppendProof appends a proof to the file and calculates its chained hash.
func (f *File) AppendProof(proof Proof) error {
	var prevHash [sha256.Size]byte

	if !f.IsEmpty() {
		prevHash = f.Proofs[len(f.Proofs)-1].Hash
	}

	proofBytes, err := json.Marshal(&proof)
	if err != nil {
		log.Println("[AppendProof] err := json.Marshal(&proof), err ", err)

		return err
	}

	f.Proofs = append(f.Proofs, &HashedProof{
		ProofBytes: proofBytes,
		Hash:       hashProof(proofBytes, prevHash),
	})

	return nil
}

func (f *File) Store() ([32]byte, error) {
	lastProof, err := f.LastProof()
	if err != nil {
		log.Println("[file.Store] get last proof fail", err)

		return [32]byte{}, err
	}

	locator := Locator{
		AssetID:   utils.ToPtr(lastProof.Asset.ID()),
		ScriptKey: lastProof.Asset.ScriptPubkey,
		OutPoint: wire.NewOutPoint(
			utils.ToPtr(lastProof.AnchorTx.TxHash()),
			lastProof.InclusionProof.OutputIndex,
		),
	}

	filenameBytes, err := locator.Hash()
	if err != nil {
		log.Println("[proof.Store] Hash locator fail", err)

		return [32]byte{}, err
	}

	filename := fmt.Sprintf(LocatorFilePath, filenameBytes)

	fileBytes, err := json.Marshal(*f)
	if err != nil {
		log.Println("fileBytes, err := json.Marshal(hashProofs) fail", err)

		return [32]byte{}, err
	}

	_ = os.Mkdir("locator", 0750)

	err = os.WriteFile(filename, fileBytes, 0666)
	if err != nil {
		log.Println("os.WriteFile(nameFile, hashProofsData fail", err)

		return [32]byte{}, err
	}

	return filenameBytes, nil
}

func (f *File) Decode(blob []byte) error {
	if err := json.Unmarshal(blob, f); err != nil {
		return err
	}

	return nil
}

func FileBytesFromName(nameFile string) ([]byte, error) {
	fileByte, err := os.ReadFile(nameFile)
	if err != nil {
		log.Println("[FileFromName] fileByte, err := os.ReadFile(nameFile)", err)

		return nil, err
	}

	return fileByte, nil
}

// AssetSnapshot commits to the result of a valid proof within a proof file.
// This represents the state of an asset's lineage at a given point in time.
type AssetSnapshot struct {
	// Asset is the resulting asset of a valid proof.
	Asset *asset.Asset

	// OutPoint is the outpoint that commits to the asset specified above.
	OutPoint wire.OutPoint

	// AnchorTx is the transaction that commits to the above asset.
	AnchorTx *wire.MsgTx

	// OutputIndex is the output index in the above transaction that
	// commits to the output.
	OutputIndex uint32

	// InternalKey is the internal key used to commit to the above asset in
	// the AnchorTx.
	InternalKey asset.SerializedKey

	// ScriptRoot is the Taproot Asset commitment root committed to using
	// the above internal key in the Anchor transaction.
	ScriptRoot *commitment.TapCommitment

	// SplitAsset is the optional indicator that the asset in the snapshot
	// resulted from splitting an asset. If this is true then the root asset
	// of the split can be found in the asset witness' split commitment.
	SplitAsset bool
}
