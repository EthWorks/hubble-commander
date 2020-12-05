package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"

	ethCmn "github.com/ethereum/go-ethereum/common"
)

var (
	ErrInvalidPubkeyLen = errors.New("invalid pubkey length")
)

type Hash ethCmn.Hash
type Address ethCmn.Address

type ByteArray [32]byte

func (b ByteArray) String() string {
	bz := b[:]
	enc := make([]byte, len(bz)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], bz)
	return string(enc)
}

// String has to be prefixed with 0x
func HexToByteArray(a string) (b ByteArray, err error) {
	bz, err := hex.DecodeString(a[2:])
	if err != nil {
		return b, err
	}
	return BytesToByteArray(bz), nil
}

func BytesToByteArray(bz []byte) ByteArray {
	var temp [32]byte
	copy(temp[:], bz)
	return temp
}

// Pubkey is an alias for public key
type Pubkey [4]*big.Int

const pubkeyLength = 128

// NewPubkeyFromBytes creates a pubkey from bytes
func NewPubkeyFromBytes(bz []byte) (pubkey Pubkey, err error) {
	if len(bz) != pubkeyLength {
		return pubkey, ErrInvalidPubkeyLen
	}
	chunkSize := 32
	for i := 0; i < pubkeyLength; i += chunkSize {
		end := i + chunkSize
		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(bz) {
			end = len(bz)
		}
		pubkeyPart := bz[i:end]
		tempPubkeyPart := big.NewInt(0)
		tempPubkeyPart = tempPubkeyPart.SetBytes(pubkeyPart)
		pubkey[i/chunkSize] = tempPubkeyPart
	}
	return pubkey, nil
}

func (p Pubkey) String() (string, error) {
	pubBytes, err := p.serialize()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(pubBytes), nil
}

// Serialize seralises a public key to bytes
func (p Pubkey) serialize() ([]byte, error) {
	var pubkey []uint64
	for i := range p {
		pub := p[i].Uint64()
		pubkey = append(pubkey, pub)
	}
	return json.Marshal(pubkey)
}

// bytesToPubkey converts bytes to Pubkey
func bytesToPubkey(b []byte) (pubkey Pubkey, err error) {
	var p []uint64
	err = json.Unmarshal(b, &p)
	if err != nil {
		return pubkey, err
	}
	var pubkeyBigInt [4]*big.Int
	for i := range p {
		temp := big.NewInt(0)
		temp.SetUint64(p[i])
		pubkeyBigInt[i] = temp
	}
	return pubkeyBigInt, nil
}

func StrToPubkey(s string) (pubkey Pubkey, err error) {
	pubBytes, err := hex.DecodeString(s)
	if err != nil {
		return
	}

	return bytesToPubkey(pubBytes)
}

type TypesUserState struct {
	PubkeyIndex *big.Int
	TokenType   *big.Int
	Balance     *big.Int
	Nonce       *big.Int
}
type TypesStateMerkleProof struct {
	State   TypesUserState
	Witness [][32]byte
}
