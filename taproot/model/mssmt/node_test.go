package mssmt

import (
	"context"
	"encoding/json"
	"log"
	"testing"
)

func TestMarshalNode(t *testing.T) {
	splitTree := NewCompactedTree(NewDefaultStore())
	rootNode, err := splitTree.Root(context.TODO())
	if err != nil {
		log.Println(" splitTree.Root(context.TODO()), err ", err)
		return
	}

	rootBytes, _ := json.Marshal(rootNode)
	log.Println("rootBytes: ", rootBytes)
}
