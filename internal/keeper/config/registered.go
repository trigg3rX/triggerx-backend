package config

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const AttestationCenterABI = `[{
	"inputs": [
	  {
		"internalType": "address",
		"name": "_operator",
		"type": "address"
	  }
	],
	"name": "operatorsIdsByAddress",
	"outputs": [
	  {
		"internalType": "uint256",
		"name": "",
		"type": "uint256"
	  }
	],
	"stateMutability": "view",
	"type": "function"
}]`

func checkKeeperRegistration() error {
	client, err := ethclient.Dial(GetBaseRPCUrl())
	if err != nil {
		return fmt.Errorf("failed to connect to L2 network: %w", err)
	}
	defer client.Close()

	parsedABI, err := abi.JSON(strings.NewReader(AttestationCenterABI))
	if err != nil {
		return fmt.Errorf("failed to parse AttestationCenter ABI: %w", err)
	}

	keeperAddr := common.HexToAddress(GetKeeperAddress())
	data, err := parsedABI.Pack("operatorsIdsByAddress", keeperAddr)
	if err != nil {
		return fmt.Errorf("failed to pack function call data: %w", err)
	}

	attestationCenterAddr := common.HexToAddress(GetAttestationCenterAddress())
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &attestationCenterAddr,
		Data: data,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to call AttestationCenter contract: %w", err)
	}

	if len(result) == 0 {
		return fmt.Errorf("empty result from contract call")
	}

	operatorID := new(big.Int).SetBytes(result)

	if operatorID.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("keeper address %s is not registered on L2. Please register the address before continuing. If registered, please wait for the registration to be confirmed", GetKeeperAddress())
	}

	fmt.Printf("Keeper address %s is registered on L2 with operator ID %s\n", GetKeeperAddress(), operatorID.String())

	return nil
}
