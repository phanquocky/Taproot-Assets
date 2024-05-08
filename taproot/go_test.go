package taproot

import (
	"encoding/json"
	"fmt"
	"github.com/quocky/taproot-asset/taproot/model/mssmt"
	"testing"
)

func TestAbc(t *testing.T) {
	key := [32]byte{1, 2, 3, 4, 5, 6, 7, 8}
	value := []byte("abc")
	leafNode := mssmt.LeafNode{
		Value: value,
	}
	tree := mssmt.NewCompactedTree(mssmt.NewDefaultStore())
	tree.Insert(nil, key, &leafNode)
	fmt.Println("abcdef ")
	b, err := json.Marshal(tree)
	if err != nil {
		fmt.Println("Marshal error: ", err)
		t.Error(err)
	}

	fmt.Println("abcbyte: ", string(b))
	t.Log("Test abc")
}
