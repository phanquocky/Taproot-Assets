package proof

import "crypto/sha256"

type HashedProof struct {
	ProofBytes []byte
	Hash       [sha256.Size]byte
}

type File struct {
	Proofs []*HashedProof
}
