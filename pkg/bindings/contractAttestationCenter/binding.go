// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractAttestationCenter

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// BLSAuthLibrarySignature is an auto generated low-level Go binding around an user-defined struct.
type BLSAuthLibrarySignature struct {
	Signature [2]*big.Int
}

// IAttestationCenterPaymentDetails is an auto generated low-level Go binding around an user-defined struct.
type IAttestationCenterPaymentDetails struct {
	Operator           common.Address
	LastPaidTaskNumber *big.Int
	FeeToClaim         *big.Int
	PaymentStatus      uint8
}

// IAttestationCenterPaymentRequestMessage is an auto generated low-level Go binding around an user-defined struct.
type IAttestationCenterPaymentRequestMessage struct {
	Operator   common.Address
	FeeToClaim *big.Int
}

// IAttestationCenterTaskInfo is an auto generated low-level Go binding around an user-defined struct.
type IAttestationCenterTaskInfo struct {
	ProofOfTask      string
	Data             []byte
	TaskPerformer    common.Address
	TaskDefinitionId uint16
}

// IAttestationCenterTaskSubmissionDetails is an auto generated low-level Go binding around an user-defined struct.
type IAttestationCenterTaskSubmissionDetails struct {
	IsApproved   bool
	TpSignature  []byte
	TaSignature  [2]*big.Int
	AttestersIds []*big.Int
}

// TaskDefinitionParams is an auto generated low-level Go binding around an user-defined struct.
type TaskDefinitionParams struct {
	BlockExpiry                *big.Int
	BaseRewardFeeForAttesters  *big.Int
	BaseRewardFeeForPerformer  *big.Int
	BaseRewardFeeForAggregator *big.Int
	DisputePeriodBlocks        *big.Int
	MinimumVotingPower         *big.Int
	RestrictedOperatorIndexes  []*big.Int
}

// AttestationCenterMetaData contains all meta data concerning the AttestationCenter contract.
var AttestationCenterMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"avsLogic\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIAvsLogic\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"baseRewardFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"beforePaymentsLogic\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIBeforePaymentsLogic\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"clearBatchPayment\",\"inputs\":[{\"name\":\"_operators\",\"type\":\"tuple[]\",\"internalType\":\"structIAttestationCenter.PaymentRequestMessage[]\",\"components\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToClaim\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"_paidTaskNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"clearPayment\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_lastPaidTaskNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_amountClaimed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createNewTaskDefinition\",\"inputs\":[{\"name\":\"_name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_taskDefinitionParams\",\"type\":\"tuple\",\"internalType\":\"structTaskDefinitionParams\",\"components\":[{\"name\":\"blockExpiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"baseRewardFeeForAttesters\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"baseRewardFeeForPerformer\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"baseRewardFeeForAggregator\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"disputePeriodBlocks\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minimumVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"restrictedOperatorIndexes\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}]}],\"outputs\":[{\"name\":\"_id\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getOperatorPaymentDetail\",\"inputs\":[{\"name\":\"_operatorId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIAttestationCenter.PaymentDetails\",\"components\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"lastPaidTaskNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeToClaim\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"paymentStatus\",\"type\":\"uint8\",\"internalType\":\"enumIAttestationCenter.PaymentStatus\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskDefinitionMinimumVotingPower\",\"inputs\":[{\"name\":\"_taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTaskDefinitionRestrictedOperators\",\"inputs\":[{\"name\":\"_taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_avsGovernanceMultisigOwner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_operationsMultisig\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_communityMultisig\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_messageHandler\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_obls\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_isRewardsOnL2\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isFlowPaused\",\"inputs\":[{\"name\":\"_pausableFlow\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"_isPaused\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numOfActiveOperators\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numOfOperators\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numOfTaskDefinitions\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numOfTotalOperators\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"obls\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIOBLS\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorsIdsByAddress\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[{\"name\":\"_pausableFlow\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerToNetwork\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_blsKey\",\"type\":\"uint256[4]\",\"internalType\":\"uint256[4]\"},{\"name\":\"_rewardsReceiver\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"requestBatchPayment\",\"inputs\":[{\"name\":\"_from\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_to\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"requestBatchPayment\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"requestPayment\",\"inputs\":[{\"name\":\"_operatorId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAvsLogic\",\"inputs\":[{\"name\":\"_avsLogic\",\"type\":\"address\",\"internalType\":\"contractIAvsLogic\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setBeforePaymentsLogic\",\"inputs\":[{\"name\":\"_beforePaymentsLogic\",\"type\":\"address\",\"internalType\":\"contractIBeforePaymentsLogic\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setFeeCalculator\",\"inputs\":[{\"name\":\"_feeCalculator\",\"type\":\"address\",\"internalType\":\"contractIFeeCalculator\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOblsSharesSyncer\",\"inputs\":[{\"name\":\"_oblsSharesSyncer\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTaskDefinitionMinVotingPower\",\"inputs\":[{\"name\":\"_taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"_minimumVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTaskDefinitionRestrictedOperators\",\"inputs\":[{\"name\":\"_taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"_restrictedOperatorIndexes\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitTask\",\"inputs\":[{\"name\":\"_taskInfo\",\"type\":\"tuple\",\"internalType\":\"structIAttestationCenter.TaskInfo\",\"components\":[{\"name\":\"proofOfTask\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"taskPerformer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"name\":\"_isApproved\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"_tpSignature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_taSignature\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"_attestersIds\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitTask\",\"inputs\":[{\"name\":\"_taskInfo\",\"type\":\"tuple\",\"internalType\":\"structIAttestationCenter.TaskInfo\",\"components\":[{\"name\":\"proofOfTask\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"taskPerformer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"name\":\"_taskSubmissionDetails\",\"type\":\"tuple\",\"internalType\":\"structIAttestationCenter.TaskSubmissionDetails\",\"components\":[{\"name\":\"isApproved\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"tpSignature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"taSignature\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"attestersIds\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"taskNumber\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferAvsGovernanceMultisig\",\"inputs\":[{\"name\":\"_newAvsGovernanceMultisig\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferMessageHandler\",\"inputs\":[{\"name\":\"_newMessageHandler\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unRegisterOperatorFromNetwork\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[{\"name\":\"_pausableFlow\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateBlsKey\",\"inputs\":[{\"name\":\"_blsKey\",\"type\":\"uint256[4]\",\"internalType\":\"uint256[4]\"},{\"name\":\"_authSignature\",\"type\":\"tuple\",\"internalType\":\"structBLSAuthLibrary.Signature\",\"components\":[{\"name\":\"signature\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"vault\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"votingPower\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"ClearPaymentRejected\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"requestedTaskNumber\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"requestedAmountClaimed\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FlowPaused\",\"inputs\":[{\"name\":\"_pausableFlow\",\"type\":\"bytes4\",\"indexed\":false,\"internalType\":\"bytes4\"},{\"name\":\"_pauser\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FlowUnpaused\",\"inputs\":[{\"name\":\"_pausableFlowFlag\",\"type\":\"bytes4\",\"indexed\":false,\"internalType\":\"bytes4\"},{\"name\":\"_unpauser\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorBlsKeyUpdated\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"blsKey\",\"type\":\"uint256[4]\",\"indexed\":false,\"internalType\":\"uint256[4]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorRegisteredToNetwork\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"votingPower\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorUnregisteredFromNetwork\",\"inputs\":[{\"name\":\"operatorId\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PaymentRequested\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"lastPaidTaskNumber\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"feeToClaim\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PaymentsRequested\",\"inputs\":[{\"name\":\"operators\",\"type\":\"tuple[]\",\"indexed\":false,\"internalType\":\"structIAttestationCenter.PaymentRequestMessage[]\",\"components\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToClaim\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"lastPaidTaskNumber\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RewardAccumulated\",\"inputs\":[{\"name\":\"_operatorId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"_baseRewardFeeForOperator\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"_taskNumber\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetAvsGovernanceMultisig\",\"inputs\":[{\"name\":\"newAvsGovernanceMultisig\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetAvsLogic\",\"inputs\":[{\"name\":\"avsLogic\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetBeforePaymentsLogic\",\"inputs\":[{\"name\":\"paymentsLogic\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetFeeCalculator\",\"inputs\":[{\"name\":\"feeCalculator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetMessageHandler\",\"inputs\":[{\"name\":\"newMessageHandler\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetMinimumTaskDefinitionVotingPower\",\"inputs\":[{\"name\":\"minimumVotingPower\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetRestrictedOperator\",\"inputs\":[{\"name\":\"taskDefinitionId\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"restrictedOperatorIndexes\",\"type\":\"uint256[]\",\"indexed\":false,\"internalType\":\"uint256[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskDefinitionCreated\",\"inputs\":[{\"name\":\"taskDefinitionId\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"blockExpiry\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"baseRewardFeeForAttesters\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"baseRewardFeeForPerformer\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"baseRewardFeeForAggregator\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"disputePeriodBlocks\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"minimumVotingPower\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"restrictedOperatorIndexes\",\"type\":\"uint256[]\",\"indexed\":false,\"internalType\":\"uint256[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskDefinitionRestrictedOperatorsModified\",\"inputs\":[{\"name\":\"taskDefinitionId\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"restrictedOperatorIndexes\",\"type\":\"uint256[]\",\"indexed\":false,\"internalType\":\"uint256[]\"},{\"name\":\"isRestricted\",\"type\":\"bool[]\",\"indexed\":false,\"internalType\":\"bool[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskRejected\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"taskNumber\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"proofOfTask\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"data\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"},{\"name\":\"taskDefinitionId\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TaskSubmitted\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"taskNumber\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"proofOfTask\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"data\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"},{\"name\":\"taskDefinitionId\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"FlowIsCurrentlyPaused\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FlowIsCurrentlyUnpaused\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InactiveAggregator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InactiveTaskPerformer\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidArrayLength\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAttesterSet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBlockExpiry\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBlsKeyUpdateSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOperatorId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOperatorsForPayment\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPaymentClaim\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPerformerSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidRangeForBatchPaymentRequest\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidRestrictedOperator\",\"inputs\":[{\"name\":\"taskDefinitionId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"operatorIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidRestrictedOperatorIndexes\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidTaskDefinition\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MessageAlreadySigned\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorNotRegistered\",\"inputs\":[{\"name\":\"_operatorAddress\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"PauseFlowIsAlreadyPaused\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PaymentClaimed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PaymentReedemed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TaskDefinitionNotFound\",\"inputs\":[{\"name\":\"taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"type\":\"error\",\"name\":\"UnpausingFlowIsAlreadyUnpaused\",\"inputs\":[]}]",
}

// AttestationCenterABI is the input ABI used to generate the binding from.
// Deprecated: Use AttestationCenterMetaData.ABI instead.
var AttestationCenterABI = AttestationCenterMetaData.ABI

// AttestationCenter is an auto generated Go binding around an Ethereum contract.
type AttestationCenter struct {
	AttestationCenterCaller     // Read-only binding to the contract
	AttestationCenterTransactor // Write-only binding to the contract
	AttestationCenterFilterer   // Log filterer for contract events
}

// AttestationCenterCaller is an auto generated read-only Go binding around an Ethereum contract.
type AttestationCenterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationCenterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AttestationCenterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationCenterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AttestationCenterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AttestationCenterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AttestationCenterSession struct {
	Contract     *AttestationCenter // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// AttestationCenterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AttestationCenterCallerSession struct {
	Contract *AttestationCenterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// AttestationCenterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AttestationCenterTransactorSession struct {
	Contract     *AttestationCenterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// AttestationCenterRaw is an auto generated low-level Go binding around an Ethereum contract.
type AttestationCenterRaw struct {
	Contract *AttestationCenter // Generic contract binding to access the raw methods on
}

// AttestationCenterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AttestationCenterCallerRaw struct {
	Contract *AttestationCenterCaller // Generic read-only contract binding to access the raw methods on
}

// AttestationCenterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AttestationCenterTransactorRaw struct {
	Contract *AttestationCenterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAttestationCenter creates a new instance of AttestationCenter, bound to a specific deployed contract.
func NewAttestationCenter(address common.Address, backend bind.ContractBackend) (*AttestationCenter, error) {
	contract, err := bindAttestationCenter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AttestationCenter{AttestationCenterCaller: AttestationCenterCaller{contract: contract}, AttestationCenterTransactor: AttestationCenterTransactor{contract: contract}, AttestationCenterFilterer: AttestationCenterFilterer{contract: contract}}, nil
}

// NewAttestationCenterCaller creates a new read-only instance of AttestationCenter, bound to a specific deployed contract.
func NewAttestationCenterCaller(address common.Address, caller bind.ContractCaller) (*AttestationCenterCaller, error) {
	contract, err := bindAttestationCenter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AttestationCenterCaller{contract: contract}, nil
}

// NewAttestationCenterTransactor creates a new write-only instance of AttestationCenter, bound to a specific deployed contract.
func NewAttestationCenterTransactor(address common.Address, transactor bind.ContractTransactor) (*AttestationCenterTransactor, error) {
	contract, err := bindAttestationCenter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AttestationCenterTransactor{contract: contract}, nil
}

// NewAttestationCenterFilterer creates a new log filterer instance of AttestationCenter, bound to a specific deployed contract.
func NewAttestationCenterFilterer(address common.Address, filterer bind.ContractFilterer) (*AttestationCenterFilterer, error) {
	contract, err := bindAttestationCenter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AttestationCenterFilterer{contract: contract}, nil
}

// bindAttestationCenter binds a generic wrapper to an already deployed contract.
func bindAttestationCenter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AttestationCenterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AttestationCenter *AttestationCenterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AttestationCenter.Contract.AttestationCenterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AttestationCenter *AttestationCenterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationCenter.Contract.AttestationCenterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AttestationCenter *AttestationCenterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AttestationCenter.Contract.AttestationCenterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AttestationCenter *AttestationCenterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AttestationCenter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AttestationCenter *AttestationCenterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationCenter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AttestationCenter *AttestationCenterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AttestationCenter.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AttestationCenter *AttestationCenterCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AttestationCenter *AttestationCenterSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AttestationCenter.Contract.DEFAULTADMINROLE(&_AttestationCenter.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AttestationCenter *AttestationCenterCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AttestationCenter.Contract.DEFAULTADMINROLE(&_AttestationCenter.CallOpts)
}

// AvsLogic is a free data retrieval call binding the contract method 0xb0817c44.
//
// Solidity: function avsLogic() view returns(address)
func (_AttestationCenter *AttestationCenterCaller) AvsLogic(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "avsLogic")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AvsLogic is a free data retrieval call binding the contract method 0xb0817c44.
//
// Solidity: function avsLogic() view returns(address)
func (_AttestationCenter *AttestationCenterSession) AvsLogic() (common.Address, error) {
	return _AttestationCenter.Contract.AvsLogic(&_AttestationCenter.CallOpts)
}

// AvsLogic is a free data retrieval call binding the contract method 0xb0817c44.
//
// Solidity: function avsLogic() view returns(address)
func (_AttestationCenter *AttestationCenterCallerSession) AvsLogic() (common.Address, error) {
	return _AttestationCenter.Contract.AvsLogic(&_AttestationCenter.CallOpts)
}

// BaseRewardFee is a free data retrieval call binding the contract method 0x3428c126.
//
// Solidity: function baseRewardFee() view returns(uint256)
func (_AttestationCenter *AttestationCenterCaller) BaseRewardFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "baseRewardFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BaseRewardFee is a free data retrieval call binding the contract method 0x3428c126.
//
// Solidity: function baseRewardFee() view returns(uint256)
func (_AttestationCenter *AttestationCenterSession) BaseRewardFee() (*big.Int, error) {
	return _AttestationCenter.Contract.BaseRewardFee(&_AttestationCenter.CallOpts)
}

// BaseRewardFee is a free data retrieval call binding the contract method 0x3428c126.
//
// Solidity: function baseRewardFee() view returns(uint256)
func (_AttestationCenter *AttestationCenterCallerSession) BaseRewardFee() (*big.Int, error) {
	return _AttestationCenter.Contract.BaseRewardFee(&_AttestationCenter.CallOpts)
}

// BeforePaymentsLogic is a free data retrieval call binding the contract method 0xc2f429f1.
//
// Solidity: function beforePaymentsLogic() view returns(address)
func (_AttestationCenter *AttestationCenterCaller) BeforePaymentsLogic(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "beforePaymentsLogic")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BeforePaymentsLogic is a free data retrieval call binding the contract method 0xc2f429f1.
//
// Solidity: function beforePaymentsLogic() view returns(address)
func (_AttestationCenter *AttestationCenterSession) BeforePaymentsLogic() (common.Address, error) {
	return _AttestationCenter.Contract.BeforePaymentsLogic(&_AttestationCenter.CallOpts)
}

// BeforePaymentsLogic is a free data retrieval call binding the contract method 0xc2f429f1.
//
// Solidity: function beforePaymentsLogic() view returns(address)
func (_AttestationCenter *AttestationCenterCallerSession) BeforePaymentsLogic() (common.Address, error) {
	return _AttestationCenter.Contract.BeforePaymentsLogic(&_AttestationCenter.CallOpts)
}

// GetOperatorPaymentDetail is a free data retrieval call binding the contract method 0x9eb72d4c.
//
// Solidity: function getOperatorPaymentDetail(uint256 _operatorId) view returns((address,uint256,uint256,uint8))
func (_AttestationCenter *AttestationCenterCaller) GetOperatorPaymentDetail(opts *bind.CallOpts, _operatorId *big.Int) (IAttestationCenterPaymentDetails, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "getOperatorPaymentDetail", _operatorId)

	if err != nil {
		return *new(IAttestationCenterPaymentDetails), err
	}

	out0 := *abi.ConvertType(out[0], new(IAttestationCenterPaymentDetails)).(*IAttestationCenterPaymentDetails)

	return out0, err

}

// GetOperatorPaymentDetail is a free data retrieval call binding the contract method 0x9eb72d4c.
//
// Solidity: function getOperatorPaymentDetail(uint256 _operatorId) view returns((address,uint256,uint256,uint8))
func (_AttestationCenter *AttestationCenterSession) GetOperatorPaymentDetail(_operatorId *big.Int) (IAttestationCenterPaymentDetails, error) {
	return _AttestationCenter.Contract.GetOperatorPaymentDetail(&_AttestationCenter.CallOpts, _operatorId)
}

// GetOperatorPaymentDetail is a free data retrieval call binding the contract method 0x9eb72d4c.
//
// Solidity: function getOperatorPaymentDetail(uint256 _operatorId) view returns((address,uint256,uint256,uint8))
func (_AttestationCenter *AttestationCenterCallerSession) GetOperatorPaymentDetail(_operatorId *big.Int) (IAttestationCenterPaymentDetails, error) {
	return _AttestationCenter.Contract.GetOperatorPaymentDetail(&_AttestationCenter.CallOpts, _operatorId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AttestationCenter *AttestationCenterCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AttestationCenter *AttestationCenterSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AttestationCenter.Contract.GetRoleAdmin(&_AttestationCenter.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AttestationCenter *AttestationCenterCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AttestationCenter.Contract.GetRoleAdmin(&_AttestationCenter.CallOpts, role)
}

// GetTaskDefinitionMinimumVotingPower is a free data retrieval call binding the contract method 0x75d9aedf.
//
// Solidity: function getTaskDefinitionMinimumVotingPower(uint16 _taskDefinitionId) view returns(uint256)
func (_AttestationCenter *AttestationCenterCaller) GetTaskDefinitionMinimumVotingPower(opts *bind.CallOpts, _taskDefinitionId uint16) (*big.Int, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "getTaskDefinitionMinimumVotingPower", _taskDefinitionId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTaskDefinitionMinimumVotingPower is a free data retrieval call binding the contract method 0x75d9aedf.
//
// Solidity: function getTaskDefinitionMinimumVotingPower(uint16 _taskDefinitionId) view returns(uint256)
func (_AttestationCenter *AttestationCenterSession) GetTaskDefinitionMinimumVotingPower(_taskDefinitionId uint16) (*big.Int, error) {
	return _AttestationCenter.Contract.GetTaskDefinitionMinimumVotingPower(&_AttestationCenter.CallOpts, _taskDefinitionId)
}

// GetTaskDefinitionMinimumVotingPower is a free data retrieval call binding the contract method 0x75d9aedf.
//
// Solidity: function getTaskDefinitionMinimumVotingPower(uint16 _taskDefinitionId) view returns(uint256)
func (_AttestationCenter *AttestationCenterCallerSession) GetTaskDefinitionMinimumVotingPower(_taskDefinitionId uint16) (*big.Int, error) {
	return _AttestationCenter.Contract.GetTaskDefinitionMinimumVotingPower(&_AttestationCenter.CallOpts, _taskDefinitionId)
}

// GetTaskDefinitionRestrictedOperators is a free data retrieval call binding the contract method 0x4e2ce53f.
//
// Solidity: function getTaskDefinitionRestrictedOperators(uint16 _taskDefinitionId) view returns(uint256[])
func (_AttestationCenter *AttestationCenterCaller) GetTaskDefinitionRestrictedOperators(opts *bind.CallOpts, _taskDefinitionId uint16) ([]*big.Int, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "getTaskDefinitionRestrictedOperators", _taskDefinitionId)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetTaskDefinitionRestrictedOperators is a free data retrieval call binding the contract method 0x4e2ce53f.
//
// Solidity: function getTaskDefinitionRestrictedOperators(uint16 _taskDefinitionId) view returns(uint256[])
func (_AttestationCenter *AttestationCenterSession) GetTaskDefinitionRestrictedOperators(_taskDefinitionId uint16) ([]*big.Int, error) {
	return _AttestationCenter.Contract.GetTaskDefinitionRestrictedOperators(&_AttestationCenter.CallOpts, _taskDefinitionId)
}

// GetTaskDefinitionRestrictedOperators is a free data retrieval call binding the contract method 0x4e2ce53f.
//
// Solidity: function getTaskDefinitionRestrictedOperators(uint16 _taskDefinitionId) view returns(uint256[])
func (_AttestationCenter *AttestationCenterCallerSession) GetTaskDefinitionRestrictedOperators(_taskDefinitionId uint16) ([]*big.Int, error) {
	return _AttestationCenter.Contract.GetTaskDefinitionRestrictedOperators(&_AttestationCenter.CallOpts, _taskDefinitionId)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AttestationCenter *AttestationCenterCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AttestationCenter *AttestationCenterSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AttestationCenter.Contract.HasRole(&_AttestationCenter.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AttestationCenter *AttestationCenterCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AttestationCenter.Contract.HasRole(&_AttestationCenter.CallOpts, role, account)
}

// IsFlowPaused is a free data retrieval call binding the contract method 0xefd96978.
//
// Solidity: function isFlowPaused(bytes4 _pausableFlow) view returns(bool _isPaused)
func (_AttestationCenter *AttestationCenterCaller) IsFlowPaused(opts *bind.CallOpts, _pausableFlow [4]byte) (bool, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "isFlowPaused", _pausableFlow)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsFlowPaused is a free data retrieval call binding the contract method 0xefd96978.
//
// Solidity: function isFlowPaused(bytes4 _pausableFlow) view returns(bool _isPaused)
func (_AttestationCenter *AttestationCenterSession) IsFlowPaused(_pausableFlow [4]byte) (bool, error) {
	return _AttestationCenter.Contract.IsFlowPaused(&_AttestationCenter.CallOpts, _pausableFlow)
}

// IsFlowPaused is a free data retrieval call binding the contract method 0xefd96978.
//
// Solidity: function isFlowPaused(bytes4 _pausableFlow) view returns(bool _isPaused)
func (_AttestationCenter *AttestationCenterCallerSession) IsFlowPaused(_pausableFlow [4]byte) (bool, error) {
	return _AttestationCenter.Contract.IsFlowPaused(&_AttestationCenter.CallOpts, _pausableFlow)
}

// NumOfActiveOperators is a free data retrieval call binding the contract method 0x7897dec3.
//
// Solidity: function numOfActiveOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterCaller) NumOfActiveOperators(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "numOfActiveOperators")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumOfActiveOperators is a free data retrieval call binding the contract method 0x7897dec3.
//
// Solidity: function numOfActiveOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterSession) NumOfActiveOperators() (*big.Int, error) {
	return _AttestationCenter.Contract.NumOfActiveOperators(&_AttestationCenter.CallOpts)
}

// NumOfActiveOperators is a free data retrieval call binding the contract method 0x7897dec3.
//
// Solidity: function numOfActiveOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterCallerSession) NumOfActiveOperators() (*big.Int, error) {
	return _AttestationCenter.Contract.NumOfActiveOperators(&_AttestationCenter.CallOpts)
}

// NumOfOperators is a free data retrieval call binding the contract method 0x6ade02da.
//
// Solidity: function numOfOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterCaller) NumOfOperators(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "numOfOperators")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumOfOperators is a free data retrieval call binding the contract method 0x6ade02da.
//
// Solidity: function numOfOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterSession) NumOfOperators() (*big.Int, error) {
	return _AttestationCenter.Contract.NumOfOperators(&_AttestationCenter.CallOpts)
}

// NumOfOperators is a free data retrieval call binding the contract method 0x6ade02da.
//
// Solidity: function numOfOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterCallerSession) NumOfOperators() (*big.Int, error) {
	return _AttestationCenter.Contract.NumOfOperators(&_AttestationCenter.CallOpts)
}

// NumOfTaskDefinitions is a free data retrieval call binding the contract method 0x34a7c391.
//
// Solidity: function numOfTaskDefinitions() view returns(uint16)
func (_AttestationCenter *AttestationCenterCaller) NumOfTaskDefinitions(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "numOfTaskDefinitions")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// NumOfTaskDefinitions is a free data retrieval call binding the contract method 0x34a7c391.
//
// Solidity: function numOfTaskDefinitions() view returns(uint16)
func (_AttestationCenter *AttestationCenterSession) NumOfTaskDefinitions() (uint16, error) {
	return _AttestationCenter.Contract.NumOfTaskDefinitions(&_AttestationCenter.CallOpts)
}

// NumOfTaskDefinitions is a free data retrieval call binding the contract method 0x34a7c391.
//
// Solidity: function numOfTaskDefinitions() view returns(uint16)
func (_AttestationCenter *AttestationCenterCallerSession) NumOfTaskDefinitions() (uint16, error) {
	return _AttestationCenter.Contract.NumOfTaskDefinitions(&_AttestationCenter.CallOpts)
}

// NumOfTotalOperators is a free data retrieval call binding the contract method 0x00028b07.
//
// Solidity: function numOfTotalOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterCaller) NumOfTotalOperators(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "numOfTotalOperators")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumOfTotalOperators is a free data retrieval call binding the contract method 0x00028b07.
//
// Solidity: function numOfTotalOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterSession) NumOfTotalOperators() (*big.Int, error) {
	return _AttestationCenter.Contract.NumOfTotalOperators(&_AttestationCenter.CallOpts)
}

// NumOfTotalOperators is a free data retrieval call binding the contract method 0x00028b07.
//
// Solidity: function numOfTotalOperators() view returns(uint256)
func (_AttestationCenter *AttestationCenterCallerSession) NumOfTotalOperators() (*big.Int, error) {
	return _AttestationCenter.Contract.NumOfTotalOperators(&_AttestationCenter.CallOpts)
}

// Obls is a free data retrieval call binding the contract method 0x659fa976.
//
// Solidity: function obls() view returns(address)
func (_AttestationCenter *AttestationCenterCaller) Obls(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "obls")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Obls is a free data retrieval call binding the contract method 0x659fa976.
//
// Solidity: function obls() view returns(address)
func (_AttestationCenter *AttestationCenterSession) Obls() (common.Address, error) {
	return _AttestationCenter.Contract.Obls(&_AttestationCenter.CallOpts)
}

// Obls is a free data retrieval call binding the contract method 0x659fa976.
//
// Solidity: function obls() view returns(address)
func (_AttestationCenter *AttestationCenterCallerSession) Obls() (common.Address, error) {
	return _AttestationCenter.Contract.Obls(&_AttestationCenter.CallOpts)
}

// OperatorsIdsByAddress is a free data retrieval call binding the contract method 0x5b15c568.
//
// Solidity: function operatorsIdsByAddress(address _operator) view returns(uint256)
func (_AttestationCenter *AttestationCenterCaller) OperatorsIdsByAddress(opts *bind.CallOpts, _operator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "operatorsIdsByAddress", _operator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OperatorsIdsByAddress is a free data retrieval call binding the contract method 0x5b15c568.
//
// Solidity: function operatorsIdsByAddress(address _operator) view returns(uint256)
func (_AttestationCenter *AttestationCenterSession) OperatorsIdsByAddress(_operator common.Address) (*big.Int, error) {
	return _AttestationCenter.Contract.OperatorsIdsByAddress(&_AttestationCenter.CallOpts, _operator)
}

// OperatorsIdsByAddress is a free data retrieval call binding the contract method 0x5b15c568.
//
// Solidity: function operatorsIdsByAddress(address _operator) view returns(uint256)
func (_AttestationCenter *AttestationCenterCallerSession) OperatorsIdsByAddress(_operator common.Address) (*big.Int, error) {
	return _AttestationCenter.Contract.OperatorsIdsByAddress(&_AttestationCenter.CallOpts, _operator)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AttestationCenter *AttestationCenterCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AttestationCenter *AttestationCenterSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AttestationCenter.Contract.SupportsInterface(&_AttestationCenter.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AttestationCenter *AttestationCenterCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AttestationCenter.Contract.SupportsInterface(&_AttestationCenter.CallOpts, interfaceId)
}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_AttestationCenter *AttestationCenterCaller) TaskNumber(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "taskNumber")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_AttestationCenter *AttestationCenterSession) TaskNumber() (uint32, error) {
	return _AttestationCenter.Contract.TaskNumber(&_AttestationCenter.CallOpts)
}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_AttestationCenter *AttestationCenterCallerSession) TaskNumber() (uint32, error) {
	return _AttestationCenter.Contract.TaskNumber(&_AttestationCenter.CallOpts)
}

// Vault is a free data retrieval call binding the contract method 0xfbfa77cf.
//
// Solidity: function vault() view returns(address)
func (_AttestationCenter *AttestationCenterCaller) Vault(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "vault")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Vault is a free data retrieval call binding the contract method 0xfbfa77cf.
//
// Solidity: function vault() view returns(address)
func (_AttestationCenter *AttestationCenterSession) Vault() (common.Address, error) {
	return _AttestationCenter.Contract.Vault(&_AttestationCenter.CallOpts)
}

// Vault is a free data retrieval call binding the contract method 0xfbfa77cf.
//
// Solidity: function vault() view returns(address)
func (_AttestationCenter *AttestationCenterCallerSession) Vault() (common.Address, error) {
	return _AttestationCenter.Contract.Vault(&_AttestationCenter.CallOpts)
}

// VotingPower is a free data retrieval call binding the contract method 0xc07473f6.
//
// Solidity: function votingPower(address _operator) view returns(uint256)
func (_AttestationCenter *AttestationCenterCaller) VotingPower(opts *bind.CallOpts, _operator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AttestationCenter.contract.Call(opts, &out, "votingPower", _operator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VotingPower is a free data retrieval call binding the contract method 0xc07473f6.
//
// Solidity: function votingPower(address _operator) view returns(uint256)
func (_AttestationCenter *AttestationCenterSession) VotingPower(_operator common.Address) (*big.Int, error) {
	return _AttestationCenter.Contract.VotingPower(&_AttestationCenter.CallOpts, _operator)
}

// VotingPower is a free data retrieval call binding the contract method 0xc07473f6.
//
// Solidity: function votingPower(address _operator) view returns(uint256)
func (_AttestationCenter *AttestationCenterCallerSession) VotingPower(_operator common.Address) (*big.Int, error) {
	return _AttestationCenter.Contract.VotingPower(&_AttestationCenter.CallOpts, _operator)
}

// ClearBatchPayment is a paid mutator transaction binding the contract method 0x915359fc.
//
// Solidity: function clearBatchPayment((address,uint256)[] _operators, uint256 _paidTaskNumber) returns()
func (_AttestationCenter *AttestationCenterTransactor) ClearBatchPayment(opts *bind.TransactOpts, _operators []IAttestationCenterPaymentRequestMessage, _paidTaskNumber *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "clearBatchPayment", _operators, _paidTaskNumber)
}

// ClearBatchPayment is a paid mutator transaction binding the contract method 0x915359fc.
//
// Solidity: function clearBatchPayment((address,uint256)[] _operators, uint256 _paidTaskNumber) returns()
func (_AttestationCenter *AttestationCenterSession) ClearBatchPayment(_operators []IAttestationCenterPaymentRequestMessage, _paidTaskNumber *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.ClearBatchPayment(&_AttestationCenter.TransactOpts, _operators, _paidTaskNumber)
}

// ClearBatchPayment is a paid mutator transaction binding the contract method 0x915359fc.
//
// Solidity: function clearBatchPayment((address,uint256)[] _operators, uint256 _paidTaskNumber) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) ClearBatchPayment(_operators []IAttestationCenterPaymentRequestMessage, _paidTaskNumber *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.ClearBatchPayment(&_AttestationCenter.TransactOpts, _operators, _paidTaskNumber)
}

// ClearPayment is a paid mutator transaction binding the contract method 0x242a76a4.
//
// Solidity: function clearPayment(address _operator, uint256 _lastPaidTaskNumber, uint256 _amountClaimed) returns()
func (_AttestationCenter *AttestationCenterTransactor) ClearPayment(opts *bind.TransactOpts, _operator common.Address, _lastPaidTaskNumber *big.Int, _amountClaimed *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "clearPayment", _operator, _lastPaidTaskNumber, _amountClaimed)
}

// ClearPayment is a paid mutator transaction binding the contract method 0x242a76a4.
//
// Solidity: function clearPayment(address _operator, uint256 _lastPaidTaskNumber, uint256 _amountClaimed) returns()
func (_AttestationCenter *AttestationCenterSession) ClearPayment(_operator common.Address, _lastPaidTaskNumber *big.Int, _amountClaimed *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.ClearPayment(&_AttestationCenter.TransactOpts, _operator, _lastPaidTaskNumber, _amountClaimed)
}

// ClearPayment is a paid mutator transaction binding the contract method 0x242a76a4.
//
// Solidity: function clearPayment(address _operator, uint256 _lastPaidTaskNumber, uint256 _amountClaimed) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) ClearPayment(_operator common.Address, _lastPaidTaskNumber *big.Int, _amountClaimed *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.ClearPayment(&_AttestationCenter.TransactOpts, _operator, _lastPaidTaskNumber, _amountClaimed)
}

// CreateNewTaskDefinition is a paid mutator transaction binding the contract method 0x0c62bf0d.
//
// Solidity: function createNewTaskDefinition(string _name, (uint256,uint256,uint256,uint256,uint256,uint256,uint256[]) _taskDefinitionParams) returns(uint16 _id)
func (_AttestationCenter *AttestationCenterTransactor) CreateNewTaskDefinition(opts *bind.TransactOpts, _name string, _taskDefinitionParams TaskDefinitionParams) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "createNewTaskDefinition", _name, _taskDefinitionParams)
}

// CreateNewTaskDefinition is a paid mutator transaction binding the contract method 0x0c62bf0d.
//
// Solidity: function createNewTaskDefinition(string _name, (uint256,uint256,uint256,uint256,uint256,uint256,uint256[]) _taskDefinitionParams) returns(uint16 _id)
func (_AttestationCenter *AttestationCenterSession) CreateNewTaskDefinition(_name string, _taskDefinitionParams TaskDefinitionParams) (*types.Transaction, error) {
	return _AttestationCenter.Contract.CreateNewTaskDefinition(&_AttestationCenter.TransactOpts, _name, _taskDefinitionParams)
}

// CreateNewTaskDefinition is a paid mutator transaction binding the contract method 0x0c62bf0d.
//
// Solidity: function createNewTaskDefinition(string _name, (uint256,uint256,uint256,uint256,uint256,uint256,uint256[]) _taskDefinitionParams) returns(uint16 _id)
func (_AttestationCenter *AttestationCenterTransactorSession) CreateNewTaskDefinition(_name string, _taskDefinitionParams TaskDefinitionParams) (*types.Transaction, error) {
	return _AttestationCenter.Contract.CreateNewTaskDefinition(&_AttestationCenter.TransactOpts, _name, _taskDefinitionParams)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AttestationCenter *AttestationCenterTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AttestationCenter *AttestationCenterSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.GrantRole(&_AttestationCenter.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.GrantRole(&_AttestationCenter.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0xd9378a59.
//
// Solidity: function initialize(address _avsGovernanceMultisigOwner, address _operationsMultisig, address _communityMultisig, address _messageHandler, address _obls, address _vault, bool _isRewardsOnL2) returns()
func (_AttestationCenter *AttestationCenterTransactor) Initialize(opts *bind.TransactOpts, _avsGovernanceMultisigOwner common.Address, _operationsMultisig common.Address, _communityMultisig common.Address, _messageHandler common.Address, _obls common.Address, _vault common.Address, _isRewardsOnL2 bool) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "initialize", _avsGovernanceMultisigOwner, _operationsMultisig, _communityMultisig, _messageHandler, _obls, _vault, _isRewardsOnL2)
}

// Initialize is a paid mutator transaction binding the contract method 0xd9378a59.
//
// Solidity: function initialize(address _avsGovernanceMultisigOwner, address _operationsMultisig, address _communityMultisig, address _messageHandler, address _obls, address _vault, bool _isRewardsOnL2) returns()
func (_AttestationCenter *AttestationCenterSession) Initialize(_avsGovernanceMultisigOwner common.Address, _operationsMultisig common.Address, _communityMultisig common.Address, _messageHandler common.Address, _obls common.Address, _vault common.Address, _isRewardsOnL2 bool) (*types.Transaction, error) {
	return _AttestationCenter.Contract.Initialize(&_AttestationCenter.TransactOpts, _avsGovernanceMultisigOwner, _operationsMultisig, _communityMultisig, _messageHandler, _obls, _vault, _isRewardsOnL2)
}

// Initialize is a paid mutator transaction binding the contract method 0xd9378a59.
//
// Solidity: function initialize(address _avsGovernanceMultisigOwner, address _operationsMultisig, address _communityMultisig, address _messageHandler, address _obls, address _vault, bool _isRewardsOnL2) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) Initialize(_avsGovernanceMultisigOwner common.Address, _operationsMultisig common.Address, _communityMultisig common.Address, _messageHandler common.Address, _obls common.Address, _vault common.Address, _isRewardsOnL2 bool) (*types.Transaction, error) {
	return _AttestationCenter.Contract.Initialize(&_AttestationCenter.TransactOpts, _avsGovernanceMultisigOwner, _operationsMultisig, _communityMultisig, _messageHandler, _obls, _vault, _isRewardsOnL2)
}

// Pause is a paid mutator transaction binding the contract method 0x3aa83ec7.
//
// Solidity: function pause(bytes4 _pausableFlow) returns()
func (_AttestationCenter *AttestationCenterTransactor) Pause(opts *bind.TransactOpts, _pausableFlow [4]byte) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "pause", _pausableFlow)
}

// Pause is a paid mutator transaction binding the contract method 0x3aa83ec7.
//
// Solidity: function pause(bytes4 _pausableFlow) returns()
func (_AttestationCenter *AttestationCenterSession) Pause(_pausableFlow [4]byte) (*types.Transaction, error) {
	return _AttestationCenter.Contract.Pause(&_AttestationCenter.TransactOpts, _pausableFlow)
}

// Pause is a paid mutator transaction binding the contract method 0x3aa83ec7.
//
// Solidity: function pause(bytes4 _pausableFlow) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) Pause(_pausableFlow [4]byte) (*types.Transaction, error) {
	return _AttestationCenter.Contract.Pause(&_AttestationCenter.TransactOpts, _pausableFlow)
}

// RegisterToNetwork is a paid mutator transaction binding the contract method 0xfcd4e66a.
//
// Solidity: function registerToNetwork(address _operator, uint256 _votingPower, uint256[4] _blsKey, address _rewardsReceiver) returns()
func (_AttestationCenter *AttestationCenterTransactor) RegisterToNetwork(opts *bind.TransactOpts, _operator common.Address, _votingPower *big.Int, _blsKey [4]*big.Int, _rewardsReceiver common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "registerToNetwork", _operator, _votingPower, _blsKey, _rewardsReceiver)
}

// RegisterToNetwork is a paid mutator transaction binding the contract method 0xfcd4e66a.
//
// Solidity: function registerToNetwork(address _operator, uint256 _votingPower, uint256[4] _blsKey, address _rewardsReceiver) returns()
func (_AttestationCenter *AttestationCenterSession) RegisterToNetwork(_operator common.Address, _votingPower *big.Int, _blsKey [4]*big.Int, _rewardsReceiver common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RegisterToNetwork(&_AttestationCenter.TransactOpts, _operator, _votingPower, _blsKey, _rewardsReceiver)
}

// RegisterToNetwork is a paid mutator transaction binding the contract method 0xfcd4e66a.
//
// Solidity: function registerToNetwork(address _operator, uint256 _votingPower, uint256[4] _blsKey, address _rewardsReceiver) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) RegisterToNetwork(_operator common.Address, _votingPower *big.Int, _blsKey [4]*big.Int, _rewardsReceiver common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RegisterToNetwork(&_AttestationCenter.TransactOpts, _operator, _votingPower, _blsKey, _rewardsReceiver)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AttestationCenter *AttestationCenterTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AttestationCenter *AttestationCenterSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RenounceRole(&_AttestationCenter.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RenounceRole(&_AttestationCenter.TransactOpts, role, callerConfirmation)
}

// RequestBatchPayment is a paid mutator transaction binding the contract method 0x6f382619.
//
// Solidity: function requestBatchPayment(uint256 _from, uint256 _to) returns()
func (_AttestationCenter *AttestationCenterTransactor) RequestBatchPayment(opts *bind.TransactOpts, _from *big.Int, _to *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "requestBatchPayment", _from, _to)
}

// RequestBatchPayment is a paid mutator transaction binding the contract method 0x6f382619.
//
// Solidity: function requestBatchPayment(uint256 _from, uint256 _to) returns()
func (_AttestationCenter *AttestationCenterSession) RequestBatchPayment(_from *big.Int, _to *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RequestBatchPayment(&_AttestationCenter.TransactOpts, _from, _to)
}

// RequestBatchPayment is a paid mutator transaction binding the contract method 0x6f382619.
//
// Solidity: function requestBatchPayment(uint256 _from, uint256 _to) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) RequestBatchPayment(_from *big.Int, _to *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RequestBatchPayment(&_AttestationCenter.TransactOpts, _from, _to)
}

// RequestBatchPayment0 is a paid mutator transaction binding the contract method 0xb7aa2fdf.
//
// Solidity: function requestBatchPayment() returns()
func (_AttestationCenter *AttestationCenterTransactor) RequestBatchPayment0(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "requestBatchPayment0")
}

// RequestBatchPayment0 is a paid mutator transaction binding the contract method 0xb7aa2fdf.
//
// Solidity: function requestBatchPayment() returns()
func (_AttestationCenter *AttestationCenterSession) RequestBatchPayment0() (*types.Transaction, error) {
	return _AttestationCenter.Contract.RequestBatchPayment0(&_AttestationCenter.TransactOpts)
}

// RequestBatchPayment0 is a paid mutator transaction binding the contract method 0xb7aa2fdf.
//
// Solidity: function requestBatchPayment() returns()
func (_AttestationCenter *AttestationCenterTransactorSession) RequestBatchPayment0() (*types.Transaction, error) {
	return _AttestationCenter.Contract.RequestBatchPayment0(&_AttestationCenter.TransactOpts)
}

// RequestPayment is a paid mutator transaction binding the contract method 0x5de988ab.
//
// Solidity: function requestPayment(uint256 _operatorId) returns()
func (_AttestationCenter *AttestationCenterTransactor) RequestPayment(opts *bind.TransactOpts, _operatorId *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "requestPayment", _operatorId)
}

// RequestPayment is a paid mutator transaction binding the contract method 0x5de988ab.
//
// Solidity: function requestPayment(uint256 _operatorId) returns()
func (_AttestationCenter *AttestationCenterSession) RequestPayment(_operatorId *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RequestPayment(&_AttestationCenter.TransactOpts, _operatorId)
}

// RequestPayment is a paid mutator transaction binding the contract method 0x5de988ab.
//
// Solidity: function requestPayment(uint256 _operatorId) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) RequestPayment(_operatorId *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RequestPayment(&_AttestationCenter.TransactOpts, _operatorId)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AttestationCenter *AttestationCenterTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AttestationCenter *AttestationCenterSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RevokeRole(&_AttestationCenter.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.RevokeRole(&_AttestationCenter.TransactOpts, role, account)
}

// SetAvsLogic is a paid mutator transaction binding the contract method 0x008fd386.
//
// Solidity: function setAvsLogic(address _avsLogic) returns()
func (_AttestationCenter *AttestationCenterTransactor) SetAvsLogic(opts *bind.TransactOpts, _avsLogic common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "setAvsLogic", _avsLogic)
}

// SetAvsLogic is a paid mutator transaction binding the contract method 0x008fd386.
//
// Solidity: function setAvsLogic(address _avsLogic) returns()
func (_AttestationCenter *AttestationCenterSession) SetAvsLogic(_avsLogic common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetAvsLogic(&_AttestationCenter.TransactOpts, _avsLogic)
}

// SetAvsLogic is a paid mutator transaction binding the contract method 0x008fd386.
//
// Solidity: function setAvsLogic(address _avsLogic) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) SetAvsLogic(_avsLogic common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetAvsLogic(&_AttestationCenter.TransactOpts, _avsLogic)
}

// SetBeforePaymentsLogic is a paid mutator transaction binding the contract method 0x11a95e38.
//
// Solidity: function setBeforePaymentsLogic(address _beforePaymentsLogic) returns()
func (_AttestationCenter *AttestationCenterTransactor) SetBeforePaymentsLogic(opts *bind.TransactOpts, _beforePaymentsLogic common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "setBeforePaymentsLogic", _beforePaymentsLogic)
}

// SetBeforePaymentsLogic is a paid mutator transaction binding the contract method 0x11a95e38.
//
// Solidity: function setBeforePaymentsLogic(address _beforePaymentsLogic) returns()
func (_AttestationCenter *AttestationCenterSession) SetBeforePaymentsLogic(_beforePaymentsLogic common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetBeforePaymentsLogic(&_AttestationCenter.TransactOpts, _beforePaymentsLogic)
}

// SetBeforePaymentsLogic is a paid mutator transaction binding the contract method 0x11a95e38.
//
// Solidity: function setBeforePaymentsLogic(address _beforePaymentsLogic) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) SetBeforePaymentsLogic(_beforePaymentsLogic common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetBeforePaymentsLogic(&_AttestationCenter.TransactOpts, _beforePaymentsLogic)
}

// SetFeeCalculator is a paid mutator transaction binding the contract method 0x8c66d04f.
//
// Solidity: function setFeeCalculator(address _feeCalculator) returns()
func (_AttestationCenter *AttestationCenterTransactor) SetFeeCalculator(opts *bind.TransactOpts, _feeCalculator common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "setFeeCalculator", _feeCalculator)
}

// SetFeeCalculator is a paid mutator transaction binding the contract method 0x8c66d04f.
//
// Solidity: function setFeeCalculator(address _feeCalculator) returns()
func (_AttestationCenter *AttestationCenterSession) SetFeeCalculator(_feeCalculator common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetFeeCalculator(&_AttestationCenter.TransactOpts, _feeCalculator)
}

// SetFeeCalculator is a paid mutator transaction binding the contract method 0x8c66d04f.
//
// Solidity: function setFeeCalculator(address _feeCalculator) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) SetFeeCalculator(_feeCalculator common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetFeeCalculator(&_AttestationCenter.TransactOpts, _feeCalculator)
}

// SetOblsSharesSyncer is a paid mutator transaction binding the contract method 0x1164224e.
//
// Solidity: function setOblsSharesSyncer(address _oblsSharesSyncer) returns()
func (_AttestationCenter *AttestationCenterTransactor) SetOblsSharesSyncer(opts *bind.TransactOpts, _oblsSharesSyncer common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "setOblsSharesSyncer", _oblsSharesSyncer)
}

// SetOblsSharesSyncer is a paid mutator transaction binding the contract method 0x1164224e.
//
// Solidity: function setOblsSharesSyncer(address _oblsSharesSyncer) returns()
func (_AttestationCenter *AttestationCenterSession) SetOblsSharesSyncer(_oblsSharesSyncer common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetOblsSharesSyncer(&_AttestationCenter.TransactOpts, _oblsSharesSyncer)
}

// SetOblsSharesSyncer is a paid mutator transaction binding the contract method 0x1164224e.
//
// Solidity: function setOblsSharesSyncer(address _oblsSharesSyncer) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) SetOblsSharesSyncer(_oblsSharesSyncer common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetOblsSharesSyncer(&_AttestationCenter.TransactOpts, _oblsSharesSyncer)
}

// SetTaskDefinitionMinVotingPower is a paid mutator transaction binding the contract method 0x64ada5d0.
//
// Solidity: function setTaskDefinitionMinVotingPower(uint16 _taskDefinitionId, uint256 _minimumVotingPower) returns()
func (_AttestationCenter *AttestationCenterTransactor) SetTaskDefinitionMinVotingPower(opts *bind.TransactOpts, _taskDefinitionId uint16, _minimumVotingPower *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "setTaskDefinitionMinVotingPower", _taskDefinitionId, _minimumVotingPower)
}

// SetTaskDefinitionMinVotingPower is a paid mutator transaction binding the contract method 0x64ada5d0.
//
// Solidity: function setTaskDefinitionMinVotingPower(uint16 _taskDefinitionId, uint256 _minimumVotingPower) returns()
func (_AttestationCenter *AttestationCenterSession) SetTaskDefinitionMinVotingPower(_taskDefinitionId uint16, _minimumVotingPower *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetTaskDefinitionMinVotingPower(&_AttestationCenter.TransactOpts, _taskDefinitionId, _minimumVotingPower)
}

// SetTaskDefinitionMinVotingPower is a paid mutator transaction binding the contract method 0x64ada5d0.
//
// Solidity: function setTaskDefinitionMinVotingPower(uint16 _taskDefinitionId, uint256 _minimumVotingPower) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) SetTaskDefinitionMinVotingPower(_taskDefinitionId uint16, _minimumVotingPower *big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetTaskDefinitionMinVotingPower(&_AttestationCenter.TransactOpts, _taskDefinitionId, _minimumVotingPower)
}

// SetTaskDefinitionRestrictedOperators is a paid mutator transaction binding the contract method 0xc8c9e7ab.
//
// Solidity: function setTaskDefinitionRestrictedOperators(uint16 _taskDefinitionId, uint256[] _restrictedOperatorIndexes) returns()
func (_AttestationCenter *AttestationCenterTransactor) SetTaskDefinitionRestrictedOperators(opts *bind.TransactOpts, _taskDefinitionId uint16, _restrictedOperatorIndexes []*big.Int) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "setTaskDefinitionRestrictedOperators", _taskDefinitionId, _restrictedOperatorIndexes)
}

// SetTaskDefinitionRestrictedOperators is a paid mutator transaction binding the contract method 0xc8c9e7ab.
//
// Solidity: function setTaskDefinitionRestrictedOperators(uint16 _taskDefinitionId, uint256[] _restrictedOperatorIndexes) returns()
func (_AttestationCenter *AttestationCenterSession) SetTaskDefinitionRestrictedOperators(_taskDefinitionId uint16, _restrictedOperatorIndexes []*big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetTaskDefinitionRestrictedOperators(&_AttestationCenter.TransactOpts, _taskDefinitionId, _restrictedOperatorIndexes)
}

// SetTaskDefinitionRestrictedOperators is a paid mutator transaction binding the contract method 0xc8c9e7ab.
//
// Solidity: function setTaskDefinitionRestrictedOperators(uint16 _taskDefinitionId, uint256[] _restrictedOperatorIndexes) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) SetTaskDefinitionRestrictedOperators(_taskDefinitionId uint16, _restrictedOperatorIndexes []*big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SetTaskDefinitionRestrictedOperators(&_AttestationCenter.TransactOpts, _taskDefinitionId, _restrictedOperatorIndexes)
}

// SubmitTask is a paid mutator transaction binding the contract method 0x7d5f32bc.
//
// Solidity: function submitTask((string,bytes,address,uint16) _taskInfo, bool _isApproved, bytes _tpSignature, uint256[2] _taSignature, uint256[] _attestersIds) returns()
func (_AttestationCenter *AttestationCenterTransactor) SubmitTask(opts *bind.TransactOpts, _taskInfo IAttestationCenterTaskInfo, _isApproved bool, _tpSignature []byte, _taSignature [2]*big.Int, _attestersIds []*big.Int) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "submitTask", _taskInfo, _isApproved, _tpSignature, _taSignature, _attestersIds)
}

// SubmitTask is a paid mutator transaction binding the contract method 0x7d5f32bc.
//
// Solidity: function submitTask((string,bytes,address,uint16) _taskInfo, bool _isApproved, bytes _tpSignature, uint256[2] _taSignature, uint256[] _attestersIds) returns()
func (_AttestationCenter *AttestationCenterSession) SubmitTask(_taskInfo IAttestationCenterTaskInfo, _isApproved bool, _tpSignature []byte, _taSignature [2]*big.Int, _attestersIds []*big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SubmitTask(&_AttestationCenter.TransactOpts, _taskInfo, _isApproved, _tpSignature, _taSignature, _attestersIds)
}

// SubmitTask is a paid mutator transaction binding the contract method 0x7d5f32bc.
//
// Solidity: function submitTask((string,bytes,address,uint16) _taskInfo, bool _isApproved, bytes _tpSignature, uint256[2] _taSignature, uint256[] _attestersIds) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) SubmitTask(_taskInfo IAttestationCenterTaskInfo, _isApproved bool, _tpSignature []byte, _taSignature [2]*big.Int, _attestersIds []*big.Int) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SubmitTask(&_AttestationCenter.TransactOpts, _taskInfo, _isApproved, _tpSignature, _taSignature, _attestersIds)
}

// SubmitTask0 is a paid mutator transaction binding the contract method 0xfff768e3.
//
// Solidity: function submitTask((string,bytes,address,uint16) _taskInfo, (bool,bytes,uint256[2],uint256[]) _taskSubmissionDetails) returns()
func (_AttestationCenter *AttestationCenterTransactor) SubmitTask0(opts *bind.TransactOpts, _taskInfo IAttestationCenterTaskInfo, _taskSubmissionDetails IAttestationCenterTaskSubmissionDetails) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "submitTask0", _taskInfo, _taskSubmissionDetails)
}

// SubmitTask0 is a paid mutator transaction binding the contract method 0xfff768e3.
//
// Solidity: function submitTask((string,bytes,address,uint16) _taskInfo, (bool,bytes,uint256[2],uint256[]) _taskSubmissionDetails) returns()
func (_AttestationCenter *AttestationCenterSession) SubmitTask0(_taskInfo IAttestationCenterTaskInfo, _taskSubmissionDetails IAttestationCenterTaskSubmissionDetails) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SubmitTask0(&_AttestationCenter.TransactOpts, _taskInfo, _taskSubmissionDetails)
}

// SubmitTask0 is a paid mutator transaction binding the contract method 0xfff768e3.
//
// Solidity: function submitTask((string,bytes,address,uint16) _taskInfo, (bool,bytes,uint256[2],uint256[]) _taskSubmissionDetails) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) SubmitTask0(_taskInfo IAttestationCenterTaskInfo, _taskSubmissionDetails IAttestationCenterTaskSubmissionDetails) (*types.Transaction, error) {
	return _AttestationCenter.Contract.SubmitTask0(&_AttestationCenter.TransactOpts, _taskInfo, _taskSubmissionDetails)
}

// TransferAvsGovernanceMultisig is a paid mutator transaction binding the contract method 0x513c52ba.
//
// Solidity: function transferAvsGovernanceMultisig(address _newAvsGovernanceMultisig) returns()
func (_AttestationCenter *AttestationCenterTransactor) TransferAvsGovernanceMultisig(opts *bind.TransactOpts, _newAvsGovernanceMultisig common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "transferAvsGovernanceMultisig", _newAvsGovernanceMultisig)
}

// TransferAvsGovernanceMultisig is a paid mutator transaction binding the contract method 0x513c52ba.
//
// Solidity: function transferAvsGovernanceMultisig(address _newAvsGovernanceMultisig) returns()
func (_AttestationCenter *AttestationCenterSession) TransferAvsGovernanceMultisig(_newAvsGovernanceMultisig common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.TransferAvsGovernanceMultisig(&_AttestationCenter.TransactOpts, _newAvsGovernanceMultisig)
}

// TransferAvsGovernanceMultisig is a paid mutator transaction binding the contract method 0x513c52ba.
//
// Solidity: function transferAvsGovernanceMultisig(address _newAvsGovernanceMultisig) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) TransferAvsGovernanceMultisig(_newAvsGovernanceMultisig common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.TransferAvsGovernanceMultisig(&_AttestationCenter.TransactOpts, _newAvsGovernanceMultisig)
}

// TransferMessageHandler is a paid mutator transaction binding the contract method 0x4d07f651.
//
// Solidity: function transferMessageHandler(address _newMessageHandler) returns()
func (_AttestationCenter *AttestationCenterTransactor) TransferMessageHandler(opts *bind.TransactOpts, _newMessageHandler common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "transferMessageHandler", _newMessageHandler)
}

// TransferMessageHandler is a paid mutator transaction binding the contract method 0x4d07f651.
//
// Solidity: function transferMessageHandler(address _newMessageHandler) returns()
func (_AttestationCenter *AttestationCenterSession) TransferMessageHandler(_newMessageHandler common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.TransferMessageHandler(&_AttestationCenter.TransactOpts, _newMessageHandler)
}

// TransferMessageHandler is a paid mutator transaction binding the contract method 0x4d07f651.
//
// Solidity: function transferMessageHandler(address _newMessageHandler) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) TransferMessageHandler(_newMessageHandler common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.TransferMessageHandler(&_AttestationCenter.TransactOpts, _newMessageHandler)
}

// UnRegisterOperatorFromNetwork is a paid mutator transaction binding the contract method 0x27bbb287.
//
// Solidity: function unRegisterOperatorFromNetwork(address _operator) returns()
func (_AttestationCenter *AttestationCenterTransactor) UnRegisterOperatorFromNetwork(opts *bind.TransactOpts, _operator common.Address) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "unRegisterOperatorFromNetwork", _operator)
}

// UnRegisterOperatorFromNetwork is a paid mutator transaction binding the contract method 0x27bbb287.
//
// Solidity: function unRegisterOperatorFromNetwork(address _operator) returns()
func (_AttestationCenter *AttestationCenterSession) UnRegisterOperatorFromNetwork(_operator common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.UnRegisterOperatorFromNetwork(&_AttestationCenter.TransactOpts, _operator)
}

// UnRegisterOperatorFromNetwork is a paid mutator transaction binding the contract method 0x27bbb287.
//
// Solidity: function unRegisterOperatorFromNetwork(address _operator) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) UnRegisterOperatorFromNetwork(_operator common.Address) (*types.Transaction, error) {
	return _AttestationCenter.Contract.UnRegisterOperatorFromNetwork(&_AttestationCenter.TransactOpts, _operator)
}

// Unpause is a paid mutator transaction binding the contract method 0xbac1e94b.
//
// Solidity: function unpause(bytes4 _pausableFlow) returns()
func (_AttestationCenter *AttestationCenterTransactor) Unpause(opts *bind.TransactOpts, _pausableFlow [4]byte) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "unpause", _pausableFlow)
}

// Unpause is a paid mutator transaction binding the contract method 0xbac1e94b.
//
// Solidity: function unpause(bytes4 _pausableFlow) returns()
func (_AttestationCenter *AttestationCenterSession) Unpause(_pausableFlow [4]byte) (*types.Transaction, error) {
	return _AttestationCenter.Contract.Unpause(&_AttestationCenter.TransactOpts, _pausableFlow)
}

// Unpause is a paid mutator transaction binding the contract method 0xbac1e94b.
//
// Solidity: function unpause(bytes4 _pausableFlow) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) Unpause(_pausableFlow [4]byte) (*types.Transaction, error) {
	return _AttestationCenter.Contract.Unpause(&_AttestationCenter.TransactOpts, _pausableFlow)
}

// UpdateBlsKey is a paid mutator transaction binding the contract method 0x6ba5aa46.
//
// Solidity: function updateBlsKey(uint256[4] _blsKey, (uint256[2]) _authSignature) returns()
func (_AttestationCenter *AttestationCenterTransactor) UpdateBlsKey(opts *bind.TransactOpts, _blsKey [4]*big.Int, _authSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AttestationCenter.contract.Transact(opts, "updateBlsKey", _blsKey, _authSignature)
}

// UpdateBlsKey is a paid mutator transaction binding the contract method 0x6ba5aa46.
//
// Solidity: function updateBlsKey(uint256[4] _blsKey, (uint256[2]) _authSignature) returns()
func (_AttestationCenter *AttestationCenterSession) UpdateBlsKey(_blsKey [4]*big.Int, _authSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AttestationCenter.Contract.UpdateBlsKey(&_AttestationCenter.TransactOpts, _blsKey, _authSignature)
}

// UpdateBlsKey is a paid mutator transaction binding the contract method 0x6ba5aa46.
//
// Solidity: function updateBlsKey(uint256[4] _blsKey, (uint256[2]) _authSignature) returns()
func (_AttestationCenter *AttestationCenterTransactorSession) UpdateBlsKey(_blsKey [4]*big.Int, _authSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AttestationCenter.Contract.UpdateBlsKey(&_AttestationCenter.TransactOpts, _blsKey, _authSignature)
}

// AttestationCenterClearPaymentRejectedIterator is returned from FilterClearPaymentRejected and is used to iterate over the raw logs and unpacked data for ClearPaymentRejected events raised by the AttestationCenter contract.
type AttestationCenterClearPaymentRejectedIterator struct {
	Event *AttestationCenterClearPaymentRejected // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterClearPaymentRejectedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterClearPaymentRejected)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterClearPaymentRejected)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterClearPaymentRejectedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterClearPaymentRejectedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterClearPaymentRejected represents a ClearPaymentRejected event raised by the AttestationCenter contract.
type AttestationCenterClearPaymentRejected struct {
	Operator               common.Address
	RequestedTaskNumber    *big.Int
	RequestedAmountClaimed *big.Int
	Raw                    types.Log // Blockchain specific contextual infos
}

// FilterClearPaymentRejected is a free log retrieval operation binding the contract event 0x1e643658b8248efd3563f24d116430bf571d036bea3721d94e848890a00a1023.
//
// Solidity: event ClearPaymentRejected(address operator, uint256 requestedTaskNumber, uint256 requestedAmountClaimed)
func (_AttestationCenter *AttestationCenterFilterer) FilterClearPaymentRejected(opts *bind.FilterOpts) (*AttestationCenterClearPaymentRejectedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "ClearPaymentRejected")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterClearPaymentRejectedIterator{contract: _AttestationCenter.contract, event: "ClearPaymentRejected", logs: logs, sub: sub}, nil
}

// WatchClearPaymentRejected is a free log subscription operation binding the contract event 0x1e643658b8248efd3563f24d116430bf571d036bea3721d94e848890a00a1023.
//
// Solidity: event ClearPaymentRejected(address operator, uint256 requestedTaskNumber, uint256 requestedAmountClaimed)
func (_AttestationCenter *AttestationCenterFilterer) WatchClearPaymentRejected(opts *bind.WatchOpts, sink chan<- *AttestationCenterClearPaymentRejected) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "ClearPaymentRejected")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterClearPaymentRejected)
				if err := _AttestationCenter.contract.UnpackLog(event, "ClearPaymentRejected", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseClearPaymentRejected is a log parse operation binding the contract event 0x1e643658b8248efd3563f24d116430bf571d036bea3721d94e848890a00a1023.
//
// Solidity: event ClearPaymentRejected(address operator, uint256 requestedTaskNumber, uint256 requestedAmountClaimed)
func (_AttestationCenter *AttestationCenterFilterer) ParseClearPaymentRejected(log types.Log) (*AttestationCenterClearPaymentRejected, error) {
	event := new(AttestationCenterClearPaymentRejected)
	if err := _AttestationCenter.contract.UnpackLog(event, "ClearPaymentRejected", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterFlowPausedIterator is returned from FilterFlowPaused and is used to iterate over the raw logs and unpacked data for FlowPaused events raised by the AttestationCenter contract.
type AttestationCenterFlowPausedIterator struct {
	Event *AttestationCenterFlowPaused // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterFlowPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterFlowPaused)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterFlowPaused)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterFlowPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterFlowPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterFlowPaused represents a FlowPaused event raised by the AttestationCenter contract.
type AttestationCenterFlowPaused struct {
	PausableFlow [4]byte
	Pauser       common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterFlowPaused is a free log retrieval operation binding the contract event 0x95c3658c5e0c74e20cf12db371b9b67d26e97a1937f6d2284f88cc44d036b4f6.
//
// Solidity: event FlowPaused(bytes4 _pausableFlow, address _pauser)
func (_AttestationCenter *AttestationCenterFilterer) FilterFlowPaused(opts *bind.FilterOpts) (*AttestationCenterFlowPausedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "FlowPaused")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterFlowPausedIterator{contract: _AttestationCenter.contract, event: "FlowPaused", logs: logs, sub: sub}, nil
}

// WatchFlowPaused is a free log subscription operation binding the contract event 0x95c3658c5e0c74e20cf12db371b9b67d26e97a1937f6d2284f88cc44d036b4f6.
//
// Solidity: event FlowPaused(bytes4 _pausableFlow, address _pauser)
func (_AttestationCenter *AttestationCenterFilterer) WatchFlowPaused(opts *bind.WatchOpts, sink chan<- *AttestationCenterFlowPaused) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "FlowPaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterFlowPaused)
				if err := _AttestationCenter.contract.UnpackLog(event, "FlowPaused", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFlowPaused is a log parse operation binding the contract event 0x95c3658c5e0c74e20cf12db371b9b67d26e97a1937f6d2284f88cc44d036b4f6.
//
// Solidity: event FlowPaused(bytes4 _pausableFlow, address _pauser)
func (_AttestationCenter *AttestationCenterFilterer) ParseFlowPaused(log types.Log) (*AttestationCenterFlowPaused, error) {
	event := new(AttestationCenterFlowPaused)
	if err := _AttestationCenter.contract.UnpackLog(event, "FlowPaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterFlowUnpausedIterator is returned from FilterFlowUnpaused and is used to iterate over the raw logs and unpacked data for FlowUnpaused events raised by the AttestationCenter contract.
type AttestationCenterFlowUnpausedIterator struct {
	Event *AttestationCenterFlowUnpaused // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterFlowUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterFlowUnpaused)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterFlowUnpaused)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterFlowUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterFlowUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterFlowUnpaused represents a FlowUnpaused event raised by the AttestationCenter contract.
type AttestationCenterFlowUnpaused struct {
	PausableFlowFlag [4]byte
	Unpauser         common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterFlowUnpaused is a free log retrieval operation binding the contract event 0xc7e56e17b0a6c4b467df6495e1eda1baecd7ba20604e80c1058ac06f4578d85e.
//
// Solidity: event FlowUnpaused(bytes4 _pausableFlowFlag, address _unpauser)
func (_AttestationCenter *AttestationCenterFilterer) FilterFlowUnpaused(opts *bind.FilterOpts) (*AttestationCenterFlowUnpausedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "FlowUnpaused")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterFlowUnpausedIterator{contract: _AttestationCenter.contract, event: "FlowUnpaused", logs: logs, sub: sub}, nil
}

// WatchFlowUnpaused is a free log subscription operation binding the contract event 0xc7e56e17b0a6c4b467df6495e1eda1baecd7ba20604e80c1058ac06f4578d85e.
//
// Solidity: event FlowUnpaused(bytes4 _pausableFlowFlag, address _unpauser)
func (_AttestationCenter *AttestationCenterFilterer) WatchFlowUnpaused(opts *bind.WatchOpts, sink chan<- *AttestationCenterFlowUnpaused) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "FlowUnpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterFlowUnpaused)
				if err := _AttestationCenter.contract.UnpackLog(event, "FlowUnpaused", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFlowUnpaused is a log parse operation binding the contract event 0xc7e56e17b0a6c4b467df6495e1eda1baecd7ba20604e80c1058ac06f4578d85e.
//
// Solidity: event FlowUnpaused(bytes4 _pausableFlowFlag, address _unpauser)
func (_AttestationCenter *AttestationCenterFilterer) ParseFlowUnpaused(log types.Log) (*AttestationCenterFlowUnpaused, error) {
	event := new(AttestationCenterFlowUnpaused)
	if err := _AttestationCenter.contract.UnpackLog(event, "FlowUnpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the AttestationCenter contract.
type AttestationCenterInitializedIterator struct {
	Event *AttestationCenterInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterInitialized represents a Initialized event raised by the AttestationCenter contract.
type AttestationCenterInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_AttestationCenter *AttestationCenterFilterer) FilterInitialized(opts *bind.FilterOpts) (*AttestationCenterInitializedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterInitializedIterator{contract: _AttestationCenter.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_AttestationCenter *AttestationCenterFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *AttestationCenterInitialized) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterInitialized)
				if err := _AttestationCenter.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_AttestationCenter *AttestationCenterFilterer) ParseInitialized(log types.Log) (*AttestationCenterInitialized, error) {
	event := new(AttestationCenterInitialized)
	if err := _AttestationCenter.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterOperatorBlsKeyUpdatedIterator is returned from FilterOperatorBlsKeyUpdated and is used to iterate over the raw logs and unpacked data for OperatorBlsKeyUpdated events raised by the AttestationCenter contract.
type AttestationCenterOperatorBlsKeyUpdatedIterator struct {
	Event *AttestationCenterOperatorBlsKeyUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterOperatorBlsKeyUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterOperatorBlsKeyUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterOperatorBlsKeyUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterOperatorBlsKeyUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterOperatorBlsKeyUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterOperatorBlsKeyUpdated represents a OperatorBlsKeyUpdated event raised by the AttestationCenter contract.
type AttestationCenterOperatorBlsKeyUpdated struct {
	Operator common.Address
	BlsKey   [4]*big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOperatorBlsKeyUpdated is a free log retrieval operation binding the contract event 0x764bc14e663abcee4585e1a92e552918c69d453c673e7161504ff52fc3d428c9.
//
// Solidity: event OperatorBlsKeyUpdated(address indexed operator, uint256[4] blsKey)
func (_AttestationCenter *AttestationCenterFilterer) FilterOperatorBlsKeyUpdated(opts *bind.FilterOpts, operator []common.Address) (*AttestationCenterOperatorBlsKeyUpdatedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "OperatorBlsKeyUpdated", operatorRule)
	if err != nil {
		return nil, err
	}
	return &AttestationCenterOperatorBlsKeyUpdatedIterator{contract: _AttestationCenter.contract, event: "OperatorBlsKeyUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorBlsKeyUpdated is a free log subscription operation binding the contract event 0x764bc14e663abcee4585e1a92e552918c69d453c673e7161504ff52fc3d428c9.
//
// Solidity: event OperatorBlsKeyUpdated(address indexed operator, uint256[4] blsKey)
func (_AttestationCenter *AttestationCenterFilterer) WatchOperatorBlsKeyUpdated(opts *bind.WatchOpts, sink chan<- *AttestationCenterOperatorBlsKeyUpdated, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "OperatorBlsKeyUpdated", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterOperatorBlsKeyUpdated)
				if err := _AttestationCenter.contract.UnpackLog(event, "OperatorBlsKeyUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOperatorBlsKeyUpdated is a log parse operation binding the contract event 0x764bc14e663abcee4585e1a92e552918c69d453c673e7161504ff52fc3d428c9.
//
// Solidity: event OperatorBlsKeyUpdated(address indexed operator, uint256[4] blsKey)
func (_AttestationCenter *AttestationCenterFilterer) ParseOperatorBlsKeyUpdated(log types.Log) (*AttestationCenterOperatorBlsKeyUpdated, error) {
	event := new(AttestationCenterOperatorBlsKeyUpdated)
	if err := _AttestationCenter.contract.UnpackLog(event, "OperatorBlsKeyUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterOperatorRegisteredToNetworkIterator is returned from FilterOperatorRegisteredToNetwork and is used to iterate over the raw logs and unpacked data for OperatorRegisteredToNetwork events raised by the AttestationCenter contract.
type AttestationCenterOperatorRegisteredToNetworkIterator struct {
	Event *AttestationCenterOperatorRegisteredToNetwork // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterOperatorRegisteredToNetworkIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterOperatorRegisteredToNetwork)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterOperatorRegisteredToNetwork)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterOperatorRegisteredToNetworkIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterOperatorRegisteredToNetworkIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterOperatorRegisteredToNetwork represents a OperatorRegisteredToNetwork event raised by the AttestationCenter contract.
type AttestationCenterOperatorRegisteredToNetwork struct {
	Operator    common.Address
	VotingPower *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterOperatorRegisteredToNetwork is a free log retrieval operation binding the contract event 0x16c1a2a8195923d655fe84191b37c746f4385a5c32c038578958b29f52daa1a8.
//
// Solidity: event OperatorRegisteredToNetwork(address operator, uint256 votingPower)
func (_AttestationCenter *AttestationCenterFilterer) FilterOperatorRegisteredToNetwork(opts *bind.FilterOpts) (*AttestationCenterOperatorRegisteredToNetworkIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "OperatorRegisteredToNetwork")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterOperatorRegisteredToNetworkIterator{contract: _AttestationCenter.contract, event: "OperatorRegisteredToNetwork", logs: logs, sub: sub}, nil
}

// WatchOperatorRegisteredToNetwork is a free log subscription operation binding the contract event 0x16c1a2a8195923d655fe84191b37c746f4385a5c32c038578958b29f52daa1a8.
//
// Solidity: event OperatorRegisteredToNetwork(address operator, uint256 votingPower)
func (_AttestationCenter *AttestationCenterFilterer) WatchOperatorRegisteredToNetwork(opts *bind.WatchOpts, sink chan<- *AttestationCenterOperatorRegisteredToNetwork) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "OperatorRegisteredToNetwork")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterOperatorRegisteredToNetwork)
				if err := _AttestationCenter.contract.UnpackLog(event, "OperatorRegisteredToNetwork", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOperatorRegisteredToNetwork is a log parse operation binding the contract event 0x16c1a2a8195923d655fe84191b37c746f4385a5c32c038578958b29f52daa1a8.
//
// Solidity: event OperatorRegisteredToNetwork(address operator, uint256 votingPower)
func (_AttestationCenter *AttestationCenterFilterer) ParseOperatorRegisteredToNetwork(log types.Log) (*AttestationCenterOperatorRegisteredToNetwork, error) {
	event := new(AttestationCenterOperatorRegisteredToNetwork)
	if err := _AttestationCenter.contract.UnpackLog(event, "OperatorRegisteredToNetwork", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterOperatorUnregisteredFromNetworkIterator is returned from FilterOperatorUnregisteredFromNetwork and is used to iterate over the raw logs and unpacked data for OperatorUnregisteredFromNetwork events raised by the AttestationCenter contract.
type AttestationCenterOperatorUnregisteredFromNetworkIterator struct {
	Event *AttestationCenterOperatorUnregisteredFromNetwork // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterOperatorUnregisteredFromNetworkIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterOperatorUnregisteredFromNetwork)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterOperatorUnregisteredFromNetwork)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterOperatorUnregisteredFromNetworkIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterOperatorUnregisteredFromNetworkIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterOperatorUnregisteredFromNetwork represents a OperatorUnregisteredFromNetwork event raised by the AttestationCenter contract.
type AttestationCenterOperatorUnregisteredFromNetwork struct {
	OperatorId *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterOperatorUnregisteredFromNetwork is a free log retrieval operation binding the contract event 0xda04f7db725bc5a9ad418baf26d08e9f24561a7cc119bfe1dd26bfebfc175db3.
//
// Solidity: event OperatorUnregisteredFromNetwork(uint256 operatorId)
func (_AttestationCenter *AttestationCenterFilterer) FilterOperatorUnregisteredFromNetwork(opts *bind.FilterOpts) (*AttestationCenterOperatorUnregisteredFromNetworkIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "OperatorUnregisteredFromNetwork")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterOperatorUnregisteredFromNetworkIterator{contract: _AttestationCenter.contract, event: "OperatorUnregisteredFromNetwork", logs: logs, sub: sub}, nil
}

// WatchOperatorUnregisteredFromNetwork is a free log subscription operation binding the contract event 0xda04f7db725bc5a9ad418baf26d08e9f24561a7cc119bfe1dd26bfebfc175db3.
//
// Solidity: event OperatorUnregisteredFromNetwork(uint256 operatorId)
func (_AttestationCenter *AttestationCenterFilterer) WatchOperatorUnregisteredFromNetwork(opts *bind.WatchOpts, sink chan<- *AttestationCenterOperatorUnregisteredFromNetwork) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "OperatorUnregisteredFromNetwork")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterOperatorUnregisteredFromNetwork)
				if err := _AttestationCenter.contract.UnpackLog(event, "OperatorUnregisteredFromNetwork", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOperatorUnregisteredFromNetwork is a log parse operation binding the contract event 0xda04f7db725bc5a9ad418baf26d08e9f24561a7cc119bfe1dd26bfebfc175db3.
//
// Solidity: event OperatorUnregisteredFromNetwork(uint256 operatorId)
func (_AttestationCenter *AttestationCenterFilterer) ParseOperatorUnregisteredFromNetwork(log types.Log) (*AttestationCenterOperatorUnregisteredFromNetwork, error) {
	event := new(AttestationCenterOperatorUnregisteredFromNetwork)
	if err := _AttestationCenter.contract.UnpackLog(event, "OperatorUnregisteredFromNetwork", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterPaymentRequestedIterator is returned from FilterPaymentRequested and is used to iterate over the raw logs and unpacked data for PaymentRequested events raised by the AttestationCenter contract.
type AttestationCenterPaymentRequestedIterator struct {
	Event *AttestationCenterPaymentRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterPaymentRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterPaymentRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterPaymentRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterPaymentRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterPaymentRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterPaymentRequested represents a PaymentRequested event raised by the AttestationCenter contract.
type AttestationCenterPaymentRequested struct {
	Operator           common.Address
	LastPaidTaskNumber *big.Int
	FeeToClaim         *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterPaymentRequested is a free log retrieval operation binding the contract event 0x34682c7a1451dbbbc9e3be4912a8f466eba3a8c72e9bcb5cb3a61e423a9c6973.
//
// Solidity: event PaymentRequested(address operator, uint256 lastPaidTaskNumber, uint256 feeToClaim)
func (_AttestationCenter *AttestationCenterFilterer) FilterPaymentRequested(opts *bind.FilterOpts) (*AttestationCenterPaymentRequestedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "PaymentRequested")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterPaymentRequestedIterator{contract: _AttestationCenter.contract, event: "PaymentRequested", logs: logs, sub: sub}, nil
}

// WatchPaymentRequested is a free log subscription operation binding the contract event 0x34682c7a1451dbbbc9e3be4912a8f466eba3a8c72e9bcb5cb3a61e423a9c6973.
//
// Solidity: event PaymentRequested(address operator, uint256 lastPaidTaskNumber, uint256 feeToClaim)
func (_AttestationCenter *AttestationCenterFilterer) WatchPaymentRequested(opts *bind.WatchOpts, sink chan<- *AttestationCenterPaymentRequested) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "PaymentRequested")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterPaymentRequested)
				if err := _AttestationCenter.contract.UnpackLog(event, "PaymentRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaymentRequested is a log parse operation binding the contract event 0x34682c7a1451dbbbc9e3be4912a8f466eba3a8c72e9bcb5cb3a61e423a9c6973.
//
// Solidity: event PaymentRequested(address operator, uint256 lastPaidTaskNumber, uint256 feeToClaim)
func (_AttestationCenter *AttestationCenterFilterer) ParsePaymentRequested(log types.Log) (*AttestationCenterPaymentRequested, error) {
	event := new(AttestationCenterPaymentRequested)
	if err := _AttestationCenter.contract.UnpackLog(event, "PaymentRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterPaymentsRequestedIterator is returned from FilterPaymentsRequested and is used to iterate over the raw logs and unpacked data for PaymentsRequested events raised by the AttestationCenter contract.
type AttestationCenterPaymentsRequestedIterator struct {
	Event *AttestationCenterPaymentsRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterPaymentsRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterPaymentsRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterPaymentsRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterPaymentsRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterPaymentsRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterPaymentsRequested represents a PaymentsRequested event raised by the AttestationCenter contract.
type AttestationCenterPaymentsRequested struct {
	Operators          []IAttestationCenterPaymentRequestMessage
	LastPaidTaskNumber *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterPaymentsRequested is a free log retrieval operation binding the contract event 0x3e17ccbb628e667c33839a666b60f15eaefb9db2cbae6f7cc9f3f223cd77fece.
//
// Solidity: event PaymentsRequested((address,uint256)[] operators, uint256 lastPaidTaskNumber)
func (_AttestationCenter *AttestationCenterFilterer) FilterPaymentsRequested(opts *bind.FilterOpts) (*AttestationCenterPaymentsRequestedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "PaymentsRequested")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterPaymentsRequestedIterator{contract: _AttestationCenter.contract, event: "PaymentsRequested", logs: logs, sub: sub}, nil
}

// WatchPaymentsRequested is a free log subscription operation binding the contract event 0x3e17ccbb628e667c33839a666b60f15eaefb9db2cbae6f7cc9f3f223cd77fece.
//
// Solidity: event PaymentsRequested((address,uint256)[] operators, uint256 lastPaidTaskNumber)
func (_AttestationCenter *AttestationCenterFilterer) WatchPaymentsRequested(opts *bind.WatchOpts, sink chan<- *AttestationCenterPaymentsRequested) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "PaymentsRequested")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterPaymentsRequested)
				if err := _AttestationCenter.contract.UnpackLog(event, "PaymentsRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaymentsRequested is a log parse operation binding the contract event 0x3e17ccbb628e667c33839a666b60f15eaefb9db2cbae6f7cc9f3f223cd77fece.
//
// Solidity: event PaymentsRequested((address,uint256)[] operators, uint256 lastPaidTaskNumber)
func (_AttestationCenter *AttestationCenterFilterer) ParsePaymentsRequested(log types.Log) (*AttestationCenterPaymentsRequested, error) {
	event := new(AttestationCenterPaymentsRequested)
	if err := _AttestationCenter.contract.UnpackLog(event, "PaymentsRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterRewardAccumulatedIterator is returned from FilterRewardAccumulated and is used to iterate over the raw logs and unpacked data for RewardAccumulated events raised by the AttestationCenter contract.
type AttestationCenterRewardAccumulatedIterator struct {
	Event *AttestationCenterRewardAccumulated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterRewardAccumulatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterRewardAccumulated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterRewardAccumulated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterRewardAccumulatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterRewardAccumulatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterRewardAccumulated represents a RewardAccumulated event raised by the AttestationCenter contract.
type AttestationCenterRewardAccumulated struct {
	OperatorId               *big.Int
	BaseRewardFeeForOperator *big.Int
	TaskNumber               uint32
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterRewardAccumulated is a free log retrieval operation binding the contract event 0xd3f16e9d8d3fe0ea8a6e5f923fe57e1ae1af6d890ac6c371e8af6cc177a49b65.
//
// Solidity: event RewardAccumulated(uint256 indexed _operatorId, uint256 _baseRewardFeeForOperator, uint32 indexed _taskNumber)
func (_AttestationCenter *AttestationCenterFilterer) FilterRewardAccumulated(opts *bind.FilterOpts, _operatorId []*big.Int, _taskNumber []uint32) (*AttestationCenterRewardAccumulatedIterator, error) {

	var _operatorIdRule []interface{}
	for _, _operatorIdItem := range _operatorId {
		_operatorIdRule = append(_operatorIdRule, _operatorIdItem)
	}

	var _taskNumberRule []interface{}
	for _, _taskNumberItem := range _taskNumber {
		_taskNumberRule = append(_taskNumberRule, _taskNumberItem)
	}

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "RewardAccumulated", _operatorIdRule, _taskNumberRule)
	if err != nil {
		return nil, err
	}
	return &AttestationCenterRewardAccumulatedIterator{contract: _AttestationCenter.contract, event: "RewardAccumulated", logs: logs, sub: sub}, nil
}

// WatchRewardAccumulated is a free log subscription operation binding the contract event 0xd3f16e9d8d3fe0ea8a6e5f923fe57e1ae1af6d890ac6c371e8af6cc177a49b65.
//
// Solidity: event RewardAccumulated(uint256 indexed _operatorId, uint256 _baseRewardFeeForOperator, uint32 indexed _taskNumber)
func (_AttestationCenter *AttestationCenterFilterer) WatchRewardAccumulated(opts *bind.WatchOpts, sink chan<- *AttestationCenterRewardAccumulated, _operatorId []*big.Int, _taskNumber []uint32) (event.Subscription, error) {

	var _operatorIdRule []interface{}
	for _, _operatorIdItem := range _operatorId {
		_operatorIdRule = append(_operatorIdRule, _operatorIdItem)
	}

	var _taskNumberRule []interface{}
	for _, _taskNumberItem := range _taskNumber {
		_taskNumberRule = append(_taskNumberRule, _taskNumberItem)
	}

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "RewardAccumulated", _operatorIdRule, _taskNumberRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterRewardAccumulated)
				if err := _AttestationCenter.contract.UnpackLog(event, "RewardAccumulated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRewardAccumulated is a log parse operation binding the contract event 0xd3f16e9d8d3fe0ea8a6e5f923fe57e1ae1af6d890ac6c371e8af6cc177a49b65.
//
// Solidity: event RewardAccumulated(uint256 indexed _operatorId, uint256 _baseRewardFeeForOperator, uint32 indexed _taskNumber)
func (_AttestationCenter *AttestationCenterFilterer) ParseRewardAccumulated(log types.Log) (*AttestationCenterRewardAccumulated, error) {
	event := new(AttestationCenterRewardAccumulated)
	if err := _AttestationCenter.contract.UnpackLog(event, "RewardAccumulated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the AttestationCenter contract.
type AttestationCenterRoleAdminChangedIterator struct {
	Event *AttestationCenterRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterRoleAdminChanged represents a RoleAdminChanged event raised by the AttestationCenter contract.
type AttestationCenterRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AttestationCenter *AttestationCenterFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*AttestationCenterRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &AttestationCenterRoleAdminChangedIterator{contract: _AttestationCenter.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AttestationCenter *AttestationCenterFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *AttestationCenterRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterRoleAdminChanged)
				if err := _AttestationCenter.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AttestationCenter *AttestationCenterFilterer) ParseRoleAdminChanged(log types.Log) (*AttestationCenterRoleAdminChanged, error) {
	event := new(AttestationCenterRoleAdminChanged)
	if err := _AttestationCenter.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the AttestationCenter contract.
type AttestationCenterRoleGrantedIterator struct {
	Event *AttestationCenterRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterRoleGranted represents a RoleGranted event raised by the AttestationCenter contract.
type AttestationCenterRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AttestationCenter *AttestationCenterFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AttestationCenterRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AttestationCenterRoleGrantedIterator{contract: _AttestationCenter.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AttestationCenter *AttestationCenterFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *AttestationCenterRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterRoleGranted)
				if err := _AttestationCenter.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AttestationCenter *AttestationCenterFilterer) ParseRoleGranted(log types.Log) (*AttestationCenterRoleGranted, error) {
	event := new(AttestationCenterRoleGranted)
	if err := _AttestationCenter.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the AttestationCenter contract.
type AttestationCenterRoleRevokedIterator struct {
	Event *AttestationCenterRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterRoleRevoked represents a RoleRevoked event raised by the AttestationCenter contract.
type AttestationCenterRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AttestationCenter *AttestationCenterFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AttestationCenterRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AttestationCenterRoleRevokedIterator{contract: _AttestationCenter.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AttestationCenter *AttestationCenterFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *AttestationCenterRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterRoleRevoked)
				if err := _AttestationCenter.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AttestationCenter *AttestationCenterFilterer) ParseRoleRevoked(log types.Log) (*AttestationCenterRoleRevoked, error) {
	event := new(AttestationCenterRoleRevoked)
	if err := _AttestationCenter.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterSetAvsGovernanceMultisigIterator is returned from FilterSetAvsGovernanceMultisig and is used to iterate over the raw logs and unpacked data for SetAvsGovernanceMultisig events raised by the AttestationCenter contract.
type AttestationCenterSetAvsGovernanceMultisigIterator struct {
	Event *AttestationCenterSetAvsGovernanceMultisig // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterSetAvsGovernanceMultisigIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterSetAvsGovernanceMultisig)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterSetAvsGovernanceMultisig)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterSetAvsGovernanceMultisigIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterSetAvsGovernanceMultisigIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterSetAvsGovernanceMultisig represents a SetAvsGovernanceMultisig event raised by the AttestationCenter contract.
type AttestationCenterSetAvsGovernanceMultisig struct {
	NewAvsGovernanceMultisig common.Address
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterSetAvsGovernanceMultisig is a free log retrieval operation binding the contract event 0x024e98b7d808a3ddb028252dc95dfdcb165a0ca59fcff8984b4fecf9a7222649.
//
// Solidity: event SetAvsGovernanceMultisig(address newAvsGovernanceMultisig)
func (_AttestationCenter *AttestationCenterFilterer) FilterSetAvsGovernanceMultisig(opts *bind.FilterOpts) (*AttestationCenterSetAvsGovernanceMultisigIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "SetAvsGovernanceMultisig")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterSetAvsGovernanceMultisigIterator{contract: _AttestationCenter.contract, event: "SetAvsGovernanceMultisig", logs: logs, sub: sub}, nil
}

// WatchSetAvsGovernanceMultisig is a free log subscription operation binding the contract event 0x024e98b7d808a3ddb028252dc95dfdcb165a0ca59fcff8984b4fecf9a7222649.
//
// Solidity: event SetAvsGovernanceMultisig(address newAvsGovernanceMultisig)
func (_AttestationCenter *AttestationCenterFilterer) WatchSetAvsGovernanceMultisig(opts *bind.WatchOpts, sink chan<- *AttestationCenterSetAvsGovernanceMultisig) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "SetAvsGovernanceMultisig")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterSetAvsGovernanceMultisig)
				if err := _AttestationCenter.contract.UnpackLog(event, "SetAvsGovernanceMultisig", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetAvsGovernanceMultisig is a log parse operation binding the contract event 0x024e98b7d808a3ddb028252dc95dfdcb165a0ca59fcff8984b4fecf9a7222649.
//
// Solidity: event SetAvsGovernanceMultisig(address newAvsGovernanceMultisig)
func (_AttestationCenter *AttestationCenterFilterer) ParseSetAvsGovernanceMultisig(log types.Log) (*AttestationCenterSetAvsGovernanceMultisig, error) {
	event := new(AttestationCenterSetAvsGovernanceMultisig)
	if err := _AttestationCenter.contract.UnpackLog(event, "SetAvsGovernanceMultisig", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterSetAvsLogicIterator is returned from FilterSetAvsLogic and is used to iterate over the raw logs and unpacked data for SetAvsLogic events raised by the AttestationCenter contract.
type AttestationCenterSetAvsLogicIterator struct {
	Event *AttestationCenterSetAvsLogic // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterSetAvsLogicIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterSetAvsLogic)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterSetAvsLogic)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterSetAvsLogicIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterSetAvsLogicIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterSetAvsLogic represents a SetAvsLogic event raised by the AttestationCenter contract.
type AttestationCenterSetAvsLogic struct {
	AvsLogic common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSetAvsLogic is a free log retrieval operation binding the contract event 0xdf0d3b0bf99a87fc195d045bc7ec61d3e2619e6a49dd3f5cb69102b8c9702034.
//
// Solidity: event SetAvsLogic(address avsLogic)
func (_AttestationCenter *AttestationCenterFilterer) FilterSetAvsLogic(opts *bind.FilterOpts) (*AttestationCenterSetAvsLogicIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "SetAvsLogic")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterSetAvsLogicIterator{contract: _AttestationCenter.contract, event: "SetAvsLogic", logs: logs, sub: sub}, nil
}

// WatchSetAvsLogic is a free log subscription operation binding the contract event 0xdf0d3b0bf99a87fc195d045bc7ec61d3e2619e6a49dd3f5cb69102b8c9702034.
//
// Solidity: event SetAvsLogic(address avsLogic)
func (_AttestationCenter *AttestationCenterFilterer) WatchSetAvsLogic(opts *bind.WatchOpts, sink chan<- *AttestationCenterSetAvsLogic) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "SetAvsLogic")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterSetAvsLogic)
				if err := _AttestationCenter.contract.UnpackLog(event, "SetAvsLogic", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetAvsLogic is a log parse operation binding the contract event 0xdf0d3b0bf99a87fc195d045bc7ec61d3e2619e6a49dd3f5cb69102b8c9702034.
//
// Solidity: event SetAvsLogic(address avsLogic)
func (_AttestationCenter *AttestationCenterFilterer) ParseSetAvsLogic(log types.Log) (*AttestationCenterSetAvsLogic, error) {
	event := new(AttestationCenterSetAvsLogic)
	if err := _AttestationCenter.contract.UnpackLog(event, "SetAvsLogic", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterSetBeforePaymentsLogicIterator is returned from FilterSetBeforePaymentsLogic and is used to iterate over the raw logs and unpacked data for SetBeforePaymentsLogic events raised by the AttestationCenter contract.
type AttestationCenterSetBeforePaymentsLogicIterator struct {
	Event *AttestationCenterSetBeforePaymentsLogic // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterSetBeforePaymentsLogicIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterSetBeforePaymentsLogic)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterSetBeforePaymentsLogic)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterSetBeforePaymentsLogicIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterSetBeforePaymentsLogicIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterSetBeforePaymentsLogic represents a SetBeforePaymentsLogic event raised by the AttestationCenter contract.
type AttestationCenterSetBeforePaymentsLogic struct {
	PaymentsLogic common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSetBeforePaymentsLogic is a free log retrieval operation binding the contract event 0x6da780d66fa2f1ae3eb780c2f39d31bec5c71c81a572640a6af0e8a443477792.
//
// Solidity: event SetBeforePaymentsLogic(address paymentsLogic)
func (_AttestationCenter *AttestationCenterFilterer) FilterSetBeforePaymentsLogic(opts *bind.FilterOpts) (*AttestationCenterSetBeforePaymentsLogicIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "SetBeforePaymentsLogic")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterSetBeforePaymentsLogicIterator{contract: _AttestationCenter.contract, event: "SetBeforePaymentsLogic", logs: logs, sub: sub}, nil
}

// WatchSetBeforePaymentsLogic is a free log subscription operation binding the contract event 0x6da780d66fa2f1ae3eb780c2f39d31bec5c71c81a572640a6af0e8a443477792.
//
// Solidity: event SetBeforePaymentsLogic(address paymentsLogic)
func (_AttestationCenter *AttestationCenterFilterer) WatchSetBeforePaymentsLogic(opts *bind.WatchOpts, sink chan<- *AttestationCenterSetBeforePaymentsLogic) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "SetBeforePaymentsLogic")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterSetBeforePaymentsLogic)
				if err := _AttestationCenter.contract.UnpackLog(event, "SetBeforePaymentsLogic", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetBeforePaymentsLogic is a log parse operation binding the contract event 0x6da780d66fa2f1ae3eb780c2f39d31bec5c71c81a572640a6af0e8a443477792.
//
// Solidity: event SetBeforePaymentsLogic(address paymentsLogic)
func (_AttestationCenter *AttestationCenterFilterer) ParseSetBeforePaymentsLogic(log types.Log) (*AttestationCenterSetBeforePaymentsLogic, error) {
	event := new(AttestationCenterSetBeforePaymentsLogic)
	if err := _AttestationCenter.contract.UnpackLog(event, "SetBeforePaymentsLogic", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterSetFeeCalculatorIterator is returned from FilterSetFeeCalculator and is used to iterate over the raw logs and unpacked data for SetFeeCalculator events raised by the AttestationCenter contract.
type AttestationCenterSetFeeCalculatorIterator struct {
	Event *AttestationCenterSetFeeCalculator // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterSetFeeCalculatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterSetFeeCalculator)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterSetFeeCalculator)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterSetFeeCalculatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterSetFeeCalculatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterSetFeeCalculator represents a SetFeeCalculator event raised by the AttestationCenter contract.
type AttestationCenterSetFeeCalculator struct {
	FeeCalculator common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSetFeeCalculator is a free log retrieval operation binding the contract event 0x83b9ee7f260088fdd4ee12aa07fa7daebc115d796b6bfb55bfb0fc839bccff2d.
//
// Solidity: event SetFeeCalculator(address feeCalculator)
func (_AttestationCenter *AttestationCenterFilterer) FilterSetFeeCalculator(opts *bind.FilterOpts) (*AttestationCenterSetFeeCalculatorIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "SetFeeCalculator")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterSetFeeCalculatorIterator{contract: _AttestationCenter.contract, event: "SetFeeCalculator", logs: logs, sub: sub}, nil
}

// WatchSetFeeCalculator is a free log subscription operation binding the contract event 0x83b9ee7f260088fdd4ee12aa07fa7daebc115d796b6bfb55bfb0fc839bccff2d.
//
// Solidity: event SetFeeCalculator(address feeCalculator)
func (_AttestationCenter *AttestationCenterFilterer) WatchSetFeeCalculator(opts *bind.WatchOpts, sink chan<- *AttestationCenterSetFeeCalculator) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "SetFeeCalculator")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterSetFeeCalculator)
				if err := _AttestationCenter.contract.UnpackLog(event, "SetFeeCalculator", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetFeeCalculator is a log parse operation binding the contract event 0x83b9ee7f260088fdd4ee12aa07fa7daebc115d796b6bfb55bfb0fc839bccff2d.
//
// Solidity: event SetFeeCalculator(address feeCalculator)
func (_AttestationCenter *AttestationCenterFilterer) ParseSetFeeCalculator(log types.Log) (*AttestationCenterSetFeeCalculator, error) {
	event := new(AttestationCenterSetFeeCalculator)
	if err := _AttestationCenter.contract.UnpackLog(event, "SetFeeCalculator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterSetMessageHandlerIterator is returned from FilterSetMessageHandler and is used to iterate over the raw logs and unpacked data for SetMessageHandler events raised by the AttestationCenter contract.
type AttestationCenterSetMessageHandlerIterator struct {
	Event *AttestationCenterSetMessageHandler // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterSetMessageHandlerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterSetMessageHandler)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterSetMessageHandler)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterSetMessageHandlerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterSetMessageHandlerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterSetMessageHandler represents a SetMessageHandler event raised by the AttestationCenter contract.
type AttestationCenterSetMessageHandler struct {
	NewMessageHandler common.Address
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterSetMessageHandler is a free log retrieval operation binding the contract event 0x997f84b541d7b68e210e6f50e3402b51d8411dbbc4d44ed81e508383126e4e94.
//
// Solidity: event SetMessageHandler(address newMessageHandler)
func (_AttestationCenter *AttestationCenterFilterer) FilterSetMessageHandler(opts *bind.FilterOpts) (*AttestationCenterSetMessageHandlerIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "SetMessageHandler")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterSetMessageHandlerIterator{contract: _AttestationCenter.contract, event: "SetMessageHandler", logs: logs, sub: sub}, nil
}

// WatchSetMessageHandler is a free log subscription operation binding the contract event 0x997f84b541d7b68e210e6f50e3402b51d8411dbbc4d44ed81e508383126e4e94.
//
// Solidity: event SetMessageHandler(address newMessageHandler)
func (_AttestationCenter *AttestationCenterFilterer) WatchSetMessageHandler(opts *bind.WatchOpts, sink chan<- *AttestationCenterSetMessageHandler) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "SetMessageHandler")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterSetMessageHandler)
				if err := _AttestationCenter.contract.UnpackLog(event, "SetMessageHandler", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetMessageHandler is a log parse operation binding the contract event 0x997f84b541d7b68e210e6f50e3402b51d8411dbbc4d44ed81e508383126e4e94.
//
// Solidity: event SetMessageHandler(address newMessageHandler)
func (_AttestationCenter *AttestationCenterFilterer) ParseSetMessageHandler(log types.Log) (*AttestationCenterSetMessageHandler, error) {
	event := new(AttestationCenterSetMessageHandler)
	if err := _AttestationCenter.contract.UnpackLog(event, "SetMessageHandler", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterSetMinimumTaskDefinitionVotingPowerIterator is returned from FilterSetMinimumTaskDefinitionVotingPower and is used to iterate over the raw logs and unpacked data for SetMinimumTaskDefinitionVotingPower events raised by the AttestationCenter contract.
type AttestationCenterSetMinimumTaskDefinitionVotingPowerIterator struct {
	Event *AttestationCenterSetMinimumTaskDefinitionVotingPower // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterSetMinimumTaskDefinitionVotingPowerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterSetMinimumTaskDefinitionVotingPower)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterSetMinimumTaskDefinitionVotingPower)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterSetMinimumTaskDefinitionVotingPowerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterSetMinimumTaskDefinitionVotingPowerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterSetMinimumTaskDefinitionVotingPower represents a SetMinimumTaskDefinitionVotingPower event raised by the AttestationCenter contract.
type AttestationCenterSetMinimumTaskDefinitionVotingPower struct {
	MinimumVotingPower *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterSetMinimumTaskDefinitionVotingPower is a free log retrieval operation binding the contract event 0x255c174d5deb340ac0a0c908147d9c66ae7d94a6c7f969f722bf5d71b92e98f8.
//
// Solidity: event SetMinimumTaskDefinitionVotingPower(uint256 minimumVotingPower)
func (_AttestationCenter *AttestationCenterFilterer) FilterSetMinimumTaskDefinitionVotingPower(opts *bind.FilterOpts) (*AttestationCenterSetMinimumTaskDefinitionVotingPowerIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "SetMinimumTaskDefinitionVotingPower")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterSetMinimumTaskDefinitionVotingPowerIterator{contract: _AttestationCenter.contract, event: "SetMinimumTaskDefinitionVotingPower", logs: logs, sub: sub}, nil
}

// WatchSetMinimumTaskDefinitionVotingPower is a free log subscription operation binding the contract event 0x255c174d5deb340ac0a0c908147d9c66ae7d94a6c7f969f722bf5d71b92e98f8.
//
// Solidity: event SetMinimumTaskDefinitionVotingPower(uint256 minimumVotingPower)
func (_AttestationCenter *AttestationCenterFilterer) WatchSetMinimumTaskDefinitionVotingPower(opts *bind.WatchOpts, sink chan<- *AttestationCenterSetMinimumTaskDefinitionVotingPower) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "SetMinimumTaskDefinitionVotingPower")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterSetMinimumTaskDefinitionVotingPower)
				if err := _AttestationCenter.contract.UnpackLog(event, "SetMinimumTaskDefinitionVotingPower", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetMinimumTaskDefinitionVotingPower is a log parse operation binding the contract event 0x255c174d5deb340ac0a0c908147d9c66ae7d94a6c7f969f722bf5d71b92e98f8.
//
// Solidity: event SetMinimumTaskDefinitionVotingPower(uint256 minimumVotingPower)
func (_AttestationCenter *AttestationCenterFilterer) ParseSetMinimumTaskDefinitionVotingPower(log types.Log) (*AttestationCenterSetMinimumTaskDefinitionVotingPower, error) {
	event := new(AttestationCenterSetMinimumTaskDefinitionVotingPower)
	if err := _AttestationCenter.contract.UnpackLog(event, "SetMinimumTaskDefinitionVotingPower", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterSetRestrictedOperatorIterator is returned from FilterSetRestrictedOperator and is used to iterate over the raw logs and unpacked data for SetRestrictedOperator events raised by the AttestationCenter contract.
type AttestationCenterSetRestrictedOperatorIterator struct {
	Event *AttestationCenterSetRestrictedOperator // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterSetRestrictedOperatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterSetRestrictedOperator)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterSetRestrictedOperator)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterSetRestrictedOperatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterSetRestrictedOperatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterSetRestrictedOperator represents a SetRestrictedOperator event raised by the AttestationCenter contract.
type AttestationCenterSetRestrictedOperator struct {
	TaskDefinitionId          uint16
	RestrictedOperatorIndexes []*big.Int
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterSetRestrictedOperator is a free log retrieval operation binding the contract event 0x364aa2fa0cf7a32b9240f9ab2bdebc99f0222262852b3b25a87388acff5a5b14.
//
// Solidity: event SetRestrictedOperator(uint16 taskDefinitionId, uint256[] restrictedOperatorIndexes)
func (_AttestationCenter *AttestationCenterFilterer) FilterSetRestrictedOperator(opts *bind.FilterOpts) (*AttestationCenterSetRestrictedOperatorIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "SetRestrictedOperator")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterSetRestrictedOperatorIterator{contract: _AttestationCenter.contract, event: "SetRestrictedOperator", logs: logs, sub: sub}, nil
}

// WatchSetRestrictedOperator is a free log subscription operation binding the contract event 0x364aa2fa0cf7a32b9240f9ab2bdebc99f0222262852b3b25a87388acff5a5b14.
//
// Solidity: event SetRestrictedOperator(uint16 taskDefinitionId, uint256[] restrictedOperatorIndexes)
func (_AttestationCenter *AttestationCenterFilterer) WatchSetRestrictedOperator(opts *bind.WatchOpts, sink chan<- *AttestationCenterSetRestrictedOperator) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "SetRestrictedOperator")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterSetRestrictedOperator)
				if err := _AttestationCenter.contract.UnpackLog(event, "SetRestrictedOperator", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetRestrictedOperator is a log parse operation binding the contract event 0x364aa2fa0cf7a32b9240f9ab2bdebc99f0222262852b3b25a87388acff5a5b14.
//
// Solidity: event SetRestrictedOperator(uint16 taskDefinitionId, uint256[] restrictedOperatorIndexes)
func (_AttestationCenter *AttestationCenterFilterer) ParseSetRestrictedOperator(log types.Log) (*AttestationCenterSetRestrictedOperator, error) {
	event := new(AttestationCenterSetRestrictedOperator)
	if err := _AttestationCenter.contract.UnpackLog(event, "SetRestrictedOperator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterTaskDefinitionCreatedIterator is returned from FilterTaskDefinitionCreated and is used to iterate over the raw logs and unpacked data for TaskDefinitionCreated events raised by the AttestationCenter contract.
type AttestationCenterTaskDefinitionCreatedIterator struct {
	Event *AttestationCenterTaskDefinitionCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterTaskDefinitionCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterTaskDefinitionCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterTaskDefinitionCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterTaskDefinitionCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterTaskDefinitionCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterTaskDefinitionCreated represents a TaskDefinitionCreated event raised by the AttestationCenter contract.
type AttestationCenterTaskDefinitionCreated struct {
	TaskDefinitionId           uint16
	Name                       string
	BlockExpiry                *big.Int
	BaseRewardFeeForAttesters  *big.Int
	BaseRewardFeeForPerformer  *big.Int
	BaseRewardFeeForAggregator *big.Int
	DisputePeriodBlocks        *big.Int
	MinimumVotingPower         *big.Int
	RestrictedOperatorIndexes  []*big.Int
	Raw                        types.Log // Blockchain specific contextual infos
}

// FilterTaskDefinitionCreated is a free log retrieval operation binding the contract event 0x4306a9df64b49fc07c0f1929d57cc2b5cdfc108656189460e6aa127754413ef1.
//
// Solidity: event TaskDefinitionCreated(uint16 taskDefinitionId, string name, uint256 blockExpiry, uint256 baseRewardFeeForAttesters, uint256 baseRewardFeeForPerformer, uint256 baseRewardFeeForAggregator, uint256 disputePeriodBlocks, uint256 minimumVotingPower, uint256[] restrictedOperatorIndexes)
func (_AttestationCenter *AttestationCenterFilterer) FilterTaskDefinitionCreated(opts *bind.FilterOpts) (*AttestationCenterTaskDefinitionCreatedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "TaskDefinitionCreated")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterTaskDefinitionCreatedIterator{contract: _AttestationCenter.contract, event: "TaskDefinitionCreated", logs: logs, sub: sub}, nil
}

// WatchTaskDefinitionCreated is a free log subscription operation binding the contract event 0x4306a9df64b49fc07c0f1929d57cc2b5cdfc108656189460e6aa127754413ef1.
//
// Solidity: event TaskDefinitionCreated(uint16 taskDefinitionId, string name, uint256 blockExpiry, uint256 baseRewardFeeForAttesters, uint256 baseRewardFeeForPerformer, uint256 baseRewardFeeForAggregator, uint256 disputePeriodBlocks, uint256 minimumVotingPower, uint256[] restrictedOperatorIndexes)
func (_AttestationCenter *AttestationCenterFilterer) WatchTaskDefinitionCreated(opts *bind.WatchOpts, sink chan<- *AttestationCenterTaskDefinitionCreated) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "TaskDefinitionCreated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterTaskDefinitionCreated)
				if err := _AttestationCenter.contract.UnpackLog(event, "TaskDefinitionCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTaskDefinitionCreated is a log parse operation binding the contract event 0x4306a9df64b49fc07c0f1929d57cc2b5cdfc108656189460e6aa127754413ef1.
//
// Solidity: event TaskDefinitionCreated(uint16 taskDefinitionId, string name, uint256 blockExpiry, uint256 baseRewardFeeForAttesters, uint256 baseRewardFeeForPerformer, uint256 baseRewardFeeForAggregator, uint256 disputePeriodBlocks, uint256 minimumVotingPower, uint256[] restrictedOperatorIndexes)
func (_AttestationCenter *AttestationCenterFilterer) ParseTaskDefinitionCreated(log types.Log) (*AttestationCenterTaskDefinitionCreated, error) {
	event := new(AttestationCenterTaskDefinitionCreated)
	if err := _AttestationCenter.contract.UnpackLog(event, "TaskDefinitionCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterTaskDefinitionRestrictedOperatorsModifiedIterator is returned from FilterTaskDefinitionRestrictedOperatorsModified and is used to iterate over the raw logs and unpacked data for TaskDefinitionRestrictedOperatorsModified events raised by the AttestationCenter contract.
type AttestationCenterTaskDefinitionRestrictedOperatorsModifiedIterator struct {
	Event *AttestationCenterTaskDefinitionRestrictedOperatorsModified // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterTaskDefinitionRestrictedOperatorsModifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterTaskDefinitionRestrictedOperatorsModified)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterTaskDefinitionRestrictedOperatorsModified)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterTaskDefinitionRestrictedOperatorsModifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterTaskDefinitionRestrictedOperatorsModifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterTaskDefinitionRestrictedOperatorsModified represents a TaskDefinitionRestrictedOperatorsModified event raised by the AttestationCenter contract.
type AttestationCenterTaskDefinitionRestrictedOperatorsModified struct {
	TaskDefinitionId          uint16
	RestrictedOperatorIndexes []*big.Int
	IsRestricted              []bool
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterTaskDefinitionRestrictedOperatorsModified is a free log retrieval operation binding the contract event 0x3a6545f49055b62a6dec3f6d6116dfcde655324bc0c3a4947266f3d11bff8239.
//
// Solidity: event TaskDefinitionRestrictedOperatorsModified(uint16 taskDefinitionId, uint256[] restrictedOperatorIndexes, bool[] isRestricted)
func (_AttestationCenter *AttestationCenterFilterer) FilterTaskDefinitionRestrictedOperatorsModified(opts *bind.FilterOpts) (*AttestationCenterTaskDefinitionRestrictedOperatorsModifiedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "TaskDefinitionRestrictedOperatorsModified")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterTaskDefinitionRestrictedOperatorsModifiedIterator{contract: _AttestationCenter.contract, event: "TaskDefinitionRestrictedOperatorsModified", logs: logs, sub: sub}, nil
}

// WatchTaskDefinitionRestrictedOperatorsModified is a free log subscription operation binding the contract event 0x3a6545f49055b62a6dec3f6d6116dfcde655324bc0c3a4947266f3d11bff8239.
//
// Solidity: event TaskDefinitionRestrictedOperatorsModified(uint16 taskDefinitionId, uint256[] restrictedOperatorIndexes, bool[] isRestricted)
func (_AttestationCenter *AttestationCenterFilterer) WatchTaskDefinitionRestrictedOperatorsModified(opts *bind.WatchOpts, sink chan<- *AttestationCenterTaskDefinitionRestrictedOperatorsModified) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "TaskDefinitionRestrictedOperatorsModified")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterTaskDefinitionRestrictedOperatorsModified)
				if err := _AttestationCenter.contract.UnpackLog(event, "TaskDefinitionRestrictedOperatorsModified", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTaskDefinitionRestrictedOperatorsModified is a log parse operation binding the contract event 0x3a6545f49055b62a6dec3f6d6116dfcde655324bc0c3a4947266f3d11bff8239.
//
// Solidity: event TaskDefinitionRestrictedOperatorsModified(uint16 taskDefinitionId, uint256[] restrictedOperatorIndexes, bool[] isRestricted)
func (_AttestationCenter *AttestationCenterFilterer) ParseTaskDefinitionRestrictedOperatorsModified(log types.Log) (*AttestationCenterTaskDefinitionRestrictedOperatorsModified, error) {
	event := new(AttestationCenterTaskDefinitionRestrictedOperatorsModified)
	if err := _AttestationCenter.contract.UnpackLog(event, "TaskDefinitionRestrictedOperatorsModified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterTaskRejectedIterator is returned from FilterTaskRejected and is used to iterate over the raw logs and unpacked data for TaskRejected events raised by the AttestationCenter contract.
type AttestationCenterTaskRejectedIterator struct {
	Event *AttestationCenterTaskRejected // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterTaskRejectedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterTaskRejected)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterTaskRejected)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterTaskRejectedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterTaskRejectedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterTaskRejected represents a TaskRejected event raised by the AttestationCenter contract.
type AttestationCenterTaskRejected struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterTaskRejected is a free log retrieval operation binding the contract event 0x5d681de577a7b3ff669022ffe837e270ef5b32ee4c1377045cf61b700d99e70a.
//
// Solidity: event TaskRejected(address operator, uint32 taskNumber, string proofOfTask, bytes data, uint16 taskDefinitionId)
func (_AttestationCenter *AttestationCenterFilterer) FilterTaskRejected(opts *bind.FilterOpts) (*AttestationCenterTaskRejectedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "TaskRejected")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterTaskRejectedIterator{contract: _AttestationCenter.contract, event: "TaskRejected", logs: logs, sub: sub}, nil
}

// WatchTaskRejected is a free log subscription operation binding the contract event 0x5d681de577a7b3ff669022ffe837e270ef5b32ee4c1377045cf61b700d99e70a.
//
// Solidity: event TaskRejected(address operator, uint32 taskNumber, string proofOfTask, bytes data, uint16 taskDefinitionId)
func (_AttestationCenter *AttestationCenterFilterer) WatchTaskRejected(opts *bind.WatchOpts, sink chan<- *AttestationCenterTaskRejected) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "TaskRejected")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterTaskRejected)
				if err := _AttestationCenter.contract.UnpackLog(event, "TaskRejected", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTaskRejected is a log parse operation binding the contract event 0x5d681de577a7b3ff669022ffe837e270ef5b32ee4c1377045cf61b700d99e70a.
//
// Solidity: event TaskRejected(address operator, uint32 taskNumber, string proofOfTask, bytes data, uint16 taskDefinitionId)
func (_AttestationCenter *AttestationCenterFilterer) ParseTaskRejected(log types.Log) (*AttestationCenterTaskRejected, error) {
	event := new(AttestationCenterTaskRejected)
	if err := _AttestationCenter.contract.UnpackLog(event, "TaskRejected", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AttestationCenterTaskSubmittedIterator is returned from FilterTaskSubmitted and is used to iterate over the raw logs and unpacked data for TaskSubmitted events raised by the AttestationCenter contract.
type AttestationCenterTaskSubmittedIterator struct {
	Event *AttestationCenterTaskSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AttestationCenterTaskSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AttestationCenterTaskSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AttestationCenterTaskSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AttestationCenterTaskSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AttestationCenterTaskSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AttestationCenterTaskSubmitted represents a TaskSubmitted event raised by the AttestationCenter contract.
type AttestationCenterTaskSubmitted struct {
	Operator         common.Address
	TaskNumber       uint32
	ProofOfTask      string
	Data             []byte
	TaskDefinitionId uint16
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterTaskSubmitted is a free log retrieval operation binding the contract event 0xde1faaa044a7216023ec32c06f91d9b5098d8bbd3b37f3662be3a729752ec9fc.
//
// Solidity: event TaskSubmitted(address operator, uint32 taskNumber, string proofOfTask, bytes data, uint16 taskDefinitionId)
func (_AttestationCenter *AttestationCenterFilterer) FilterTaskSubmitted(opts *bind.FilterOpts) (*AttestationCenterTaskSubmittedIterator, error) {

	logs, sub, err := _AttestationCenter.contract.FilterLogs(opts, "TaskSubmitted")
	if err != nil {
		return nil, err
	}
	return &AttestationCenterTaskSubmittedIterator{contract: _AttestationCenter.contract, event: "TaskSubmitted", logs: logs, sub: sub}, nil
}

// WatchTaskSubmitted is a free log subscription operation binding the contract event 0xde1faaa044a7216023ec32c06f91d9b5098d8bbd3b37f3662be3a729752ec9fc.
//
// Solidity: event TaskSubmitted(address operator, uint32 taskNumber, string proofOfTask, bytes data, uint16 taskDefinitionId)
func (_AttestationCenter *AttestationCenterFilterer) WatchTaskSubmitted(opts *bind.WatchOpts, sink chan<- *AttestationCenterTaskSubmitted) (event.Subscription, error) {

	logs, sub, err := _AttestationCenter.contract.WatchLogs(opts, "TaskSubmitted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AttestationCenterTaskSubmitted)
				if err := _AttestationCenter.contract.UnpackLog(event, "TaskSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTaskSubmitted is a log parse operation binding the contract event 0xde1faaa044a7216023ec32c06f91d9b5098d8bbd3b37f3662be3a729752ec9fc.
//
// Solidity: event TaskSubmitted(address operator, uint32 taskNumber, string proofOfTask, bytes data, uint16 taskDefinitionId)
func (_AttestationCenter *AttestationCenterFilterer) ParseTaskSubmitted(log types.Log) (*AttestationCenterTaskSubmitted, error) {
	event := new(AttestationCenterTaskSubmitted)
	if err := _AttestationCenter.contract.UnpackLog(event, "TaskSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
