package types

type RegisterKeeperRequest struct {
	KeeperAddress     string   `json:"keeper_address"`
	Signature         string   `json:"signature"`
	Salt              string   `json:"salt"`
	Expiry            string   `json:"expiry"`
	BlsPublicKey      string   `json:"bls_public_key"`
	// TokenStrategyAddr string   `json:"token_strategy_addr"`
	// StakeAmount       string   `json:"stake_amount"`
}

type RegisterKeeperResponse struct {
	KeeperID          int64  `json:"keeper_id"`
	RegisteredTx     string `json:"registered_tx"`
	PeerID           string `json:"peer_id"`
}

type DeregisterKeeperRequest struct {
	KeeperID int64 `json:"keeper_id"`
}

type DeregisterKeeperResponse struct {
	DeregisteredTx string `json:"deregistered_tx"`
}
