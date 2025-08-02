package aggregator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	bn254utils "github.com/Layr-Labs/eigensdk-go/crypto/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/ethereum/go-ethereum/crypto"
)


// BN254 field order (same as JavaScript FIELD_ORDER)
var FIELD_ORDER, _ = new(big.Int).SetString("30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47", 16)

func getKeyPair(privateKey string) (*bls.KeyPair, error) {
	// Add 0x prefix if not present (same as JS prefixSeed function)
	var prefixedPrivateKey string
	if len(privateKey) >= 2 && privateKey[:2] == "0x" {
		prefixedPrivateKey = privateKey
	} else {
		prefixedPrivateKey = "0x" + privateKey
	}

	log.Println("Go prefixedPrivateKey:", prefixedPrivateKey)

	// Create Fr element from hashed private key (like JavaScript fr.setHashOf)
	// Use SHA256 like the expandMsg algorithm to be consistent
	frElement := new(fr.Element)
	hasher := sha256.New()
	hasher.Write([]byte(prefixedPrivateKey))
	frElement.SetBytes(hasher.Sum(nil))

	// Create KeyPair from Fr element
	keyPair := bls.NewKeyPair(frElement)

	return keyPair, nil
}

func getDomain() []byte {
	return crypto.Keccak256([]byte("TasksManager"))
}

// Exact implementation of JavaScript expandMsg function
func expandMsg(domain, msg []byte, outLen int) ([]byte, error) {
	if len(domain) > 32 {
		return nil, fmt.Errorf("bad domain size")
	}

	out := make([]byte, outLen)

	// Build in0: 64 zero bytes + msg + l_i_b_str + I2OSP(0,1) + DST_prime + len(domain)
	len0 := 64 + len(msg) + 2 + 1 + len(domain) + 1
	in0 := make([]byte, len0)

	// Zero pad (first 64 bytes are already zero)
	off := 64

	// msg
	copy(in0[off:], msg)
	off += len(msg)

	// l_i_b_str (outLen as 2 bytes big endian)
	in0[off] = byte((outLen >> 8) & 0xff)
	in0[off+1] = byte(outLen & 0xff)
	off += 2

	// I2OSP(0, 1)
	in0[off] = 0
	off += 1

	// DST_prime
	copy(in0[off:], domain)
	off += len(domain)
	in0[off] = byte(len(domain))

	// b0 = SHA256(in0) - use SHA256 like JavaScript ethers.sha256
	b0Hasher := sha256.New()
	b0Hasher.Write(in0)
	b0 := b0Hasher.Sum(nil)

	// Build in1: b0 + I2OSP(1,1) + DST_prime + len(domain)
	len1 := 32 + 1 + len(domain) + 1
	in1 := make([]byte, len1)

	// b0
	copy(in1[0:], b0)
	off = 32

	// I2OSP(1, 1)
	in1[off] = 1
	off += 1

	// DST_prime
	copy(in1[off:], domain)
	off += len(domain)
	in1[off] = byte(len(domain))

	// b1 = SHA256(in1) - use SHA256 like JavaScript ethers.sha256
	b1Hasher := sha256.New()
	b1Hasher.Write(in1)
	b1 := b1Hasher.Sum(nil)

	// ell = ceil(outLen / 32)
	ell := (outLen + 32 - 1) / 32
	bi := b1

	for i := 1; i < ell; i++ {
		// ini = strxor(b0, bi) + I2OSP(1+i, 1) + DST_prime + len(domain)
		ini := make([]byte, 32+1+len(domain)+1)

		// strxor(b0, bi)
		for j := 0; j < 32; j++ {
			ini[j] = b0[j] ^ bi[j]
		}
		off = 32

		// I2OSP(1+i, 1)
		ini[off] = byte(1 + i)
		off += 1

		// DST_prime
		copy(ini[off:], domain)
		off += len(domain)
		ini[off] = byte(len(domain))

		// Copy previous bi to output
		copy(out[32*(i-1):], bi)

		// bi = SHA256(ini) - use SHA256 like JavaScript ethers.sha256
		biHasher := sha256.New()
		biHasher.Write(ini)
		bi = biHasher.Sum(nil)
	}

	// Copy final bi to output
	copy(out[32*(ell-1):], bi)
	return out, nil
}

// Exact implementation of JavaScript hashToField function
func hashToField(domain, msg []byte, count int) ([]*big.Int, error) {
	u := 48 // Length for BN254
	expandedMsg, err := expandMsg(domain, msg, count*u)
	if err != nil {
		return nil, err
	}

	els := make([]*big.Int, count)
	for i := 0; i < count; i++ {
		slice := expandedMsg[i*u : (i+1)*u]
		el := new(big.Int).SetBytes(slice)
		el.Mod(el, FIELD_ORDER)
		els[i] = el
	}
	return els, nil
}

// Exact implementation of JavaScript mapToPoint function
func mapToPoint(e *big.Int) (*bn254.G1Affine, error) {
	// Create Fp element from big.Int (like JS e1.setStr)
	var fpElement fp.Element
	fpElement.SetBigInt(e)

	// Use eigensdk-go's MapToCurve which should be equivalent to mcl's mapToG1
	var hashArray [32]byte
	eBytes := e.Bytes()
	// Pad to 32 bytes if needed
	if len(eBytes) <= 32 {
		copy(hashArray[32-len(eBytes):], eBytes)
	} else {
		copy(hashArray[:], eBytes[len(eBytes)-32:])
	}

	return bn254utils.MapToCurve(hashArray), nil
}

// Exact implementation of JavaScript hashToPoint function
func hashToPoint(messageHex string, domainBytes []byte) (*bn254.G1Affine, error) {
	// Convert hex message to bytes (like JavaScript ethers.getBytes)
	messageBytes, err := hex.DecodeString(messageHex[2:]) // Remove 0x prefix
	if err != nil {
		return nil, err
	}

	// hashToField(domain, msg, 2) - get two field elements
	fieldElements, err := hashToField(domainBytes, messageBytes, 2)
	if err != nil {
		return nil, err
	}

	// mapToPoint for each element
	p0, err := mapToPoint(fieldElements[0])
	if err != nil {
		return nil, err
	}

	p1, err := mapToPoint(fieldElements[1])
	if err != nil {
		return nil, err
	}

	// Add the two points (like JavaScript mcl.add(p0, p1))
	result := new(bn254.G1Affine)
	result.Add(p0, p1)

	return result, nil
}
