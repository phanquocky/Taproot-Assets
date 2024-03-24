package taproot

import (
	"context"
	"errors"
	"fmt"
	"github.com/quocky/taproot-asset/taproot/model/asset"
	"github.com/quocky/taproot-asset/taproot/model/commitment"
	"github.com/quocky/taproot-asset/taproot/model/proof"
	"github.com/quocky/taproot-asset/taproot/onchain"
	"github.com/quocky/taproot-asset/taproot/utils"
	"log"
)

func (t *Taproot) MintAsset(names []string, amounts []int32) error {
	if len(names) != len(amounts) {
		return errors.New("len names and amount is different")
	}

	ctx := context.Background()

	var (
		// TODO: update expected amount when have multiple mint assets
		expectedAmount = int32(DEFAULT_OUTPUT_AMOUNT + DEFAULT_FEE)
		pubkey         = asset.ToSerialized(t.wif.PrivKey.PubKey())

		mintAssets       = make([]*asset.Asset, len(names))
		assetCommitments = make([]*commitment.AssetCommitment, len(names))
	)

	utxosResult, err := t.btcClient.GetUTXOByAmount(expectedAmount)
	if err != nil {
		return err
	}

	for i, name := range names {
		amount := amounts[i]

		if len(name) < 6 {
			return fmt.Errorf("expected len your asset name is greater than or equal 6 (but len = %d) ", len(name))
		}

		if amount == 0 {
			return fmt.Errorf("expected amount greater than 0 (but amount = %d)", len(name))
		}

		log.Printf("[Mint Asset] mintint asset! name : %v, amount : %v ! \n", name, amount)

		mintAssets[i], err = asset.New(*utxosResult.UTXOs[0], name,
			DEFAULT_MINTING_OUTPUT_INDEX, amount,
			pubkey, nil,
		)
		if err != nil {
			return err
		}
		fmt.Print("[Mint Asset] New asset created: ")
		utils.PrintStruct(mintAssets[i])

		assetCommitments[i], err = commitment.NewAssetCommitment(ctx, mintAssets[i])
		if err != nil {
			return err
		}
		fmt.Print("[Mint Asset] New asset commitment created: ")
		utils.PrintStruct(assetCommitments[i])
	}

	tapCommitment, err := commitment.NewTapCommitment(assetCommitments...)
	if err != nil {
		log.Println("[Mint Asset] create top commitment fail", "err", err)

		return err
	}

	mintAddressResult, err := t.addressMaker.CreateTapAddrByCommitment(pubkey, tapCommitment)
	if err != nil {
		return err
	}

	receivers := []*onchain.Receiver{
		onchain.NewReceiver(mintAddressResult, mintAssets...),
	}

	txIncludeOutInternalKey, err := t.prepareTx(utxosResult, nil, DEFAULT_OUTPUT_AMOUNT,
		receivers, true,
	)
	if err != nil {
		return err
	}

	mintProof, err := createMintProof(
		txIncludeOutInternalKey,
		DEFAULT_MINTING_OUTPUT_INDEX,
		pubkey,
		tapCommitment,
	)

	fmt.Println("Mint Proof: ", mintProof)

	//fmt.Println("Tx Include Out Internal Key: ", hex.EncodeToString(rawTxBytes.Bytes()))
	//
	//hash, err := t.btcClient.SendRawTx(txIncludeOutInternalKey.Tx)
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Println("Mint asset success! Tx hash: ", hash.String())
	return nil
}

// createMintProof create mint proof and store data to locator.
func createMintProof(
	txIncludeOutPubKey *onchain.TxIncludeOutInternalKey,
	outputIndex int32,
	pubkey asset.SerializedKey,
	tapCommitment *commitment.TapCommitment,
) (proof.AssetProofs, error) {
	baseProof := &proof.MintParams{
		BaseProofParams: proof.BaseProofParams{
			Tx:               txIncludeOutPubKey.Tx,
			OutputIndex:      outputIndex,
			InternalKey:      pubkey,
			TaprootAssetRoot: tapCommitment,
		},
		GenesisPoint: txIncludeOutPubKey.Tx.TxIn[0].PreviousOutPoint,
	}

	fmt.Println("Base Proof: ", baseProof)

	err := baseProof.BaseProofParams.AddExclusionProofs(
		txIncludeOutPubKey,
		func(idx uint32) bool {
			return idx == uint32(outputIndex)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to add exclusion proofs: "+
			"%w", err)
	}

	fmt.Println("Base Proof: ", baseProof)

	mintProofs, err := proof.NewMintingBlobs(baseProof)
	if err != nil {
		return nil, fmt.Errorf("unable to construct minting "+
			"proofs: %w", err)
	}

	fmt.Println("Mint Proofs: ", mintProofs)

	return mintProofs, nil
}
