package asset

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/quocky/taproot-asset/taproot/model/mssmt"
)

type SplitCommitment struct {
	Proof     mssmt.Proof
	RootAsset Asset
}

// DeepEqual returns true if this split commitment is equal with the given split
// commitment.
func (s *SplitCommitment) DeepEqual(o *SplitCommitment) bool {
	if s == nil || o == nil {
		return s == o
	}

	if len(s.Proof.Nodes) != len(o.Proof.Nodes) {
		return false
	}

	for i := range s.Proof.Nodes {
		nodeA := s.Proof.Nodes[i]
		nodeB := o.Proof.Nodes[i]
		if !mssmt.IsEqualNode(nodeA, nodeB) {
			return false
		}
	}

	// We can't directly compare the root assets, as some non-TLV fields
	// might be different in unit tests. To avoid introducing flakes, we
	// only compare the encoded TLV data.
	// var bufA, bufB bytes.Buffer

	// We ignore errors here, these possible errors (incorrect TLV stream
	// being created) are covered in unit tests.
	leafS, err := s.RootAsset.Leaf()
	if err != nil {
		log.Println("leafS, err := s.RootAsset.Leaf()", err)

		return false
	}

	leadO, err := o.RootAsset.Leaf()
	if err != nil {
		log.Println("leadO, err = o.RootAsset.Leaf()", err)

		return false
	}

	return bytes.Equal(leadO.Value, leafS.Value)
}

type SplitCommitmentByte struct {
	Proof []byte

	RootAsset Asset
}

// MarshalJSON function custom CommitmentProof when using json.Marshal function
func (u SplitCommitment) MarshalJSON() ([]byte, error) {
	var proofBytes bytes.Buffer

	err := u.Proof.Compress().Encode(&proofBytes)
	if err != nil {
		log.Println("[MarshalJSON] u.Proof.Compress().Encode(&proofBytes), err ", err)

		return nil, err
	}

	return json.Marshal(SplitCommitmentByte{
		Proof:     proofBytes.Bytes(),
		RootAsset: u.RootAsset,
	})
}

func (b *SplitCommitment) UnmarshalJSON(data []byte) error {

	var (
		commitBytes   SplitCommitmentByte
		compressProof mssmt.CompressedProof
	)

	if err := json.Unmarshal(data, &commitBytes); err != nil {
		log.Println("err := json.Unmarshal(data, &commitBytes), err ", err)

		return err
	}

	b.Proof.Compress()
	// commitBytes.Proof
	compressProof.Decode(bytes.NewReader(commitBytes.Proof))
	proof, err := compressProof.Decompress()
	if err != nil {
		log.Println("err := compressProof.Decompress(), err ", err)

		return err
	}

	b.Proof = *proof
	b.RootAsset = commitBytes.RootAsset

	return nil
}
