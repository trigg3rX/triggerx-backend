package config

import (
	// "context"
	// "github.com/ethereum/go-ethereum"
	// "github.com/ethereum/go-ethereum/accounts/abi"
	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/ethclient"
	// "log"
	// "math/big"
	// "strings"
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

func checkKeeperRegistration() bool {
	// client, err := ethclient.Dial(GetBaseRPCUrl())
	// if err != nil {
	// 	log.Fatal("Failed to connect to L2 network", "error", err)
	// 	return false
	// }
	// defer client.Close()

	// parsedABI, err := abi.JSON(strings.NewReader(AttestationCenterABI))
	// if err != nil {
	// 	log.Fatal("Failed to parse AttestationCenter ABI", "error", err)
	// 	return false
	// }

	// keeperAddr := common.HexToAddress(GetKeeperAddress())
	// data, err := parsedABI.Pack("operatorsIdsByAddress", keeperAddr)
	// if err != nil {
	// 	log.Fatal("Failed to pack function call data", "error", err)
	// 	return false
	// }

	// attestationCenterAddr := common.HexToAddress(GetAttestationCenterAddress())
	// result, err := client.CallContract(context.Background(), ethereum.CallMsg{
	// 	To:   &attestationCenterAddr,
	// 	Data: data,
	// }, nil)
	// if err != nil {
	// 	log.Fatal("Failed to call AttestationCenter contract", "error", err)
	// 	return false
	// }

	// if len(result) == 0 {
	// 	log.Fatal("Empty result from contract call")
	// 	return false
	// }

	// operatorID := new(big.Int).SetBytes(result)

	// if operatorID.Cmp(big.NewInt(0)) == 0 {
	// 	log.Fatal("Keeper address is not registered on L2")
	// 	return false
	// }

	// log.Println("Keeper address", GetKeeperAddress(), "is registered on L2 with operator ID", operatorID)

	return true
}
