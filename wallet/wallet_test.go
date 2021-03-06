package wallet

import (
	"encoding/hex"
	"fmt"
	"testing"

	blswallet "github.com/kilic/bn254/bls"
	"github.com/stretchr/testify/require"
)

func TestSignAndVerify(t *testing.T) {
	wallet, err := NewWallet()
	require.Equal(t, err, nil, "error creating wallet")
	signBytes := []byte("0x123222")
	signature, err := wallet.Sign(signBytes)
	require.Equal(t, err, nil, "error signing transaction")
	fmt.Println(hex.EncodeToString(signature.ToBytes()))
	valid, err := wallet.VerifySignature(signBytes, signature, *wallet.signer.Account.Public)
	require.Equal(t, err, nil, "error verifying signature")
	require.Equal(t, valid, true, "error verifying signature")
}

func TestVerifyAggregated(t *testing.T) {
	signBytes := []byte("0x123222")
	signerSize := 2
	publicKeys := make([]*blswallet.PublicKey, signerSize)
	messages := make([]blswallet.Message, signerSize)
	signatures := make([]*blswallet.Signature, signerSize)
	for i := 0; i < signerSize; i++ {
		wallet, err := NewWallet()
		if err != nil {
			t.Fatal(err)
		}
		accountSignature, err := wallet.Sign(signBytes)
		if err != nil {
			t.Fatal(err)
		}
		messages[i] = signBytes
		publicKeys[i] = wallet.signer.Account.Public
		signatures[i] = &accountSignature
	}

	aggregatedSignature := blswallet.AggregateSignatures(signatures)
	aggregatedSignatureWallet, err := NewAggregateSignature(signatures)
	if err != nil {
		panic(err)
	}
	fmt.Println("aggregated sig", hex.EncodeToString(aggregatedSignature.ToBytes()), hex.EncodeToString(aggregatedSignatureWallet.ToBytes()))
	verified, err := VerifyAggregatedSignature(messages, publicKeys, aggregatedSignatureWallet)
	if err != nil {
		t.Fatal(err)
	}
	if !verified {
		t.Fatalf("signature is not verified")
	}
}
