package handlers

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
)

type ClaimFundRequest struct {
	WalletAddress string `json:"wallet_address"`
	Network       string `json:"network"`
}

type ClaimFundResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	TransactionHash string `json:"transaction_hash,omitempty"`
}

func (h *Handler) ClaimFund(w http.ResponseWriter, r *http.Request) {
	var req ClaimFundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !common.IsHexAddress(req.WalletAddress) {
		http.Error(w, "Invalid wallet address", http.StatusBadRequest)
		return
	}

	var rpcURL string
	switch req.Network {
	case "op_sepolia":
		rpcURL = fmt.Sprintf("https://optimism-sepolia.g.alchemy.com/v2/%s", config.AlchemyAPIKey)
	case "base_sepolia":
		rpcURL = fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", config.AlchemyAPIKey)
	default:
		http.Error(w, "Invalid network specified", http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		h.logger.Errorf("Failed to connect to network: %v", err)
		http.Error(w, "Failed to connect to network", http.StatusInternalServerError)
		return
	}

	address := common.HexToAddress(req.WalletAddress)
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		h.logger.Errorf("Failed to get balance: %v", err)
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)
		return
	}

	thresholdWei, ok := new(big.Int).SetString(config.FaucetFundAmount, 10)
	if !ok {
		h.logger.Warnf("Failed to parse FaucetFundAmount: %s", config.FaucetFundAmount)
	}

	if balance.Cmp(thresholdWei) >= 0 {
		json.NewEncoder(w).Encode(ClaimFundResponse{
			Success: false,
			Message: "Account already has sufficient funds",
		})
		return
	}

	privateKey, err := crypto.HexToECDSA(config.FaucetPrivateKey)
	if err != nil {
		h.logger.Errorf("Failed to load private key: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tx, err := h.sendFunds(client, privateKey, address)
	if err != nil {
		h.logger.Errorf("Failed to send funds: %v", err)
		http.Error(w, "Failed to send funds", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(ClaimFundResponse{
		Success:         true,
		Message:         "Funds sent successfully",
		TransactionHash: tx.Hash().Hex(),
	})
}

func (h *Handler) sendFunds(client *ethclient.Client, privateKey *ecdsa.PrivateKey, to common.Address) (*types.Transaction, error) {
	ctx := context.Background()

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	from := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	value, ok := new(big.Int).SetString(config.FaucetFundAmount, 10)
	if !ok {
		h.logger.Warnf("Failed to parse FaucetFundAmount: %s", config.FaucetFundAmount)
	}

	gasLimit := uint64(21000)

	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}
