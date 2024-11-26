package crypto

import (
	"fmt"

	blst "github.com/supranational/blst/bindings/go"
	"github.com/umbracle/go-eth-consensus/bls"
)

type BLSCurve string

const (
	BLS12381 BLSCurve = "BLS12-381"
	BN254    BLSCurve = "BN254"
)

type BLSScheme interface {
	VerifySignature(pubKey, message, signature []byte, isG1 bool) (bool, error)
	AggregateSignatures(signatures [][]byte, isG1 bool) ([]byte, error)
	AggregatePublicKeys(pubKeys [][]byte, isG1 bool) ([]byte, error)
	VerifyAggregatedSignature(pubKeys [][]byte, message, signature []byte, isG1 bool) (bool, error)
	Sign(privKey, message []byte, isG1 bool) ([]byte, error)
	GenerateRandomKey() ([]byte, error)
	GetPublicKey(privKey []byte, isCompressed, isG1 bool) ([]byte, error)
	ConvertPublicKey(pubKey []byte, isCompressed, isG1 bool) ([]byte, error)
}

func NewBLSScheme(curve BLSCurve) BLSScheme {
	switch curve {
	case BLS12381:
		return &BLS12381Scheme{}
	case BN254:
		return &BN254Scheme{}
	default:
		panic("invalid curve: " + curve)
	}
}

type BLS12381Scheme struct {
}

var _ BLSScheme = (*BLS12381Scheme)(nil)

func (s *BLS12381Scheme) VerifySignature(pubKey, message, signature []byte, _ bool) (bool, error) {
	pubK := new(bls.PublicKey)
	if err := pubK.Deserialize(pubKey); err != nil {
		return false, err
	}
	sig := new(bls.Signature)
	if err := sig.Deserialize(signature); err != nil {
		return false, err
	}
	return sig.VerifyByte(pubK, message)
}

func (s *BLS12381Scheme) AggregateSignatures(signatures [][]byte, _ bool) ([]byte, error) {
	sigs := make([]*blst.P2Affine, len(signatures))
	for i, sig := range signatures {
		s := new(blst.P2Affine).Uncompress(sig)
		if s == nil {
			return nil, fmt.Errorf("failed to deserialize signature %d", i)
		}
		sigs[i] = s
	}
	
	aggregator := new(blst.P2Aggregate)
	if !aggregator.Aggregate(sigs, false) {
		return nil, fmt.Errorf("failed to aggregate signatures")
	}
	
	return aggregator.ToAffine().Compress(), nil
}

func (s *BLS12381Scheme) aggregatePublicKeys(pubKeys [][]byte) (*blst.P1Affine, error) {
	pks := make([]*blst.P1Affine, len(pubKeys))
	for i, pk := range pubKeys {
		pub := new(blst.P1Affine).Uncompress(pk)
		if pub == nil {
			return nil, fmt.Errorf("failed to deserialize the public key")
		}
		if !pub.KeyValidate() {
			return nil, fmt.Errorf("public key not in group")
		}
		pks[i] = pub
	}
	aggregator := new(blst.P1Aggregate)
	if !aggregator.Aggregate(pks, false) {
		return nil, fmt.Errorf("failed to aggregate public keys")
	}

	return aggregator.ToAffine(), nil
}

func (s *BLS12381Scheme) AggregatePublicKeys(pubKeys [][]byte, _ bool) ([]byte, error) {
	aggPk, err := s.aggregatePublicKeys(pubKeys)
	if err != nil {
		return nil, err
	}

	aggPkRaw := aggPk.Compress()
	return aggPkRaw[:], nil
}

func (s *BLS12381Scheme) VerifyAggregatedSignature(pubKeys [][]byte, message, signature []byte, _ bool) (bool, error) {
	sig := new(blst.P2Affine).Uncompress(signature)
	if sig == nil {
		return false, fmt.Errorf("failed to deserialize signature")
	}

	aggPk, err := s.aggregatePublicKeys(pubKeys)
	if err != nil {
		return false, err
	}

	dst := []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")
	return sig.Verify(true, aggPk, true, message, dst), nil
}

func (s *BLS12381Scheme) Sign(privKey, message []byte, _ bool) ([]byte, error) {
	priv := new(bls.SecretKey)
	if err := priv.Unmarshal(privKey); err != nil {
		return nil, err
	}
	sig, err := priv.Sign(message)
	if err != nil {
		return nil, err
	}
	sigRaw := sig.Serialize()
	return sigRaw[:], nil
}

func (s *BLS12381Scheme) GenerateRandomKey() ([]byte, error) {
	priv := bls.RandomKey()
	privRaw, err := priv.Marshal()
	if err != nil {
		return nil, err
	}
	return privRaw[:], nil
}

func (s *BLS12381Scheme) GetPublicKey(privKey []byte, isCompressed bool, _ bool) ([]byte, error) {
	priv := new(bls.SecretKey)
	if err := priv.Unmarshal(privKey); err != nil {
		return nil, err
	}

	pubRaw := priv.GetPublicKey().Serialize()
	if isCompressed {
		return pubRaw[:], nil
	}

	pubKey := new(blst.P1Affine).Uncompress(pubRaw[:])
	return pubKey.Serialize(), nil
}

func (s *BLS12381Scheme) ConvertPublicKey(pubKey []byte, isCompressed bool, _ bool) ([]byte, error) {
	if isCompressed {
		pubKey = new(blst.P1Affine).Deserialize(pubKey).Compress()

	} else {
		pubKey = new(blst.P1Affine).Uncompress(pubKey).Serialize()
	}

	return pubKey, nil
}