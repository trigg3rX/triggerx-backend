package proof

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"time"
)

type TLSProof struct {
	CertificateHash string `json:"certificateHash"`
	ResponseHash    string `json:"responseHash"`
	Timestamp       string `json:"timestamp"`
}

type KeeperResponse interface {
	GetData() []byte
}

func GenerateProof(response KeeperResponse, connState *tls.ConnectionState) (*TLSProof, error) {
	if connState == nil || len(connState.PeerCertificates) == 0 {
		return nil, errors.New("no TLS certificates found")
	}

	certHash := sha256.Sum256(connState.PeerCertificates[0].Raw)
	certHashStr := hex.EncodeToString(certHash[:])

	respHash := sha256.Sum256(response.GetData())
	respHashStr := hex.EncodeToString(respHash[:])

	return &TLSProof{
		CertificateHash: certHashStr,
		ResponseHash:    respHashStr,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
	}, nil
}
