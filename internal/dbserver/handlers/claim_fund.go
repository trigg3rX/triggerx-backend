package handlers

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/http"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ClaimFundRequest struct {
	WalletAddress string `json:"wallet_address" validate:"required,eth_addr"`
	Network       string `json:"network" validate:"required,oneof=op_sepolia base_sepolia arbitrum_sepolia"`
}

type ClaimFundResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	TransactionHash string `json:"transaction_hash,omitempty"`
}

func (h *Handler) ClaimFund(c *gin.Context) {
	logger := h.getLogger(c)
	var req ClaimFundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Errorf("%s: %v", errors.ErrInvalidRequestBody, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	logger.Debugf("POST [ClaimFund] Wallet %s on network %s", req.WalletAddress, req.Network)

	var rpcURL string
	switch req.Network {
	case "op_sepolia":
		rpcURL = fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "base_sepolia":
		rpcURL = fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	case "arbitrum_sepolia":
		rpcURL = fmt.Sprintf("https://arb-sepolia.g.alchemy.com/v2/%s", config.GetAlchemyAPIKey())
	default:
		logger.Debugf("Invalid network specified: %s", req.Network)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid network specified"})
		return
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		logger.Debugf("Failed to connect to network: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to network"})
		return
	}

	address := common.HexToAddress(req.WalletAddress)
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		logger.Debugf("Failed to get balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance"})
		return
	}

	if types.IsGreater(balance.String(), config.GetFaucetFundAmount()) || types.IsEqual(balance.String(), config.GetFaucetFundAmount()) {
		logger.Debugf("Wallet balance is above the threshold")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Wallet balance is above the threshold",
		})
		return
	}

	privateKey, err := crypto.HexToECDSA(config.GetFaucetPrivateKey())
	if err != nil {
		logger.Debugf("Failed to parse private key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Debugf("Failed to cast public key to ECDSA")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logger.Debugf("Failed to get nonce: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get nonce"})
		return
	}

	thresholdWei, err := types.ParseBigInt(types.Sub(config.GetFaucetFundAmount(), balance.String()))
	if err != nil {
		logger.Debugf("Failed to parse deposit amount: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse deposit amount"})
		return
	}

	// Estimate gas limit for the transfer to avoid hardcoded 21000
	callMsg := ethereum.CallMsg{From: fromAddress, To: &address, Value: thresholdWei}
	gasLimit, err := client.EstimateGas(context.Background(), callMsg)
	if err != nil {
		logger.Debugf("Failed to estimate gas, falling back to 21000: %v", err)
		gasLimit = 21000
	}

	// Use EIP-1559 dynamic fees
	maxPriorityFeePerGas, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		logger.Debugf("Failed to get gas tip cap: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get gas tip cap"})
		return
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		logger.Debugf("Failed to get latest header: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest header"})
		return
	}
	baseFee := big.NewInt(0)
	if header.BaseFee != nil {
		baseFee = new(big.Int).Set(header.BaseFee)
	}
	maxFeePerGas := new(big.Int).Add(new(big.Int).Mul(baseFee, big.NewInt(2)), maxPriorityFeePerGas)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logger.Debugf("Failed to get chain ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chain ID"})
		return
	}

	tx := ethTypes.NewTx(&ethTypes.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: maxPriorityFeePerGas,
		GasFeeCap: maxFeePerGas,
		Gas:       gasLimit,
		To:        &address,
		Value:     thresholdWei,
		Data:      nil,
	})

	signedTx, err := ethTypes.SignTx(tx, ethTypes.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		logger.Debugf("Failed to sign transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign transaction"})
		return
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		logger.Errorf("Failed to send transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send transaction"})
		return
	}

	logger.Infof("Fund (%s) sent successfully to %s", thresholdWei.String(), req.WalletAddress)
	c.JSON(http.StatusOK, ClaimFundResponse{
		Success:         true,
		Message:         "Funds sent successfully",
		TransactionHash: signedTx.Hash().Hex(),
	})
}
