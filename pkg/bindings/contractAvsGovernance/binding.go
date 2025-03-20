// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractAvsGovernance

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

// IAvsGovernanceInitializationParams is an auto generated low-level Go binding around an user-defined struct.
type IAvsGovernanceInitializationParams struct {
	AvsGovernanceMultisigOwner common.Address
	OperationsMultisig         common.Address
	CommunityMultisig          common.Address
	OthenticRegistry           common.Address
	MessageHandler             common.Address
	Vault                      common.Address
	AvsDirectoryContract       common.Address
	AllowlistSigner            common.Address
	AvsName                    string
	BlsAuthSingleton           common.Address
}

// IAvsGovernancePaymentRequestMessage is an auto generated low-level Go binding around an user-defined struct.
type IAvsGovernancePaymentRequestMessage struct {
	Operator   common.Address
	FeeToClaim *big.Int
}

// IAvsGovernanceStrategyMultiplier is an auto generated low-level Go binding around an user-defined struct.
type IAvsGovernanceStrategyMultiplier struct {
	Strategy   common.Address
	Multiplier *big.Int
}

// ISignatureUtilsSignatureWithSaltAndExpiry is an auto generated low-level Go binding around an user-defined struct.
type ISignatureUtilsSignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    *big.Int
}

// AvsGovernanceMetaData contains all meta data concerning the AvsGovernance contract.
var AvsGovernanceMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"avsDirectory\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"avsName\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"completeRewardsReceiverModification\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"depositERC20\",\"inputs\":[{\"name\":\"_amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getDefaultStrategies\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getIsAllowlisted\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNumOfOperatorsLimit\",\"inputs\":[],\"outputs\":[{\"name\":\"numOfOperatorsLimitView\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorRestakedStrategies\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRestakeableStrategies\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRewardsReceiver\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_initializationParams\",\"type\":\"tuple\",\"internalType\":\"structIAvsGovernance.InitializationParams\",\"components\":[{\"name\":\"avsGovernanceMultisigOwner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"operationsMultisig\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"communityMultisig\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"othenticRegistry\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"messageHandler\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"vault\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsDirectoryContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"allowlistSigner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"avsName\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"blsAuthSingleton\",\"type\":\"address\",\"internalType\":\"address\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isFlowPaused\",\"inputs\":[{\"name\":\"_pausableFlow\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"_isPaused\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOperatorRegistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxEffectiveBalance\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minSharesForStrategy\",\"inputs\":[{\"name\":\"_strategy\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minVotingPower\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numOfActiveOperators\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numOfOperators\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numOfShares\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[{\"name\":\"_pausableFlow\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"queueRewardsReceiverModification\",\"inputs\":[{\"name\":\"_newRewardsReceiver\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerAsAllowedOperator\",\"inputs\":[{\"name\":\"_blsKey\",\"type\":\"uint256[4]\",\"internalType\":\"uint256[4]\"},{\"name\":\"_authToken\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_rewardsReceiver\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_operatorSignature\",\"type\":\"tuple\",\"internalType\":\"structISignatureUtils.SignatureWithSaltAndExpiry\",\"components\":[{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"_blsRegistrationSignature\",\"type\":\"tuple\",\"internalType\":\"structBLSAuthLibrary.Signature\",\"components\":[{\"name\":\"signature\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerAsOperator\",\"inputs\":[{\"name\":\"_blsKey\",\"type\":\"uint256[4]\",\"internalType\":\"uint256[4]\"},{\"name\":\"_rewardsReceiver\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_operatorSignature\",\"type\":\"tuple\",\"internalType\":\"structISignatureUtils.SignatureWithSaltAndExpiry\",\"components\":[{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"_blsRegistrationSignature\",\"type\":\"tuple\",\"internalType\":\"structBLSAuthLibrary.Signature\",\"components\":[{\"name\":\"signature\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAllowlistSigner\",\"inputs\":[{\"name\":\"_allowlistSigner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAvsGovernanceLogic\",\"inputs\":[{\"name\":\"_avsGovernanceLogic\",\"type\":\"address\",\"internalType\":\"contractIAvsGovernanceLogic\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAvsGovernanceMultiplierSyncer\",\"inputs\":[{\"name\":\"_newAvsGovernanceMultiplierSyncer\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAvsName\",\"inputs\":[{\"name\":\"_avsName\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setBLSAuthSingleton\",\"inputs\":[{\"name\":\"_blsAuthSingleton\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setIsAllowlisted\",\"inputs\":[{\"name\":\"_isAllowlisted\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxEffectiveBalance\",\"inputs\":[{\"name\":\"_maxBalance\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMinSharesForStrategy\",\"inputs\":[{\"name\":\"_strategy\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_minShares\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMinVotingPower\",\"inputs\":[{\"name\":\"_minVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setNumOfOperatorsLimit\",\"inputs\":[{\"name\":\"_newLimitOfNumOfOperators\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOthenticRegistry\",\"inputs\":[{\"name\":\"_othenticRegistry\",\"type\":\"address\",\"internalType\":\"contractIOthenticRegistry\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRewardsReceiverModificationDelay\",\"inputs\":[{\"name\":\"_rewardsReceiverModificationDelay\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setStrategyMultiplier\",\"inputs\":[{\"name\":\"_strategyMultiplier\",\"type\":\"tuple\",\"internalType\":\"structIAvsGovernance.StrategyMultiplier\",\"components\":[{\"name\":\"strategy\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"multiplier\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setStrategyMultiplierBatch\",\"inputs\":[{\"name\":\"_strategyMultipliers\",\"type\":\"tuple[]\",\"internalType\":\"structIAvsGovernance.StrategyMultiplier[]\",\"components\":[{\"name\":\"strategy\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"multiplier\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSupportedStrategies\",\"inputs\":[{\"name\":\"_strategies\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"strategies\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"strategyMultiplier\",\"inputs\":[{\"name\":\"_strategy\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferAvsGovernanceMultisig\",\"inputs\":[{\"name\":\"_newAvsGovernanceMultisig\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferMessageHandler\",\"inputs\":[{\"name\":\"_newMessageHandler\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[{\"name\":\"_pausableFlow\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unregisterAsOperator\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateAVSMetadataURI\",\"inputs\":[{\"name\":\"metadataURI\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"vault\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"votingPower\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdrawBatchRewards\",\"inputs\":[{\"name\":\"_operators\",\"type\":\"tuple[]\",\"internalType\":\"structIAvsGovernance.PaymentRequestMessage[]\",\"components\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeToClaim\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"_lastPayedTask\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawRewards\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_lastPayedTask\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeToClaim\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"BLSAuthSingletonSet\",\"inputs\":[{\"name\":\"blsAuthSingleton\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FlowPaused\",\"inputs\":[{\"name\":\"_pausableFlow\",\"type\":\"bytes4\",\"indexed\":false,\"internalType\":\"bytes4\"},{\"name\":\"_pauser\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FlowUnpaused\",\"inputs\":[{\"name\":\"_pausableFlowFlag\",\"type\":\"bytes4\",\"indexed\":false,\"internalType\":\"bytes4\"},{\"name\":\"_unpauser\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MaxEffectiveBalanceSet\",\"inputs\":[{\"name\":\"maxEffectiveBalance\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MinSharesPerStrategySet\",\"inputs\":[{\"name\":\"strategy\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"minShares\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MinVotingPowerSet\",\"inputs\":[{\"name\":\"minVotingPower\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorRegistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"blsKey\",\"type\":\"uint256[4]\",\"indexed\":false,\"internalType\":\"uint256[4]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorUnregistered\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"QueuedRewardsReceiverModification\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"receiver\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"delay\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetAllowlistSigner\",\"inputs\":[{\"name\":\"allowlistSigner\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetAvsGovernanceLogic\",\"inputs\":[{\"name\":\"avsGovernanceLogic\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetAvsGovernanceMultiplierSyncer\",\"inputs\":[{\"name\":\"avsGovernanceMultiplierSyncer\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetAvsGovernanceMultisig\",\"inputs\":[{\"name\":\"newAvsGovernanceMultisig\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetAvsName\",\"inputs\":[{\"name\":\"avsName\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetIsAllowlisted\",\"inputs\":[{\"name\":\"isAllowlisted\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetMessageHandler\",\"inputs\":[{\"name\":\"newMessageHandler\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetNumOfOperatorsLimit\",\"inputs\":[{\"name\":\"newLimitOfNumOfOperators\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetOthenticRegistry\",\"inputs\":[{\"name\":\"othenticRegistry\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetRewardsReceiver\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"receiver\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetRewardsReceiverModificationDelay\",\"inputs\":[{\"name\":\"modificationDelay\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetStrategyMultiplier\",\"inputs\":[{\"name\":\"strategy\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"multiplier\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetSupportedStrategies\",\"inputs\":[{\"name\":\"strategies\",\"type\":\"address[]\",\"indexed\":false,\"internalType\":\"address[]\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetToken\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlInvalidMultiplierSyncer\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AllowlistDisabled\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AllowlistEnabled\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"FlowIsCurrentlyPaused\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FlowIsCurrentlyUnpaused\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAllowlistAuthToken\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidBlsRegistrationSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidMultiplierNotSet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidRewardsReceiver\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSlashingRate\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidStrategy\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ModificationDelayNotPassed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotEnoughVotingPower\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NumOfActiveOperatorsIsGreaterThanNumOfOperatorLimit\",\"inputs\":[{\"name\":\"numOfOperatorsLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"numOfActiveOperators\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"NumOfOperatorsLimitReached\",\"inputs\":[{\"name\":\"numOfOperatorsLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"OperatorAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PauseFlowIsAlreadyPaused\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"Unauthorized\",\"inputs\":[{\"name\":\"message\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"UnpausingFlowIsAlreadyUnpaused\",\"inputs\":[]}]",
}

// AvsGovernanceABI is the input ABI used to generate the binding from.
// Deprecated: Use AvsGovernanceMetaData.ABI instead.
var AvsGovernanceABI = AvsGovernanceMetaData.ABI

// AvsGovernance is an auto generated Go binding around an Ethereum contract.
type AvsGovernance struct {
	AvsGovernanceCaller     // Read-only binding to the contract
	AvsGovernanceTransactor // Write-only binding to the contract
	AvsGovernanceFilterer   // Log filterer for contract events
}

// AvsGovernanceCaller is an auto generated read-only Go binding around an Ethereum contract.
type AvsGovernanceCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AvsGovernanceTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AvsGovernanceTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AvsGovernanceFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AvsGovernanceFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AvsGovernanceSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AvsGovernanceSession struct {
	Contract     *AvsGovernance    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AvsGovernanceCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AvsGovernanceCallerSession struct {
	Contract *AvsGovernanceCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// AvsGovernanceTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AvsGovernanceTransactorSession struct {
	Contract     *AvsGovernanceTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// AvsGovernanceRaw is an auto generated low-level Go binding around an Ethereum contract.
type AvsGovernanceRaw struct {
	Contract *AvsGovernance // Generic contract binding to access the raw methods on
}

// AvsGovernanceCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AvsGovernanceCallerRaw struct {
	Contract *AvsGovernanceCaller // Generic read-only contract binding to access the raw methods on
}

// AvsGovernanceTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AvsGovernanceTransactorRaw struct {
	Contract *AvsGovernanceTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAvsGovernance creates a new instance of AvsGovernance, bound to a specific deployed contract.
func NewAvsGovernance(address common.Address, backend bind.ContractBackend) (*AvsGovernance, error) {
	contract, err := bindAvsGovernance(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AvsGovernance{AvsGovernanceCaller: AvsGovernanceCaller{contract: contract}, AvsGovernanceTransactor: AvsGovernanceTransactor{contract: contract}, AvsGovernanceFilterer: AvsGovernanceFilterer{contract: contract}}, nil
}

// NewAvsGovernanceCaller creates a new read-only instance of AvsGovernance, bound to a specific deployed contract.
func NewAvsGovernanceCaller(address common.Address, caller bind.ContractCaller) (*AvsGovernanceCaller, error) {
	contract, err := bindAvsGovernance(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceCaller{contract: contract}, nil
}

// NewAvsGovernanceTransactor creates a new write-only instance of AvsGovernance, bound to a specific deployed contract.
func NewAvsGovernanceTransactor(address common.Address, transactor bind.ContractTransactor) (*AvsGovernanceTransactor, error) {
	contract, err := bindAvsGovernance(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceTransactor{contract: contract}, nil
}

// NewAvsGovernanceFilterer creates a new log filterer instance of AvsGovernance, bound to a specific deployed contract.
func NewAvsGovernanceFilterer(address common.Address, filterer bind.ContractFilterer) (*AvsGovernanceFilterer, error) {
	contract, err := bindAvsGovernance(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceFilterer{contract: contract}, nil
}

// bindAvsGovernance binds a generic wrapper to an already deployed contract.
func bindAvsGovernance(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AvsGovernanceMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AvsGovernance *AvsGovernanceRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AvsGovernance.Contract.AvsGovernanceCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AvsGovernance *AvsGovernanceRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AvsGovernance.Contract.AvsGovernanceTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AvsGovernance *AvsGovernanceRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AvsGovernance.Contract.AvsGovernanceTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AvsGovernance *AvsGovernanceCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AvsGovernance.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AvsGovernance *AvsGovernanceTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AvsGovernance.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AvsGovernance *AvsGovernanceTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AvsGovernance.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AvsGovernance *AvsGovernanceCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AvsGovernance *AvsGovernanceSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AvsGovernance.Contract.DEFAULTADMINROLE(&_AvsGovernance.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AvsGovernance *AvsGovernanceCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AvsGovernance.Contract.DEFAULTADMINROLE(&_AvsGovernance.CallOpts)
}

// AvsDirectory is a free data retrieval call binding the contract method 0x6b3aa72e.
//
// Solidity: function avsDirectory() view returns(address)
func (_AvsGovernance *AvsGovernanceCaller) AvsDirectory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "avsDirectory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AvsDirectory is a free data retrieval call binding the contract method 0x6b3aa72e.
//
// Solidity: function avsDirectory() view returns(address)
func (_AvsGovernance *AvsGovernanceSession) AvsDirectory() (common.Address, error) {
	return _AvsGovernance.Contract.AvsDirectory(&_AvsGovernance.CallOpts)
}

// AvsDirectory is a free data retrieval call binding the contract method 0x6b3aa72e.
//
// Solidity: function avsDirectory() view returns(address)
func (_AvsGovernance *AvsGovernanceCallerSession) AvsDirectory() (common.Address, error) {
	return _AvsGovernance.Contract.AvsDirectory(&_AvsGovernance.CallOpts)
}

// AvsName is a free data retrieval call binding the contract method 0x41b92a29.
//
// Solidity: function avsName() view returns(string)
func (_AvsGovernance *AvsGovernanceCaller) AvsName(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "avsName")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// AvsName is a free data retrieval call binding the contract method 0x41b92a29.
//
// Solidity: function avsName() view returns(string)
func (_AvsGovernance *AvsGovernanceSession) AvsName() (string, error) {
	return _AvsGovernance.Contract.AvsName(&_AvsGovernance.CallOpts)
}

// AvsName is a free data retrieval call binding the contract method 0x41b92a29.
//
// Solidity: function avsName() view returns(string)
func (_AvsGovernance *AvsGovernanceCallerSession) AvsName() (string, error) {
	return _AvsGovernance.Contract.AvsName(&_AvsGovernance.CallOpts)
}

// GetDefaultStrategies is a free data retrieval call binding the contract method 0xe86685d9.
//
// Solidity: function getDefaultStrategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceCaller) GetDefaultStrategies(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "getDefaultStrategies")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetDefaultStrategies is a free data retrieval call binding the contract method 0xe86685d9.
//
// Solidity: function getDefaultStrategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceSession) GetDefaultStrategies() ([]common.Address, error) {
	return _AvsGovernance.Contract.GetDefaultStrategies(&_AvsGovernance.CallOpts)
}

// GetDefaultStrategies is a free data retrieval call binding the contract method 0xe86685d9.
//
// Solidity: function getDefaultStrategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceCallerSession) GetDefaultStrategies() ([]common.Address, error) {
	return _AvsGovernance.Contract.GetDefaultStrategies(&_AvsGovernance.CallOpts)
}

// GetIsAllowlisted is a free data retrieval call binding the contract method 0xb525fa88.
//
// Solidity: function getIsAllowlisted() view returns(bool)
func (_AvsGovernance *AvsGovernanceCaller) GetIsAllowlisted(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "getIsAllowlisted")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetIsAllowlisted is a free data retrieval call binding the contract method 0xb525fa88.
//
// Solidity: function getIsAllowlisted() view returns(bool)
func (_AvsGovernance *AvsGovernanceSession) GetIsAllowlisted() (bool, error) {
	return _AvsGovernance.Contract.GetIsAllowlisted(&_AvsGovernance.CallOpts)
}

// GetIsAllowlisted is a free data retrieval call binding the contract method 0xb525fa88.
//
// Solidity: function getIsAllowlisted() view returns(bool)
func (_AvsGovernance *AvsGovernanceCallerSession) GetIsAllowlisted() (bool, error) {
	return _AvsGovernance.Contract.GetIsAllowlisted(&_AvsGovernance.CallOpts)
}

// GetNumOfOperatorsLimit is a free data retrieval call binding the contract method 0xf251c9a6.
//
// Solidity: function getNumOfOperatorsLimit() view returns(uint256 numOfOperatorsLimitView)
func (_AvsGovernance *AvsGovernanceCaller) GetNumOfOperatorsLimit(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "getNumOfOperatorsLimit")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNumOfOperatorsLimit is a free data retrieval call binding the contract method 0xf251c9a6.
//
// Solidity: function getNumOfOperatorsLimit() view returns(uint256 numOfOperatorsLimitView)
func (_AvsGovernance *AvsGovernanceSession) GetNumOfOperatorsLimit() (*big.Int, error) {
	return _AvsGovernance.Contract.GetNumOfOperatorsLimit(&_AvsGovernance.CallOpts)
}

// GetNumOfOperatorsLimit is a free data retrieval call binding the contract method 0xf251c9a6.
//
// Solidity: function getNumOfOperatorsLimit() view returns(uint256 numOfOperatorsLimitView)
func (_AvsGovernance *AvsGovernanceCallerSession) GetNumOfOperatorsLimit() (*big.Int, error) {
	return _AvsGovernance.Contract.GetNumOfOperatorsLimit(&_AvsGovernance.CallOpts)
}

// GetOperatorRestakedStrategies is a free data retrieval call binding the contract method 0x33cfb7b7.
//
// Solidity: function getOperatorRestakedStrategies(address _operator) view returns(address[])
func (_AvsGovernance *AvsGovernanceCaller) GetOperatorRestakedStrategies(opts *bind.CallOpts, _operator common.Address) ([]common.Address, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "getOperatorRestakedStrategies", _operator)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOperatorRestakedStrategies is a free data retrieval call binding the contract method 0x33cfb7b7.
//
// Solidity: function getOperatorRestakedStrategies(address _operator) view returns(address[])
func (_AvsGovernance *AvsGovernanceSession) GetOperatorRestakedStrategies(_operator common.Address) ([]common.Address, error) {
	return _AvsGovernance.Contract.GetOperatorRestakedStrategies(&_AvsGovernance.CallOpts, _operator)
}

// GetOperatorRestakedStrategies is a free data retrieval call binding the contract method 0x33cfb7b7.
//
// Solidity: function getOperatorRestakedStrategies(address _operator) view returns(address[])
func (_AvsGovernance *AvsGovernanceCallerSession) GetOperatorRestakedStrategies(_operator common.Address) ([]common.Address, error) {
	return _AvsGovernance.Contract.GetOperatorRestakedStrategies(&_AvsGovernance.CallOpts, _operator)
}

// GetRestakeableStrategies is a free data retrieval call binding the contract method 0xe481af9d.
//
// Solidity: function getRestakeableStrategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceCaller) GetRestakeableStrategies(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "getRestakeableStrategies")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetRestakeableStrategies is a free data retrieval call binding the contract method 0xe481af9d.
//
// Solidity: function getRestakeableStrategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceSession) GetRestakeableStrategies() ([]common.Address, error) {
	return _AvsGovernance.Contract.GetRestakeableStrategies(&_AvsGovernance.CallOpts)
}

// GetRestakeableStrategies is a free data retrieval call binding the contract method 0xe481af9d.
//
// Solidity: function getRestakeableStrategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceCallerSession) GetRestakeableStrategies() ([]common.Address, error) {
	return _AvsGovernance.Contract.GetRestakeableStrategies(&_AvsGovernance.CallOpts)
}

// GetRewardsReceiver is a free data retrieval call binding the contract method 0x5e95cee2.
//
// Solidity: function getRewardsReceiver(address _operator) view returns(address)
func (_AvsGovernance *AvsGovernanceCaller) GetRewardsReceiver(opts *bind.CallOpts, _operator common.Address) (common.Address, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "getRewardsReceiver", _operator)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRewardsReceiver is a free data retrieval call binding the contract method 0x5e95cee2.
//
// Solidity: function getRewardsReceiver(address _operator) view returns(address)
func (_AvsGovernance *AvsGovernanceSession) GetRewardsReceiver(_operator common.Address) (common.Address, error) {
	return _AvsGovernance.Contract.GetRewardsReceiver(&_AvsGovernance.CallOpts, _operator)
}

// GetRewardsReceiver is a free data retrieval call binding the contract method 0x5e95cee2.
//
// Solidity: function getRewardsReceiver(address _operator) view returns(address)
func (_AvsGovernance *AvsGovernanceCallerSession) GetRewardsReceiver(_operator common.Address) (common.Address, error) {
	return _AvsGovernance.Contract.GetRewardsReceiver(&_AvsGovernance.CallOpts, _operator)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AvsGovernance *AvsGovernanceCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AvsGovernance *AvsGovernanceSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AvsGovernance.Contract.GetRoleAdmin(&_AvsGovernance.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AvsGovernance *AvsGovernanceCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AvsGovernance.Contract.GetRoleAdmin(&_AvsGovernance.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AvsGovernance *AvsGovernanceCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AvsGovernance *AvsGovernanceSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AvsGovernance.Contract.HasRole(&_AvsGovernance.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AvsGovernance *AvsGovernanceCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AvsGovernance.Contract.HasRole(&_AvsGovernance.CallOpts, role, account)
}

// IsFlowPaused is a free data retrieval call binding the contract method 0xefd96978.
//
// Solidity: function isFlowPaused(bytes4 _pausableFlow) view returns(bool _isPaused)
func (_AvsGovernance *AvsGovernanceCaller) IsFlowPaused(opts *bind.CallOpts, _pausableFlow [4]byte) (bool, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "isFlowPaused", _pausableFlow)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsFlowPaused is a free data retrieval call binding the contract method 0xefd96978.
//
// Solidity: function isFlowPaused(bytes4 _pausableFlow) view returns(bool _isPaused)
func (_AvsGovernance *AvsGovernanceSession) IsFlowPaused(_pausableFlow [4]byte) (bool, error) {
	return _AvsGovernance.Contract.IsFlowPaused(&_AvsGovernance.CallOpts, _pausableFlow)
}

// IsFlowPaused is a free data retrieval call binding the contract method 0xefd96978.
//
// Solidity: function isFlowPaused(bytes4 _pausableFlow) view returns(bool _isPaused)
func (_AvsGovernance *AvsGovernanceCallerSession) IsFlowPaused(_pausableFlow [4]byte) (bool, error) {
	return _AvsGovernance.Contract.IsFlowPaused(&_AvsGovernance.CallOpts, _pausableFlow)
}

// IsOperatorRegistered is a free data retrieval call binding the contract method 0x6b1906f8.
//
// Solidity: function isOperatorRegistered(address operator) view returns(bool)
func (_AvsGovernance *AvsGovernanceCaller) IsOperatorRegistered(opts *bind.CallOpts, operator common.Address) (bool, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "isOperatorRegistered", operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperatorRegistered is a free data retrieval call binding the contract method 0x6b1906f8.
//
// Solidity: function isOperatorRegistered(address operator) view returns(bool)
func (_AvsGovernance *AvsGovernanceSession) IsOperatorRegistered(operator common.Address) (bool, error) {
	return _AvsGovernance.Contract.IsOperatorRegistered(&_AvsGovernance.CallOpts, operator)
}

// IsOperatorRegistered is a free data retrieval call binding the contract method 0x6b1906f8.
//
// Solidity: function isOperatorRegistered(address operator) view returns(bool)
func (_AvsGovernance *AvsGovernanceCallerSession) IsOperatorRegistered(operator common.Address) (bool, error) {
	return _AvsGovernance.Contract.IsOperatorRegistered(&_AvsGovernance.CallOpts, operator)
}

// MaxEffectiveBalance is a free data retrieval call binding the contract method 0xa88171ee.
//
// Solidity: function maxEffectiveBalance() view returns(uint256)
func (_AvsGovernance *AvsGovernanceCaller) MaxEffectiveBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "maxEffectiveBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxEffectiveBalance is a free data retrieval call binding the contract method 0xa88171ee.
//
// Solidity: function maxEffectiveBalance() view returns(uint256)
func (_AvsGovernance *AvsGovernanceSession) MaxEffectiveBalance() (*big.Int, error) {
	return _AvsGovernance.Contract.MaxEffectiveBalance(&_AvsGovernance.CallOpts)
}

// MaxEffectiveBalance is a free data retrieval call binding the contract method 0xa88171ee.
//
// Solidity: function maxEffectiveBalance() view returns(uint256)
func (_AvsGovernance *AvsGovernanceCallerSession) MaxEffectiveBalance() (*big.Int, error) {
	return _AvsGovernance.Contract.MaxEffectiveBalance(&_AvsGovernance.CallOpts)
}

// MinSharesForStrategy is a free data retrieval call binding the contract method 0xc3814e5b.
//
// Solidity: function minSharesForStrategy(address _strategy) view returns(uint256)
func (_AvsGovernance *AvsGovernanceCaller) MinSharesForStrategy(opts *bind.CallOpts, _strategy common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "minSharesForStrategy", _strategy)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinSharesForStrategy is a free data retrieval call binding the contract method 0xc3814e5b.
//
// Solidity: function minSharesForStrategy(address _strategy) view returns(uint256)
func (_AvsGovernance *AvsGovernanceSession) MinSharesForStrategy(_strategy common.Address) (*big.Int, error) {
	return _AvsGovernance.Contract.MinSharesForStrategy(&_AvsGovernance.CallOpts, _strategy)
}

// MinSharesForStrategy is a free data retrieval call binding the contract method 0xc3814e5b.
//
// Solidity: function minSharesForStrategy(address _strategy) view returns(uint256)
func (_AvsGovernance *AvsGovernanceCallerSession) MinSharesForStrategy(_strategy common.Address) (*big.Int, error) {
	return _AvsGovernance.Contract.MinSharesForStrategy(&_AvsGovernance.CallOpts, _strategy)
}

// MinVotingPower is a free data retrieval call binding the contract method 0x36fffde0.
//
// Solidity: function minVotingPower() view returns(uint256)
func (_AvsGovernance *AvsGovernanceCaller) MinVotingPower(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "minVotingPower")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinVotingPower is a free data retrieval call binding the contract method 0x36fffde0.
//
// Solidity: function minVotingPower() view returns(uint256)
func (_AvsGovernance *AvsGovernanceSession) MinVotingPower() (*big.Int, error) {
	return _AvsGovernance.Contract.MinVotingPower(&_AvsGovernance.CallOpts)
}

// MinVotingPower is a free data retrieval call binding the contract method 0x36fffde0.
//
// Solidity: function minVotingPower() view returns(uint256)
func (_AvsGovernance *AvsGovernanceCallerSession) MinVotingPower() (*big.Int, error) {
	return _AvsGovernance.Contract.MinVotingPower(&_AvsGovernance.CallOpts)
}

// NumOfActiveOperators is a free data retrieval call binding the contract method 0x7897dec3.
//
// Solidity: function numOfActiveOperators() view returns(uint256)
func (_AvsGovernance *AvsGovernanceCaller) NumOfActiveOperators(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "numOfActiveOperators")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumOfActiveOperators is a free data retrieval call binding the contract method 0x7897dec3.
//
// Solidity: function numOfActiveOperators() view returns(uint256)
func (_AvsGovernance *AvsGovernanceSession) NumOfActiveOperators() (*big.Int, error) {
	return _AvsGovernance.Contract.NumOfActiveOperators(&_AvsGovernance.CallOpts)
}

// NumOfActiveOperators is a free data retrieval call binding the contract method 0x7897dec3.
//
// Solidity: function numOfActiveOperators() view returns(uint256)
func (_AvsGovernance *AvsGovernanceCallerSession) NumOfActiveOperators() (*big.Int, error) {
	return _AvsGovernance.Contract.NumOfActiveOperators(&_AvsGovernance.CallOpts)
}

// NumOfOperators is a free data retrieval call binding the contract method 0x6ade02da.
//
// Solidity: function numOfOperators() view returns(uint256)
func (_AvsGovernance *AvsGovernanceCaller) NumOfOperators(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "numOfOperators")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumOfOperators is a free data retrieval call binding the contract method 0x6ade02da.
//
// Solidity: function numOfOperators() view returns(uint256)
func (_AvsGovernance *AvsGovernanceSession) NumOfOperators() (*big.Int, error) {
	return _AvsGovernance.Contract.NumOfOperators(&_AvsGovernance.CallOpts)
}

// NumOfOperators is a free data retrieval call binding the contract method 0x6ade02da.
//
// Solidity: function numOfOperators() view returns(uint256)
func (_AvsGovernance *AvsGovernanceCallerSession) NumOfOperators() (*big.Int, error) {
	return _AvsGovernance.Contract.NumOfOperators(&_AvsGovernance.CallOpts)
}

// NumOfShares is a free data retrieval call binding the contract method 0x6a907803.
//
// Solidity: function numOfShares(address _operator) view returns(uint256)
func (_AvsGovernance *AvsGovernanceCaller) NumOfShares(opts *bind.CallOpts, _operator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "numOfShares", _operator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumOfShares is a free data retrieval call binding the contract method 0x6a907803.
//
// Solidity: function numOfShares(address _operator) view returns(uint256)
func (_AvsGovernance *AvsGovernanceSession) NumOfShares(_operator common.Address) (*big.Int, error) {
	return _AvsGovernance.Contract.NumOfShares(&_AvsGovernance.CallOpts, _operator)
}

// NumOfShares is a free data retrieval call binding the contract method 0x6a907803.
//
// Solidity: function numOfShares(address _operator) view returns(uint256)
func (_AvsGovernance *AvsGovernanceCallerSession) NumOfShares(_operator common.Address) (*big.Int, error) {
	return _AvsGovernance.Contract.NumOfShares(&_AvsGovernance.CallOpts, _operator)
}

// Strategies is a free data retrieval call binding the contract method 0xd9f9027f.
//
// Solidity: function strategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceCaller) Strategies(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "strategies")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// Strategies is a free data retrieval call binding the contract method 0xd9f9027f.
//
// Solidity: function strategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceSession) Strategies() ([]common.Address, error) {
	return _AvsGovernance.Contract.Strategies(&_AvsGovernance.CallOpts)
}

// Strategies is a free data retrieval call binding the contract method 0xd9f9027f.
//
// Solidity: function strategies() view returns(address[])
func (_AvsGovernance *AvsGovernanceCallerSession) Strategies() ([]common.Address, error) {
	return _AvsGovernance.Contract.Strategies(&_AvsGovernance.CallOpts)
}

// StrategyMultiplier is a free data retrieval call binding the contract method 0x8f53bc50.
//
// Solidity: function strategyMultiplier(address _strategy) view returns(uint256)
func (_AvsGovernance *AvsGovernanceCaller) StrategyMultiplier(opts *bind.CallOpts, _strategy common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "strategyMultiplier", _strategy)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StrategyMultiplier is a free data retrieval call binding the contract method 0x8f53bc50.
//
// Solidity: function strategyMultiplier(address _strategy) view returns(uint256)
func (_AvsGovernance *AvsGovernanceSession) StrategyMultiplier(_strategy common.Address) (*big.Int, error) {
	return _AvsGovernance.Contract.StrategyMultiplier(&_AvsGovernance.CallOpts, _strategy)
}

// StrategyMultiplier is a free data retrieval call binding the contract method 0x8f53bc50.
//
// Solidity: function strategyMultiplier(address _strategy) view returns(uint256)
func (_AvsGovernance *AvsGovernanceCallerSession) StrategyMultiplier(_strategy common.Address) (*big.Int, error) {
	return _AvsGovernance.Contract.StrategyMultiplier(&_AvsGovernance.CallOpts, _strategy)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AvsGovernance *AvsGovernanceCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AvsGovernance *AvsGovernanceSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AvsGovernance.Contract.SupportsInterface(&_AvsGovernance.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AvsGovernance *AvsGovernanceCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AvsGovernance.Contract.SupportsInterface(&_AvsGovernance.CallOpts, interfaceId)
}

// Vault is a free data retrieval call binding the contract method 0xfbfa77cf.
//
// Solidity: function vault() view returns(address)
func (_AvsGovernance *AvsGovernanceCaller) Vault(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "vault")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Vault is a free data retrieval call binding the contract method 0xfbfa77cf.
//
// Solidity: function vault() view returns(address)
func (_AvsGovernance *AvsGovernanceSession) Vault() (common.Address, error) {
	return _AvsGovernance.Contract.Vault(&_AvsGovernance.CallOpts)
}

// Vault is a free data retrieval call binding the contract method 0xfbfa77cf.
//
// Solidity: function vault() view returns(address)
func (_AvsGovernance *AvsGovernanceCallerSession) Vault() (common.Address, error) {
	return _AvsGovernance.Contract.Vault(&_AvsGovernance.CallOpts)
}

// VotingPower is a free data retrieval call binding the contract method 0xc07473f6.
//
// Solidity: function votingPower(address _operator) view returns(uint256)
func (_AvsGovernance *AvsGovernanceCaller) VotingPower(opts *bind.CallOpts, _operator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AvsGovernance.contract.Call(opts, &out, "votingPower", _operator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VotingPower is a free data retrieval call binding the contract method 0xc07473f6.
//
// Solidity: function votingPower(address _operator) view returns(uint256)
func (_AvsGovernance *AvsGovernanceSession) VotingPower(_operator common.Address) (*big.Int, error) {
	return _AvsGovernance.Contract.VotingPower(&_AvsGovernance.CallOpts, _operator)
}

// VotingPower is a free data retrieval call binding the contract method 0xc07473f6.
//
// Solidity: function votingPower(address _operator) view returns(uint256)
func (_AvsGovernance *AvsGovernanceCallerSession) VotingPower(_operator common.Address) (*big.Int, error) {
	return _AvsGovernance.Contract.VotingPower(&_AvsGovernance.CallOpts, _operator)
}

// CompleteRewardsReceiverModification is a paid mutator transaction binding the contract method 0xe6474b0f.
//
// Solidity: function completeRewardsReceiverModification() returns()
func (_AvsGovernance *AvsGovernanceTransactor) CompleteRewardsReceiverModification(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "completeRewardsReceiverModification")
}

// CompleteRewardsReceiverModification is a paid mutator transaction binding the contract method 0xe6474b0f.
//
// Solidity: function completeRewardsReceiverModification() returns()
func (_AvsGovernance *AvsGovernanceSession) CompleteRewardsReceiverModification() (*types.Transaction, error) {
	return _AvsGovernance.Contract.CompleteRewardsReceiverModification(&_AvsGovernance.TransactOpts)
}

// CompleteRewardsReceiverModification is a paid mutator transaction binding the contract method 0xe6474b0f.
//
// Solidity: function completeRewardsReceiverModification() returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) CompleteRewardsReceiverModification() (*types.Transaction, error) {
	return _AvsGovernance.Contract.CompleteRewardsReceiverModification(&_AvsGovernance.TransactOpts)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0xb79092fd.
//
// Solidity: function depositERC20(uint256 _amount) returns()
func (_AvsGovernance *AvsGovernanceTransactor) DepositERC20(opts *bind.TransactOpts, _amount *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "depositERC20", _amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0xb79092fd.
//
// Solidity: function depositERC20(uint256 _amount) returns()
func (_AvsGovernance *AvsGovernanceSession) DepositERC20(_amount *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.DepositERC20(&_AvsGovernance.TransactOpts, _amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0xb79092fd.
//
// Solidity: function depositERC20(uint256 _amount) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) DepositERC20(_amount *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.DepositERC20(&_AvsGovernance.TransactOpts, _amount)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AvsGovernance *AvsGovernanceTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AvsGovernance *AvsGovernanceSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.GrantRole(&_AvsGovernance.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.GrantRole(&_AvsGovernance.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0xfab57b8f.
//
// Solidity: function initialize((address,address,address,address,address,address,address,address,string,address) _initializationParams) returns()
func (_AvsGovernance *AvsGovernanceTransactor) Initialize(opts *bind.TransactOpts, _initializationParams IAvsGovernanceInitializationParams) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "initialize", _initializationParams)
}

// Initialize is a paid mutator transaction binding the contract method 0xfab57b8f.
//
// Solidity: function initialize((address,address,address,address,address,address,address,address,string,address) _initializationParams) returns()
func (_AvsGovernance *AvsGovernanceSession) Initialize(_initializationParams IAvsGovernanceInitializationParams) (*types.Transaction, error) {
	return _AvsGovernance.Contract.Initialize(&_AvsGovernance.TransactOpts, _initializationParams)
}

// Initialize is a paid mutator transaction binding the contract method 0xfab57b8f.
//
// Solidity: function initialize((address,address,address,address,address,address,address,address,string,address) _initializationParams) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) Initialize(_initializationParams IAvsGovernanceInitializationParams) (*types.Transaction, error) {
	return _AvsGovernance.Contract.Initialize(&_AvsGovernance.TransactOpts, _initializationParams)
}

// Pause is a paid mutator transaction binding the contract method 0x3aa83ec7.
//
// Solidity: function pause(bytes4 _pausableFlow) returns()
func (_AvsGovernance *AvsGovernanceTransactor) Pause(opts *bind.TransactOpts, _pausableFlow [4]byte) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "pause", _pausableFlow)
}

// Pause is a paid mutator transaction binding the contract method 0x3aa83ec7.
//
// Solidity: function pause(bytes4 _pausableFlow) returns()
func (_AvsGovernance *AvsGovernanceSession) Pause(_pausableFlow [4]byte) (*types.Transaction, error) {
	return _AvsGovernance.Contract.Pause(&_AvsGovernance.TransactOpts, _pausableFlow)
}

// Pause is a paid mutator transaction binding the contract method 0x3aa83ec7.
//
// Solidity: function pause(bytes4 _pausableFlow) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) Pause(_pausableFlow [4]byte) (*types.Transaction, error) {
	return _AvsGovernance.Contract.Pause(&_AvsGovernance.TransactOpts, _pausableFlow)
}

// QueueRewardsReceiverModification is a paid mutator transaction binding the contract method 0x1b21ba72.
//
// Solidity: function queueRewardsReceiverModification(address _newRewardsReceiver) returns()
func (_AvsGovernance *AvsGovernanceTransactor) QueueRewardsReceiverModification(opts *bind.TransactOpts, _newRewardsReceiver common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "queueRewardsReceiverModification", _newRewardsReceiver)
}

// QueueRewardsReceiverModification is a paid mutator transaction binding the contract method 0x1b21ba72.
//
// Solidity: function queueRewardsReceiverModification(address _newRewardsReceiver) returns()
func (_AvsGovernance *AvsGovernanceSession) QueueRewardsReceiverModification(_newRewardsReceiver common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.QueueRewardsReceiverModification(&_AvsGovernance.TransactOpts, _newRewardsReceiver)
}

// QueueRewardsReceiverModification is a paid mutator transaction binding the contract method 0x1b21ba72.
//
// Solidity: function queueRewardsReceiverModification(address _newRewardsReceiver) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) QueueRewardsReceiverModification(_newRewardsReceiver common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.QueueRewardsReceiverModification(&_AvsGovernance.TransactOpts, _newRewardsReceiver)
}

// RegisterAsAllowedOperator is a paid mutator transaction binding the contract method 0x93304a9d.
//
// Solidity: function registerAsAllowedOperator(uint256[4] _blsKey, bytes _authToken, address _rewardsReceiver, (bytes,bytes32,uint256) _operatorSignature, (uint256[2]) _blsRegistrationSignature) returns()
func (_AvsGovernance *AvsGovernanceTransactor) RegisterAsAllowedOperator(opts *bind.TransactOpts, _blsKey [4]*big.Int, _authToken []byte, _rewardsReceiver common.Address, _operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _blsRegistrationSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "registerAsAllowedOperator", _blsKey, _authToken, _rewardsReceiver, _operatorSignature, _blsRegistrationSignature)
}

// RegisterAsAllowedOperator is a paid mutator transaction binding the contract method 0x93304a9d.
//
// Solidity: function registerAsAllowedOperator(uint256[4] _blsKey, bytes _authToken, address _rewardsReceiver, (bytes,bytes32,uint256) _operatorSignature, (uint256[2]) _blsRegistrationSignature) returns()
func (_AvsGovernance *AvsGovernanceSession) RegisterAsAllowedOperator(_blsKey [4]*big.Int, _authToken []byte, _rewardsReceiver common.Address, _operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _blsRegistrationSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AvsGovernance.Contract.RegisterAsAllowedOperator(&_AvsGovernance.TransactOpts, _blsKey, _authToken, _rewardsReceiver, _operatorSignature, _blsRegistrationSignature)
}

// RegisterAsAllowedOperator is a paid mutator transaction binding the contract method 0x93304a9d.
//
// Solidity: function registerAsAllowedOperator(uint256[4] _blsKey, bytes _authToken, address _rewardsReceiver, (bytes,bytes32,uint256) _operatorSignature, (uint256[2]) _blsRegistrationSignature) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) RegisterAsAllowedOperator(_blsKey [4]*big.Int, _authToken []byte, _rewardsReceiver common.Address, _operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _blsRegistrationSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AvsGovernance.Contract.RegisterAsAllowedOperator(&_AvsGovernance.TransactOpts, _blsKey, _authToken, _rewardsReceiver, _operatorSignature, _blsRegistrationSignature)
}

// RegisterAsOperator is a paid mutator transaction binding the contract method 0x22609a4d.
//
// Solidity: function registerAsOperator(uint256[4] _blsKey, address _rewardsReceiver, (bytes,bytes32,uint256) _operatorSignature, (uint256[2]) _blsRegistrationSignature) returns()
func (_AvsGovernance *AvsGovernanceTransactor) RegisterAsOperator(opts *bind.TransactOpts, _blsKey [4]*big.Int, _rewardsReceiver common.Address, _operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _blsRegistrationSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "registerAsOperator", _blsKey, _rewardsReceiver, _operatorSignature, _blsRegistrationSignature)
}

// RegisterAsOperator is a paid mutator transaction binding the contract method 0x22609a4d.
//
// Solidity: function registerAsOperator(uint256[4] _blsKey, address _rewardsReceiver, (bytes,bytes32,uint256) _operatorSignature, (uint256[2]) _blsRegistrationSignature) returns()
func (_AvsGovernance *AvsGovernanceSession) RegisterAsOperator(_blsKey [4]*big.Int, _rewardsReceiver common.Address, _operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _blsRegistrationSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AvsGovernance.Contract.RegisterAsOperator(&_AvsGovernance.TransactOpts, _blsKey, _rewardsReceiver, _operatorSignature, _blsRegistrationSignature)
}

// RegisterAsOperator is a paid mutator transaction binding the contract method 0x22609a4d.
//
// Solidity: function registerAsOperator(uint256[4] _blsKey, address _rewardsReceiver, (bytes,bytes32,uint256) _operatorSignature, (uint256[2]) _blsRegistrationSignature) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) RegisterAsOperator(_blsKey [4]*big.Int, _rewardsReceiver common.Address, _operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _blsRegistrationSignature BLSAuthLibrarySignature) (*types.Transaction, error) {
	return _AvsGovernance.Contract.RegisterAsOperator(&_AvsGovernance.TransactOpts, _blsKey, _rewardsReceiver, _operatorSignature, _blsRegistrationSignature)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AvsGovernance *AvsGovernanceTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AvsGovernance *AvsGovernanceSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.RenounceRole(&_AvsGovernance.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.RenounceRole(&_AvsGovernance.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AvsGovernance *AvsGovernanceTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AvsGovernance *AvsGovernanceSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.RevokeRole(&_AvsGovernance.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.RevokeRole(&_AvsGovernance.TransactOpts, role, account)
}

// SetAllowlistSigner is a paid mutator transaction binding the contract method 0xe474def4.
//
// Solidity: function setAllowlistSigner(address _allowlistSigner) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetAllowlistSigner(opts *bind.TransactOpts, _allowlistSigner common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setAllowlistSigner", _allowlistSigner)
}

// SetAllowlistSigner is a paid mutator transaction binding the contract method 0xe474def4.
//
// Solidity: function setAllowlistSigner(address _allowlistSigner) returns()
func (_AvsGovernance *AvsGovernanceSession) SetAllowlistSigner(_allowlistSigner common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetAllowlistSigner(&_AvsGovernance.TransactOpts, _allowlistSigner)
}

// SetAllowlistSigner is a paid mutator transaction binding the contract method 0xe474def4.
//
// Solidity: function setAllowlistSigner(address _allowlistSigner) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetAllowlistSigner(_allowlistSigner common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetAllowlistSigner(&_AvsGovernance.TransactOpts, _allowlistSigner)
}

// SetAvsGovernanceLogic is a paid mutator transaction binding the contract method 0x8987c767.
//
// Solidity: function setAvsGovernanceLogic(address _avsGovernanceLogic) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetAvsGovernanceLogic(opts *bind.TransactOpts, _avsGovernanceLogic common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setAvsGovernanceLogic", _avsGovernanceLogic)
}

// SetAvsGovernanceLogic is a paid mutator transaction binding the contract method 0x8987c767.
//
// Solidity: function setAvsGovernanceLogic(address _avsGovernanceLogic) returns()
func (_AvsGovernance *AvsGovernanceSession) SetAvsGovernanceLogic(_avsGovernanceLogic common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetAvsGovernanceLogic(&_AvsGovernance.TransactOpts, _avsGovernanceLogic)
}

// SetAvsGovernanceLogic is a paid mutator transaction binding the contract method 0x8987c767.
//
// Solidity: function setAvsGovernanceLogic(address _avsGovernanceLogic) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetAvsGovernanceLogic(_avsGovernanceLogic common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetAvsGovernanceLogic(&_AvsGovernance.TransactOpts, _avsGovernanceLogic)
}

// SetAvsGovernanceMultiplierSyncer is a paid mutator transaction binding the contract method 0x3425e8d8.
//
// Solidity: function setAvsGovernanceMultiplierSyncer(address _newAvsGovernanceMultiplierSyncer) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetAvsGovernanceMultiplierSyncer(opts *bind.TransactOpts, _newAvsGovernanceMultiplierSyncer common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setAvsGovernanceMultiplierSyncer", _newAvsGovernanceMultiplierSyncer)
}

// SetAvsGovernanceMultiplierSyncer is a paid mutator transaction binding the contract method 0x3425e8d8.
//
// Solidity: function setAvsGovernanceMultiplierSyncer(address _newAvsGovernanceMultiplierSyncer) returns()
func (_AvsGovernance *AvsGovernanceSession) SetAvsGovernanceMultiplierSyncer(_newAvsGovernanceMultiplierSyncer common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetAvsGovernanceMultiplierSyncer(&_AvsGovernance.TransactOpts, _newAvsGovernanceMultiplierSyncer)
}

// SetAvsGovernanceMultiplierSyncer is a paid mutator transaction binding the contract method 0x3425e8d8.
//
// Solidity: function setAvsGovernanceMultiplierSyncer(address _newAvsGovernanceMultiplierSyncer) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetAvsGovernanceMultiplierSyncer(_newAvsGovernanceMultiplierSyncer common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetAvsGovernanceMultiplierSyncer(&_AvsGovernance.TransactOpts, _newAvsGovernanceMultiplierSyncer)
}

// SetAvsName is a paid mutator transaction binding the contract method 0x7d38e926.
//
// Solidity: function setAvsName(string _avsName) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetAvsName(opts *bind.TransactOpts, _avsName string) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setAvsName", _avsName)
}

// SetAvsName is a paid mutator transaction binding the contract method 0x7d38e926.
//
// Solidity: function setAvsName(string _avsName) returns()
func (_AvsGovernance *AvsGovernanceSession) SetAvsName(_avsName string) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetAvsName(&_AvsGovernance.TransactOpts, _avsName)
}

// SetAvsName is a paid mutator transaction binding the contract method 0x7d38e926.
//
// Solidity: function setAvsName(string _avsName) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetAvsName(_avsName string) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetAvsName(&_AvsGovernance.TransactOpts, _avsName)
}

// SetBLSAuthSingleton is a paid mutator transaction binding the contract method 0x4ef1476e.
//
// Solidity: function setBLSAuthSingleton(address _blsAuthSingleton) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetBLSAuthSingleton(opts *bind.TransactOpts, _blsAuthSingleton common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setBLSAuthSingleton", _blsAuthSingleton)
}

// SetBLSAuthSingleton is a paid mutator transaction binding the contract method 0x4ef1476e.
//
// Solidity: function setBLSAuthSingleton(address _blsAuthSingleton) returns()
func (_AvsGovernance *AvsGovernanceSession) SetBLSAuthSingleton(_blsAuthSingleton common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetBLSAuthSingleton(&_AvsGovernance.TransactOpts, _blsAuthSingleton)
}

// SetBLSAuthSingleton is a paid mutator transaction binding the contract method 0x4ef1476e.
//
// Solidity: function setBLSAuthSingleton(address _blsAuthSingleton) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetBLSAuthSingleton(_blsAuthSingleton common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetBLSAuthSingleton(&_AvsGovernance.TransactOpts, _blsAuthSingleton)
}

// SetIsAllowlisted is a paid mutator transaction binding the contract method 0x9e965cc1.
//
// Solidity: function setIsAllowlisted(bool _isAllowlisted) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetIsAllowlisted(opts *bind.TransactOpts, _isAllowlisted bool) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setIsAllowlisted", _isAllowlisted)
}

// SetIsAllowlisted is a paid mutator transaction binding the contract method 0x9e965cc1.
//
// Solidity: function setIsAllowlisted(bool _isAllowlisted) returns()
func (_AvsGovernance *AvsGovernanceSession) SetIsAllowlisted(_isAllowlisted bool) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetIsAllowlisted(&_AvsGovernance.TransactOpts, _isAllowlisted)
}

// SetIsAllowlisted is a paid mutator transaction binding the contract method 0x9e965cc1.
//
// Solidity: function setIsAllowlisted(bool _isAllowlisted) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetIsAllowlisted(_isAllowlisted bool) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetIsAllowlisted(&_AvsGovernance.TransactOpts, _isAllowlisted)
}

// SetMaxEffectiveBalance is a paid mutator transaction binding the contract method 0x76086c70.
//
// Solidity: function setMaxEffectiveBalance(uint256 _maxBalance) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetMaxEffectiveBalance(opts *bind.TransactOpts, _maxBalance *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setMaxEffectiveBalance", _maxBalance)
}

// SetMaxEffectiveBalance is a paid mutator transaction binding the contract method 0x76086c70.
//
// Solidity: function setMaxEffectiveBalance(uint256 _maxBalance) returns()
func (_AvsGovernance *AvsGovernanceSession) SetMaxEffectiveBalance(_maxBalance *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetMaxEffectiveBalance(&_AvsGovernance.TransactOpts, _maxBalance)
}

// SetMaxEffectiveBalance is a paid mutator transaction binding the contract method 0x76086c70.
//
// Solidity: function setMaxEffectiveBalance(uint256 _maxBalance) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetMaxEffectiveBalance(_maxBalance *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetMaxEffectiveBalance(&_AvsGovernance.TransactOpts, _maxBalance)
}

// SetMinSharesForStrategy is a paid mutator transaction binding the contract method 0x305df58a.
//
// Solidity: function setMinSharesForStrategy(address _strategy, uint256 _minShares) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetMinSharesForStrategy(opts *bind.TransactOpts, _strategy common.Address, _minShares *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setMinSharesForStrategy", _strategy, _minShares)
}

// SetMinSharesForStrategy is a paid mutator transaction binding the contract method 0x305df58a.
//
// Solidity: function setMinSharesForStrategy(address _strategy, uint256 _minShares) returns()
func (_AvsGovernance *AvsGovernanceSession) SetMinSharesForStrategy(_strategy common.Address, _minShares *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetMinSharesForStrategy(&_AvsGovernance.TransactOpts, _strategy, _minShares)
}

// SetMinSharesForStrategy is a paid mutator transaction binding the contract method 0x305df58a.
//
// Solidity: function setMinSharesForStrategy(address _strategy, uint256 _minShares) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetMinSharesForStrategy(_strategy common.Address, _minShares *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetMinSharesForStrategy(&_AvsGovernance.TransactOpts, _strategy, _minShares)
}

// SetMinVotingPower is a paid mutator transaction binding the contract method 0x55e48918.
//
// Solidity: function setMinVotingPower(uint256 _minVotingPower) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetMinVotingPower(opts *bind.TransactOpts, _minVotingPower *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setMinVotingPower", _minVotingPower)
}

// SetMinVotingPower is a paid mutator transaction binding the contract method 0x55e48918.
//
// Solidity: function setMinVotingPower(uint256 _minVotingPower) returns()
func (_AvsGovernance *AvsGovernanceSession) SetMinVotingPower(_minVotingPower *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetMinVotingPower(&_AvsGovernance.TransactOpts, _minVotingPower)
}

// SetMinVotingPower is a paid mutator transaction binding the contract method 0x55e48918.
//
// Solidity: function setMinVotingPower(uint256 _minVotingPower) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetMinVotingPower(_minVotingPower *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetMinVotingPower(&_AvsGovernance.TransactOpts, _minVotingPower)
}

// SetNumOfOperatorsLimit is a paid mutator transaction binding the contract method 0x9d79e4a7.
//
// Solidity: function setNumOfOperatorsLimit(uint256 _newLimitOfNumOfOperators) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetNumOfOperatorsLimit(opts *bind.TransactOpts, _newLimitOfNumOfOperators *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setNumOfOperatorsLimit", _newLimitOfNumOfOperators)
}

// SetNumOfOperatorsLimit is a paid mutator transaction binding the contract method 0x9d79e4a7.
//
// Solidity: function setNumOfOperatorsLimit(uint256 _newLimitOfNumOfOperators) returns()
func (_AvsGovernance *AvsGovernanceSession) SetNumOfOperatorsLimit(_newLimitOfNumOfOperators *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetNumOfOperatorsLimit(&_AvsGovernance.TransactOpts, _newLimitOfNumOfOperators)
}

// SetNumOfOperatorsLimit is a paid mutator transaction binding the contract method 0x9d79e4a7.
//
// Solidity: function setNumOfOperatorsLimit(uint256 _newLimitOfNumOfOperators) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetNumOfOperatorsLimit(_newLimitOfNumOfOperators *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetNumOfOperatorsLimit(&_AvsGovernance.TransactOpts, _newLimitOfNumOfOperators)
}

// SetOthenticRegistry is a paid mutator transaction binding the contract method 0x45a022fa.
//
// Solidity: function setOthenticRegistry(address _othenticRegistry) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetOthenticRegistry(opts *bind.TransactOpts, _othenticRegistry common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setOthenticRegistry", _othenticRegistry)
}

// SetOthenticRegistry is a paid mutator transaction binding the contract method 0x45a022fa.
//
// Solidity: function setOthenticRegistry(address _othenticRegistry) returns()
func (_AvsGovernance *AvsGovernanceSession) SetOthenticRegistry(_othenticRegistry common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetOthenticRegistry(&_AvsGovernance.TransactOpts, _othenticRegistry)
}

// SetOthenticRegistry is a paid mutator transaction binding the contract method 0x45a022fa.
//
// Solidity: function setOthenticRegistry(address _othenticRegistry) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetOthenticRegistry(_othenticRegistry common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetOthenticRegistry(&_AvsGovernance.TransactOpts, _othenticRegistry)
}

// SetRewardsReceiverModificationDelay is a paid mutator transaction binding the contract method 0x8a70469a.
//
// Solidity: function setRewardsReceiverModificationDelay(uint256 _rewardsReceiverModificationDelay) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetRewardsReceiverModificationDelay(opts *bind.TransactOpts, _rewardsReceiverModificationDelay *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setRewardsReceiverModificationDelay", _rewardsReceiverModificationDelay)
}

// SetRewardsReceiverModificationDelay is a paid mutator transaction binding the contract method 0x8a70469a.
//
// Solidity: function setRewardsReceiverModificationDelay(uint256 _rewardsReceiverModificationDelay) returns()
func (_AvsGovernance *AvsGovernanceSession) SetRewardsReceiverModificationDelay(_rewardsReceiverModificationDelay *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetRewardsReceiverModificationDelay(&_AvsGovernance.TransactOpts, _rewardsReceiverModificationDelay)
}

// SetRewardsReceiverModificationDelay is a paid mutator transaction binding the contract method 0x8a70469a.
//
// Solidity: function setRewardsReceiverModificationDelay(uint256 _rewardsReceiverModificationDelay) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetRewardsReceiverModificationDelay(_rewardsReceiverModificationDelay *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetRewardsReceiverModificationDelay(&_AvsGovernance.TransactOpts, _rewardsReceiverModificationDelay)
}

// SetStrategyMultiplier is a paid mutator transaction binding the contract method 0x076400d5.
//
// Solidity: function setStrategyMultiplier((address,uint256) _strategyMultiplier) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetStrategyMultiplier(opts *bind.TransactOpts, _strategyMultiplier IAvsGovernanceStrategyMultiplier) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setStrategyMultiplier", _strategyMultiplier)
}

// SetStrategyMultiplier is a paid mutator transaction binding the contract method 0x076400d5.
//
// Solidity: function setStrategyMultiplier((address,uint256) _strategyMultiplier) returns()
func (_AvsGovernance *AvsGovernanceSession) SetStrategyMultiplier(_strategyMultiplier IAvsGovernanceStrategyMultiplier) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetStrategyMultiplier(&_AvsGovernance.TransactOpts, _strategyMultiplier)
}

// SetStrategyMultiplier is a paid mutator transaction binding the contract method 0x076400d5.
//
// Solidity: function setStrategyMultiplier((address,uint256) _strategyMultiplier) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetStrategyMultiplier(_strategyMultiplier IAvsGovernanceStrategyMultiplier) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetStrategyMultiplier(&_AvsGovernance.TransactOpts, _strategyMultiplier)
}

// SetStrategyMultiplierBatch is a paid mutator transaction binding the contract method 0xd94a2e1d.
//
// Solidity: function setStrategyMultiplierBatch((address,uint256)[] _strategyMultipliers) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetStrategyMultiplierBatch(opts *bind.TransactOpts, _strategyMultipliers []IAvsGovernanceStrategyMultiplier) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setStrategyMultiplierBatch", _strategyMultipliers)
}

// SetStrategyMultiplierBatch is a paid mutator transaction binding the contract method 0xd94a2e1d.
//
// Solidity: function setStrategyMultiplierBatch((address,uint256)[] _strategyMultipliers) returns()
func (_AvsGovernance *AvsGovernanceSession) SetStrategyMultiplierBatch(_strategyMultipliers []IAvsGovernanceStrategyMultiplier) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetStrategyMultiplierBatch(&_AvsGovernance.TransactOpts, _strategyMultipliers)
}

// SetStrategyMultiplierBatch is a paid mutator transaction binding the contract method 0xd94a2e1d.
//
// Solidity: function setStrategyMultiplierBatch((address,uint256)[] _strategyMultipliers) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetStrategyMultiplierBatch(_strategyMultipliers []IAvsGovernanceStrategyMultiplier) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetStrategyMultiplierBatch(&_AvsGovernance.TransactOpts, _strategyMultipliers)
}

// SetSupportedStrategies is a paid mutator transaction binding the contract method 0x312c150b.
//
// Solidity: function setSupportedStrategies(address[] _strategies) returns()
func (_AvsGovernance *AvsGovernanceTransactor) SetSupportedStrategies(opts *bind.TransactOpts, _strategies []common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "setSupportedStrategies", _strategies)
}

// SetSupportedStrategies is a paid mutator transaction binding the contract method 0x312c150b.
//
// Solidity: function setSupportedStrategies(address[] _strategies) returns()
func (_AvsGovernance *AvsGovernanceSession) SetSupportedStrategies(_strategies []common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetSupportedStrategies(&_AvsGovernance.TransactOpts, _strategies)
}

// SetSupportedStrategies is a paid mutator transaction binding the contract method 0x312c150b.
//
// Solidity: function setSupportedStrategies(address[] _strategies) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) SetSupportedStrategies(_strategies []common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.SetSupportedStrategies(&_AvsGovernance.TransactOpts, _strategies)
}

// TransferAvsGovernanceMultisig is a paid mutator transaction binding the contract method 0x513c52ba.
//
// Solidity: function transferAvsGovernanceMultisig(address _newAvsGovernanceMultisig) returns()
func (_AvsGovernance *AvsGovernanceTransactor) TransferAvsGovernanceMultisig(opts *bind.TransactOpts, _newAvsGovernanceMultisig common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "transferAvsGovernanceMultisig", _newAvsGovernanceMultisig)
}

// TransferAvsGovernanceMultisig is a paid mutator transaction binding the contract method 0x513c52ba.
//
// Solidity: function transferAvsGovernanceMultisig(address _newAvsGovernanceMultisig) returns()
func (_AvsGovernance *AvsGovernanceSession) TransferAvsGovernanceMultisig(_newAvsGovernanceMultisig common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.TransferAvsGovernanceMultisig(&_AvsGovernance.TransactOpts, _newAvsGovernanceMultisig)
}

// TransferAvsGovernanceMultisig is a paid mutator transaction binding the contract method 0x513c52ba.
//
// Solidity: function transferAvsGovernanceMultisig(address _newAvsGovernanceMultisig) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) TransferAvsGovernanceMultisig(_newAvsGovernanceMultisig common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.TransferAvsGovernanceMultisig(&_AvsGovernance.TransactOpts, _newAvsGovernanceMultisig)
}

// TransferMessageHandler is a paid mutator transaction binding the contract method 0x4d07f651.
//
// Solidity: function transferMessageHandler(address _newMessageHandler) returns()
func (_AvsGovernance *AvsGovernanceTransactor) TransferMessageHandler(opts *bind.TransactOpts, _newMessageHandler common.Address) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "transferMessageHandler", _newMessageHandler)
}

// TransferMessageHandler is a paid mutator transaction binding the contract method 0x4d07f651.
//
// Solidity: function transferMessageHandler(address _newMessageHandler) returns()
func (_AvsGovernance *AvsGovernanceSession) TransferMessageHandler(_newMessageHandler common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.TransferMessageHandler(&_AvsGovernance.TransactOpts, _newMessageHandler)
}

// TransferMessageHandler is a paid mutator transaction binding the contract method 0x4d07f651.
//
// Solidity: function transferMessageHandler(address _newMessageHandler) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) TransferMessageHandler(_newMessageHandler common.Address) (*types.Transaction, error) {
	return _AvsGovernance.Contract.TransferMessageHandler(&_AvsGovernance.TransactOpts, _newMessageHandler)
}

// Unpause is a paid mutator transaction binding the contract method 0xbac1e94b.
//
// Solidity: function unpause(bytes4 _pausableFlow) returns()
func (_AvsGovernance *AvsGovernanceTransactor) Unpause(opts *bind.TransactOpts, _pausableFlow [4]byte) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "unpause", _pausableFlow)
}

// Unpause is a paid mutator transaction binding the contract method 0xbac1e94b.
//
// Solidity: function unpause(bytes4 _pausableFlow) returns()
func (_AvsGovernance *AvsGovernanceSession) Unpause(_pausableFlow [4]byte) (*types.Transaction, error) {
	return _AvsGovernance.Contract.Unpause(&_AvsGovernance.TransactOpts, _pausableFlow)
}

// Unpause is a paid mutator transaction binding the contract method 0xbac1e94b.
//
// Solidity: function unpause(bytes4 _pausableFlow) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) Unpause(_pausableFlow [4]byte) (*types.Transaction, error) {
	return _AvsGovernance.Contract.Unpause(&_AvsGovernance.TransactOpts, _pausableFlow)
}

// UnregisterAsOperator is a paid mutator transaction binding the contract method 0x09869442.
//
// Solidity: function unregisterAsOperator() returns()
func (_AvsGovernance *AvsGovernanceTransactor) UnregisterAsOperator(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "unregisterAsOperator")
}

// UnregisterAsOperator is a paid mutator transaction binding the contract method 0x09869442.
//
// Solidity: function unregisterAsOperator() returns()
func (_AvsGovernance *AvsGovernanceSession) UnregisterAsOperator() (*types.Transaction, error) {
	return _AvsGovernance.Contract.UnregisterAsOperator(&_AvsGovernance.TransactOpts)
}

// UnregisterAsOperator is a paid mutator transaction binding the contract method 0x09869442.
//
// Solidity: function unregisterAsOperator() returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) UnregisterAsOperator() (*types.Transaction, error) {
	return _AvsGovernance.Contract.UnregisterAsOperator(&_AvsGovernance.TransactOpts)
}

// UpdateAVSMetadataURI is a paid mutator transaction binding the contract method 0xa98fb355.
//
// Solidity: function updateAVSMetadataURI(string metadataURI) returns()
func (_AvsGovernance *AvsGovernanceTransactor) UpdateAVSMetadataURI(opts *bind.TransactOpts, metadataURI string) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "updateAVSMetadataURI", metadataURI)
}

// UpdateAVSMetadataURI is a paid mutator transaction binding the contract method 0xa98fb355.
//
// Solidity: function updateAVSMetadataURI(string metadataURI) returns()
func (_AvsGovernance *AvsGovernanceSession) UpdateAVSMetadataURI(metadataURI string) (*types.Transaction, error) {
	return _AvsGovernance.Contract.UpdateAVSMetadataURI(&_AvsGovernance.TransactOpts, metadataURI)
}

// UpdateAVSMetadataURI is a paid mutator transaction binding the contract method 0xa98fb355.
//
// Solidity: function updateAVSMetadataURI(string metadataURI) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) UpdateAVSMetadataURI(metadataURI string) (*types.Transaction, error) {
	return _AvsGovernance.Contract.UpdateAVSMetadataURI(&_AvsGovernance.TransactOpts, metadataURI)
}

// WithdrawBatchRewards is a paid mutator transaction binding the contract method 0xbc8be0c8.
//
// Solidity: function withdrawBatchRewards((address,uint256)[] _operators, uint256 _lastPayedTask) returns()
func (_AvsGovernance *AvsGovernanceTransactor) WithdrawBatchRewards(opts *bind.TransactOpts, _operators []IAvsGovernancePaymentRequestMessage, _lastPayedTask *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "withdrawBatchRewards", _operators, _lastPayedTask)
}

// WithdrawBatchRewards is a paid mutator transaction binding the contract method 0xbc8be0c8.
//
// Solidity: function withdrawBatchRewards((address,uint256)[] _operators, uint256 _lastPayedTask) returns()
func (_AvsGovernance *AvsGovernanceSession) WithdrawBatchRewards(_operators []IAvsGovernancePaymentRequestMessage, _lastPayedTask *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.WithdrawBatchRewards(&_AvsGovernance.TransactOpts, _operators, _lastPayedTask)
}

// WithdrawBatchRewards is a paid mutator transaction binding the contract method 0xbc8be0c8.
//
// Solidity: function withdrawBatchRewards((address,uint256)[] _operators, uint256 _lastPayedTask) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) WithdrawBatchRewards(_operators []IAvsGovernancePaymentRequestMessage, _lastPayedTask *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.WithdrawBatchRewards(&_AvsGovernance.TransactOpts, _operators, _lastPayedTask)
}

// WithdrawRewards is a paid mutator transaction binding the contract method 0x3256b4d1.
//
// Solidity: function withdrawRewards(address _operator, uint256 _lastPayedTask, uint256 _feeToClaim) returns()
func (_AvsGovernance *AvsGovernanceTransactor) WithdrawRewards(opts *bind.TransactOpts, _operator common.Address, _lastPayedTask *big.Int, _feeToClaim *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.contract.Transact(opts, "withdrawRewards", _operator, _lastPayedTask, _feeToClaim)
}

// WithdrawRewards is a paid mutator transaction binding the contract method 0x3256b4d1.
//
// Solidity: function withdrawRewards(address _operator, uint256 _lastPayedTask, uint256 _feeToClaim) returns()
func (_AvsGovernance *AvsGovernanceSession) WithdrawRewards(_operator common.Address, _lastPayedTask *big.Int, _feeToClaim *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.WithdrawRewards(&_AvsGovernance.TransactOpts, _operator, _lastPayedTask, _feeToClaim)
}

// WithdrawRewards is a paid mutator transaction binding the contract method 0x3256b4d1.
//
// Solidity: function withdrawRewards(address _operator, uint256 _lastPayedTask, uint256 _feeToClaim) returns()
func (_AvsGovernance *AvsGovernanceTransactorSession) WithdrawRewards(_operator common.Address, _lastPayedTask *big.Int, _feeToClaim *big.Int) (*types.Transaction, error) {
	return _AvsGovernance.Contract.WithdrawRewards(&_AvsGovernance.TransactOpts, _operator, _lastPayedTask, _feeToClaim)
}

// AvsGovernanceBLSAuthSingletonSetIterator is returned from FilterBLSAuthSingletonSet and is used to iterate over the raw logs and unpacked data for BLSAuthSingletonSet events raised by the AvsGovernance contract.
type AvsGovernanceBLSAuthSingletonSetIterator struct {
	Event *AvsGovernanceBLSAuthSingletonSet // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceBLSAuthSingletonSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceBLSAuthSingletonSet)
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
		it.Event = new(AvsGovernanceBLSAuthSingletonSet)
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
func (it *AvsGovernanceBLSAuthSingletonSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceBLSAuthSingletonSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceBLSAuthSingletonSet represents a BLSAuthSingletonSet event raised by the AvsGovernance contract.
type AvsGovernanceBLSAuthSingletonSet struct {
	BlsAuthSingleton common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterBLSAuthSingletonSet is a free log retrieval operation binding the contract event 0x4cbffdecf3b5e4b22bfb2bdec99a66f8fcf81e19b060682afd9645c729da1472.
//
// Solidity: event BLSAuthSingletonSet(address blsAuthSingleton)
func (_AvsGovernance *AvsGovernanceFilterer) FilterBLSAuthSingletonSet(opts *bind.FilterOpts) (*AvsGovernanceBLSAuthSingletonSetIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "BLSAuthSingletonSet")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceBLSAuthSingletonSetIterator{contract: _AvsGovernance.contract, event: "BLSAuthSingletonSet", logs: logs, sub: sub}, nil
}

// WatchBLSAuthSingletonSet is a free log subscription operation binding the contract event 0x4cbffdecf3b5e4b22bfb2bdec99a66f8fcf81e19b060682afd9645c729da1472.
//
// Solidity: event BLSAuthSingletonSet(address blsAuthSingleton)
func (_AvsGovernance *AvsGovernanceFilterer) WatchBLSAuthSingletonSet(opts *bind.WatchOpts, sink chan<- *AvsGovernanceBLSAuthSingletonSet) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "BLSAuthSingletonSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceBLSAuthSingletonSet)
				if err := _AvsGovernance.contract.UnpackLog(event, "BLSAuthSingletonSet", log); err != nil {
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

// ParseBLSAuthSingletonSet is a log parse operation binding the contract event 0x4cbffdecf3b5e4b22bfb2bdec99a66f8fcf81e19b060682afd9645c729da1472.
//
// Solidity: event BLSAuthSingletonSet(address blsAuthSingleton)
func (_AvsGovernance *AvsGovernanceFilterer) ParseBLSAuthSingletonSet(log types.Log) (*AvsGovernanceBLSAuthSingletonSet, error) {
	event := new(AvsGovernanceBLSAuthSingletonSet)
	if err := _AvsGovernance.contract.UnpackLog(event, "BLSAuthSingletonSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceFlowPausedIterator is returned from FilterFlowPaused and is used to iterate over the raw logs and unpacked data for FlowPaused events raised by the AvsGovernance contract.
type AvsGovernanceFlowPausedIterator struct {
	Event *AvsGovernanceFlowPaused // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceFlowPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceFlowPaused)
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
		it.Event = new(AvsGovernanceFlowPaused)
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
func (it *AvsGovernanceFlowPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceFlowPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceFlowPaused represents a FlowPaused event raised by the AvsGovernance contract.
type AvsGovernanceFlowPaused struct {
	PausableFlow [4]byte
	Pauser       common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterFlowPaused is a free log retrieval operation binding the contract event 0x95c3658c5e0c74e20cf12db371b9b67d26e97a1937f6d2284f88cc44d036b4f6.
//
// Solidity: event FlowPaused(bytes4 _pausableFlow, address _pauser)
func (_AvsGovernance *AvsGovernanceFilterer) FilterFlowPaused(opts *bind.FilterOpts) (*AvsGovernanceFlowPausedIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "FlowPaused")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceFlowPausedIterator{contract: _AvsGovernance.contract, event: "FlowPaused", logs: logs, sub: sub}, nil
}

// WatchFlowPaused is a free log subscription operation binding the contract event 0x95c3658c5e0c74e20cf12db371b9b67d26e97a1937f6d2284f88cc44d036b4f6.
//
// Solidity: event FlowPaused(bytes4 _pausableFlow, address _pauser)
func (_AvsGovernance *AvsGovernanceFilterer) WatchFlowPaused(opts *bind.WatchOpts, sink chan<- *AvsGovernanceFlowPaused) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "FlowPaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceFlowPaused)
				if err := _AvsGovernance.contract.UnpackLog(event, "FlowPaused", log); err != nil {
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
func (_AvsGovernance *AvsGovernanceFilterer) ParseFlowPaused(log types.Log) (*AvsGovernanceFlowPaused, error) {
	event := new(AvsGovernanceFlowPaused)
	if err := _AvsGovernance.contract.UnpackLog(event, "FlowPaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceFlowUnpausedIterator is returned from FilterFlowUnpaused and is used to iterate over the raw logs and unpacked data for FlowUnpaused events raised by the AvsGovernance contract.
type AvsGovernanceFlowUnpausedIterator struct {
	Event *AvsGovernanceFlowUnpaused // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceFlowUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceFlowUnpaused)
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
		it.Event = new(AvsGovernanceFlowUnpaused)
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
func (it *AvsGovernanceFlowUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceFlowUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceFlowUnpaused represents a FlowUnpaused event raised by the AvsGovernance contract.
type AvsGovernanceFlowUnpaused struct {
	PausableFlowFlag [4]byte
	Unpauser         common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterFlowUnpaused is a free log retrieval operation binding the contract event 0xc7e56e17b0a6c4b467df6495e1eda1baecd7ba20604e80c1058ac06f4578d85e.
//
// Solidity: event FlowUnpaused(bytes4 _pausableFlowFlag, address _unpauser)
func (_AvsGovernance *AvsGovernanceFilterer) FilterFlowUnpaused(opts *bind.FilterOpts) (*AvsGovernanceFlowUnpausedIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "FlowUnpaused")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceFlowUnpausedIterator{contract: _AvsGovernance.contract, event: "FlowUnpaused", logs: logs, sub: sub}, nil
}

// WatchFlowUnpaused is a free log subscription operation binding the contract event 0xc7e56e17b0a6c4b467df6495e1eda1baecd7ba20604e80c1058ac06f4578d85e.
//
// Solidity: event FlowUnpaused(bytes4 _pausableFlowFlag, address _unpauser)
func (_AvsGovernance *AvsGovernanceFilterer) WatchFlowUnpaused(opts *bind.WatchOpts, sink chan<- *AvsGovernanceFlowUnpaused) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "FlowUnpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceFlowUnpaused)
				if err := _AvsGovernance.contract.UnpackLog(event, "FlowUnpaused", log); err != nil {
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
func (_AvsGovernance *AvsGovernanceFilterer) ParseFlowUnpaused(log types.Log) (*AvsGovernanceFlowUnpaused, error) {
	event := new(AvsGovernanceFlowUnpaused)
	if err := _AvsGovernance.contract.UnpackLog(event, "FlowUnpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the AvsGovernance contract.
type AvsGovernanceInitializedIterator struct {
	Event *AvsGovernanceInitialized // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceInitialized)
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
		it.Event = new(AvsGovernanceInitialized)
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
func (it *AvsGovernanceInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceInitialized represents a Initialized event raised by the AvsGovernance contract.
type AvsGovernanceInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_AvsGovernance *AvsGovernanceFilterer) FilterInitialized(opts *bind.FilterOpts) (*AvsGovernanceInitializedIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceInitializedIterator{contract: _AvsGovernance.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_AvsGovernance *AvsGovernanceFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *AvsGovernanceInitialized) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceInitialized)
				if err := _AvsGovernance.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_AvsGovernance *AvsGovernanceFilterer) ParseInitialized(log types.Log) (*AvsGovernanceInitialized, error) {
	event := new(AvsGovernanceInitialized)
	if err := _AvsGovernance.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceMaxEffectiveBalanceSetIterator is returned from FilterMaxEffectiveBalanceSet and is used to iterate over the raw logs and unpacked data for MaxEffectiveBalanceSet events raised by the AvsGovernance contract.
type AvsGovernanceMaxEffectiveBalanceSetIterator struct {
	Event *AvsGovernanceMaxEffectiveBalanceSet // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceMaxEffectiveBalanceSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceMaxEffectiveBalanceSet)
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
		it.Event = new(AvsGovernanceMaxEffectiveBalanceSet)
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
func (it *AvsGovernanceMaxEffectiveBalanceSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceMaxEffectiveBalanceSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceMaxEffectiveBalanceSet represents a MaxEffectiveBalanceSet event raised by the AvsGovernance contract.
type AvsGovernanceMaxEffectiveBalanceSet struct {
	MaxEffectiveBalance *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterMaxEffectiveBalanceSet is a free log retrieval operation binding the contract event 0x00c6fb6db9c52d89a1eaf84e0470a3304db2086d0ac44d64ebf4ea35a905a7d0.
//
// Solidity: event MaxEffectiveBalanceSet(uint256 maxEffectiveBalance)
func (_AvsGovernance *AvsGovernanceFilterer) FilterMaxEffectiveBalanceSet(opts *bind.FilterOpts) (*AvsGovernanceMaxEffectiveBalanceSetIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "MaxEffectiveBalanceSet")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceMaxEffectiveBalanceSetIterator{contract: _AvsGovernance.contract, event: "MaxEffectiveBalanceSet", logs: logs, sub: sub}, nil
}

// WatchMaxEffectiveBalanceSet is a free log subscription operation binding the contract event 0x00c6fb6db9c52d89a1eaf84e0470a3304db2086d0ac44d64ebf4ea35a905a7d0.
//
// Solidity: event MaxEffectiveBalanceSet(uint256 maxEffectiveBalance)
func (_AvsGovernance *AvsGovernanceFilterer) WatchMaxEffectiveBalanceSet(opts *bind.WatchOpts, sink chan<- *AvsGovernanceMaxEffectiveBalanceSet) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "MaxEffectiveBalanceSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceMaxEffectiveBalanceSet)
				if err := _AvsGovernance.contract.UnpackLog(event, "MaxEffectiveBalanceSet", log); err != nil {
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

// ParseMaxEffectiveBalanceSet is a log parse operation binding the contract event 0x00c6fb6db9c52d89a1eaf84e0470a3304db2086d0ac44d64ebf4ea35a905a7d0.
//
// Solidity: event MaxEffectiveBalanceSet(uint256 maxEffectiveBalance)
func (_AvsGovernance *AvsGovernanceFilterer) ParseMaxEffectiveBalanceSet(log types.Log) (*AvsGovernanceMaxEffectiveBalanceSet, error) {
	event := new(AvsGovernanceMaxEffectiveBalanceSet)
	if err := _AvsGovernance.contract.UnpackLog(event, "MaxEffectiveBalanceSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceMinSharesPerStrategySetIterator is returned from FilterMinSharesPerStrategySet and is used to iterate over the raw logs and unpacked data for MinSharesPerStrategySet events raised by the AvsGovernance contract.
type AvsGovernanceMinSharesPerStrategySetIterator struct {
	Event *AvsGovernanceMinSharesPerStrategySet // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceMinSharesPerStrategySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceMinSharesPerStrategySet)
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
		it.Event = new(AvsGovernanceMinSharesPerStrategySet)
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
func (it *AvsGovernanceMinSharesPerStrategySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceMinSharesPerStrategySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceMinSharesPerStrategySet represents a MinSharesPerStrategySet event raised by the AvsGovernance contract.
type AvsGovernanceMinSharesPerStrategySet struct {
	Strategy  common.Address
	MinShares *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMinSharesPerStrategySet is a free log retrieval operation binding the contract event 0x3a6c52328a7b3b726d0ec757d68f416b26ec2991ac4d4f95d450c504f5a0e521.
//
// Solidity: event MinSharesPerStrategySet(address strategy, uint256 minShares)
func (_AvsGovernance *AvsGovernanceFilterer) FilterMinSharesPerStrategySet(opts *bind.FilterOpts) (*AvsGovernanceMinSharesPerStrategySetIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "MinSharesPerStrategySet")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceMinSharesPerStrategySetIterator{contract: _AvsGovernance.contract, event: "MinSharesPerStrategySet", logs: logs, sub: sub}, nil
}

// WatchMinSharesPerStrategySet is a free log subscription operation binding the contract event 0x3a6c52328a7b3b726d0ec757d68f416b26ec2991ac4d4f95d450c504f5a0e521.
//
// Solidity: event MinSharesPerStrategySet(address strategy, uint256 minShares)
func (_AvsGovernance *AvsGovernanceFilterer) WatchMinSharesPerStrategySet(opts *bind.WatchOpts, sink chan<- *AvsGovernanceMinSharesPerStrategySet) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "MinSharesPerStrategySet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceMinSharesPerStrategySet)
				if err := _AvsGovernance.contract.UnpackLog(event, "MinSharesPerStrategySet", log); err != nil {
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

// ParseMinSharesPerStrategySet is a log parse operation binding the contract event 0x3a6c52328a7b3b726d0ec757d68f416b26ec2991ac4d4f95d450c504f5a0e521.
//
// Solidity: event MinSharesPerStrategySet(address strategy, uint256 minShares)
func (_AvsGovernance *AvsGovernanceFilterer) ParseMinSharesPerStrategySet(log types.Log) (*AvsGovernanceMinSharesPerStrategySet, error) {
	event := new(AvsGovernanceMinSharesPerStrategySet)
	if err := _AvsGovernance.contract.UnpackLog(event, "MinSharesPerStrategySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceMinVotingPowerSetIterator is returned from FilterMinVotingPowerSet and is used to iterate over the raw logs and unpacked data for MinVotingPowerSet events raised by the AvsGovernance contract.
type AvsGovernanceMinVotingPowerSetIterator struct {
	Event *AvsGovernanceMinVotingPowerSet // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceMinVotingPowerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceMinVotingPowerSet)
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
		it.Event = new(AvsGovernanceMinVotingPowerSet)
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
func (it *AvsGovernanceMinVotingPowerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceMinVotingPowerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceMinVotingPowerSet represents a MinVotingPowerSet event raised by the AvsGovernance contract.
type AvsGovernanceMinVotingPowerSet struct {
	MinVotingPower *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterMinVotingPowerSet is a free log retrieval operation binding the contract event 0x10203ddc048c86cf14172a6ea2565c805ce7320b22d6941b2eb396d0ee077983.
//
// Solidity: event MinVotingPowerSet(uint256 minVotingPower)
func (_AvsGovernance *AvsGovernanceFilterer) FilterMinVotingPowerSet(opts *bind.FilterOpts) (*AvsGovernanceMinVotingPowerSetIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "MinVotingPowerSet")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceMinVotingPowerSetIterator{contract: _AvsGovernance.contract, event: "MinVotingPowerSet", logs: logs, sub: sub}, nil
}

// WatchMinVotingPowerSet is a free log subscription operation binding the contract event 0x10203ddc048c86cf14172a6ea2565c805ce7320b22d6941b2eb396d0ee077983.
//
// Solidity: event MinVotingPowerSet(uint256 minVotingPower)
func (_AvsGovernance *AvsGovernanceFilterer) WatchMinVotingPowerSet(opts *bind.WatchOpts, sink chan<- *AvsGovernanceMinVotingPowerSet) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "MinVotingPowerSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceMinVotingPowerSet)
				if err := _AvsGovernance.contract.UnpackLog(event, "MinVotingPowerSet", log); err != nil {
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

// ParseMinVotingPowerSet is a log parse operation binding the contract event 0x10203ddc048c86cf14172a6ea2565c805ce7320b22d6941b2eb396d0ee077983.
//
// Solidity: event MinVotingPowerSet(uint256 minVotingPower)
func (_AvsGovernance *AvsGovernanceFilterer) ParseMinVotingPowerSet(log types.Log) (*AvsGovernanceMinVotingPowerSet, error) {
	event := new(AvsGovernanceMinVotingPowerSet)
	if err := _AvsGovernance.contract.UnpackLog(event, "MinVotingPowerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceOperatorRegisteredIterator is returned from FilterOperatorRegistered and is used to iterate over the raw logs and unpacked data for OperatorRegistered events raised by the AvsGovernance contract.
type AvsGovernanceOperatorRegisteredIterator struct {
	Event *AvsGovernanceOperatorRegistered // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceOperatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceOperatorRegistered)
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
		it.Event = new(AvsGovernanceOperatorRegistered)
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
func (it *AvsGovernanceOperatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceOperatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceOperatorRegistered represents a OperatorRegistered event raised by the AvsGovernance contract.
type AvsGovernanceOperatorRegistered struct {
	Operator common.Address
	BlsKey   [4]*big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOperatorRegistered is a free log retrieval operation binding the contract event 0x54bc9cf83c2eb0f2ad1abf6e4fab882964404622ba2df6b5a9356a18d3aac055.
//
// Solidity: event OperatorRegistered(address indexed operator, uint256[4] blsKey)
func (_AvsGovernance *AvsGovernanceFilterer) FilterOperatorRegistered(opts *bind.FilterOpts, operator []common.Address) (*AvsGovernanceOperatorRegisteredIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "OperatorRegistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceOperatorRegisteredIterator{contract: _AvsGovernance.contract, event: "OperatorRegistered", logs: logs, sub: sub}, nil
}

// WatchOperatorRegistered is a free log subscription operation binding the contract event 0x54bc9cf83c2eb0f2ad1abf6e4fab882964404622ba2df6b5a9356a18d3aac055.
//
// Solidity: event OperatorRegistered(address indexed operator, uint256[4] blsKey)
func (_AvsGovernance *AvsGovernanceFilterer) WatchOperatorRegistered(opts *bind.WatchOpts, sink chan<- *AvsGovernanceOperatorRegistered, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "OperatorRegistered", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceOperatorRegistered)
				if err := _AvsGovernance.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
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

// ParseOperatorRegistered is a log parse operation binding the contract event 0x54bc9cf83c2eb0f2ad1abf6e4fab882964404622ba2df6b5a9356a18d3aac055.
//
// Solidity: event OperatorRegistered(address indexed operator, uint256[4] blsKey)
func (_AvsGovernance *AvsGovernanceFilterer) ParseOperatorRegistered(log types.Log) (*AvsGovernanceOperatorRegistered, error) {
	event := new(AvsGovernanceOperatorRegistered)
	if err := _AvsGovernance.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceOperatorUnregisteredIterator is returned from FilterOperatorUnregistered and is used to iterate over the raw logs and unpacked data for OperatorUnregistered events raised by the AvsGovernance contract.
type AvsGovernanceOperatorUnregisteredIterator struct {
	Event *AvsGovernanceOperatorUnregistered // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceOperatorUnregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceOperatorUnregistered)
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
		it.Event = new(AvsGovernanceOperatorUnregistered)
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
func (it *AvsGovernanceOperatorUnregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceOperatorUnregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceOperatorUnregistered represents a OperatorUnregistered event raised by the AvsGovernance contract.
type AvsGovernanceOperatorUnregistered struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOperatorUnregistered is a free log retrieval operation binding the contract event 0x6f42117a557500c705ddf040a619d86f39101e6b74ac20d7b3e5943ba473fc7f.
//
// Solidity: event OperatorUnregistered(address operator)
func (_AvsGovernance *AvsGovernanceFilterer) FilterOperatorUnregistered(opts *bind.FilterOpts) (*AvsGovernanceOperatorUnregisteredIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "OperatorUnregistered")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceOperatorUnregisteredIterator{contract: _AvsGovernance.contract, event: "OperatorUnregistered", logs: logs, sub: sub}, nil
}

// WatchOperatorUnregistered is a free log subscription operation binding the contract event 0x6f42117a557500c705ddf040a619d86f39101e6b74ac20d7b3e5943ba473fc7f.
//
// Solidity: event OperatorUnregistered(address operator)
func (_AvsGovernance *AvsGovernanceFilterer) WatchOperatorUnregistered(opts *bind.WatchOpts, sink chan<- *AvsGovernanceOperatorUnregistered) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "OperatorUnregistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceOperatorUnregistered)
				if err := _AvsGovernance.contract.UnpackLog(event, "OperatorUnregistered", log); err != nil {
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

// ParseOperatorUnregistered is a log parse operation binding the contract event 0x6f42117a557500c705ddf040a619d86f39101e6b74ac20d7b3e5943ba473fc7f.
//
// Solidity: event OperatorUnregistered(address operator)
func (_AvsGovernance *AvsGovernanceFilterer) ParseOperatorUnregistered(log types.Log) (*AvsGovernanceOperatorUnregistered, error) {
	event := new(AvsGovernanceOperatorUnregistered)
	if err := _AvsGovernance.contract.UnpackLog(event, "OperatorUnregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceQueuedRewardsReceiverModificationIterator is returned from FilterQueuedRewardsReceiverModification and is used to iterate over the raw logs and unpacked data for QueuedRewardsReceiverModification events raised by the AvsGovernance contract.
type AvsGovernanceQueuedRewardsReceiverModificationIterator struct {
	Event *AvsGovernanceQueuedRewardsReceiverModification // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceQueuedRewardsReceiverModificationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceQueuedRewardsReceiverModification)
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
		it.Event = new(AvsGovernanceQueuedRewardsReceiverModification)
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
func (it *AvsGovernanceQueuedRewardsReceiverModificationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceQueuedRewardsReceiverModificationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceQueuedRewardsReceiverModification represents a QueuedRewardsReceiverModification event raised by the AvsGovernance contract.
type AvsGovernanceQueuedRewardsReceiverModification struct {
	Operator common.Address
	Receiver common.Address
	Delay    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterQueuedRewardsReceiverModification is a free log retrieval operation binding the contract event 0x0d8cfa10a3087b28d3c226ad9a37314860e7c3c0505a25a39e3cdefb3118a98a.
//
// Solidity: event QueuedRewardsReceiverModification(address operator, address receiver, uint256 delay)
func (_AvsGovernance *AvsGovernanceFilterer) FilterQueuedRewardsReceiverModification(opts *bind.FilterOpts) (*AvsGovernanceQueuedRewardsReceiverModificationIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "QueuedRewardsReceiverModification")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceQueuedRewardsReceiverModificationIterator{contract: _AvsGovernance.contract, event: "QueuedRewardsReceiverModification", logs: logs, sub: sub}, nil
}

// WatchQueuedRewardsReceiverModification is a free log subscription operation binding the contract event 0x0d8cfa10a3087b28d3c226ad9a37314860e7c3c0505a25a39e3cdefb3118a98a.
//
// Solidity: event QueuedRewardsReceiverModification(address operator, address receiver, uint256 delay)
func (_AvsGovernance *AvsGovernanceFilterer) WatchQueuedRewardsReceiverModification(opts *bind.WatchOpts, sink chan<- *AvsGovernanceQueuedRewardsReceiverModification) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "QueuedRewardsReceiverModification")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceQueuedRewardsReceiverModification)
				if err := _AvsGovernance.contract.UnpackLog(event, "QueuedRewardsReceiverModification", log); err != nil {
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

// ParseQueuedRewardsReceiverModification is a log parse operation binding the contract event 0x0d8cfa10a3087b28d3c226ad9a37314860e7c3c0505a25a39e3cdefb3118a98a.
//
// Solidity: event QueuedRewardsReceiverModification(address operator, address receiver, uint256 delay)
func (_AvsGovernance *AvsGovernanceFilterer) ParseQueuedRewardsReceiverModification(log types.Log) (*AvsGovernanceQueuedRewardsReceiverModification, error) {
	event := new(AvsGovernanceQueuedRewardsReceiverModification)
	if err := _AvsGovernance.contract.UnpackLog(event, "QueuedRewardsReceiverModification", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the AvsGovernance contract.
type AvsGovernanceRoleAdminChangedIterator struct {
	Event *AvsGovernanceRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceRoleAdminChanged)
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
		it.Event = new(AvsGovernanceRoleAdminChanged)
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
func (it *AvsGovernanceRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceRoleAdminChanged represents a RoleAdminChanged event raised by the AvsGovernance contract.
type AvsGovernanceRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AvsGovernance *AvsGovernanceFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*AvsGovernanceRoleAdminChangedIterator, error) {

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

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceRoleAdminChangedIterator{contract: _AvsGovernance.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AvsGovernance *AvsGovernanceFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *AvsGovernanceRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceRoleAdminChanged)
				if err := _AvsGovernance.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_AvsGovernance *AvsGovernanceFilterer) ParseRoleAdminChanged(log types.Log) (*AvsGovernanceRoleAdminChanged, error) {
	event := new(AvsGovernanceRoleAdminChanged)
	if err := _AvsGovernance.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the AvsGovernance contract.
type AvsGovernanceRoleGrantedIterator struct {
	Event *AvsGovernanceRoleGranted // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceRoleGranted)
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
		it.Event = new(AvsGovernanceRoleGranted)
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
func (it *AvsGovernanceRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceRoleGranted represents a RoleGranted event raised by the AvsGovernance contract.
type AvsGovernanceRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AvsGovernance *AvsGovernanceFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AvsGovernanceRoleGrantedIterator, error) {

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

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceRoleGrantedIterator{contract: _AvsGovernance.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AvsGovernance *AvsGovernanceFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *AvsGovernanceRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceRoleGranted)
				if err := _AvsGovernance.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_AvsGovernance *AvsGovernanceFilterer) ParseRoleGranted(log types.Log) (*AvsGovernanceRoleGranted, error) {
	event := new(AvsGovernanceRoleGranted)
	if err := _AvsGovernance.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the AvsGovernance contract.
type AvsGovernanceRoleRevokedIterator struct {
	Event *AvsGovernanceRoleRevoked // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceRoleRevoked)
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
		it.Event = new(AvsGovernanceRoleRevoked)
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
func (it *AvsGovernanceRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceRoleRevoked represents a RoleRevoked event raised by the AvsGovernance contract.
type AvsGovernanceRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AvsGovernance *AvsGovernanceFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AvsGovernanceRoleRevokedIterator, error) {

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

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceRoleRevokedIterator{contract: _AvsGovernance.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AvsGovernance *AvsGovernanceFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *AvsGovernanceRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceRoleRevoked)
				if err := _AvsGovernance.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_AvsGovernance *AvsGovernanceFilterer) ParseRoleRevoked(log types.Log) (*AvsGovernanceRoleRevoked, error) {
	event := new(AvsGovernanceRoleRevoked)
	if err := _AvsGovernance.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetAllowlistSignerIterator is returned from FilterSetAllowlistSigner and is used to iterate over the raw logs and unpacked data for SetAllowlistSigner events raised by the AvsGovernance contract.
type AvsGovernanceSetAllowlistSignerIterator struct {
	Event *AvsGovernanceSetAllowlistSigner // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetAllowlistSignerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetAllowlistSigner)
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
		it.Event = new(AvsGovernanceSetAllowlistSigner)
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
func (it *AvsGovernanceSetAllowlistSignerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetAllowlistSignerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetAllowlistSigner represents a SetAllowlistSigner event raised by the AvsGovernance contract.
type AvsGovernanceSetAllowlistSigner struct {
	AllowlistSigner common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSetAllowlistSigner is a free log retrieval operation binding the contract event 0xfa4acc0aaeb2714e420e9c8339167ddef7bc66c0f94a0c5a7722de21dcb7508c.
//
// Solidity: event SetAllowlistSigner(address allowlistSigner)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetAllowlistSigner(opts *bind.FilterOpts) (*AvsGovernanceSetAllowlistSignerIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetAllowlistSigner")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetAllowlistSignerIterator{contract: _AvsGovernance.contract, event: "SetAllowlistSigner", logs: logs, sub: sub}, nil
}

// WatchSetAllowlistSigner is a free log subscription operation binding the contract event 0xfa4acc0aaeb2714e420e9c8339167ddef7bc66c0f94a0c5a7722de21dcb7508c.
//
// Solidity: event SetAllowlistSigner(address allowlistSigner)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetAllowlistSigner(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetAllowlistSigner) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetAllowlistSigner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetAllowlistSigner)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetAllowlistSigner", log); err != nil {
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

// ParseSetAllowlistSigner is a log parse operation binding the contract event 0xfa4acc0aaeb2714e420e9c8339167ddef7bc66c0f94a0c5a7722de21dcb7508c.
//
// Solidity: event SetAllowlistSigner(address allowlistSigner)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetAllowlistSigner(log types.Log) (*AvsGovernanceSetAllowlistSigner, error) {
	event := new(AvsGovernanceSetAllowlistSigner)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetAllowlistSigner", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetAvsGovernanceLogicIterator is returned from FilterSetAvsGovernanceLogic and is used to iterate over the raw logs and unpacked data for SetAvsGovernanceLogic events raised by the AvsGovernance contract.
type AvsGovernanceSetAvsGovernanceLogicIterator struct {
	Event *AvsGovernanceSetAvsGovernanceLogic // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetAvsGovernanceLogicIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetAvsGovernanceLogic)
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
		it.Event = new(AvsGovernanceSetAvsGovernanceLogic)
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
func (it *AvsGovernanceSetAvsGovernanceLogicIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetAvsGovernanceLogicIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetAvsGovernanceLogic represents a SetAvsGovernanceLogic event raised by the AvsGovernance contract.
type AvsGovernanceSetAvsGovernanceLogic struct {
	AvsGovernanceLogic common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterSetAvsGovernanceLogic is a free log retrieval operation binding the contract event 0x7c36ee80df183e227956a9f387a48d26bbf4d2f1526410493d11126de5a8942c.
//
// Solidity: event SetAvsGovernanceLogic(address avsGovernanceLogic)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetAvsGovernanceLogic(opts *bind.FilterOpts) (*AvsGovernanceSetAvsGovernanceLogicIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetAvsGovernanceLogic")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetAvsGovernanceLogicIterator{contract: _AvsGovernance.contract, event: "SetAvsGovernanceLogic", logs: logs, sub: sub}, nil
}

// WatchSetAvsGovernanceLogic is a free log subscription operation binding the contract event 0x7c36ee80df183e227956a9f387a48d26bbf4d2f1526410493d11126de5a8942c.
//
// Solidity: event SetAvsGovernanceLogic(address avsGovernanceLogic)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetAvsGovernanceLogic(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetAvsGovernanceLogic) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetAvsGovernanceLogic")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetAvsGovernanceLogic)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetAvsGovernanceLogic", log); err != nil {
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

// ParseSetAvsGovernanceLogic is a log parse operation binding the contract event 0x7c36ee80df183e227956a9f387a48d26bbf4d2f1526410493d11126de5a8942c.
//
// Solidity: event SetAvsGovernanceLogic(address avsGovernanceLogic)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetAvsGovernanceLogic(log types.Log) (*AvsGovernanceSetAvsGovernanceLogic, error) {
	event := new(AvsGovernanceSetAvsGovernanceLogic)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetAvsGovernanceLogic", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetAvsGovernanceMultiplierSyncerIterator is returned from FilterSetAvsGovernanceMultiplierSyncer and is used to iterate over the raw logs and unpacked data for SetAvsGovernanceMultiplierSyncer events raised by the AvsGovernance contract.
type AvsGovernanceSetAvsGovernanceMultiplierSyncerIterator struct {
	Event *AvsGovernanceSetAvsGovernanceMultiplierSyncer // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetAvsGovernanceMultiplierSyncerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetAvsGovernanceMultiplierSyncer)
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
		it.Event = new(AvsGovernanceSetAvsGovernanceMultiplierSyncer)
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
func (it *AvsGovernanceSetAvsGovernanceMultiplierSyncerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetAvsGovernanceMultiplierSyncerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetAvsGovernanceMultiplierSyncer represents a SetAvsGovernanceMultiplierSyncer event raised by the AvsGovernance contract.
type AvsGovernanceSetAvsGovernanceMultiplierSyncer struct {
	AvsGovernanceMultiplierSyncer common.Address
	Raw                           types.Log // Blockchain specific contextual infos
}

// FilterSetAvsGovernanceMultiplierSyncer is a free log retrieval operation binding the contract event 0xb73a70f24733a9265231de5807eae76d1740a9974b31a142ef9e243508987bbe.
//
// Solidity: event SetAvsGovernanceMultiplierSyncer(address avsGovernanceMultiplierSyncer)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetAvsGovernanceMultiplierSyncer(opts *bind.FilterOpts) (*AvsGovernanceSetAvsGovernanceMultiplierSyncerIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetAvsGovernanceMultiplierSyncer")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetAvsGovernanceMultiplierSyncerIterator{contract: _AvsGovernance.contract, event: "SetAvsGovernanceMultiplierSyncer", logs: logs, sub: sub}, nil
}

// WatchSetAvsGovernanceMultiplierSyncer is a free log subscription operation binding the contract event 0xb73a70f24733a9265231de5807eae76d1740a9974b31a142ef9e243508987bbe.
//
// Solidity: event SetAvsGovernanceMultiplierSyncer(address avsGovernanceMultiplierSyncer)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetAvsGovernanceMultiplierSyncer(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetAvsGovernanceMultiplierSyncer) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetAvsGovernanceMultiplierSyncer")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetAvsGovernanceMultiplierSyncer)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetAvsGovernanceMultiplierSyncer", log); err != nil {
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

// ParseSetAvsGovernanceMultiplierSyncer is a log parse operation binding the contract event 0xb73a70f24733a9265231de5807eae76d1740a9974b31a142ef9e243508987bbe.
//
// Solidity: event SetAvsGovernanceMultiplierSyncer(address avsGovernanceMultiplierSyncer)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetAvsGovernanceMultiplierSyncer(log types.Log) (*AvsGovernanceSetAvsGovernanceMultiplierSyncer, error) {
	event := new(AvsGovernanceSetAvsGovernanceMultiplierSyncer)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetAvsGovernanceMultiplierSyncer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetAvsGovernanceMultisigIterator is returned from FilterSetAvsGovernanceMultisig and is used to iterate over the raw logs and unpacked data for SetAvsGovernanceMultisig events raised by the AvsGovernance contract.
type AvsGovernanceSetAvsGovernanceMultisigIterator struct {
	Event *AvsGovernanceSetAvsGovernanceMultisig // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetAvsGovernanceMultisigIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetAvsGovernanceMultisig)
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
		it.Event = new(AvsGovernanceSetAvsGovernanceMultisig)
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
func (it *AvsGovernanceSetAvsGovernanceMultisigIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetAvsGovernanceMultisigIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetAvsGovernanceMultisig represents a SetAvsGovernanceMultisig event raised by the AvsGovernance contract.
type AvsGovernanceSetAvsGovernanceMultisig struct {
	NewAvsGovernanceMultisig common.Address
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterSetAvsGovernanceMultisig is a free log retrieval operation binding the contract event 0x024e98b7d808a3ddb028252dc95dfdcb165a0ca59fcff8984b4fecf9a7222649.
//
// Solidity: event SetAvsGovernanceMultisig(address newAvsGovernanceMultisig)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetAvsGovernanceMultisig(opts *bind.FilterOpts) (*AvsGovernanceSetAvsGovernanceMultisigIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetAvsGovernanceMultisig")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetAvsGovernanceMultisigIterator{contract: _AvsGovernance.contract, event: "SetAvsGovernanceMultisig", logs: logs, sub: sub}, nil
}

// WatchSetAvsGovernanceMultisig is a free log subscription operation binding the contract event 0x024e98b7d808a3ddb028252dc95dfdcb165a0ca59fcff8984b4fecf9a7222649.
//
// Solidity: event SetAvsGovernanceMultisig(address newAvsGovernanceMultisig)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetAvsGovernanceMultisig(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetAvsGovernanceMultisig) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetAvsGovernanceMultisig")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetAvsGovernanceMultisig)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetAvsGovernanceMultisig", log); err != nil {
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
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetAvsGovernanceMultisig(log types.Log) (*AvsGovernanceSetAvsGovernanceMultisig, error) {
	event := new(AvsGovernanceSetAvsGovernanceMultisig)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetAvsGovernanceMultisig", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetAvsNameIterator is returned from FilterSetAvsName and is used to iterate over the raw logs and unpacked data for SetAvsName events raised by the AvsGovernance contract.
type AvsGovernanceSetAvsNameIterator struct {
	Event *AvsGovernanceSetAvsName // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetAvsNameIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetAvsName)
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
		it.Event = new(AvsGovernanceSetAvsName)
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
func (it *AvsGovernanceSetAvsNameIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetAvsNameIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetAvsName represents a SetAvsName event raised by the AvsGovernance contract.
type AvsGovernanceSetAvsName struct {
	AvsName string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterSetAvsName is a free log retrieval operation binding the contract event 0x7f63aacad63bc1693280450d5c3612ccd4efc53e46d69f3a537db102cd66290c.
//
// Solidity: event SetAvsName(string avsName)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetAvsName(opts *bind.FilterOpts) (*AvsGovernanceSetAvsNameIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetAvsName")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetAvsNameIterator{contract: _AvsGovernance.contract, event: "SetAvsName", logs: logs, sub: sub}, nil
}

// WatchSetAvsName is a free log subscription operation binding the contract event 0x7f63aacad63bc1693280450d5c3612ccd4efc53e46d69f3a537db102cd66290c.
//
// Solidity: event SetAvsName(string avsName)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetAvsName(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetAvsName) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetAvsName")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetAvsName)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetAvsName", log); err != nil {
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

// ParseSetAvsName is a log parse operation binding the contract event 0x7f63aacad63bc1693280450d5c3612ccd4efc53e46d69f3a537db102cd66290c.
//
// Solidity: event SetAvsName(string avsName)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetAvsName(log types.Log) (*AvsGovernanceSetAvsName, error) {
	event := new(AvsGovernanceSetAvsName)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetAvsName", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetIsAllowlistedIterator is returned from FilterSetIsAllowlisted and is used to iterate over the raw logs and unpacked data for SetIsAllowlisted events raised by the AvsGovernance contract.
type AvsGovernanceSetIsAllowlistedIterator struct {
	Event *AvsGovernanceSetIsAllowlisted // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetIsAllowlistedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetIsAllowlisted)
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
		it.Event = new(AvsGovernanceSetIsAllowlisted)
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
func (it *AvsGovernanceSetIsAllowlistedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetIsAllowlistedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetIsAllowlisted represents a SetIsAllowlisted event raised by the AvsGovernance contract.
type AvsGovernanceSetIsAllowlisted struct {
	IsAllowlisted bool
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSetIsAllowlisted is a free log retrieval operation binding the contract event 0x2dcb3282f9b7aa18e1bf7fa254c45f3e270e8f26d9a37ae590d5d8125b58d1b1.
//
// Solidity: event SetIsAllowlisted(bool isAllowlisted)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetIsAllowlisted(opts *bind.FilterOpts) (*AvsGovernanceSetIsAllowlistedIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetIsAllowlisted")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetIsAllowlistedIterator{contract: _AvsGovernance.contract, event: "SetIsAllowlisted", logs: logs, sub: sub}, nil
}

// WatchSetIsAllowlisted is a free log subscription operation binding the contract event 0x2dcb3282f9b7aa18e1bf7fa254c45f3e270e8f26d9a37ae590d5d8125b58d1b1.
//
// Solidity: event SetIsAllowlisted(bool isAllowlisted)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetIsAllowlisted(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetIsAllowlisted) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetIsAllowlisted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetIsAllowlisted)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetIsAllowlisted", log); err != nil {
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

// ParseSetIsAllowlisted is a log parse operation binding the contract event 0x2dcb3282f9b7aa18e1bf7fa254c45f3e270e8f26d9a37ae590d5d8125b58d1b1.
//
// Solidity: event SetIsAllowlisted(bool isAllowlisted)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetIsAllowlisted(log types.Log) (*AvsGovernanceSetIsAllowlisted, error) {
	event := new(AvsGovernanceSetIsAllowlisted)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetIsAllowlisted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetMessageHandlerIterator is returned from FilterSetMessageHandler and is used to iterate over the raw logs and unpacked data for SetMessageHandler events raised by the AvsGovernance contract.
type AvsGovernanceSetMessageHandlerIterator struct {
	Event *AvsGovernanceSetMessageHandler // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetMessageHandlerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetMessageHandler)
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
		it.Event = new(AvsGovernanceSetMessageHandler)
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
func (it *AvsGovernanceSetMessageHandlerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetMessageHandlerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetMessageHandler represents a SetMessageHandler event raised by the AvsGovernance contract.
type AvsGovernanceSetMessageHandler struct {
	NewMessageHandler common.Address
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterSetMessageHandler is a free log retrieval operation binding the contract event 0x997f84b541d7b68e210e6f50e3402b51d8411dbbc4d44ed81e508383126e4e94.
//
// Solidity: event SetMessageHandler(address newMessageHandler)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetMessageHandler(opts *bind.FilterOpts) (*AvsGovernanceSetMessageHandlerIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetMessageHandler")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetMessageHandlerIterator{contract: _AvsGovernance.contract, event: "SetMessageHandler", logs: logs, sub: sub}, nil
}

// WatchSetMessageHandler is a free log subscription operation binding the contract event 0x997f84b541d7b68e210e6f50e3402b51d8411dbbc4d44ed81e508383126e4e94.
//
// Solidity: event SetMessageHandler(address newMessageHandler)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetMessageHandler(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetMessageHandler) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetMessageHandler")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetMessageHandler)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetMessageHandler", log); err != nil {
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
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetMessageHandler(log types.Log) (*AvsGovernanceSetMessageHandler, error) {
	event := new(AvsGovernanceSetMessageHandler)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetMessageHandler", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetNumOfOperatorsLimitIterator is returned from FilterSetNumOfOperatorsLimit and is used to iterate over the raw logs and unpacked data for SetNumOfOperatorsLimit events raised by the AvsGovernance contract.
type AvsGovernanceSetNumOfOperatorsLimitIterator struct {
	Event *AvsGovernanceSetNumOfOperatorsLimit // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetNumOfOperatorsLimitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetNumOfOperatorsLimit)
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
		it.Event = new(AvsGovernanceSetNumOfOperatorsLimit)
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
func (it *AvsGovernanceSetNumOfOperatorsLimitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetNumOfOperatorsLimitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetNumOfOperatorsLimit represents a SetNumOfOperatorsLimit event raised by the AvsGovernance contract.
type AvsGovernanceSetNumOfOperatorsLimit struct {
	NewLimitOfNumOfOperators *big.Int
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterSetNumOfOperatorsLimit is a free log retrieval operation binding the contract event 0xc0dd1d82df4ae12576f7a7912395305cf73deae26c764dd74a945cd6ba81591b.
//
// Solidity: event SetNumOfOperatorsLimit(uint256 newLimitOfNumOfOperators)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetNumOfOperatorsLimit(opts *bind.FilterOpts) (*AvsGovernanceSetNumOfOperatorsLimitIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetNumOfOperatorsLimit")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetNumOfOperatorsLimitIterator{contract: _AvsGovernance.contract, event: "SetNumOfOperatorsLimit", logs: logs, sub: sub}, nil
}

// WatchSetNumOfOperatorsLimit is a free log subscription operation binding the contract event 0xc0dd1d82df4ae12576f7a7912395305cf73deae26c764dd74a945cd6ba81591b.
//
// Solidity: event SetNumOfOperatorsLimit(uint256 newLimitOfNumOfOperators)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetNumOfOperatorsLimit(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetNumOfOperatorsLimit) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetNumOfOperatorsLimit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetNumOfOperatorsLimit)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetNumOfOperatorsLimit", log); err != nil {
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

// ParseSetNumOfOperatorsLimit is a log parse operation binding the contract event 0xc0dd1d82df4ae12576f7a7912395305cf73deae26c764dd74a945cd6ba81591b.
//
// Solidity: event SetNumOfOperatorsLimit(uint256 newLimitOfNumOfOperators)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetNumOfOperatorsLimit(log types.Log) (*AvsGovernanceSetNumOfOperatorsLimit, error) {
	event := new(AvsGovernanceSetNumOfOperatorsLimit)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetNumOfOperatorsLimit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetOthenticRegistryIterator is returned from FilterSetOthenticRegistry and is used to iterate over the raw logs and unpacked data for SetOthenticRegistry events raised by the AvsGovernance contract.
type AvsGovernanceSetOthenticRegistryIterator struct {
	Event *AvsGovernanceSetOthenticRegistry // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetOthenticRegistryIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetOthenticRegistry)
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
		it.Event = new(AvsGovernanceSetOthenticRegistry)
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
func (it *AvsGovernanceSetOthenticRegistryIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetOthenticRegistryIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetOthenticRegistry represents a SetOthenticRegistry event raised by the AvsGovernance contract.
type AvsGovernanceSetOthenticRegistry struct {
	OthenticRegistry common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterSetOthenticRegistry is a free log retrieval operation binding the contract event 0xf9855cc914fefc396bdeb5a4dcb97a2f6c75f4d6f00a8e71d6f9a40e474afe8d.
//
// Solidity: event SetOthenticRegistry(address othenticRegistry)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetOthenticRegistry(opts *bind.FilterOpts) (*AvsGovernanceSetOthenticRegistryIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetOthenticRegistry")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetOthenticRegistryIterator{contract: _AvsGovernance.contract, event: "SetOthenticRegistry", logs: logs, sub: sub}, nil
}

// WatchSetOthenticRegistry is a free log subscription operation binding the contract event 0xf9855cc914fefc396bdeb5a4dcb97a2f6c75f4d6f00a8e71d6f9a40e474afe8d.
//
// Solidity: event SetOthenticRegistry(address othenticRegistry)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetOthenticRegistry(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetOthenticRegistry) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetOthenticRegistry")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetOthenticRegistry)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetOthenticRegistry", log); err != nil {
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

// ParseSetOthenticRegistry is a log parse operation binding the contract event 0xf9855cc914fefc396bdeb5a4dcb97a2f6c75f4d6f00a8e71d6f9a40e474afe8d.
//
// Solidity: event SetOthenticRegistry(address othenticRegistry)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetOthenticRegistry(log types.Log) (*AvsGovernanceSetOthenticRegistry, error) {
	event := new(AvsGovernanceSetOthenticRegistry)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetOthenticRegistry", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetRewardsReceiverIterator is returned from FilterSetRewardsReceiver and is used to iterate over the raw logs and unpacked data for SetRewardsReceiver events raised by the AvsGovernance contract.
type AvsGovernanceSetRewardsReceiverIterator struct {
	Event *AvsGovernanceSetRewardsReceiver // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetRewardsReceiverIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetRewardsReceiver)
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
		it.Event = new(AvsGovernanceSetRewardsReceiver)
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
func (it *AvsGovernanceSetRewardsReceiverIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetRewardsReceiverIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetRewardsReceiver represents a SetRewardsReceiver event raised by the AvsGovernance contract.
type AvsGovernanceSetRewardsReceiver struct {
	Operator common.Address
	Receiver common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSetRewardsReceiver is a free log retrieval operation binding the contract event 0xe906feea2ef60b474e22b4169bdd4de6906a84cd448cbcee99593526fe87082d.
//
// Solidity: event SetRewardsReceiver(address operator, address receiver)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetRewardsReceiver(opts *bind.FilterOpts) (*AvsGovernanceSetRewardsReceiverIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetRewardsReceiver")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetRewardsReceiverIterator{contract: _AvsGovernance.contract, event: "SetRewardsReceiver", logs: logs, sub: sub}, nil
}

// WatchSetRewardsReceiver is a free log subscription operation binding the contract event 0xe906feea2ef60b474e22b4169bdd4de6906a84cd448cbcee99593526fe87082d.
//
// Solidity: event SetRewardsReceiver(address operator, address receiver)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetRewardsReceiver(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetRewardsReceiver) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetRewardsReceiver")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetRewardsReceiver)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetRewardsReceiver", log); err != nil {
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

// ParseSetRewardsReceiver is a log parse operation binding the contract event 0xe906feea2ef60b474e22b4169bdd4de6906a84cd448cbcee99593526fe87082d.
//
// Solidity: event SetRewardsReceiver(address operator, address receiver)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetRewardsReceiver(log types.Log) (*AvsGovernanceSetRewardsReceiver, error) {
	event := new(AvsGovernanceSetRewardsReceiver)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetRewardsReceiver", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetRewardsReceiverModificationDelayIterator is returned from FilterSetRewardsReceiverModificationDelay and is used to iterate over the raw logs and unpacked data for SetRewardsReceiverModificationDelay events raised by the AvsGovernance contract.
type AvsGovernanceSetRewardsReceiverModificationDelayIterator struct {
	Event *AvsGovernanceSetRewardsReceiverModificationDelay // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetRewardsReceiverModificationDelayIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetRewardsReceiverModificationDelay)
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
		it.Event = new(AvsGovernanceSetRewardsReceiverModificationDelay)
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
func (it *AvsGovernanceSetRewardsReceiverModificationDelayIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetRewardsReceiverModificationDelayIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetRewardsReceiverModificationDelay represents a SetRewardsReceiverModificationDelay event raised by the AvsGovernance contract.
type AvsGovernanceSetRewardsReceiverModificationDelay struct {
	ModificationDelay *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterSetRewardsReceiverModificationDelay is a free log retrieval operation binding the contract event 0x47c8c3268759fc47868c5e319217a2e85d47bd3935a4108debe246f6025fb88b.
//
// Solidity: event SetRewardsReceiverModificationDelay(uint256 modificationDelay)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetRewardsReceiverModificationDelay(opts *bind.FilterOpts) (*AvsGovernanceSetRewardsReceiverModificationDelayIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetRewardsReceiverModificationDelay")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetRewardsReceiverModificationDelayIterator{contract: _AvsGovernance.contract, event: "SetRewardsReceiverModificationDelay", logs: logs, sub: sub}, nil
}

// WatchSetRewardsReceiverModificationDelay is a free log subscription operation binding the contract event 0x47c8c3268759fc47868c5e319217a2e85d47bd3935a4108debe246f6025fb88b.
//
// Solidity: event SetRewardsReceiverModificationDelay(uint256 modificationDelay)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetRewardsReceiverModificationDelay(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetRewardsReceiverModificationDelay) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetRewardsReceiverModificationDelay")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetRewardsReceiverModificationDelay)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetRewardsReceiverModificationDelay", log); err != nil {
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

// ParseSetRewardsReceiverModificationDelay is a log parse operation binding the contract event 0x47c8c3268759fc47868c5e319217a2e85d47bd3935a4108debe246f6025fb88b.
//
// Solidity: event SetRewardsReceiverModificationDelay(uint256 modificationDelay)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetRewardsReceiverModificationDelay(log types.Log) (*AvsGovernanceSetRewardsReceiverModificationDelay, error) {
	event := new(AvsGovernanceSetRewardsReceiverModificationDelay)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetRewardsReceiverModificationDelay", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetStrategyMultiplierIterator is returned from FilterSetStrategyMultiplier and is used to iterate over the raw logs and unpacked data for SetStrategyMultiplier events raised by the AvsGovernance contract.
type AvsGovernanceSetStrategyMultiplierIterator struct {
	Event *AvsGovernanceSetStrategyMultiplier // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetStrategyMultiplierIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetStrategyMultiplier)
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
		it.Event = new(AvsGovernanceSetStrategyMultiplier)
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
func (it *AvsGovernanceSetStrategyMultiplierIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetStrategyMultiplierIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetStrategyMultiplier represents a SetStrategyMultiplier event raised by the AvsGovernance contract.
type AvsGovernanceSetStrategyMultiplier struct {
	Strategy   common.Address
	Multiplier *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSetStrategyMultiplier is a free log retrieval operation binding the contract event 0x8ae53ffd0ebc018acb19342fba690554d49ae9a467a9606a38b49cb5ad775c81.
//
// Solidity: event SetStrategyMultiplier(address strategy, uint256 multiplier)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetStrategyMultiplier(opts *bind.FilterOpts) (*AvsGovernanceSetStrategyMultiplierIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetStrategyMultiplier")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetStrategyMultiplierIterator{contract: _AvsGovernance.contract, event: "SetStrategyMultiplier", logs: logs, sub: sub}, nil
}

// WatchSetStrategyMultiplier is a free log subscription operation binding the contract event 0x8ae53ffd0ebc018acb19342fba690554d49ae9a467a9606a38b49cb5ad775c81.
//
// Solidity: event SetStrategyMultiplier(address strategy, uint256 multiplier)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetStrategyMultiplier(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetStrategyMultiplier) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetStrategyMultiplier")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetStrategyMultiplier)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetStrategyMultiplier", log); err != nil {
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

// ParseSetStrategyMultiplier is a log parse operation binding the contract event 0x8ae53ffd0ebc018acb19342fba690554d49ae9a467a9606a38b49cb5ad775c81.
//
// Solidity: event SetStrategyMultiplier(address strategy, uint256 multiplier)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetStrategyMultiplier(log types.Log) (*AvsGovernanceSetStrategyMultiplier, error) {
	event := new(AvsGovernanceSetStrategyMultiplier)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetStrategyMultiplier", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetSupportedStrategiesIterator is returned from FilterSetSupportedStrategies and is used to iterate over the raw logs and unpacked data for SetSupportedStrategies events raised by the AvsGovernance contract.
type AvsGovernanceSetSupportedStrategiesIterator struct {
	Event *AvsGovernanceSetSupportedStrategies // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetSupportedStrategiesIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetSupportedStrategies)
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
		it.Event = new(AvsGovernanceSetSupportedStrategies)
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
func (it *AvsGovernanceSetSupportedStrategiesIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetSupportedStrategiesIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetSupportedStrategies represents a SetSupportedStrategies event raised by the AvsGovernance contract.
type AvsGovernanceSetSupportedStrategies struct {
	Strategies []common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSetSupportedStrategies is a free log retrieval operation binding the contract event 0xf009a6ffded424f714e8904d643a1ea4479453188faf08a3996121996b76684f.
//
// Solidity: event SetSupportedStrategies(address[] strategies)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetSupportedStrategies(opts *bind.FilterOpts) (*AvsGovernanceSetSupportedStrategiesIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetSupportedStrategies")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetSupportedStrategiesIterator{contract: _AvsGovernance.contract, event: "SetSupportedStrategies", logs: logs, sub: sub}, nil
}

// WatchSetSupportedStrategies is a free log subscription operation binding the contract event 0xf009a6ffded424f714e8904d643a1ea4479453188faf08a3996121996b76684f.
//
// Solidity: event SetSupportedStrategies(address[] strategies)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetSupportedStrategies(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetSupportedStrategies) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetSupportedStrategies")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetSupportedStrategies)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetSupportedStrategies", log); err != nil {
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

// ParseSetSupportedStrategies is a log parse operation binding the contract event 0xf009a6ffded424f714e8904d643a1ea4479453188faf08a3996121996b76684f.
//
// Solidity: event SetSupportedStrategies(address[] strategies)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetSupportedStrategies(log types.Log) (*AvsGovernanceSetSupportedStrategies, error) {
	event := new(AvsGovernanceSetSupportedStrategies)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetSupportedStrategies", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AvsGovernanceSetTokenIterator is returned from FilterSetToken and is used to iterate over the raw logs and unpacked data for SetToken events raised by the AvsGovernance contract.
type AvsGovernanceSetTokenIterator struct {
	Event *AvsGovernanceSetToken // Event containing the contract specifics and raw log

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
func (it *AvsGovernanceSetTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AvsGovernanceSetToken)
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
		it.Event = new(AvsGovernanceSetToken)
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
func (it *AvsGovernanceSetTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AvsGovernanceSetTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AvsGovernanceSetToken represents a SetToken event raised by the AvsGovernance contract.
type AvsGovernanceSetToken struct {
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSetToken is a free log retrieval operation binding the contract event 0xefc1fd16ea80a922086ee4e995739d59b025c1bcea6d1f67855747480c83214b.
//
// Solidity: event SetToken(address token)
func (_AvsGovernance *AvsGovernanceFilterer) FilterSetToken(opts *bind.FilterOpts) (*AvsGovernanceSetTokenIterator, error) {

	logs, sub, err := _AvsGovernance.contract.FilterLogs(opts, "SetToken")
	if err != nil {
		return nil, err
	}
	return &AvsGovernanceSetTokenIterator{contract: _AvsGovernance.contract, event: "SetToken", logs: logs, sub: sub}, nil
}

// WatchSetToken is a free log subscription operation binding the contract event 0xefc1fd16ea80a922086ee4e995739d59b025c1bcea6d1f67855747480c83214b.
//
// Solidity: event SetToken(address token)
func (_AvsGovernance *AvsGovernanceFilterer) WatchSetToken(opts *bind.WatchOpts, sink chan<- *AvsGovernanceSetToken) (event.Subscription, error) {

	logs, sub, err := _AvsGovernance.contract.WatchLogs(opts, "SetToken")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AvsGovernanceSetToken)
				if err := _AvsGovernance.contract.UnpackLog(event, "SetToken", log); err != nil {
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

// ParseSetToken is a log parse operation binding the contract event 0xefc1fd16ea80a922086ee4e995739d59b025c1bcea6d1f67855747480c83214b.
//
// Solidity: event SetToken(address token)
func (_AvsGovernance *AvsGovernanceFilterer) ParseSetToken(log types.Log) (*AvsGovernanceSetToken, error) {
	event := new(AvsGovernanceSetToken)
	if err := _AvsGovernance.contract.UnpackLog(event, "SetToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
