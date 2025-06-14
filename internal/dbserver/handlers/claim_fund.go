package handlers

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
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

func (h *Handler) ClaimFund(c *gin.Context) {
	var req ClaimFundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !common.IsHexAddress(req.WalletAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet address"})
		return
	}

	// Track database operation for checking wallet balance
	trackDBOp := metrics.TrackDBOperation("read", "wallet_balance")

	var rpcURL string
	switch req.Network {
	case "op_sepolia":
		rpcURL = fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "base_sepolia":
		rpcURL = fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid network specified"})
		return
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		h.logger.Errorf("Failed to connect to network: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to network"})
		return
	}

	address := common.HexToAddress(req.WalletAddress)
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		h.logger.Errorf("Failed to get balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance"})
		return
	}

	thresholdWei, ok := new(big.Int).SetString(config.GetFaucetFundAmount(), 10)
	if !ok {
		h.logger.Error("Failed to parse threshold amount")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if balance.Cmp(thresholdWei) >= 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Wallet balance is above the threshold",
		})
		return
	}

	privateKey, err := crypto.HexToECDSA(config.GetFaucetPrivateKey())
	if err != nil {
		h.logger.Errorf("Failed to parse private key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		h.logger.Error("Failed to cast public key to ECDSA")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		h.logger.Errorf("Failed to get nonce: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get nonce"})
		return
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get gas price: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get gas price"})
		return
	}

	tx := types.NewTransaction(nonce, address, thresholdWei, 21000, gasPrice, nil)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		h.logger.Errorf("Failed to get chain ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chain ID"})
		return
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		h.logger.Errorf("Failed to sign transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign transaction"})
		return
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		h.logger.Errorf("Failed to send transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send transaction"})
		return
	}

	c.JSON(http.StatusOK, ClaimFundResponse{
		Success:         true,
		Message:         "Funds sent successfully",
		TransactionHash: signedTx.Hash().Hex(),
	})

	trackDBOp(nil) // No error if we reach this point
}
