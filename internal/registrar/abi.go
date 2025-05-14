package registrar

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	AvsGovernanceABI     abi.ABI
	AttestationCenterABI abi.ABI
)

type OperatorRegisteredEvent struct {
	Operator common.Address
	BlsKey   [4]*big.Int
	Raw      types.Log
}

type OperatorUnregisteredEvent struct {
	Operator common.Address
	Raw      types.Log
}

type TaskSubmittedEvent struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	AttestersIds     []*big.Int
	Raw              types.Log
}

type TaskRejectedEvent struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	AttestersIds     []*big.Int
	Raw              types.Log
}

func InitABI() error {
	avsGovernanceABIJSON, err := os.ReadFile("pkg/bindings/abi/AvsGovernance.json")
	if err != nil {
		logger.Fatalf("Failed to read AvsGovernance ABI: %v", err)
		return fmt.Errorf("failed to read AvsGovernance ABI: %v", err)
	}
	AvsGovernanceABI, err = abi.JSON(strings.NewReader(string(avsGovernanceABIJSON)))
	if err != nil {
		return fmt.Errorf("failed to parse AvsGovernance ABI: %v", err)
	}

	attestationCenterABIJSON, err := os.ReadFile("pkg/bindings/abi/AttestationCenter.json")
	if err != nil {
		logger.Fatalf("Failed to read AttestationCenter ABI: %v", err)
		return fmt.Errorf("failed to read AttestationCenter ABI: %v", err)
	}
	AttestationCenterABI, err = abi.JSON(strings.NewReader(string(attestationCenterABIJSON)))
	if err != nil {
		return fmt.Errorf("failed to parse AttestationCenter ABI: %v", err)
	}

	return nil
}
