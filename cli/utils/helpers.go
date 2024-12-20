package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
)

func RandomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func Uint64ToBytes(num uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, num)
	return b
}

func Hex2Bytes(hex string) []byte {
	return common.FromHex(hex)
}

func Bytes2Hex(bytes []byte) string {
	return common.Bytes2Hex(bytes)
}

func GetValidAddress(hex string) string {
	return common.HexToAddress(hex).Hex()
}