package handlers

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ClaimFundRequest struct {
	WalletAddress string `json:"wallet_address"`
	Network       string `json:"network"` // "op_sepolia" or "base_sepolia"
}

type ClaimFundResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	TransactionHash string `json:"transaction_hash,omitempty"`
}

const (
	OP_SEPOLIA_RPC   = "https://sepolia.optimism.io"
	BASE_SEPOLIA_RPC = "https://sepolia.base.org"
	FUND_AMOUNT      = 0.05 // ETH
)

func (h *Handler) ClaimFund(w http.ResponseWriter, r *http.Request) {
	var req ClaimFundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate wallet address
	if !common.IsHexAddress(req.WalletAddress) {
		http.Error(w, "Invalid wallet address", http.StatusBadRequest)
		return
	}

	// Select RPC URL based on network
	var rpcURL string
	switch req.Network {
	case "op_sepolia":
		rpcURL = OP_SEPOLIA_RPC
	case "base_sepolia":
		rpcURL = BASE_SEPOLIA_RPC
	default:
		http.Error(w, "Invalid network specified", http.StatusBadRequest)
		return
	}

	// Connect to the network
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		h.logger.Errorf("Failed to connect to network: %v", err)
		http.Error(w, "Failed to connect to network", http.StatusInternalServerError)
		return
	}

	// Check recipient's balance
	address := common.HexToAddress(req.WalletAddress)
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		h.logger.Errorf("Failed to get balance: %v", err)
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)
		return
	}

	// Convert 0.005 ETH to Wei
	threshold := new(big.Float).Mul(big.NewFloat(FUND_AMOUNT), big.NewFloat(1e18))
	thresholdWei, _ := threshold.Int(nil)

	if balance.Cmp(thresholdWei) >= 0 {
		json.NewEncoder(w).Encode(ClaimFundResponse{
			Success: false,
			Message: "Account already has sufficient funds",
		})
		return
	}

	// Get the private key for the funding wallet
	privateKey, err := crypto.HexToECDSA(os.Getenv("FUNDER_PRIVATE_KEY"))
	if err != nil {
		h.logger.Errorf("Failed to load private key: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create and send the transaction
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

	// Get the funding wallet address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	from := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, err
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	// Create transaction data
	value := new(big.Int)
	fundAmountWei := new(big.Float).Mul(big.NewFloat(FUND_AMOUNT), big.NewFloat(1e18))
	value, _ = fundAmountWei.Int(value) // Convert to Wei

	// Estimate gas limit
	gasLimit := uint64(21000) // Standard ETH transfer gas limit

	// Create transaction
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, nil)

	// Sign transaction
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	// Send transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}
