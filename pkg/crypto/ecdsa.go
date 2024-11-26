package crypto

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"golang.org/x/crypto/sha3"
)

func Hash(data ...[]byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	for _, d := range data {
		hash.Write(d[:])
	}
	return hash.Sum(nil)
}

func PoseidonHash(data ...[]byte) []byte {
	msg := []byte{}
	for _, d := range data {
		msg = append(msg, d...)
	}
	hash, err := poseidon.HashBytes(msg)
	if err != nil {
		panic(fmt.Errorf("poseidon hash failed: %v", err))
	}
	return hash.Bytes()
}

func VerifyECDSASignature(message, signature []byte) (bool, common.Address, error) {
	publicKey, err := crypto.SigToPub(message, signature)
	if err != nil {
		return false, common.Address{}, err
	}
	pubKey := crypto.FromECDSAPub(publicKey)
	addr := crypto.PubkeyToAddress(*publicKey)
	return crypto.VerifySignature(pubKey, message, signature[:len(signature)-1]), common.BytesToAddress(addr[:]), nil
}

func GetSigner(ctx context.Context, c *ethclient.Client, accHexPrivateKey string) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(accHexPrivateKey, "0x"))
	if err != nil {
		return nil, err
	}
	chainID, err := c.NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	return bind.NewKeyedTransactorWithChainID(privateKey, chainID)
}