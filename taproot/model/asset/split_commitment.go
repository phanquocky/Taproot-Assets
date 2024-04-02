package asset

import (
	"bytes"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"log"
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
