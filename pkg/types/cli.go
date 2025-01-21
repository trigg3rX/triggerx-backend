package types

import (
	regcoord "github.com/trigg3rX/triggerx-contracts/bindings/contracts/RegistryCoordinator"
)

type RegisterKeeperRequest struct {
	KeeperAddress     string   `json:"keeper_address"`
	PubkeyRegistrationParams regcoord.IBLSApkRegistryPubkeyRegistrationParams `json:"pubkey_registration_params"`
	SignatureWithSaltAndExpiry regcoord.ISignatureUtilsSignatureWithSaltAndExpiry `json:"signature_with_salt_and_expiry"`
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
