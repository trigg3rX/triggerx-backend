package events

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// ABI definitions
var (
	AvsGovernanceABI     abi.ABI
	AttestationCenterABI abi.ABI
)

// Event type definitions
type TaskSubmittedEvent struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	AttestersIds     []*big.Int
	Raw              ethtypes.Log
}

type TaskRejectedEvent struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	AttestersIds     []*big.Int
	Raw              ethtypes.Log
}

type OperatorRegisteredEvent struct {
	Operator common.Address
	BlsKey   [4]*big.Int
	Raw      ethtypes.Log
}

type OperatorUnregisteredEvent struct {
	Operator common.Address
	Raw      ethtypes.Log
}
