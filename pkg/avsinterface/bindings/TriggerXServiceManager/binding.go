// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractTriggerXServiceManager

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

// BN254G1Point is an auto generated low-level Go binding around an user-defined struct.
type BN254G1Point struct {
	X *big.Int
	Y *big.Int
}

// BN254G2Point is an auto generated low-level Go binding around an user-defined struct.
type BN254G2Point struct {
	X [2]*big.Int
	Y [2]*big.Int
}

// IBLSSignatureCheckerNonSignerStakesAndSignature is an auto generated low-level Go binding around an user-defined struct.
type IBLSSignatureCheckerNonSignerStakesAndSignature struct {
	NonSignerQuorumBitmapIndices []uint32
	NonSignerPubkeys             []BN254G1Point
	QuorumApks                   []BN254G1Point
	ApkG2                        BN254G2Point
	Sigma                        BN254G1Point
	QuorumApkIndices             []uint32
	TotalStakeIndices            []uint32
	NonSignerStakeIndices        [][]uint32
}

// IBLSSignatureCheckerQuorumStakeTotals is an auto generated low-level Go binding around an user-defined struct.
type IBLSSignatureCheckerQuorumStakeTotals struct {
	SignedStakeForQuorum []*big.Int
	TotalStakeForQuorum  []*big.Int
}

// IRewardsCoordinatorOperatorDirectedRewardsSubmission is an auto generated low-level Go binding around an user-defined struct.
type IRewardsCoordinatorOperatorDirectedRewardsSubmission struct {
	StrategiesAndMultipliers []IRewardsCoordinatorStrategyAndMultiplier
	Token                    common.Address
	OperatorRewards          []IRewardsCoordinatorOperatorReward
	StartTimestamp           uint32
	Duration                 uint32
	Description              string
}

// IRewardsCoordinatorOperatorReward is an auto generated low-level Go binding around an user-defined struct.
type IRewardsCoordinatorOperatorReward struct {
	Operator common.Address
	Amount   *big.Int
}

// IRewardsCoordinatorRewardsSubmission is an auto generated low-level Go binding around an user-defined struct.
type IRewardsCoordinatorRewardsSubmission struct {
	StrategiesAndMultipliers []IRewardsCoordinatorStrategyAndMultiplier
	Token                    common.Address
	Amount                   *big.Int
	StartTimestamp           uint32
	Duration                 uint32
}

// IRewardsCoordinatorStrategyAndMultiplier is an auto generated low-level Go binding around an user-defined struct.
type IRewardsCoordinatorStrategyAndMultiplier struct {
	Strategy   common.Address
	Multiplier *big.Int
}

// ISignatureUtilsSignatureWithSaltAndExpiry is an auto generated low-level Go binding around an user-defined struct.
type ISignatureUtilsSignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    *big.Int
}

// ContractTriggerXServiceManagerMetaData contains all meta data concerning the ContractTriggerXServiceManager contract.
var ContractTriggerXServiceManagerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIAVSDirectory\",\"name\":\"_avsDirectory\",\"type\":\"address\"},{\"internalType\":\"contractIRewardsCoordinator\",\"name\":\"_rewardsCoordinator\",\"type\":\"address\"},{\"internalType\":\"contractIRegistryCoordinator\",\"name\":\"_registryCoordinator\",\"type\":\"address\"},{\"internalType\":\"contractIStakeRegistry\",\"name\":\"_stakeRegistry\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"KeeperAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"KeeperBlacklisted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"KeeperRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"KeeperUnblacklisted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newPausedStatus\",\"type\":\"uint256\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"contractIPauserRegistry\",\"name\":\"pauserRegistry\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"contractIPauserRegistry\",\"name\":\"newPauserRegistry\",\"type\":\"address\"}],\"name\":\"PauserRegistrySet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldQuorumManager\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newQuorumManager\",\"type\":\"address\"}],\"name\":\"QuorumManagerSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"prevRewardsInitiator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newRewardsInitiator\",\"type\":\"address\"}],\"name\":\"RewardsInitiatorUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"value\",\"type\":\"bool\"}],\"name\":\"StaleStakesForbiddenUpdate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldTaskManager\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newTaskManager\",\"type\":\"address\"}],\"name\":\"TaskManagerContractUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldTaskManager\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newTaskManager\",\"type\":\"address\"}],\"name\":\"TaskManagerSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldTaskValidator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newTaskValidator\",\"type\":\"address\"}],\"name\":\"TaskValidatorSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newPausedStatus\",\"type\":\"uint256\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"avsDirectory\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"}],\"name\":\"blacklistKeeper\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"blsApkRegistry\",\"outputs\":[{\"internalType\":\"contractIBLSApkRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"quorumNumbers\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"referenceBlockNumber\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"uint32[]\",\"name\":\"nonSignerQuorumBitmapIndices\",\"type\":\"uint32[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point[]\",\"name\":\"nonSignerPubkeys\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point[]\",\"name\":\"quorumApks\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structBN254.G2Point\",\"name\":\"apkG2\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point\",\"name\":\"sigma\",\"type\":\"tuple\"},{\"internalType\":\"uint32[]\",\"name\":\"quorumApkIndices\",\"type\":\"uint32[]\"},{\"internalType\":\"uint32[]\",\"name\":\"totalStakeIndices\",\"type\":\"uint32[]\"},{\"internalType\":\"uint32[][]\",\"name\":\"nonSignerStakeIndices\",\"type\":\"uint32[][]\"}],\"internalType\":\"structIBLSSignatureChecker.NonSignerStakesAndSignature\",\"name\":\"params\",\"type\":\"tuple\"}],\"name\":\"checkSignatures\",\"outputs\":[{\"components\":[{\"internalType\":\"uint96[]\",\"name\":\"signedStakeForQuorum\",\"type\":\"uint96[]\"},{\"internalType\":\"uint96[]\",\"name\":\"totalStakeForQuorum\",\"type\":\"uint96[]\"}],\"internalType\":\"structIBLSSignatureChecker.QuorumStakeTotals\",\"name\":\"\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"contractIStrategy\",\"name\":\"strategy\",\"type\":\"address\"},{\"internalType\":\"uint96\",\"name\":\"multiplier\",\"type\":\"uint96\"}],\"internalType\":\"structIRewardsCoordinator.StrategyAndMultiplier[]\",\"name\":\"strategiesAndMultipliers\",\"type\":\"tuple[]\"},{\"internalType\":\"contractIERC20\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"startTimestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"duration\",\"type\":\"uint32\"}],\"internalType\":\"structIRewardsCoordinator.RewardsSubmission[]\",\"name\":\"rewardsSubmissions\",\"type\":\"tuple[]\"}],\"name\":\"createAVSRewardsSubmission\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"contractIStrategy\",\"name\":\"strategy\",\"type\":\"address\"},{\"internalType\":\"uint96\",\"name\":\"multiplier\",\"type\":\"uint96\"}],\"internalType\":\"structIRewardsCoordinator.StrategyAndMultiplier[]\",\"name\":\"strategiesAndMultipliers\",\"type\":\"tuple[]\"},{\"internalType\":\"contractIERC20\",\"name\":\"token\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIRewardsCoordinator.OperatorReward[]\",\"name\":\"operatorRewards\",\"type\":\"tuple[]\"},{\"internalType\":\"uint32\",\"name\":\"startTimestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"duration\",\"type\":\"uint32\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"}],\"internalType\":\"structIRewardsCoordinator.OperatorDirectedRewardsSubmission[]\",\"name\":\"operatorDirectedRewardsSubmissions\",\"type\":\"tuple[]\"}],\"name\":\"createOperatorDirectedAVSRewardsSubmission\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"delegation\",\"outputs\":[{\"internalType\":\"contractIDelegationManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"deregisterKeeperFromTriggerX\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"deregisterOperatorFromAVS\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"getOperatorRestakedStrategies\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRestakeableStrategies\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractITriggerXTaskManager\",\"name\":\"_taskManagerContract\",\"type\":\"address\"},{\"internalType\":\"contractIPauserRegistry\",\"name\":\"_pauserRegistry\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_initialPausedStatus\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"initialOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"rewardsInitiator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_taskManager\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_taskValidator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_quorumManager\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"isBlackListed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newPausedStatus\",\"type\":\"uint256\"}],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pauseAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"index\",\"type\":\"uint8\"}],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pauserRegistry\",\"outputs\":[{\"internalType\":\"contractIPauserRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"quorumManager\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"expiry\",\"type\":\"uint256\"}],\"internalType\":\"structISignatureUtils.SignatureWithSaltAndExpiry\",\"name\":\"operatorSignature\",\"type\":\"tuple\"}],\"name\":\"registerKeeperToTriggerX\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"expiry\",\"type\":\"uint256\"}],\"internalType\":\"structISignatureUtils.SignatureWithSaltAndExpiry\",\"name\":\"operatorSignature\",\"type\":\"tuple\"}],\"name\":\"registerOperatorToAVS\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registryCoordinator\",\"outputs\":[{\"internalType\":\"contractIRegistryCoordinator\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardsInitiator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"claimer\",\"type\":\"address\"}],\"name\":\"setClaimerFor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIPauserRegistry\",\"name\":\"newPauserRegistry\",\"type\":\"address\"}],\"name\":\"setPauserRegistry\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_quorumManager\",\"type\":\"address\"}],\"name\":\"setQuorumManager\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newRewardsInitiator\",\"type\":\"address\"}],\"name\":\"setRewardsInitiator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"value\",\"type\":\"bool\"}],\"name\":\"setStaleStakesForbidden\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_taskManager\",\"type\":\"address\"}],\"name\":\"setTaskManager\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_taskValidator\",\"type\":\"address\"}],\"name\":\"setTaskValidator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakeRegistry\",\"outputs\":[{\"internalType\":\"contractIStakeRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"staleStakesForbidden\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"taskManager\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"taskManagerContract\",\"outputs\":[{\"internalType\":\"contractITriggerXTaskManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"taskValidator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point\",\"name\":\"apk\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structBN254.G2Point\",\"name\":\"apkG2\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point\",\"name\":\"sigma\",\"type\":\"tuple\"}],\"name\":\"trySignatureAndApkVerification\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"pairingSuccessful\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"siganatureIsValid\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"}],\"name\":\"unblacklistKeeper\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newPausedStatus\",\"type\":\"uint256\"}],\"name\":\"unpause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_metadataURI\",\"type\":\"string\"}],\"name\":\"updateAVSMetadataURI\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_taskManager\",\"type\":\"address\"}],\"name\":\"updateTaskManager\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052600436101561001257600080fd5b60003560e01c806310d67a2f146102c7578063136439dd146102c2578063171f1d5b146102bd57806326a965f0146102b8578063327d0a60146102b357806333cfb7b7146102ae5780633bc28c8c146102a9578063416c7e5e146102a4578063500c8dd31461029f57806358ac4a1e1461029a578063595c6a67146102955780635ac86ab7146102905780635c975abb1461028b5780635df45946146102865780635edd78ea14610281578063683048351461027c5780636b3aa72e146102775780636cf5ca21146102725780636d14a9871461026d5780636efb463614610268578063715018a614610263578063763ac9571461025e57806383a9432214610259578063886f1195146102545780638da5cb5b1461024f5780639926ee7d1461024a5780639cd0fb8614610245578063a0169ddd14610240578063a20b99bf1461023b578063a364f4da14610236578063a41d3f9414610231578063a50a640e1461022c578063a98fb35514610227578063b5f7eb6b14610222578063b98d09081461021d578063df5cf72314610218578063e47d606014610213578063e481af9d1461020e578063f2fde38b14610209578063fabc1cbc14610204578063fc299dee146101ff578063fce36c7d146101fa5763fd38ec8c146101f557600080fd5b61206c565b611e99565b611e70565b611d61565b611cd0565b611cb4565b611c72565b611c2d565b611c0a565b611be0565b611b22565b611af9565b611a8f565b6119d0565b611744565b611650565b6115e1565b611525565b6114fc565b6114d3565b611396565b61129a565b611159565b6110c4565b610d70565b610c56565b610c11565b610bcc565b610b5e565b610b19565b610afb565b610ac8565b610a33565b6109c4565b61099a565b61085b565b610824565b6107ec565b61073b565b6106ce565b610666565b610387565b6102e2565b6001600160a01b038116036102dd57565b600080fd5b346102dd5760203660031901126102dd5760048035610300816102cc565b60c95460405163755b36bd60e11b81529260209184919082906001600160a01b03165afa918215610382576103519261034c91600091610353575b506001600160a01b031633146120b9565b613e39565b005b610375915060203d60201161037b575b61036d8183610513565b810190612095565b3861033b565b503d610363565b6120ad565b346102dd5760203660031901126102dd5760043560c95460405163237dfb4760e11b815233600482015290602090829060249082906001600160a01b03165afa8015610382576103df91600091610493575b5061212d565b60ca5481811603610428578060ca557fab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d604051806104233394829190602083019252565b0390a2005b60405162461bcd60e51b815260206004820152603860248201527f5061757361626c652e70617573653a20696e76616c696420617474656d70742060448201527f746f20756e70617573652066756e6374696f6e616c69747900000000000000006064820152608490fd5b6104b5915060203d6020116104bb575b6104ad8183610513565b810190612118565b386103d9565b503d6104a3565b634e487b7160e01b600052604160045260246000fd5b604081019081106001600160401b038211176104f357604052565b6104c2565b606081019081106001600160401b038211176104f357604052565b90601f801991011681019081106001600160401b038211176104f357604052565b6040519061054461010083610513565b565b60405190610544604083610513565b906105446040519283610513565b60409060e31901126102dd576040519061057c826104d8565b60e4358252610104356020830152565b91908260409103126102dd576040516105a4816104d8565b6020808294803584520135910152565b9080601f830112156102dd5760408051926105cf8285610513565b839181019283116102dd57905b8282106105e95750505090565b81358152602091820191016105dc565b9060806063198301126102dd57604051610612816104d8565b602061062d82946106248160646105b4565b845260a46105b4565b910152565b91906080838203126102dd57602061062d60405192610650846104d8565b6040849661065e83826105b4565b8652016105b4565b346102dd576101203660031901126102dd5760043560403660231901126102dd576106be6040918251610698816104d8565b602435815260443560208201526106ae366105f9565b906106b836610563565b926121cc565b8251911515825215156020820152f35b346102dd5760203660031901126102dd576004356106eb816102cc565b6106f36143d4565b6001600160a01b0316600081815260fd60205260408120805460ff191660011790557f79bf5efab685b39bdd2b7b1aba09ffa5e9dfbd86f7aaebab4b52d7ef679d28449080a2005b346102dd5760203660031901126102dd57600435610758816102cc565b6107606143d4565b60fe80546001600160a01b039283166001600160a01b0319821681179092559091167fdae652729007c1e5ea13dfe9e126a1b72358db8db69c68b96f316dcea59c1c5f600080a3005b602060408183019282815284518094520192019060005b8181106107cd5750505090565b82516001600160a01b03168452602093840193909201916001016107c0565b346102dd5760203660031901126102dd5761082061081460043561080f816102cc565b61243d565b604051918291826107a9565b0390f35b346102dd5760203660031901126102dd57610351600435610844816102cc565b61084c6143d4565b6144cc565b801515036102dd57565b346102dd5760203660031901126102dd5760043561087881610851565b604051638da5cb5b60e01b81526020816004817f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be75770386001600160a01b03165afa90811561038257600091610970575b506001600160a01b031633036108df576103519061452a565b60405162461bcd60e51b815260206004820152605c60248201527f424c535369676e6174757265436865636b65722e6f6e6c79436f6f7264696e6160448201527f746f724f776e65723a2063616c6c6572206973206e6f7420746865206f776e6560648201527f72206f6620746865207265676973747279436f6f7264696e61746f7200000000608482015260a490fd5b610989915060203d60201161037b5761036d8183610513565b386108c6565b60009103126102dd57565b346102dd5760003660031901126102dd57610101546040516001600160a01b039091168152602090f35b346102dd5760203660031901126102dd576004356109e1816102cc565b6109e96143d4565b61010180546001600160a01b039283166001600160a01b0319821681179092559091167fecc36ab0616c35e413d90730ada58214cbe0188c5f26632fbb32451c9a135784600080a3005b346102dd5760003660031901126102dd5760c95460405163237dfb4760e11b815233600482015290602090829060249082906001600160a01b03165afa801561038257610a8791600091610493575061212d565b60001960ca5560405160001981527fab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d60203392a2005b60ff8116036102dd57565b346102dd5760203660031901126102dd576020600160ff600435610aeb81610abd565b161b8060ca541614604051908152f35b346102dd5760003660031901126102dd57602060ca54604051908152f35b346102dd5760003660031901126102dd576040517f000000000000000000000000d569dd2958373e2d880089ac9892238bc91ed8fe6001600160a01b03168152602090f35b346102dd5760203660031901126102dd57600435610b7b816102cc565b610b836143d4565b60ff80546001600160a01b039283166001600160a01b0319821681179092559091167f86cebbccc5bbc9ac544f7e5361e6c5e4acf14f39d85c855032a815125cb97fd8600080a3005b346102dd5760003660031901126102dd576040517f0000000000000000000000006ad0086321a7a247887982c4b5bfbb3005e0f1986001600160a01b03168152602090f35b346102dd5760003660031901126102dd576040517f000000000000000000000000055733000064333caddbc92763c58bf0192ffebf6001600160a01b03168152602090f35b346102dd5760203660031901126102dd57600435610c73816102cc565b7f000000000000000000000000055733000064333caddbc92763c58bf0192ffebf6001600160a01b0316803b156102dd576040516351b27a6d60e11b81526001600160a01b0383166004820152906000908290602490829084905af1801561038257610d55575b50610cfd610cf86001600160a01b0383165b6001600160a01b031690565b61495b565b506001600160a01b038116600090815260fd60205260409020610d25905b805460ff19169055565b6001600160a01b03167fa7a775c2c8141f7985c111748ec31c11e5e44b83528e105c8d1d4e8e6b81cf80600080a2005b80610d646000610d6a93610513565b8061098f565b38610cda565b346102dd5760003660031901126102dd576040517f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be75770386001600160a01b03168152602090f35b6044359063ffffffff821682036102dd57565b359063ffffffff821682036102dd57565b6001600160401b0381116104f35760051b60200190565b9080601f830112156102dd578135610e0781610dd9565b92610e156040519485610513565b81845260208085019260051b8201019283116102dd57602001905b828210610e3d5750505090565b60208091610e4a84610dc8565b815201910190610e30565b81601f820112156102dd578035610e6b81610dd9565b92610e796040519485610513565b81845260208085019260061b840101928184116102dd57602001915b838310610ea3575050505090565b6020604091610eb2848661058c565b815201920191610e95565b9080601f830112156102dd578135610ed481610dd9565b92610ee26040519485610513565b81845260208085019260051b820101918383116102dd5760208201905b838210610f0e57505050505090565b81356001600160401b0381116102dd57602091610f3087848094880101610df0565b815201910190610eff565b919091610180818403126102dd57610f51610534565b9281356001600160401b0381116102dd5781610f6e918401610df0565b845260208201356001600160401b0381116102dd5781610f8f918401610e55565b602085015260408201356001600160401b0381116102dd5781610fb3918401610e55565b6040850152610fc58160608401610632565b6060850152610fd78160e0840161058c565b60808501526101208201356001600160401b0381116102dd5781610ffc918401610df0565b60a08501526101408201356001600160401b0381116102dd5781611021918401610df0565b60c08501526101608201356001600160401b0381116102dd576110449201610ebd565b60e0830152565b906020808351928381520192019060005b8181106110695750505090565b82516001600160601b031684526020938401939092019160010161105c565b9291906110bf60209160408652826110ab82516040808a0152608089019061104b565b910151868203603f1901606088015261104b565b930152565b346102dd5760803660031901126102dd576004356024356001600160401b0381116102dd57366023820112156102dd5780600401356001600160401b0381116102dd5736602482840101116102dd5761111b610db5565b90606435936001600160401b0385116102dd576024611141611149963690600401610f3b565b940190612c75565b9061082060405192839283611088565b346102dd5760003660031901126102dd576111726143d4565b603380546001600160a01b031981169091556000906001600160a01b03167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e08280a3005b6001600160401b0381116104f357601f01601f191660200190565b9291926111dd826111b6565b916111eb6040519384610513565b8294818452818301116102dd578281602093846000960137010152565b9060406003198301126102dd57600435611221816102cc565b91602435906001600160401b0382116102dd57606082820360031901126102dd576040519161124f836104f8565b80600401356001600160401b0381116102dd578101826023820112156102dd57604492816024600461128494013591016111d1565b8352602481013560208401520135604082015290565b346102dd576112a836611208565b7f000000000000000000000000055733000064333caddbc92763c58bf0192ffebf6001600160a01b0316803b156102dd57604051639926ee7d60e01b8152916000918391829084908290611300908960048401613592565b03925af180156103825761137c575b5061132a6113256001600160a01b038316610cec565b614b6d565b506001600160a01b038116600090815260fd6020526040902061134c90610d1b565b6001600160a01b03167f1584773458d98c71b34a270ee1100b3a42889bf91e3b7a858563b684c24d838e600080a2005b8061138960008093610513565b8003126102dd573861130f565b346102dd576101003660031901126102dd576004356113b4816102cc565b6114476024356113c3816102cc565b6044356064356113d2816102cc565b6084356113de816102cc565b60a435916113eb836102cc565b60c435936113f8856102cc565b60e43595611405876102cc565b6000549861142b60ff8b60081c16151515809b816114c5575b81156114a5575b506135d0565b8961143e600160ff196000541617600055565b61148c57613633565b61144d57005b61145d61ff001960005416600055565b604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb384740249890602090a1005b6114a061010061ff00196000541617600055565b613633565b303b159150816114b7575b5038611425565b60ff166001149050386114b0565b600160ff821610915061141e565b346102dd5760003660031901126102dd5760c9546040516001600160a01b039091168152602090f35b346102dd5760003660031901126102dd576033546040516001600160a01b039091168152602090f35b346102dd5761153336611208565b611567337f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be75770386001600160a01b0316146137b7565b7f000000000000000000000000055733000064333caddbc92763c58bf0192ffebf6001600160a01b0316803b156102dd57604051639926ee7d60e01b815292839283916115b8919060048401613592565b918160008096819503925af180156103825782906115d35780f35b6115dc91610513565b388180f35b346102dd5760203660031901126102dd576004356115fe816102cc565b6116066143d4565b61010080546001600160a01b039283166001600160a01b0319821681179092559091167f5d68f69aebb8a1d75f8b2a87dddd6e8db842f9fa3bb63745ea7f49ea5bdf03a9600080a3005b346102dd57600060203660031901126116e75780600435611670816102cc565b6116786143d4565b7f000000000000000000000000acc1fb458a1317e886db376fc8141540537e68fe6001600160a01b0316803b156116e35760405163a0169ddd60e01b81526001600160a01b0390921660048301529091908290602490829084905af180156103825782906115d35780f35b5050fd5b80fd5b9060206003198301126102dd576004356001600160401b0381116102dd5760040160009280601f83011215611740578135936001600160401b0385116116e757506020808301928560051b0101116102dd579190565b8380fd5b346102dd57611752366116ea565b9061175b614852565b7f000000000000000000000000acc1fb458a1317e886db376fc8141540537e68fe6001600160a01b03169160005b8181106117dc5750823b156102dd576117bd9260009283604051809681958294634e5cd2fd60e11b84523060048501613999565b03925af18015610382576117cd57005b80610d64600061035193610513565b600091825b6117f96117ef848488613844565b6040810190613866565b90508410156118375761182c60019160206118248761181e6117ef89898d9e9d613844565b9061389b565b0135906123c2565b9301929493946117e1565b91909250611854610cec602061184e848789613844565b016138ab565b6040516323b872dd60e01b81523360048201523060248201526044810184905290602090829060649082906000905af18015610382576119b4575b506118a3610cec602061184e848789613844565b604051636eb1769f60e11b81523060048201526001600160a01b038716602482015290602090829060449082905afa80156103825760009361190861193b9285888a6020978a94611973575b5061184e61190293610cec938a93613844565b926123c2565b60405163095ea7b360e01b81526001600160a01b038a166004820152602481019190915294859283919082906044820190565b03925af191821561038257600192611955575b5001611789565b61196c9060203d81116104bb576104ad8183610513565b503861194e565b8891945061190293610cec936119a161184e93853d81116119ad575b6119998183610513565b8101906122b2565b969350935093506118ef565b503d61198f565b6119cb9060203d81116104bb576104ad8183610513565b61188f565b346102dd57600060203660031901126116e757806004356119f0816102cc565b611a24337f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be75770386001600160a01b0316146137b7565b7f000000000000000000000000055733000064333caddbc92763c58bf0192ffebf6001600160a01b0316803b156116e3576040516351b27a6d60e11b81526001600160a01b0390921660048301529091908290602490829084905af180156103825782906115d35780f35b346102dd5760203660031901126102dd57600435611aac816102cc565b611ab46143d4565b6001600160a01b0316600081815260fd60205260408120805460ff191690557f8511046d95f95689c89b4807d91820adbc266db891b2c5c25a42f812781dde579080a2005b346102dd5760003660031901126102dd5760fe546040516001600160a01b039091168152602090f35b346102dd57600060203660031901126116e757806004356001600160401b038111611bdd5736602382011215611bdd57611b669036906024816004013591016111d1565b611b6e6143d4565b7f000000000000000000000000055733000064333caddbc92763c58bf0192ffebf6001600160a01b031690813b156116e3578291611bca9160405194858094819363a98fb35560e01b8352602060048401526024830190613551565b03925af180156103825782906115d35780f35b50fd5b346102dd5760003660031901126102dd57610100546040516001600160a01b039091168152602090f35b346102dd5760003660031901126102dd57602060ff609754166040519015158152f35b346102dd5760003660031901126102dd576040517f000000000000000000000000a44151489861fe9e3055d95adc98fbd462b948e76001600160a01b03168152602090f35b346102dd5760203660031901126102dd57600435611c8f816102cc565b60018060a01b031660005260fd602052602060ff604060002054166040519015158152f35b346102dd5760003660031901126102dd57610820610814613b0b565b346102dd5760203660031901126102dd57600435611ced816102cc565b611cf56143d4565b6001600160a01b03811615611d0d5761035190614795565b60405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b6064820152608490fd5b346102dd5760203660031901126102dd5760c95460405163755b36bd60e11b81526004803592602091839182906001600160a01b03165afa801561038257611dbb9160009161035357506001600160a01b031633146120b9565b60ca54198119811603611e0557611dd18160ca55565b60405190815233907f3582d1828e26bf56bd801502bc021ac0bc8afb57c826e4986b45593c8fad389c908060208101610423565b60405162461bcd60e51b815260206004820152603860248201527f5061757361626c652e756e70617573653a20696e76616c696420617474656d7060448201527f7420746f2070617573652066756e6374696f6e616c69747900000000000000006064820152608490fd5b346102dd5760003660031901126102dd576065546040516001600160a01b039091168152602090f35b346102dd57611ea7366116ea565b90611eb0614852565b7f000000000000000000000000acc1fb458a1317e886db376fc8141540537e68fe6001600160a01b03169160005b818110611f115750823b156102dd576117bd926000928360405180968195829463fce36c7d60e01b845260048401613d63565b60006020611f67611f2a610cec8361184e87898b613d41565b6040611f3786888a613d41565b6040516323b872dd60e01b8152336004820152306024820152910135604482015293849283919082906064820190565b03925af1801561038257612050575b50611f8a610cec602061184e848688613d41565b604051636eb1769f60e11b81523060048201526001600160a01b03861660248201529190602090839060449082905afa801561038257611ffb6020916000948591612033575b50611908611fe6610cec8561184e888b8d613d41565b916040611ff4878a8c613d41565b01356123c2565b03925af191821561038257600192612015575b5001611ede565b61202c9060203d81116104bb576104ad8183610513565b503861200e565b61204a9150833d81116119ad576119998183610513565b38611fd0565b6120679060203d81116104bb576104ad8183610513565b611f76565b346102dd5760003660031901126102dd5760ff546040516001600160a01b039091168152602090f35b908160209103126102dd57516120aa816102cc565b90565b6040513d6000823e3d90fd5b156120c057565b60405162461bcd60e51b815260206004820152602a60248201527f6d73672e73656e646572206973206e6f74207065726d697373696f6e6564206160448201526939903ab73830bab9b2b960b11b6064820152608490fd5b908160209103126102dd57516120aa81610851565b1561213457565b60405162461bcd60e51b815260206004820152602860248201527f6d73672e73656e646572206973206e6f74207065726d697373696f6e6564206160448201526739903830bab9b2b960c11b6064820152608490fd5b634e487b7160e01b600052603260045260246000fd5b9060028110156121b15760051b0190565b61218a565b634e487b7160e01b600052601260045260246000fd5b6122a86122856122ae9561227f61227885875160208901518a515160208c51015160208d016020815151915101519189519360208b0151956040519760208901998a5260208a015260408901526060880152608087015260a086015260c085015260e084015261010083015261224f81610120840103601f198101835282610513565b5190207f30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001900690565b8096613f60565b90613fa8565b9261227f61229a612294614032565b94614129565b916122a3614250565b613f60565b9161429a565b9091565b908160209103126102dd575190565b908160209103126102dd57516001600160c01b03811681036102dd5790565b908160209103126102dd57516120aa81610abd565b60405190612304602083610513565b6000808352366020840137565b9061231b82610dd9565b6123286040519182610513565b8281528092612339601f1991610dd9565b0190602036910137565b8051156121b15760200190565b9081518110156121b1570160200190565b634e487b7160e01b600052601160045260246000fd5b906001820180921161238557565b612361565b906002820180921161238557565b906003820180921161238557565b906004820180921161238557565b906005820180921161238557565b9190820180921161238557565b6001600160601b038116036102dd57565b908160409103126102dd576020604051916123fa836104d8565b8051612405816102cc565b83520151612412816123cf565b602082015290565b80518210156121b15760209160051b010190565b60001981146123855760010190565b6040516309aa152760e11b81526001600160a01b0391821660048201527f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be757703890911690602081602481855afa908115610382576124be916020916000916127a7575b506040518093819263871ef04960e01b8352600483019190602083019252565b0381855afa90811561038257600091612778575b506001600160c01b0316908115908115612714575b5061270b576124f59061442c565b600091907f0000000000000000000000006ad0086321a7a247887982c4b5bfbb3005e0f1986001600160a01b031690835b81518510156125c35761257b60206125586125526125448987612350565b516001600160f81b03191690565b60f81c90565b604051633ca5a5f560e01b815260ff909116600482015291829081906024820190565b0381875afa80156103825760019261259b926000926125a3575b506123c2565b940193612526565b6125bc91925060203d81116119ad576119998183610513565b9038612595565b6125ce919450612311565b9260009060005b8151811015612705576125ee6125526125448385612350565b604051633ca5a5f560e01b815260ff8216600482015290602082602481895afa918215610382576000926126e5575b50906000915b818310612635575050506001016125d5565b604080516356e4026d60e11b815260ff83166004820152602481018590529396929391929190816044818b5afa918215610382576126a98b61269a83612694610cec6001986126ae986000916126b7575b50516001600160a01b031690565b9261241a565b6001600160a01b039091169052565b61242e565b95019190612623565b6126d8915060403d81116126de575b6126d08183610513565b8101906123e0565b38612686565b503d6126c6565b6126fe91925060203d81116119ad576119998183610513565b903861261d565b50505050565b506120aa6122f5565b604051639aa1653d60e01b81529150602090829060049082905afa80156103825760ff91600091612749575b501615386124e7565b61276b915060203d602011612771575b6127638183610513565b8101906122e0565b38612740565b503d612759565b61279a915060203d6020116127a0575b6127928183610513565b8101906122c1565b386124d2565b503d612788565b6127be9150823d84116119ad576119998183610513565b3861249e565b604051906127d1826104d8565b60606020838281520152565b156127e457565b60405162461bcd60e51b81526020600482015260376024820152600080516020614c5e83398151915260448201527f7265733a20656d7074792071756f72756d20696e7075740000000000000000006064820152608490fd5b1561284457565b60405162461bcd60e51b81526020600482015260416024820152600080516020614c5e83398151915260448201527f7265733a20696e7075742071756f72756d206c656e677468206d69736d6174636064820152600d60fb1b608482015260a490fd5b156128ae57565b60a460405162461bcd60e51b81526020600482015260446024820152600080516020614c5e83398151915260448201527f7265733a20696e707574206e6f6e7369676e6572206c656e677468206d69736d6064820152630c2e8c6d60e31b6084820152fd5b1561291a57565b60405162461bcd60e51b815260206004820152603c6024820152600080516020614c5e83398151915260448201527f7265733a20696e76616c6964207265666572656e636520626c6f636b000000006064820152608490fd5b60001981019190821161238557565b1561298957565b608460405162461bcd60e51b81526020600482015260406024820152600080516020614c5e83398151915260448201527f7265733a206e6f6e5369676e65725075626b657973206e6f7420736f727465646064820152fd5b908210156121b1570190565b156129f457565b60405162461bcd60e51b81526020600482015260666024820152600080516020614c5e83398151915260448201527f7265733a205374616b6552656769737472792075706461746573206d7573742060648201527f62652077697468696e207769746864726177616c44656c6179426c6f636b732060848201526577696e646f7760d01b60a482015260c490fd5b908160209103126102dd575167ffffffffffffffff19811681036102dd5790565b15612aaa57565b60405162461bcd60e51b81526020600482015260616024820152600080516020614c5e83398151915260448201527f7265733a2071756f72756d41706b206861736820696e2073746f72616765206460648201527f6f6573206e6f74206d617463682070726f76696465642071756f72756d2061706084820152606b60f81b60a482015260c490fd5b908160209103126102dd57516120aa816123cf565b906001600160601b03809116911603906001600160601b03821161238557565b15612b6f57565b60405162461bcd60e51b81526020600482015260436024820152600080516020614c5e83398151915260448201527f7265733a2070616972696e6720707265636f6d70696c652063616c6c206661696064820152621b195960ea1b608482015260a490fd5b15612bdb57565b60405162461bcd60e51b81526020600482015260396024820152600080516020614c5e83398151915260448201527f7265733a207369676e617475726520697320696e76616c6964000000000000006064820152608490fd5b60049163ffffffff60e01b9060e01b16815201602082519192019060005b818110612c5f5750505090565b8251845260209384019390920191600101612c52565b949392909193612c836127c4565b50612c8f8515156127dd565b604084015151851480613543575b80613535575b80613527575b612cb29061283d565b612cc4602085015151855151146128a7565b612cdb63ffffffff431663ffffffff841610612913565b612ce3610546565b600081526000602082015292612cf76127c4565b612d0087612311565b6020820152612d0e87612311565b8152612d186127c4565b92612d27602088015151612311565b8452612d37602088015151612311565b602085810191909152604051639aa1653d60e01b815290816004817f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be75770386001600160a01b03165afa801561038257612da191600091613508575b50612d9c368b876111d1565b614568565b986000965b60208901518051891015612f2157602088612e15612e0b8c612e038f96868e612de8612dd386809561241a565b51805160005260200151602052604060002090565b612df5848484015161241a565b5282612eee575b015161241a565b51955161241a565b5163ffffffff1690565b6040516304ec635160e01b8152600481019490945263ffffffff9182166024850152166044830152816064816001600160a01b037f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be7577038165afa9182156103825761227f8a612ec38f612ebc8f8460208f92612eb393612eab8460019e612ec99e600091612ed1575b508f8060c01b0316925161241a565b52015161241a565b51938d5161241a565b51166145ef565b90614622565b970196612da6565b612ee89150863d81116127a0576127928183610513565b38612e9c565b612f1c612efe848484015161241a565b51612f1584840151612f0f87612973565b9061241a565b5110612982565b612dfc565b50909597949650612f3691989392995061470b565b91612f4360975460ff1690565b9081156134ff576040516318891fd760e31b81526020816004817f000000000000000000000000a44151489861fe9e3055d95adc98fbd462b948e76001600160a01b03165afa908115610382576000916134e0575b5091905b6000925b81841061300457505050505092612fdd612fd8612fd1612ffe9585612ff098608060606020990151920151926121cc565b9190612b68565b612bd4565b0151604051928391602083019586612c34565b03601f198101835282610513565b51902090565b92989596909399919794878b888c888d6133dc575b612e0b8260a06130676125526130598461306f97613053613045612dd38f9c604060209f9e015161241a565b67ffffffffffffffff191690565b9b6129e1565b356001600160f81b03191690565b97015161241a565b604051631a2f32ab60e21b815260ff95909516600486015263ffffffff9182166024860152166044840152826064816001600160a01b037f000000000000000000000000d569dd2958373e2d880089ac9892238bc91ed8fe165afa90811561038257613134612e0b8f958f9061312c8f978f96848f61312660c09661311f848f60209f90612dfc613059996040936125529c6000916133ae575b5067ffffffffffffffff19918216911614612aa3565b5190613fa8565b9c6129e1565b96015161241a565b604051636414a62b60e11b815260ff94909416600485015263ffffffff9182166024850152166044830152816064816001600160a01b037f0000000000000000000000006ad0086321a7a247887982c4b5bfbb3005e0f198165afa908115610382576131c2918c8f9260009261338a575b5060206131b49293015161241a565b906001600160601b03169052565b6131ef8c6131b48c6131e86131db82602086015161241a565b516001600160601b031690565b925161241a565b600098895b60208a015151811015613371578b8d61323289613225612552613059868f8961321d915161241a565b5194876129e1565b60ff161c60019081161490565b613241575b50506001016131f4565b8a8a6132c9859f948f96866132838f9360e061327a612e0b956020613272612552613059839f6132899c89916129e1565b9a015161241a565b519b015161241a565b5161241a565b60405163795f4a5760e11b815260ff909316600484015263ffffffff93841660248401526044830196909652919094166064850152839081906084820190565b03817f0000000000000000000000006ad0086321a7a247887982c4b5bfbb3005e0f1986001600160a01b03165afa908115610382578f613330908f9360019594869560009261333b575b506126946131b49293519361332b6131db848761241a565b612b48565b019a90508b8d613237565b6131b492506133636126949160203d811161336a575b61335b8183610513565b810190612b33565b9250613313565b503d613351565b5093919796996001919699509a94929a01929190612fa0565b6131b492506133a7602091823d811161336a5761335b8183610513565b92506131a5565b60206133cf92503d81116133d5575b6133c78183610513565b810190612a82565b38613109565b503d6133bd565b61341994506133f6925061255291613059916020956129e1565b60405163124d062160e11b815260ff909116600482015291829081906024820190565b03817f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be75770386001600160a01b03165afa80156103825760208961306f8f938f60a08f976125526130598f8f90613053613045612dd38f60408b96918f6134a190612e0b9f8a956130679e6000926134b7575b5063ffffffff61349b929316926123c2565b116129ed565b5050505050509750505050505092935050613019565b602061349b935063ffffffff916134d9913d81116119ad576119998183610513565b9250613489565b6134f9915060203d6020116119ad576119998183610513565b38612f98565b60009190612f9c565b613521915060203d602011612771576127638183610513565b38612d90565b5060e0840151518514612ca9565b5060c0840151518514612ca3565b5060a0840151518514612c9d565b919082519283825260005b84811061357d575050826000602080949584010152601f8019910116010190565b8060208092840101518282860101520161355c565b9060018060a01b0316815260406020820152608060406135bd845160608386015260a0850190613551565b9360208101516060850152015191015290565b156135d757565b60405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b6064820152608490fd5b60c954919796949592916001600160a01b031615806137a5575b1561372a576136b261370b966136ee956136a46105449b866136716136d19860ca55565b60405190815233907fab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d90602090a2613e39565b6136ad81614795565b6147de565b60018060a01b03166001600160601b0360a01b61010154161761010155565b60018060a01b03166001600160601b0360a01b60fe54161760fe55565b60018060a01b03166001600160601b0360a01b60ff54161760ff55565b60018060a01b03166001600160601b0360a01b61010054161761010055565b60405162461bcd60e51b815260206004820152604760248201527f5061757361626c652e5f696e697469616c697a655061757365723a205f696e6960448201527f7469616c697a6550617573657228292063616e206f6e6c792062652063616c6c6064820152666564206f6e636560c81b608482015260a490fd5b506001600160a01b038816151561364d565b156137be57565b60405162461bcd60e51b815260206004820152605260248201527f536572766963654d616e61676572426173652e6f6e6c7952656769737472794360448201527f6f6f7264696e61746f723a2063616c6c6572206973206e6f742074686520726560648201527133b4b9ba393c9031b7b7b93234b730ba37b960711b608482015260a490fd5b91908110156121b15760051b8101359060be19813603018212156102dd570190565b903590601e19813603018212156102dd57018035906001600160401b0382116102dd57602001918160061b360383136102dd57565b91908110156121b15760061b0190565b356120aa816102cc565b9035601e19823603018112156102dd5701602081359101916001600160401b0382116102dd578160061b360383136102dd57565b9160209082815201919060005b8181106139035750505090565b9091926040806001928635613917816102cc565b848060a01b031681526001600160601b036020880135613936816123cf565b1660208201520194019291016138f6565b9035601e19823603018112156102dd5701602081359101916001600160401b0382116102dd5781360383136102dd57565b908060209392818452848401376000828201840152601f01601f1916010190565b6001600160a01b0390911681526040602082018190528101839052600583901b810160609081019383923684900360be190192600091908101905b8383106139e5575050505050505090565b90919293949596605f198282030183528735868112156102dd57870190613a1d613a0f83806138b5565b60c0845260c08401916138e9565b916020810135613a2c816102cc565b6001600160a01b0316602083810191909152613a4b60408301836138b5565b8486036040860152808652949091019360005b818110613ad757505050613ac6600193602093613ab884613a92613a856060899801610dc8565b63ffffffff166060850152565b613aae613aa160808301610dc8565b63ffffffff166080850152565b60a0810190613947565b9160a0818503910152613978565b9901930193019195949392906139d4565b9091946040806001928835613aeb816102cc565b848060a01b03168152602089013560208201520196019101919091613a5e565b604051639aa1653d60e01b81527f0000000000000000000000003fe0ebc2e40be247fefe4c4636a6477be75770386001600160a01b031690602081600481855afa80156103825760ff91600091613d22575b50168015613d18577f0000000000000000000000006ad0086321a7a247887982c4b5bfbb3005e0f1986001600160a01b03169060009081905b808310613cd35750613ba89150612311565b9260009060005b604051639aa1653d60e01b8152602081600481895afa80156103825760ff91600091613cb5575b5016811015613cae57604051633ca5a5f560e01b815260ff821660048201819052602082602481895afa91821561038257600092613c8e575b50906000915b818310613c2757505050600101613baf565b604080516356e4026d60e11b815260ff83166004820152602481018590529396929391929190816044818b5afa918215610382576126a98b61269a83612694610cec600198613c85986000916126b75750516001600160a01b031690565b95019190613c15565b613ca791925060203d81116119ad576119998183610513565b9038613c0f565b5092505050565b613ccd915060203d8111612771576127638183610513565b38613bd6565b604051633ca5a5f560e01b815260ff84166004820152909190602081602481885afa801561038257600192613d0f926000926125a357506123c2565b92019190613b96565b50506120aa6122f5565b613d3b915060203d602011612771576127638183610513565b38613b5d565b91908110156121b15760051b81013590609e19813603018212156102dd570190565b909180602083016020845252604082019260408260051b840101938193600091609e1984360301915b858410613d9d575050505050505090565b90919293949596603f19828203018352873590848212156102dd5760208091886001940190608063ffffffff613e2482613de8613dda87806138b5565b60a0885260a08801916138e9565b9587810135613df6816102cc565b8a8060a01b0316888701526040810135604087015283613e1860608301610dc8565b16606087015201610dc8565b16910152990193019401929195949390613d8c565b6001600160a01b03811615613ea45760c954604080516001600160a01b03928316815291831660208301527f6e9fcd539896fca60e8b0f01dd580233e48a6b0f7df013b89ba7f565869acdb691a160018060a01b03166001600160601b0360a01b60c954161760c955565b60405162461bcd60e51b815260206004820152604960248201527f5061757361626c652e5f73657450617573657252656769737472793a206e657760448201527f50617573657252656769737472792063616e6e6f7420626520746865207a65726064820152686f206164647265737360b81b608482015260a490fd5b60405190613f2e826104d8565b60006020838281520152565b60405190610180613f4b8184610513565b368337565b604051906020613f4b8184610513565b604090929192613f6e613f21565b9384916060916020855192613f838585610513565b8436853780518452015160208301528482015260076107cf195a01fa15613fa657565bfe5b604090929192613fb6613f21565b938491602060809281865193613fcc8686610513565b85368637805185520151828401528051868401520151606082015260066107cf195a01fa8015613fa65715613ffd57565b60405162461bcd60e51b815260206004820152600d60248201526c1958cb5859190b59985a5b1959609a1b6044820152606490fd5b60405161403e816104d8565b604090815161404d8382610513565b82368237815260208251916140628484610513565b83368437015280516140748282610513565b7f198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c281527f1800deef121f1e76426a00665e5c4479674322d4f75edadd46debd5cd992f6ed60208201528151906140ca8383610513565b7f275dc4a288d1afb3cbb1ac09187524c7db36395df7be3b99e673b13a075a65ec82527f1d9befcd05a5323e6da4d435f3b617cdb3af83285c2df711ef39c01571827f9d602083015261411f83519384610513565b8252602082015290565b600080516020614c3e83398151915290614141613f21565b5006906000908192602060c0945b61424b57600093600080516020614c3e8339815191526003818581818009090860405161417c8482610513565b833682378381896040516141908282610513565b813682378381528360208201528360408201528560608201527f0c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f526080820152600080516020614c3e83398151915260a082015260056107cf195a01fa8015613fa6576141fb90614bf1565b519161424b57600080516020614c3e833981519152828009146142345750600080516020614c3e8339815191526001859308919361414f565b929450909250614242610546565b92835282015290565b6121b6565b614258613f21565b50604051614265816104d8565b600181526002602082015290565b9060068202918083046006149015171561238557565b90600c8110156121b15760051b0190565b939290916142a86040610555565b94855260208501526142ba6040610555565b91825260208201526142ca613f3a565b9260005b600281106142f8575050506020610180926142e7613f50565b93849160086201d4c0fa9151151590565b80614304600192614273565b61430e82856121a0565b515160009061431d838a614289565b52602061432a84876121a0565b51015161433f61433984612377565b8a614289565b5261434a83876121a0565b5151516143596143398461238a565b5261436f61436784886121a0565b515160200190565b5161437c61433984612398565b52602061438984886121a0565b5101519050516143a161439b836123a6565b89614289565b526143cd6143c76143c060206143b7868a6121a0565b51015160200190565b51926123b4565b88614289565b52016142ce565b6033546001600160a01b031633036143e857565b606460405162461bcd60e51b815260206004820152602060248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152fd5b61ffff614438826145ef565b16614442816111b6565b906144506040519283610513565b80825261445f601f19916111b6565b013660208301376000805b82518210806144c1575b156144ba576001811b8416614492575b61448d9061242e565b61446a565b90600161448d9160ff60f81b8460f81b1660001a6144b08287612350565b5301919050614484565b5050905090565b506101008110614474565b606554604080516001600160a01b038084168252841660208201529192917fe11cddf1816a43318ca175bbc52cd0185436e9cbead7c83acc54a73e461717e39190a16001600160a01b03166001600160a01b03199190911617606555565b60207f40e4ed880a29e0f6ddce307457fb75cddf4feef7d3ecb0301bfdf4976a0e2dfc91151560ff196097541660ff821617609755604051908152a1565b90600161457660ff93614a85565b928392161b11156145845790565b60405162461bcd60e51b815260206004820152603f60248201527f4269746d61705574696c732e6f72646572656442797465734172726179546f4260448201527f69746d61703a206269746d61702065786365656473206d61782076616c7565006064820152608490fd5b806000915b6145fc575090565b60001981018181116123855761ffff9116911661ffff81146123855760010190806145f4565b9061462b613f21565b5061ffff8116906102008210156146d357600182146146ce5761464c610546565b600081526000602082015292906001906000925b61ffff831685101561467457505050505090565b600161ffff831660ff86161c8116146146ae575b60016146a46146998360ff94613fa8565b9460011b61fffe1690565b9401169291614660565b9460016146a46146996146c38960ff95613fa8565b989350505050614688565b505090565b60405162461bcd60e51b815260206004820152601060248201526f7363616c61722d746f6f2d6c6172676560801b6044820152606490fd5b614713613f21565b50805190811580614789575b15614742575050604051614734604082610513565b600081526000602082015290565b6020600080516020614c3e83398151915291015106600080516020614c3e83398151915203600080516020614c3e8339815191528111612385576040519161411f836104d8565b5060208101511561471f565b603380546001600160a01b039283166001600160a01b0319821681179092559091167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0600080a3565b9060ff60005460081c16156147f95761084c61054492614795565b60405162461bcd60e51b815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201526a6e697469616c697a696e6760a81b6064820152608490fd5b6065546001600160a01b0316330361486657565b60405162461bcd60e51b815260206004820152604c60248201527f536572766963654d616e61676572426173652e6f6e6c7952657761726473496e60448201527f69746961746f723a2063616c6c6572206973206e6f742074686520726577617260648201526b32399034b734ba34b0ba37b960a11b608482015260a490fd5b80548210156121b15760005260206000200190600090565b916149189183549060031b91821b91600019901b19161790565b9055565b8054801561494557600019019061493382826148e6565b8154906000199060031b1b1916905555565b634e487b7160e01b600052603160045260246000fd5b600081815260fc60205260409020549081156149fc576000198201908282116123855760fb546000198101939084116123855783836149bb94600096036149c1575b5050506149aa60fb61491c565b60fc90600052602052604060002090565b55600190565b6149aa6149ed916149e36149d96149f39560fb6148e6565b90549060031b1c90565b92839160fb6148e6565b906148fe565b5538808061499d565b5050600090565b15614a0a57565b60405162461bcd60e51b815260206004820152604760248201527f4269746d61705574696c732e6f72646572656442797465734172726179546f4260448201527f69746d61703a206f72646572656442797465734172726179206973206e6f74206064820152661bdc99195c995960ca1b608482015260a490fd5b90610100825111614af657815115614af057614ab3614aa961255261254485612343565b60ff600191161b90565b6001905b8351821015614aeb57600190614ad6614aa96125526125448689612350565b90614ae2818311614a03565b17910190614ab7565b925050565b60009150565b60a460405162461bcd60e51b815260206004820152604460248201527f4269746d61705574696c732e6f72646572656442797465734172726179546f4260448201527f69746d61703a206f7264657265644279746573417272617920697320746f6f206064820152636c6f6e6760e01b6084820152fd5b8060005260fc60205260406000205415600014614beb5760fb54680100000000000000008110156104f3576001810160fb55600060fb548210156121b15760fb90527f3e7257b7272bb46d49cd6019b04ddee20da7c0cb13f7c1ec3391291b2ccebabc0181905560fb549060005260fc602052604060002055600190565b50600090565b15614bf857565b60405162461bcd60e51b815260206004820152601a60248201527f424e3235342e6578704d6f643a2063616c6c206661696c7572650000000000006044820152606490fdfe30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47424c535369676e6174757265436865636b65722e636865636b5369676e617475a2646970667358221220890048e1eb21e67a2aaed7a131a484acbb50e4673db7c4f8ed8b8e06635c649f64736f6c634300081a0033",
}

// ContractTriggerXServiceManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractTriggerXServiceManagerMetaData.ABI instead.
var ContractTriggerXServiceManagerABI = ContractTriggerXServiceManagerMetaData.ABI

// ContractTriggerXServiceManagerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractTriggerXServiceManagerMetaData.Bin instead.
var ContractTriggerXServiceManagerBin = ContractTriggerXServiceManagerMetaData.Bin

// DeployContractTriggerXServiceManager deploys a new Ethereum contract, binding an instance of ContractTriggerXServiceManager to it.
func DeployContractTriggerXServiceManager(auth *bind.TransactOpts, backend bind.ContractBackend, _avsDirectory common.Address, _rewardsCoordinator common.Address, _registryCoordinator common.Address, _stakeRegistry common.Address) (common.Address, *types.Transaction, *ContractTriggerXServiceManager, error) {
	parsed, err := ContractTriggerXServiceManagerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractTriggerXServiceManagerBin), backend, _avsDirectory, _rewardsCoordinator, _registryCoordinator, _stakeRegistry)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractTriggerXServiceManager{ContractTriggerXServiceManagerCaller: ContractTriggerXServiceManagerCaller{contract: contract}, ContractTriggerXServiceManagerTransactor: ContractTriggerXServiceManagerTransactor{contract: contract}, ContractTriggerXServiceManagerFilterer: ContractTriggerXServiceManagerFilterer{contract: contract}}, nil
}

// ContractTriggerXServiceManagerMethods is an auto generated interface around an Ethereum contract.
type ContractTriggerXServiceManagerMethods interface {
	ContractTriggerXServiceManagerCalls
	ContractTriggerXServiceManagerTransacts
	ContractTriggerXServiceManagerFilters
}

// ContractTriggerXServiceManagerCalls is an auto generated interface that defines the call methods available for an Ethereum contract.
type ContractTriggerXServiceManagerCalls interface {
	AvsDirectory(opts *bind.CallOpts) (common.Address, error)

	BlsApkRegistry(opts *bind.CallOpts) (common.Address, error)

	CheckSignatures(opts *bind.CallOpts, msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, params IBLSSignatureCheckerNonSignerStakesAndSignature) (IBLSSignatureCheckerQuorumStakeTotals, [32]byte, error)

	Delegation(opts *bind.CallOpts) (common.Address, error)

	GetOperatorRestakedStrategies(opts *bind.CallOpts, operator common.Address) ([]common.Address, error)

	GetRestakeableStrategies(opts *bind.CallOpts) ([]common.Address, error)

	IsBlackListed(opts *bind.CallOpts, arg0 common.Address) (bool, error)

	Owner(opts *bind.CallOpts) (common.Address, error)

	Paused(opts *bind.CallOpts, index uint8) (bool, error)

	Paused0(opts *bind.CallOpts) (*big.Int, error)

	PauserRegistry(opts *bind.CallOpts) (common.Address, error)

	QuorumManager(opts *bind.CallOpts) (common.Address, error)

	RegistryCoordinator(opts *bind.CallOpts) (common.Address, error)

	RewardsInitiator(opts *bind.CallOpts) (common.Address, error)

	StakeRegistry(opts *bind.CallOpts) (common.Address, error)

	StaleStakesForbidden(opts *bind.CallOpts) (bool, error)

	TaskManager(opts *bind.CallOpts) (common.Address, error)

	TaskManagerContract(opts *bind.CallOpts) (common.Address, error)

	TaskValidator(opts *bind.CallOpts) (common.Address, error)

	TrySignatureAndApkVerification(opts *bind.CallOpts, msgHash [32]byte, apk BN254G1Point, apkG2 BN254G2Point, sigma BN254G1Point) (struct {
		PairingSuccessful bool
		SiganatureIsValid bool
	}, error)
}

// ContractTriggerXServiceManagerTransacts is an auto generated interface that defines the transact methods available for an Ethereum contract.
type ContractTriggerXServiceManagerTransacts interface {
	BlacklistKeeper(opts *bind.TransactOpts, _operator common.Address) (*types.Transaction, error)

	CreateAVSRewardsSubmission(opts *bind.TransactOpts, rewardsSubmissions []IRewardsCoordinatorRewardsSubmission) (*types.Transaction, error)

	CreateOperatorDirectedAVSRewardsSubmission(opts *bind.TransactOpts, operatorDirectedRewardsSubmissions []IRewardsCoordinatorOperatorDirectedRewardsSubmission) (*types.Transaction, error)

	DeregisterKeeperFromTriggerX(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error)

	DeregisterOperatorFromAVS(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error)

	Initialize(opts *bind.TransactOpts, _taskManagerContract common.Address, _pauserRegistry common.Address, _initialPausedStatus *big.Int, initialOwner common.Address, rewardsInitiator common.Address, _taskManager common.Address, _taskValidator common.Address, _quorumManager common.Address) (*types.Transaction, error)

	Pause(opts *bind.TransactOpts, newPausedStatus *big.Int) (*types.Transaction, error)

	PauseAll(opts *bind.TransactOpts) (*types.Transaction, error)

	RegisterKeeperToTriggerX(opts *bind.TransactOpts, operator common.Address, operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error)

	RegisterOperatorToAVS(opts *bind.TransactOpts, operator common.Address, operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error)

	RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error)

	SetClaimerFor(opts *bind.TransactOpts, claimer common.Address) (*types.Transaction, error)

	SetPauserRegistry(opts *bind.TransactOpts, newPauserRegistry common.Address) (*types.Transaction, error)

	SetQuorumManager(opts *bind.TransactOpts, _quorumManager common.Address) (*types.Transaction, error)

	SetRewardsInitiator(opts *bind.TransactOpts, newRewardsInitiator common.Address) (*types.Transaction, error)

	SetStaleStakesForbidden(opts *bind.TransactOpts, value bool) (*types.Transaction, error)

	SetTaskManager(opts *bind.TransactOpts, _taskManager common.Address) (*types.Transaction, error)

	SetTaskValidator(opts *bind.TransactOpts, _taskValidator common.Address) (*types.Transaction, error)

	TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error)

	UnblacklistKeeper(opts *bind.TransactOpts, _operator common.Address) (*types.Transaction, error)

	Unpause(opts *bind.TransactOpts, newPausedStatus *big.Int) (*types.Transaction, error)

	UpdateAVSMetadataURI(opts *bind.TransactOpts, _metadataURI string) (*types.Transaction, error)

	UpdateTaskManager(opts *bind.TransactOpts, _taskManager common.Address) (*types.Transaction, error)
}

// ContractTriggerXServiceManagerFilterer is an auto generated interface that defines the log filtering methods available for an Ethereum contract.
type ContractTriggerXServiceManagerFilters interface {
	FilterInitialized(opts *bind.FilterOpts) (*ContractTriggerXServiceManagerInitializedIterator, error)
	WatchInitialized(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerInitialized) (event.Subscription, error)
	ParseInitialized(log types.Log) (*ContractTriggerXServiceManagerInitialized, error)

	FilterKeeperAdded(opts *bind.FilterOpts, operator []common.Address) (*ContractTriggerXServiceManagerKeeperAddedIterator, error)
	WatchKeeperAdded(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerKeeperAdded, operator []common.Address) (event.Subscription, error)
	ParseKeeperAdded(log types.Log) (*ContractTriggerXServiceManagerKeeperAdded, error)

	FilterKeeperBlacklisted(opts *bind.FilterOpts, operator []common.Address) (*ContractTriggerXServiceManagerKeeperBlacklistedIterator, error)
	WatchKeeperBlacklisted(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerKeeperBlacklisted, operator []common.Address) (event.Subscription, error)
	ParseKeeperBlacklisted(log types.Log) (*ContractTriggerXServiceManagerKeeperBlacklisted, error)

	FilterKeeperRemoved(opts *bind.FilterOpts, operator []common.Address) (*ContractTriggerXServiceManagerKeeperRemovedIterator, error)
	WatchKeeperRemoved(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerKeeperRemoved, operator []common.Address) (event.Subscription, error)
	ParseKeeperRemoved(log types.Log) (*ContractTriggerXServiceManagerKeeperRemoved, error)

	FilterKeeperUnblacklisted(opts *bind.FilterOpts, operator []common.Address) (*ContractTriggerXServiceManagerKeeperUnblacklistedIterator, error)
	WatchKeeperUnblacklisted(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerKeeperUnblacklisted, operator []common.Address) (event.Subscription, error)
	ParseKeeperUnblacklisted(log types.Log) (*ContractTriggerXServiceManagerKeeperUnblacklisted, error)

	FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractTriggerXServiceManagerOwnershipTransferredIterator, error)
	WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error)
	ParseOwnershipTransferred(log types.Log) (*ContractTriggerXServiceManagerOwnershipTransferred, error)

	FilterPaused(opts *bind.FilterOpts, account []common.Address) (*ContractTriggerXServiceManagerPausedIterator, error)
	WatchPaused(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerPaused, account []common.Address) (event.Subscription, error)
	ParsePaused(log types.Log) (*ContractTriggerXServiceManagerPaused, error)

	FilterPauserRegistrySet(opts *bind.FilterOpts) (*ContractTriggerXServiceManagerPauserRegistrySetIterator, error)
	WatchPauserRegistrySet(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerPauserRegistrySet) (event.Subscription, error)
	ParsePauserRegistrySet(log types.Log) (*ContractTriggerXServiceManagerPauserRegistrySet, error)

	FilterQuorumManagerSet(opts *bind.FilterOpts, oldQuorumManager []common.Address, newQuorumManager []common.Address) (*ContractTriggerXServiceManagerQuorumManagerSetIterator, error)
	WatchQuorumManagerSet(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerQuorumManagerSet, oldQuorumManager []common.Address, newQuorumManager []common.Address) (event.Subscription, error)
	ParseQuorumManagerSet(log types.Log) (*ContractTriggerXServiceManagerQuorumManagerSet, error)

	FilterRewardsInitiatorUpdated(opts *bind.FilterOpts) (*ContractTriggerXServiceManagerRewardsInitiatorUpdatedIterator, error)
	WatchRewardsInitiatorUpdated(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerRewardsInitiatorUpdated) (event.Subscription, error)
	ParseRewardsInitiatorUpdated(log types.Log) (*ContractTriggerXServiceManagerRewardsInitiatorUpdated, error)

	FilterStaleStakesForbiddenUpdate(opts *bind.FilterOpts) (*ContractTriggerXServiceManagerStaleStakesForbiddenUpdateIterator, error)
	WatchStaleStakesForbiddenUpdate(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerStaleStakesForbiddenUpdate) (event.Subscription, error)
	ParseStaleStakesForbiddenUpdate(log types.Log) (*ContractTriggerXServiceManagerStaleStakesForbiddenUpdate, error)

	FilterTaskManagerContractUpdated(opts *bind.FilterOpts, oldTaskManager []common.Address, newTaskManager []common.Address) (*ContractTriggerXServiceManagerTaskManagerContractUpdatedIterator, error)
	WatchTaskManagerContractUpdated(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerTaskManagerContractUpdated, oldTaskManager []common.Address, newTaskManager []common.Address) (event.Subscription, error)
	ParseTaskManagerContractUpdated(log types.Log) (*ContractTriggerXServiceManagerTaskManagerContractUpdated, error)

	FilterTaskManagerSet(opts *bind.FilterOpts, oldTaskManager []common.Address, newTaskManager []common.Address) (*ContractTriggerXServiceManagerTaskManagerSetIterator, error)
	WatchTaskManagerSet(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerTaskManagerSet, oldTaskManager []common.Address, newTaskManager []common.Address) (event.Subscription, error)
	ParseTaskManagerSet(log types.Log) (*ContractTriggerXServiceManagerTaskManagerSet, error)

	FilterTaskValidatorSet(opts *bind.FilterOpts, oldTaskValidator []common.Address, newTaskValidator []common.Address) (*ContractTriggerXServiceManagerTaskValidatorSetIterator, error)
	WatchTaskValidatorSet(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerTaskValidatorSet, oldTaskValidator []common.Address, newTaskValidator []common.Address) (event.Subscription, error)
	ParseTaskValidatorSet(log types.Log) (*ContractTriggerXServiceManagerTaskValidatorSet, error)

	FilterUnpaused(opts *bind.FilterOpts, account []common.Address) (*ContractTriggerXServiceManagerUnpausedIterator, error)
	WatchUnpaused(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerUnpaused, account []common.Address) (event.Subscription, error)
	ParseUnpaused(log types.Log) (*ContractTriggerXServiceManagerUnpaused, error)
}

// ContractTriggerXServiceManager is an auto generated Go binding around an Ethereum contract.
type ContractTriggerXServiceManager struct {
	ContractTriggerXServiceManagerCaller     // Read-only binding to the contract
	ContractTriggerXServiceManagerTransactor // Write-only binding to the contract
	ContractTriggerXServiceManagerFilterer   // Log filterer for contract events
}

// ContractTriggerXServiceManager implements the ContractTriggerXServiceManagerMethods interface.
var _ ContractTriggerXServiceManagerMethods = (*ContractTriggerXServiceManager)(nil)

// ContractTriggerXServiceManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractTriggerXServiceManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTriggerXServiceManagerCaller implements the ContractTriggerXServiceManagerCalls interface.
var _ ContractTriggerXServiceManagerCalls = (*ContractTriggerXServiceManagerCaller)(nil)

// ContractTriggerXServiceManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTriggerXServiceManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTriggerXServiceManagerTransactor implements the ContractTriggerXServiceManagerTransacts interface.
var _ ContractTriggerXServiceManagerTransacts = (*ContractTriggerXServiceManagerTransactor)(nil)

// ContractTriggerXServiceManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractTriggerXServiceManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTriggerXServiceManagerFilterer implements the ContractTriggerXServiceManagerFilters interface.
var _ ContractTriggerXServiceManagerFilters = (*ContractTriggerXServiceManagerFilterer)(nil)

// ContractTriggerXServiceManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractTriggerXServiceManagerSession struct {
	Contract     *ContractTriggerXServiceManager // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                   // Call options to use throughout this session
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// ContractTriggerXServiceManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractTriggerXServiceManagerCallerSession struct {
	Contract *ContractTriggerXServiceManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                         // Call options to use throughout this session
}

// ContractTriggerXServiceManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTriggerXServiceManagerTransactorSession struct {
	Contract     *ContractTriggerXServiceManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                         // Transaction auth options to use throughout this session
}

// ContractTriggerXServiceManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractTriggerXServiceManagerRaw struct {
	Contract *ContractTriggerXServiceManager // Generic contract binding to access the raw methods on
}

// ContractTriggerXServiceManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractTriggerXServiceManagerCallerRaw struct {
	Contract *ContractTriggerXServiceManagerCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTriggerXServiceManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTriggerXServiceManagerTransactorRaw struct {
	Contract *ContractTriggerXServiceManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractTriggerXServiceManager creates a new instance of ContractTriggerXServiceManager, bound to a specific deployed contract.
func NewContractTriggerXServiceManager(address common.Address, backend bind.ContractBackend) (*ContractTriggerXServiceManager, error) {
	contract, err := bindContractTriggerXServiceManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManager{ContractTriggerXServiceManagerCaller: ContractTriggerXServiceManagerCaller{contract: contract}, ContractTriggerXServiceManagerTransactor: ContractTriggerXServiceManagerTransactor{contract: contract}, ContractTriggerXServiceManagerFilterer: ContractTriggerXServiceManagerFilterer{contract: contract}}, nil
}

// NewContractTriggerXServiceManagerCaller creates a new read-only instance of ContractTriggerXServiceManager, bound to a specific deployed contract.
func NewContractTriggerXServiceManagerCaller(address common.Address, caller bind.ContractCaller) (*ContractTriggerXServiceManagerCaller, error) {
	contract, err := bindContractTriggerXServiceManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerCaller{contract: contract}, nil
}

// NewContractTriggerXServiceManagerTransactor creates a new write-only instance of ContractTriggerXServiceManager, bound to a specific deployed contract.
func NewContractTriggerXServiceManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTriggerXServiceManagerTransactor, error) {
	contract, err := bindContractTriggerXServiceManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerTransactor{contract: contract}, nil
}

// NewContractTriggerXServiceManagerFilterer creates a new log filterer instance of ContractTriggerXServiceManager, bound to a specific deployed contract.
func NewContractTriggerXServiceManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractTriggerXServiceManagerFilterer, error) {
	contract, err := bindContractTriggerXServiceManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerFilterer{contract: contract}, nil
}

// bindContractTriggerXServiceManager binds a generic wrapper to an already deployed contract.
func bindContractTriggerXServiceManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractTriggerXServiceManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractTriggerXServiceManager.Contract.ContractTriggerXServiceManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.ContractTriggerXServiceManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.ContractTriggerXServiceManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractTriggerXServiceManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.contract.Transact(opts, method, params...)
}

// AvsDirectory is a free data retrieval call binding the contract method 0x6b3aa72e.
//
// Solidity: function avsDirectory() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) AvsDirectory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "avsDirectory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AvsDirectory is a free data retrieval call binding the contract method 0x6b3aa72e.
//
// Solidity: function avsDirectory() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) AvsDirectory() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.AvsDirectory(&_ContractTriggerXServiceManager.CallOpts)
}

// AvsDirectory is a free data retrieval call binding the contract method 0x6b3aa72e.
//
// Solidity: function avsDirectory() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) AvsDirectory() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.AvsDirectory(&_ContractTriggerXServiceManager.CallOpts)
}

// BlsApkRegistry is a free data retrieval call binding the contract method 0x5df45946.
//
// Solidity: function blsApkRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) BlsApkRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "blsApkRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BlsApkRegistry is a free data retrieval call binding the contract method 0x5df45946.
//
// Solidity: function blsApkRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) BlsApkRegistry() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.BlsApkRegistry(&_ContractTriggerXServiceManager.CallOpts)
}

// BlsApkRegistry is a free data retrieval call binding the contract method 0x5df45946.
//
// Solidity: function blsApkRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) BlsApkRegistry() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.BlsApkRegistry(&_ContractTriggerXServiceManager.CallOpts)
}

// CheckSignatures is a free data retrieval call binding the contract method 0x6efb4636.
//
// Solidity: function checkSignatures(bytes32 msgHash, bytes quorumNumbers, uint32 referenceBlockNumber, (uint32[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2]),(uint256,uint256),uint32[],uint32[],uint32[][]) params) view returns((uint96[],uint96[]), bytes32)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) CheckSignatures(opts *bind.CallOpts, msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, params IBLSSignatureCheckerNonSignerStakesAndSignature) (IBLSSignatureCheckerQuorumStakeTotals, [32]byte, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "checkSignatures", msgHash, quorumNumbers, referenceBlockNumber, params)

	if err != nil {
		return *new(IBLSSignatureCheckerQuorumStakeTotals), *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(IBLSSignatureCheckerQuorumStakeTotals)).(*IBLSSignatureCheckerQuorumStakeTotals)
	out1 := *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)

	return out0, out1, err

}

// CheckSignatures is a free data retrieval call binding the contract method 0x6efb4636.
//
// Solidity: function checkSignatures(bytes32 msgHash, bytes quorumNumbers, uint32 referenceBlockNumber, (uint32[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2]),(uint256,uint256),uint32[],uint32[],uint32[][]) params) view returns((uint96[],uint96[]), bytes32)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) CheckSignatures(msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, params IBLSSignatureCheckerNonSignerStakesAndSignature) (IBLSSignatureCheckerQuorumStakeTotals, [32]byte, error) {
	return _ContractTriggerXServiceManager.Contract.CheckSignatures(&_ContractTriggerXServiceManager.CallOpts, msgHash, quorumNumbers, referenceBlockNumber, params)
}

// CheckSignatures is a free data retrieval call binding the contract method 0x6efb4636.
//
// Solidity: function checkSignatures(bytes32 msgHash, bytes quorumNumbers, uint32 referenceBlockNumber, (uint32[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2]),(uint256,uint256),uint32[],uint32[],uint32[][]) params) view returns((uint96[],uint96[]), bytes32)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) CheckSignatures(msgHash [32]byte, quorumNumbers []byte, referenceBlockNumber uint32, params IBLSSignatureCheckerNonSignerStakesAndSignature) (IBLSSignatureCheckerQuorumStakeTotals, [32]byte, error) {
	return _ContractTriggerXServiceManager.Contract.CheckSignatures(&_ContractTriggerXServiceManager.CallOpts, msgHash, quorumNumbers, referenceBlockNumber, params)
}

// Delegation is a free data retrieval call binding the contract method 0xdf5cf723.
//
// Solidity: function delegation() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) Delegation(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "delegation")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Delegation is a free data retrieval call binding the contract method 0xdf5cf723.
//
// Solidity: function delegation() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) Delegation() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.Delegation(&_ContractTriggerXServiceManager.CallOpts)
}

// Delegation is a free data retrieval call binding the contract method 0xdf5cf723.
//
// Solidity: function delegation() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) Delegation() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.Delegation(&_ContractTriggerXServiceManager.CallOpts)
}

// GetOperatorRestakedStrategies is a free data retrieval call binding the contract method 0x33cfb7b7.
//
// Solidity: function getOperatorRestakedStrategies(address operator) view returns(address[])
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) GetOperatorRestakedStrategies(opts *bind.CallOpts, operator common.Address) ([]common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "getOperatorRestakedStrategies", operator)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOperatorRestakedStrategies is a free data retrieval call binding the contract method 0x33cfb7b7.
//
// Solidity: function getOperatorRestakedStrategies(address operator) view returns(address[])
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) GetOperatorRestakedStrategies(operator common.Address) ([]common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.GetOperatorRestakedStrategies(&_ContractTriggerXServiceManager.CallOpts, operator)
}

// GetOperatorRestakedStrategies is a free data retrieval call binding the contract method 0x33cfb7b7.
//
// Solidity: function getOperatorRestakedStrategies(address operator) view returns(address[])
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) GetOperatorRestakedStrategies(operator common.Address) ([]common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.GetOperatorRestakedStrategies(&_ContractTriggerXServiceManager.CallOpts, operator)
}

// GetRestakeableStrategies is a free data retrieval call binding the contract method 0xe481af9d.
//
// Solidity: function getRestakeableStrategies() view returns(address[])
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) GetRestakeableStrategies(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "getRestakeableStrategies")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetRestakeableStrategies is a free data retrieval call binding the contract method 0xe481af9d.
//
// Solidity: function getRestakeableStrategies() view returns(address[])
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) GetRestakeableStrategies() ([]common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.GetRestakeableStrategies(&_ContractTriggerXServiceManager.CallOpts)
}

// GetRestakeableStrategies is a free data retrieval call binding the contract method 0xe481af9d.
//
// Solidity: function getRestakeableStrategies() view returns(address[])
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) GetRestakeableStrategies() ([]common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.GetRestakeableStrategies(&_ContractTriggerXServiceManager.CallOpts)
}

// IsBlackListed is a free data retrieval call binding the contract method 0xe47d6060.
//
// Solidity: function isBlackListed(address ) view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) IsBlackListed(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "isBlackListed", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBlackListed is a free data retrieval call binding the contract method 0xe47d6060.
//
// Solidity: function isBlackListed(address ) view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) IsBlackListed(arg0 common.Address) (bool, error) {
	return _ContractTriggerXServiceManager.Contract.IsBlackListed(&_ContractTriggerXServiceManager.CallOpts, arg0)
}

// IsBlackListed is a free data retrieval call binding the contract method 0xe47d6060.
//
// Solidity: function isBlackListed(address ) view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) IsBlackListed(arg0 common.Address) (bool, error) {
	return _ContractTriggerXServiceManager.Contract.IsBlackListed(&_ContractTriggerXServiceManager.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) Owner() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.Owner(&_ContractTriggerXServiceManager.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) Owner() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.Owner(&_ContractTriggerXServiceManager.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5ac86ab7.
//
// Solidity: function paused(uint8 index) view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) Paused(opts *bind.CallOpts, index uint8) (bool, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "paused", index)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5ac86ab7.
//
// Solidity: function paused(uint8 index) view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) Paused(index uint8) (bool, error) {
	return _ContractTriggerXServiceManager.Contract.Paused(&_ContractTriggerXServiceManager.CallOpts, index)
}

// Paused is a free data retrieval call binding the contract method 0x5ac86ab7.
//
// Solidity: function paused(uint8 index) view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) Paused(index uint8) (bool, error) {
	return _ContractTriggerXServiceManager.Contract.Paused(&_ContractTriggerXServiceManager.CallOpts, index)
}

// Paused0 is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(uint256)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) Paused0(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "paused0")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Paused0 is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(uint256)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) Paused0() (*big.Int, error) {
	return _ContractTriggerXServiceManager.Contract.Paused0(&_ContractTriggerXServiceManager.CallOpts)
}

// Paused0 is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(uint256)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) Paused0() (*big.Int, error) {
	return _ContractTriggerXServiceManager.Contract.Paused0(&_ContractTriggerXServiceManager.CallOpts)
}

// PauserRegistry is a free data retrieval call binding the contract method 0x886f1195.
//
// Solidity: function pauserRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) PauserRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "pauserRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PauserRegistry is a free data retrieval call binding the contract method 0x886f1195.
//
// Solidity: function pauserRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) PauserRegistry() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.PauserRegistry(&_ContractTriggerXServiceManager.CallOpts)
}

// PauserRegistry is a free data retrieval call binding the contract method 0x886f1195.
//
// Solidity: function pauserRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) PauserRegistry() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.PauserRegistry(&_ContractTriggerXServiceManager.CallOpts)
}

// QuorumManager is a free data retrieval call binding the contract method 0xb5f7eb6b.
//
// Solidity: function quorumManager() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) QuorumManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "quorumManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// QuorumManager is a free data retrieval call binding the contract method 0xb5f7eb6b.
//
// Solidity: function quorumManager() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) QuorumManager() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.QuorumManager(&_ContractTriggerXServiceManager.CallOpts)
}

// QuorumManager is a free data retrieval call binding the contract method 0xb5f7eb6b.
//
// Solidity: function quorumManager() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) QuorumManager() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.QuorumManager(&_ContractTriggerXServiceManager.CallOpts)
}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) RegistryCoordinator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "registryCoordinator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) RegistryCoordinator() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.RegistryCoordinator(&_ContractTriggerXServiceManager.CallOpts)
}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) RegistryCoordinator() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.RegistryCoordinator(&_ContractTriggerXServiceManager.CallOpts)
}

// RewardsInitiator is a free data retrieval call binding the contract method 0xfc299dee.
//
// Solidity: function rewardsInitiator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) RewardsInitiator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "rewardsInitiator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RewardsInitiator is a free data retrieval call binding the contract method 0xfc299dee.
//
// Solidity: function rewardsInitiator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) RewardsInitiator() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.RewardsInitiator(&_ContractTriggerXServiceManager.CallOpts)
}

// RewardsInitiator is a free data retrieval call binding the contract method 0xfc299dee.
//
// Solidity: function rewardsInitiator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) RewardsInitiator() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.RewardsInitiator(&_ContractTriggerXServiceManager.CallOpts)
}

// StakeRegistry is a free data retrieval call binding the contract method 0x68304835.
//
// Solidity: function stakeRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) StakeRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "stakeRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakeRegistry is a free data retrieval call binding the contract method 0x68304835.
//
// Solidity: function stakeRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) StakeRegistry() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.StakeRegistry(&_ContractTriggerXServiceManager.CallOpts)
}

// StakeRegistry is a free data retrieval call binding the contract method 0x68304835.
//
// Solidity: function stakeRegistry() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) StakeRegistry() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.StakeRegistry(&_ContractTriggerXServiceManager.CallOpts)
}

// StaleStakesForbidden is a free data retrieval call binding the contract method 0xb98d0908.
//
// Solidity: function staleStakesForbidden() view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) StaleStakesForbidden(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "staleStakesForbidden")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// StaleStakesForbidden is a free data retrieval call binding the contract method 0xb98d0908.
//
// Solidity: function staleStakesForbidden() view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) StaleStakesForbidden() (bool, error) {
	return _ContractTriggerXServiceManager.Contract.StaleStakesForbidden(&_ContractTriggerXServiceManager.CallOpts)
}

// StaleStakesForbidden is a free data retrieval call binding the contract method 0xb98d0908.
//
// Solidity: function staleStakesForbidden() view returns(bool)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) StaleStakesForbidden() (bool, error) {
	return _ContractTriggerXServiceManager.Contract.StaleStakesForbidden(&_ContractTriggerXServiceManager.CallOpts)
}

// TaskManager is a free data retrieval call binding the contract method 0xa50a640e.
//
// Solidity: function taskManager() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) TaskManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "taskManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TaskManager is a free data retrieval call binding the contract method 0xa50a640e.
//
// Solidity: function taskManager() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) TaskManager() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.TaskManager(&_ContractTriggerXServiceManager.CallOpts)
}

// TaskManager is a free data retrieval call binding the contract method 0xa50a640e.
//
// Solidity: function taskManager() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) TaskManager() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.TaskManager(&_ContractTriggerXServiceManager.CallOpts)
}

// TaskManagerContract is a free data retrieval call binding the contract method 0x500c8dd3.
//
// Solidity: function taskManagerContract() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) TaskManagerContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "taskManagerContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TaskManagerContract is a free data retrieval call binding the contract method 0x500c8dd3.
//
// Solidity: function taskManagerContract() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) TaskManagerContract() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.TaskManagerContract(&_ContractTriggerXServiceManager.CallOpts)
}

// TaskManagerContract is a free data retrieval call binding the contract method 0x500c8dd3.
//
// Solidity: function taskManagerContract() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) TaskManagerContract() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.TaskManagerContract(&_ContractTriggerXServiceManager.CallOpts)
}

// TaskValidator is a free data retrieval call binding the contract method 0xfd38ec8c.
//
// Solidity: function taskValidator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) TaskValidator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "taskValidator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TaskValidator is a free data retrieval call binding the contract method 0xfd38ec8c.
//
// Solidity: function taskValidator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) TaskValidator() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.TaskValidator(&_ContractTriggerXServiceManager.CallOpts)
}

// TaskValidator is a free data retrieval call binding the contract method 0xfd38ec8c.
//
// Solidity: function taskValidator() view returns(address)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) TaskValidator() (common.Address, error) {
	return _ContractTriggerXServiceManager.Contract.TaskValidator(&_ContractTriggerXServiceManager.CallOpts)
}

// TrySignatureAndApkVerification is a free data retrieval call binding the contract method 0x171f1d5b.
//
// Solidity: function trySignatureAndApkVerification(bytes32 msgHash, (uint256,uint256) apk, (uint256[2],uint256[2]) apkG2, (uint256,uint256) sigma) view returns(bool pairingSuccessful, bool siganatureIsValid)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCaller) TrySignatureAndApkVerification(opts *bind.CallOpts, msgHash [32]byte, apk BN254G1Point, apkG2 BN254G2Point, sigma BN254G1Point) (struct {
	PairingSuccessful bool
	SiganatureIsValid bool
}, error) {
	var out []interface{}
	err := _ContractTriggerXServiceManager.contract.Call(opts, &out, "trySignatureAndApkVerification", msgHash, apk, apkG2, sigma)

	outstruct := new(struct {
		PairingSuccessful bool
		SiganatureIsValid bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PairingSuccessful = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.SiganatureIsValid = *abi.ConvertType(out[1], new(bool)).(*bool)

	return *outstruct, err

}

// TrySignatureAndApkVerification is a free data retrieval call binding the contract method 0x171f1d5b.
//
// Solidity: function trySignatureAndApkVerification(bytes32 msgHash, (uint256,uint256) apk, (uint256[2],uint256[2]) apkG2, (uint256,uint256) sigma) view returns(bool pairingSuccessful, bool siganatureIsValid)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) TrySignatureAndApkVerification(msgHash [32]byte, apk BN254G1Point, apkG2 BN254G2Point, sigma BN254G1Point) (struct {
	PairingSuccessful bool
	SiganatureIsValid bool
}, error) {
	return _ContractTriggerXServiceManager.Contract.TrySignatureAndApkVerification(&_ContractTriggerXServiceManager.CallOpts, msgHash, apk, apkG2, sigma)
}

// TrySignatureAndApkVerification is a free data retrieval call binding the contract method 0x171f1d5b.
//
// Solidity: function trySignatureAndApkVerification(bytes32 msgHash, (uint256,uint256) apk, (uint256[2],uint256[2]) apkG2, (uint256,uint256) sigma) view returns(bool pairingSuccessful, bool siganatureIsValid)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerCallerSession) TrySignatureAndApkVerification(msgHash [32]byte, apk BN254G1Point, apkG2 BN254G2Point, sigma BN254G1Point) (struct {
	PairingSuccessful bool
	SiganatureIsValid bool
}, error) {
	return _ContractTriggerXServiceManager.Contract.TrySignatureAndApkVerification(&_ContractTriggerXServiceManager.CallOpts, msgHash, apk, apkG2, sigma)
}

// BlacklistKeeper is a paid mutator transaction binding the contract method 0x26a965f0.
//
// Solidity: function blacklistKeeper(address _operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) BlacklistKeeper(opts *bind.TransactOpts, _operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "blacklistKeeper", _operator)
}

// BlacklistKeeper is a paid mutator transaction binding the contract method 0x26a965f0.
//
// Solidity: function blacklistKeeper(address _operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) BlacklistKeeper(_operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.BlacklistKeeper(&_ContractTriggerXServiceManager.TransactOpts, _operator)
}

// BlacklistKeeper is a paid mutator transaction binding the contract method 0x26a965f0.
//
// Solidity: function blacklistKeeper(address _operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) BlacklistKeeper(_operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.BlacklistKeeper(&_ContractTriggerXServiceManager.TransactOpts, _operator)
}

// CreateAVSRewardsSubmission is a paid mutator transaction binding the contract method 0xfce36c7d.
//
// Solidity: function createAVSRewardsSubmission(((address,uint96)[],address,uint256,uint32,uint32)[] rewardsSubmissions) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) CreateAVSRewardsSubmission(opts *bind.TransactOpts, rewardsSubmissions []IRewardsCoordinatorRewardsSubmission) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "createAVSRewardsSubmission", rewardsSubmissions)
}

// CreateAVSRewardsSubmission is a paid mutator transaction binding the contract method 0xfce36c7d.
//
// Solidity: function createAVSRewardsSubmission(((address,uint96)[],address,uint256,uint32,uint32)[] rewardsSubmissions) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) CreateAVSRewardsSubmission(rewardsSubmissions []IRewardsCoordinatorRewardsSubmission) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.CreateAVSRewardsSubmission(&_ContractTriggerXServiceManager.TransactOpts, rewardsSubmissions)
}

// CreateAVSRewardsSubmission is a paid mutator transaction binding the contract method 0xfce36c7d.
//
// Solidity: function createAVSRewardsSubmission(((address,uint96)[],address,uint256,uint32,uint32)[] rewardsSubmissions) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) CreateAVSRewardsSubmission(rewardsSubmissions []IRewardsCoordinatorRewardsSubmission) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.CreateAVSRewardsSubmission(&_ContractTriggerXServiceManager.TransactOpts, rewardsSubmissions)
}

// CreateOperatorDirectedAVSRewardsSubmission is a paid mutator transaction binding the contract method 0xa20b99bf.
//
// Solidity: function createOperatorDirectedAVSRewardsSubmission(((address,uint96)[],address,(address,uint256)[],uint32,uint32,string)[] operatorDirectedRewardsSubmissions) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) CreateOperatorDirectedAVSRewardsSubmission(opts *bind.TransactOpts, operatorDirectedRewardsSubmissions []IRewardsCoordinatorOperatorDirectedRewardsSubmission) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "createOperatorDirectedAVSRewardsSubmission", operatorDirectedRewardsSubmissions)
}

// CreateOperatorDirectedAVSRewardsSubmission is a paid mutator transaction binding the contract method 0xa20b99bf.
//
// Solidity: function createOperatorDirectedAVSRewardsSubmission(((address,uint96)[],address,(address,uint256)[],uint32,uint32,string)[] operatorDirectedRewardsSubmissions) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) CreateOperatorDirectedAVSRewardsSubmission(operatorDirectedRewardsSubmissions []IRewardsCoordinatorOperatorDirectedRewardsSubmission) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.CreateOperatorDirectedAVSRewardsSubmission(&_ContractTriggerXServiceManager.TransactOpts, operatorDirectedRewardsSubmissions)
}

// CreateOperatorDirectedAVSRewardsSubmission is a paid mutator transaction binding the contract method 0xa20b99bf.
//
// Solidity: function createOperatorDirectedAVSRewardsSubmission(((address,uint96)[],address,(address,uint256)[],uint32,uint32,string)[] operatorDirectedRewardsSubmissions) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) CreateOperatorDirectedAVSRewardsSubmission(operatorDirectedRewardsSubmissions []IRewardsCoordinatorOperatorDirectedRewardsSubmission) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.CreateOperatorDirectedAVSRewardsSubmission(&_ContractTriggerXServiceManager.TransactOpts, operatorDirectedRewardsSubmissions)
}

// DeregisterKeeperFromTriggerX is a paid mutator transaction binding the contract method 0x6cf5ca21.
//
// Solidity: function deregisterKeeperFromTriggerX(address operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) DeregisterKeeperFromTriggerX(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "deregisterKeeperFromTriggerX", operator)
}

// DeregisterKeeperFromTriggerX is a paid mutator transaction binding the contract method 0x6cf5ca21.
//
// Solidity: function deregisterKeeperFromTriggerX(address operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) DeregisterKeeperFromTriggerX(operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.DeregisterKeeperFromTriggerX(&_ContractTriggerXServiceManager.TransactOpts, operator)
}

// DeregisterKeeperFromTriggerX is a paid mutator transaction binding the contract method 0x6cf5ca21.
//
// Solidity: function deregisterKeeperFromTriggerX(address operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) DeregisterKeeperFromTriggerX(operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.DeregisterKeeperFromTriggerX(&_ContractTriggerXServiceManager.TransactOpts, operator)
}

// DeregisterOperatorFromAVS is a paid mutator transaction binding the contract method 0xa364f4da.
//
// Solidity: function deregisterOperatorFromAVS(address operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) DeregisterOperatorFromAVS(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "deregisterOperatorFromAVS", operator)
}

// DeregisterOperatorFromAVS is a paid mutator transaction binding the contract method 0xa364f4da.
//
// Solidity: function deregisterOperatorFromAVS(address operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) DeregisterOperatorFromAVS(operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.DeregisterOperatorFromAVS(&_ContractTriggerXServiceManager.TransactOpts, operator)
}

// DeregisterOperatorFromAVS is a paid mutator transaction binding the contract method 0xa364f4da.
//
// Solidity: function deregisterOperatorFromAVS(address operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) DeregisterOperatorFromAVS(operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.DeregisterOperatorFromAVS(&_ContractTriggerXServiceManager.TransactOpts, operator)
}

// Initialize is a paid mutator transaction binding the contract method 0x83a94322.
//
// Solidity: function initialize(address _taskManagerContract, address _pauserRegistry, uint256 _initialPausedStatus, address initialOwner, address rewardsInitiator, address _taskManager, address _taskValidator, address _quorumManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) Initialize(opts *bind.TransactOpts, _taskManagerContract common.Address, _pauserRegistry common.Address, _initialPausedStatus *big.Int, initialOwner common.Address, rewardsInitiator common.Address, _taskManager common.Address, _taskValidator common.Address, _quorumManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "initialize", _taskManagerContract, _pauserRegistry, _initialPausedStatus, initialOwner, rewardsInitiator, _taskManager, _taskValidator, _quorumManager)
}

// Initialize is a paid mutator transaction binding the contract method 0x83a94322.
//
// Solidity: function initialize(address _taskManagerContract, address _pauserRegistry, uint256 _initialPausedStatus, address initialOwner, address rewardsInitiator, address _taskManager, address _taskValidator, address _quorumManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) Initialize(_taskManagerContract common.Address, _pauserRegistry common.Address, _initialPausedStatus *big.Int, initialOwner common.Address, rewardsInitiator common.Address, _taskManager common.Address, _taskValidator common.Address, _quorumManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.Initialize(&_ContractTriggerXServiceManager.TransactOpts, _taskManagerContract, _pauserRegistry, _initialPausedStatus, initialOwner, rewardsInitiator, _taskManager, _taskValidator, _quorumManager)
}

// Initialize is a paid mutator transaction binding the contract method 0x83a94322.
//
// Solidity: function initialize(address _taskManagerContract, address _pauserRegistry, uint256 _initialPausedStatus, address initialOwner, address rewardsInitiator, address _taskManager, address _taskValidator, address _quorumManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) Initialize(_taskManagerContract common.Address, _pauserRegistry common.Address, _initialPausedStatus *big.Int, initialOwner common.Address, rewardsInitiator common.Address, _taskManager common.Address, _taskValidator common.Address, _quorumManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.Initialize(&_ContractTriggerXServiceManager.TransactOpts, _taskManagerContract, _pauserRegistry, _initialPausedStatus, initialOwner, rewardsInitiator, _taskManager, _taskValidator, _quorumManager)
}

// Pause is a paid mutator transaction binding the contract method 0x136439dd.
//
// Solidity: function pause(uint256 newPausedStatus) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) Pause(opts *bind.TransactOpts, newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "pause", newPausedStatus)
}

// Pause is a paid mutator transaction binding the contract method 0x136439dd.
//
// Solidity: function pause(uint256 newPausedStatus) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) Pause(newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.Pause(&_ContractTriggerXServiceManager.TransactOpts, newPausedStatus)
}

// Pause is a paid mutator transaction binding the contract method 0x136439dd.
//
// Solidity: function pause(uint256 newPausedStatus) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) Pause(newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.Pause(&_ContractTriggerXServiceManager.TransactOpts, newPausedStatus)
}

// PauseAll is a paid mutator transaction binding the contract method 0x595c6a67.
//
// Solidity: function pauseAll() returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) PauseAll(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "pauseAll")
}

// PauseAll is a paid mutator transaction binding the contract method 0x595c6a67.
//
// Solidity: function pauseAll() returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) PauseAll() (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.PauseAll(&_ContractTriggerXServiceManager.TransactOpts)
}

// PauseAll is a paid mutator transaction binding the contract method 0x595c6a67.
//
// Solidity: function pauseAll() returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) PauseAll() (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.PauseAll(&_ContractTriggerXServiceManager.TransactOpts)
}

// RegisterKeeperToTriggerX is a paid mutator transaction binding the contract method 0x763ac957.
//
// Solidity: function registerKeeperToTriggerX(address operator, (bytes,bytes32,uint256) operatorSignature) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) RegisterKeeperToTriggerX(opts *bind.TransactOpts, operator common.Address, operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "registerKeeperToTriggerX", operator, operatorSignature)
}

// RegisterKeeperToTriggerX is a paid mutator transaction binding the contract method 0x763ac957.
//
// Solidity: function registerKeeperToTriggerX(address operator, (bytes,bytes32,uint256) operatorSignature) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) RegisterKeeperToTriggerX(operator common.Address, operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.RegisterKeeperToTriggerX(&_ContractTriggerXServiceManager.TransactOpts, operator, operatorSignature)
}

// RegisterKeeperToTriggerX is a paid mutator transaction binding the contract method 0x763ac957.
//
// Solidity: function registerKeeperToTriggerX(address operator, (bytes,bytes32,uint256) operatorSignature) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) RegisterKeeperToTriggerX(operator common.Address, operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.RegisterKeeperToTriggerX(&_ContractTriggerXServiceManager.TransactOpts, operator, operatorSignature)
}

// RegisterOperatorToAVS is a paid mutator transaction binding the contract method 0x9926ee7d.
//
// Solidity: function registerOperatorToAVS(address operator, (bytes,bytes32,uint256) operatorSignature) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) RegisterOperatorToAVS(opts *bind.TransactOpts, operator common.Address, operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "registerOperatorToAVS", operator, operatorSignature)
}

// RegisterOperatorToAVS is a paid mutator transaction binding the contract method 0x9926ee7d.
//
// Solidity: function registerOperatorToAVS(address operator, (bytes,bytes32,uint256) operatorSignature) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) RegisterOperatorToAVS(operator common.Address, operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.RegisterOperatorToAVS(&_ContractTriggerXServiceManager.TransactOpts, operator, operatorSignature)
}

// RegisterOperatorToAVS is a paid mutator transaction binding the contract method 0x9926ee7d.
//
// Solidity: function registerOperatorToAVS(address operator, (bytes,bytes32,uint256) operatorSignature) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) RegisterOperatorToAVS(operator common.Address, operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.RegisterOperatorToAVS(&_ContractTriggerXServiceManager.TransactOpts, operator, operatorSignature)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) RenounceOwnership() (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.RenounceOwnership(&_ContractTriggerXServiceManager.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.RenounceOwnership(&_ContractTriggerXServiceManager.TransactOpts)
}

// SetClaimerFor is a paid mutator transaction binding the contract method 0xa0169ddd.
//
// Solidity: function setClaimerFor(address claimer) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) SetClaimerFor(opts *bind.TransactOpts, claimer common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "setClaimerFor", claimer)
}

// SetClaimerFor is a paid mutator transaction binding the contract method 0xa0169ddd.
//
// Solidity: function setClaimerFor(address claimer) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) SetClaimerFor(claimer common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetClaimerFor(&_ContractTriggerXServiceManager.TransactOpts, claimer)
}

// SetClaimerFor is a paid mutator transaction binding the contract method 0xa0169ddd.
//
// Solidity: function setClaimerFor(address claimer) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) SetClaimerFor(claimer common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetClaimerFor(&_ContractTriggerXServiceManager.TransactOpts, claimer)
}

// SetPauserRegistry is a paid mutator transaction binding the contract method 0x10d67a2f.
//
// Solidity: function setPauserRegistry(address newPauserRegistry) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) SetPauserRegistry(opts *bind.TransactOpts, newPauserRegistry common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "setPauserRegistry", newPauserRegistry)
}

// SetPauserRegistry is a paid mutator transaction binding the contract method 0x10d67a2f.
//
// Solidity: function setPauserRegistry(address newPauserRegistry) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) SetPauserRegistry(newPauserRegistry common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetPauserRegistry(&_ContractTriggerXServiceManager.TransactOpts, newPauserRegistry)
}

// SetPauserRegistry is a paid mutator transaction binding the contract method 0x10d67a2f.
//
// Solidity: function setPauserRegistry(address newPauserRegistry) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) SetPauserRegistry(newPauserRegistry common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetPauserRegistry(&_ContractTriggerXServiceManager.TransactOpts, newPauserRegistry)
}

// SetQuorumManager is a paid mutator transaction binding the contract method 0x9cd0fb86.
//
// Solidity: function setQuorumManager(address _quorumManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) SetQuorumManager(opts *bind.TransactOpts, _quorumManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "setQuorumManager", _quorumManager)
}

// SetQuorumManager is a paid mutator transaction binding the contract method 0x9cd0fb86.
//
// Solidity: function setQuorumManager(address _quorumManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) SetQuorumManager(_quorumManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetQuorumManager(&_ContractTriggerXServiceManager.TransactOpts, _quorumManager)
}

// SetQuorumManager is a paid mutator transaction binding the contract method 0x9cd0fb86.
//
// Solidity: function setQuorumManager(address _quorumManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) SetQuorumManager(_quorumManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetQuorumManager(&_ContractTriggerXServiceManager.TransactOpts, _quorumManager)
}

// SetRewardsInitiator is a paid mutator transaction binding the contract method 0x3bc28c8c.
//
// Solidity: function setRewardsInitiator(address newRewardsInitiator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) SetRewardsInitiator(opts *bind.TransactOpts, newRewardsInitiator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "setRewardsInitiator", newRewardsInitiator)
}

// SetRewardsInitiator is a paid mutator transaction binding the contract method 0x3bc28c8c.
//
// Solidity: function setRewardsInitiator(address newRewardsInitiator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) SetRewardsInitiator(newRewardsInitiator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetRewardsInitiator(&_ContractTriggerXServiceManager.TransactOpts, newRewardsInitiator)
}

// SetRewardsInitiator is a paid mutator transaction binding the contract method 0x3bc28c8c.
//
// Solidity: function setRewardsInitiator(address newRewardsInitiator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) SetRewardsInitiator(newRewardsInitiator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetRewardsInitiator(&_ContractTriggerXServiceManager.TransactOpts, newRewardsInitiator)
}

// SetStaleStakesForbidden is a paid mutator transaction binding the contract method 0x416c7e5e.
//
// Solidity: function setStaleStakesForbidden(bool value) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) SetStaleStakesForbidden(opts *bind.TransactOpts, value bool) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "setStaleStakesForbidden", value)
}

// SetStaleStakesForbidden is a paid mutator transaction binding the contract method 0x416c7e5e.
//
// Solidity: function setStaleStakesForbidden(bool value) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) SetStaleStakesForbidden(value bool) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetStaleStakesForbidden(&_ContractTriggerXServiceManager.TransactOpts, value)
}

// SetStaleStakesForbidden is a paid mutator transaction binding the contract method 0x416c7e5e.
//
// Solidity: function setStaleStakesForbidden(bool value) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) SetStaleStakesForbidden(value bool) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetStaleStakesForbidden(&_ContractTriggerXServiceManager.TransactOpts, value)
}

// SetTaskManager is a paid mutator transaction binding the contract method 0x327d0a60.
//
// Solidity: function setTaskManager(address _taskManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) SetTaskManager(opts *bind.TransactOpts, _taskManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "setTaskManager", _taskManager)
}

// SetTaskManager is a paid mutator transaction binding the contract method 0x327d0a60.
//
// Solidity: function setTaskManager(address _taskManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) SetTaskManager(_taskManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetTaskManager(&_ContractTriggerXServiceManager.TransactOpts, _taskManager)
}

// SetTaskManager is a paid mutator transaction binding the contract method 0x327d0a60.
//
// Solidity: function setTaskManager(address _taskManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) SetTaskManager(_taskManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetTaskManager(&_ContractTriggerXServiceManager.TransactOpts, _taskManager)
}

// SetTaskValidator is a paid mutator transaction binding the contract method 0x5edd78ea.
//
// Solidity: function setTaskValidator(address _taskValidator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) SetTaskValidator(opts *bind.TransactOpts, _taskValidator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "setTaskValidator", _taskValidator)
}

// SetTaskValidator is a paid mutator transaction binding the contract method 0x5edd78ea.
//
// Solidity: function setTaskValidator(address _taskValidator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) SetTaskValidator(_taskValidator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetTaskValidator(&_ContractTriggerXServiceManager.TransactOpts, _taskValidator)
}

// SetTaskValidator is a paid mutator transaction binding the contract method 0x5edd78ea.
//
// Solidity: function setTaskValidator(address _taskValidator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) SetTaskValidator(_taskValidator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.SetTaskValidator(&_ContractTriggerXServiceManager.TransactOpts, _taskValidator)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.TransferOwnership(&_ContractTriggerXServiceManager.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.TransferOwnership(&_ContractTriggerXServiceManager.TransactOpts, newOwner)
}

// UnblacklistKeeper is a paid mutator transaction binding the contract method 0xa41d3f94.
//
// Solidity: function unblacklistKeeper(address _operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) UnblacklistKeeper(opts *bind.TransactOpts, _operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "unblacklistKeeper", _operator)
}

// UnblacklistKeeper is a paid mutator transaction binding the contract method 0xa41d3f94.
//
// Solidity: function unblacklistKeeper(address _operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) UnblacklistKeeper(_operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.UnblacklistKeeper(&_ContractTriggerXServiceManager.TransactOpts, _operator)
}

// UnblacklistKeeper is a paid mutator transaction binding the contract method 0xa41d3f94.
//
// Solidity: function unblacklistKeeper(address _operator) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) UnblacklistKeeper(_operator common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.UnblacklistKeeper(&_ContractTriggerXServiceManager.TransactOpts, _operator)
}

// Unpause is a paid mutator transaction binding the contract method 0xfabc1cbc.
//
// Solidity: function unpause(uint256 newPausedStatus) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) Unpause(opts *bind.TransactOpts, newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "unpause", newPausedStatus)
}

// Unpause is a paid mutator transaction binding the contract method 0xfabc1cbc.
//
// Solidity: function unpause(uint256 newPausedStatus) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) Unpause(newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.Unpause(&_ContractTriggerXServiceManager.TransactOpts, newPausedStatus)
}

// Unpause is a paid mutator transaction binding the contract method 0xfabc1cbc.
//
// Solidity: function unpause(uint256 newPausedStatus) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) Unpause(newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.Unpause(&_ContractTriggerXServiceManager.TransactOpts, newPausedStatus)
}

// UpdateAVSMetadataURI is a paid mutator transaction binding the contract method 0xa98fb355.
//
// Solidity: function updateAVSMetadataURI(string _metadataURI) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) UpdateAVSMetadataURI(opts *bind.TransactOpts, _metadataURI string) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "updateAVSMetadataURI", _metadataURI)
}

// UpdateAVSMetadataURI is a paid mutator transaction binding the contract method 0xa98fb355.
//
// Solidity: function updateAVSMetadataURI(string _metadataURI) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) UpdateAVSMetadataURI(_metadataURI string) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.UpdateAVSMetadataURI(&_ContractTriggerXServiceManager.TransactOpts, _metadataURI)
}

// UpdateAVSMetadataURI is a paid mutator transaction binding the contract method 0xa98fb355.
//
// Solidity: function updateAVSMetadataURI(string _metadataURI) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) UpdateAVSMetadataURI(_metadataURI string) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.UpdateAVSMetadataURI(&_ContractTriggerXServiceManager.TransactOpts, _metadataURI)
}

// UpdateTaskManager is a paid mutator transaction binding the contract method 0x58ac4a1e.
//
// Solidity: function updateTaskManager(address _taskManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactor) UpdateTaskManager(opts *bind.TransactOpts, _taskManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.contract.Transact(opts, "updateTaskManager", _taskManager)
}

// UpdateTaskManager is a paid mutator transaction binding the contract method 0x58ac4a1e.
//
// Solidity: function updateTaskManager(address _taskManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerSession) UpdateTaskManager(_taskManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.UpdateTaskManager(&_ContractTriggerXServiceManager.TransactOpts, _taskManager)
}

// UpdateTaskManager is a paid mutator transaction binding the contract method 0x58ac4a1e.
//
// Solidity: function updateTaskManager(address _taskManager) returns()
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerTransactorSession) UpdateTaskManager(_taskManager common.Address) (*types.Transaction, error) {
	return _ContractTriggerXServiceManager.Contract.UpdateTaskManager(&_ContractTriggerXServiceManager.TransactOpts, _taskManager)
}

// ContractTriggerXServiceManagerInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerInitializedIterator struct {
	Event *ContractTriggerXServiceManagerInitialized // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerInitialized)
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
		it.Event = new(ContractTriggerXServiceManagerInitialized)
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
func (it *ContractTriggerXServiceManagerInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerInitialized represents a Initialized event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterInitialized(opts *bind.FilterOpts) (*ContractTriggerXServiceManagerInitializedIterator, error) {

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerInitializedIterator{contract: _ContractTriggerXServiceManager.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerInitialized) (event.Subscription, error) {

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerInitialized)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseInitialized(log types.Log) (*ContractTriggerXServiceManagerInitialized, error) {
	event := new(ContractTriggerXServiceManagerInitialized)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerKeeperAddedIterator is returned from FilterKeeperAdded and is used to iterate over the raw logs and unpacked data for KeeperAdded events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerKeeperAddedIterator struct {
	Event *ContractTriggerXServiceManagerKeeperAdded // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerKeeperAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerKeeperAdded)
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
		it.Event = new(ContractTriggerXServiceManagerKeeperAdded)
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
func (it *ContractTriggerXServiceManagerKeeperAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerKeeperAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerKeeperAdded represents a KeeperAdded event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerKeeperAdded struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterKeeperAdded is a free log retrieval operation binding the contract event 0x1584773458d98c71b34a270ee1100b3a42889bf91e3b7a858563b684c24d838e.
//
// Solidity: event KeeperAdded(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterKeeperAdded(opts *bind.FilterOpts, operator []common.Address) (*ContractTriggerXServiceManagerKeeperAddedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "KeeperAdded", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerKeeperAddedIterator{contract: _ContractTriggerXServiceManager.contract, event: "KeeperAdded", logs: logs, sub: sub}, nil
}

// WatchKeeperAdded is a free log subscription operation binding the contract event 0x1584773458d98c71b34a270ee1100b3a42889bf91e3b7a858563b684c24d838e.
//
// Solidity: event KeeperAdded(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchKeeperAdded(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerKeeperAdded, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "KeeperAdded", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerKeeperAdded)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "KeeperAdded", log); err != nil {
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

// ParseKeeperAdded is a log parse operation binding the contract event 0x1584773458d98c71b34a270ee1100b3a42889bf91e3b7a858563b684c24d838e.
//
// Solidity: event KeeperAdded(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseKeeperAdded(log types.Log) (*ContractTriggerXServiceManagerKeeperAdded, error) {
	event := new(ContractTriggerXServiceManagerKeeperAdded)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "KeeperAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerKeeperBlacklistedIterator is returned from FilterKeeperBlacklisted and is used to iterate over the raw logs and unpacked data for KeeperBlacklisted events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerKeeperBlacklistedIterator struct {
	Event *ContractTriggerXServiceManagerKeeperBlacklisted // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerKeeperBlacklistedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerKeeperBlacklisted)
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
		it.Event = new(ContractTriggerXServiceManagerKeeperBlacklisted)
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
func (it *ContractTriggerXServiceManagerKeeperBlacklistedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerKeeperBlacklistedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerKeeperBlacklisted represents a KeeperBlacklisted event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerKeeperBlacklisted struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterKeeperBlacklisted is a free log retrieval operation binding the contract event 0x79bf5efab685b39bdd2b7b1aba09ffa5e9dfbd86f7aaebab4b52d7ef679d2844.
//
// Solidity: event KeeperBlacklisted(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterKeeperBlacklisted(opts *bind.FilterOpts, operator []common.Address) (*ContractTriggerXServiceManagerKeeperBlacklistedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "KeeperBlacklisted", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerKeeperBlacklistedIterator{contract: _ContractTriggerXServiceManager.contract, event: "KeeperBlacklisted", logs: logs, sub: sub}, nil
}

// WatchKeeperBlacklisted is a free log subscription operation binding the contract event 0x79bf5efab685b39bdd2b7b1aba09ffa5e9dfbd86f7aaebab4b52d7ef679d2844.
//
// Solidity: event KeeperBlacklisted(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchKeeperBlacklisted(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerKeeperBlacklisted, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "KeeperBlacklisted", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerKeeperBlacklisted)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "KeeperBlacklisted", log); err != nil {
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

// ParseKeeperBlacklisted is a log parse operation binding the contract event 0x79bf5efab685b39bdd2b7b1aba09ffa5e9dfbd86f7aaebab4b52d7ef679d2844.
//
// Solidity: event KeeperBlacklisted(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseKeeperBlacklisted(log types.Log) (*ContractTriggerXServiceManagerKeeperBlacklisted, error) {
	event := new(ContractTriggerXServiceManagerKeeperBlacklisted)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "KeeperBlacklisted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerKeeperRemovedIterator is returned from FilterKeeperRemoved and is used to iterate over the raw logs and unpacked data for KeeperRemoved events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerKeeperRemovedIterator struct {
	Event *ContractTriggerXServiceManagerKeeperRemoved // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerKeeperRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerKeeperRemoved)
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
		it.Event = new(ContractTriggerXServiceManagerKeeperRemoved)
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
func (it *ContractTriggerXServiceManagerKeeperRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerKeeperRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerKeeperRemoved represents a KeeperRemoved event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerKeeperRemoved struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterKeeperRemoved is a free log retrieval operation binding the contract event 0xa7a775c2c8141f7985c111748ec31c11e5e44b83528e105c8d1d4e8e6b81cf80.
//
// Solidity: event KeeperRemoved(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterKeeperRemoved(opts *bind.FilterOpts, operator []common.Address) (*ContractTriggerXServiceManagerKeeperRemovedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "KeeperRemoved", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerKeeperRemovedIterator{contract: _ContractTriggerXServiceManager.contract, event: "KeeperRemoved", logs: logs, sub: sub}, nil
}

// WatchKeeperRemoved is a free log subscription operation binding the contract event 0xa7a775c2c8141f7985c111748ec31c11e5e44b83528e105c8d1d4e8e6b81cf80.
//
// Solidity: event KeeperRemoved(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchKeeperRemoved(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerKeeperRemoved, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "KeeperRemoved", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerKeeperRemoved)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "KeeperRemoved", log); err != nil {
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

// ParseKeeperRemoved is a log parse operation binding the contract event 0xa7a775c2c8141f7985c111748ec31c11e5e44b83528e105c8d1d4e8e6b81cf80.
//
// Solidity: event KeeperRemoved(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseKeeperRemoved(log types.Log) (*ContractTriggerXServiceManagerKeeperRemoved, error) {
	event := new(ContractTriggerXServiceManagerKeeperRemoved)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "KeeperRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerKeeperUnblacklistedIterator is returned from FilterKeeperUnblacklisted and is used to iterate over the raw logs and unpacked data for KeeperUnblacklisted events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerKeeperUnblacklistedIterator struct {
	Event *ContractTriggerXServiceManagerKeeperUnblacklisted // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerKeeperUnblacklistedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerKeeperUnblacklisted)
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
		it.Event = new(ContractTriggerXServiceManagerKeeperUnblacklisted)
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
func (it *ContractTriggerXServiceManagerKeeperUnblacklistedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerKeeperUnblacklistedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerKeeperUnblacklisted represents a KeeperUnblacklisted event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerKeeperUnblacklisted struct {
	Operator common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterKeeperUnblacklisted is a free log retrieval operation binding the contract event 0x8511046d95f95689c89b4807d91820adbc266db891b2c5c25a42f812781dde57.
//
// Solidity: event KeeperUnblacklisted(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterKeeperUnblacklisted(opts *bind.FilterOpts, operator []common.Address) (*ContractTriggerXServiceManagerKeeperUnblacklistedIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "KeeperUnblacklisted", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerKeeperUnblacklistedIterator{contract: _ContractTriggerXServiceManager.contract, event: "KeeperUnblacklisted", logs: logs, sub: sub}, nil
}

// WatchKeeperUnblacklisted is a free log subscription operation binding the contract event 0x8511046d95f95689c89b4807d91820adbc266db891b2c5c25a42f812781dde57.
//
// Solidity: event KeeperUnblacklisted(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchKeeperUnblacklisted(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerKeeperUnblacklisted, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "KeeperUnblacklisted", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerKeeperUnblacklisted)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "KeeperUnblacklisted", log); err != nil {
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

// ParseKeeperUnblacklisted is a log parse operation binding the contract event 0x8511046d95f95689c89b4807d91820adbc266db891b2c5c25a42f812781dde57.
//
// Solidity: event KeeperUnblacklisted(address indexed operator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseKeeperUnblacklisted(log types.Log) (*ContractTriggerXServiceManagerKeeperUnblacklisted, error) {
	event := new(ContractTriggerXServiceManagerKeeperUnblacklisted)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "KeeperUnblacklisted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerOwnershipTransferredIterator struct {
	Event *ContractTriggerXServiceManagerOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerOwnershipTransferred)
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
		it.Event = new(ContractTriggerXServiceManagerOwnershipTransferred)
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
func (it *ContractTriggerXServiceManagerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerOwnershipTransferred represents a OwnershipTransferred event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractTriggerXServiceManagerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerOwnershipTransferredIterator{contract: _ContractTriggerXServiceManager.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerOwnershipTransferred)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseOwnershipTransferred(log types.Log) (*ContractTriggerXServiceManagerOwnershipTransferred, error) {
	event := new(ContractTriggerXServiceManagerOwnershipTransferred)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerPausedIterator struct {
	Event *ContractTriggerXServiceManagerPaused // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerPaused)
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
		it.Event = new(ContractTriggerXServiceManagerPaused)
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
func (it *ContractTriggerXServiceManagerPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerPaused represents a Paused event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerPaused struct {
	Account         common.Address
	NewPausedStatus *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0xab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d.
//
// Solidity: event Paused(address indexed account, uint256 newPausedStatus)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterPaused(opts *bind.FilterOpts, account []common.Address) (*ContractTriggerXServiceManagerPausedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "Paused", accountRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerPausedIterator{contract: _ContractTriggerXServiceManager.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0xab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d.
//
// Solidity: event Paused(address indexed account, uint256 newPausedStatus)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerPaused, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "Paused", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerPaused)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "Paused", log); err != nil {
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

// ParsePaused is a log parse operation binding the contract event 0xab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d.
//
// Solidity: event Paused(address indexed account, uint256 newPausedStatus)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParsePaused(log types.Log) (*ContractTriggerXServiceManagerPaused, error) {
	event := new(ContractTriggerXServiceManagerPaused)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerPauserRegistrySetIterator is returned from FilterPauserRegistrySet and is used to iterate over the raw logs and unpacked data for PauserRegistrySet events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerPauserRegistrySetIterator struct {
	Event *ContractTriggerXServiceManagerPauserRegistrySet // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerPauserRegistrySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerPauserRegistrySet)
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
		it.Event = new(ContractTriggerXServiceManagerPauserRegistrySet)
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
func (it *ContractTriggerXServiceManagerPauserRegistrySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerPauserRegistrySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerPauserRegistrySet represents a PauserRegistrySet event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerPauserRegistrySet struct {
	PauserRegistry    common.Address
	NewPauserRegistry common.Address
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterPauserRegistrySet is a free log retrieval operation binding the contract event 0x6e9fcd539896fca60e8b0f01dd580233e48a6b0f7df013b89ba7f565869acdb6.
//
// Solidity: event PauserRegistrySet(address pauserRegistry, address newPauserRegistry)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterPauserRegistrySet(opts *bind.FilterOpts) (*ContractTriggerXServiceManagerPauserRegistrySetIterator, error) {

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "PauserRegistrySet")
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerPauserRegistrySetIterator{contract: _ContractTriggerXServiceManager.contract, event: "PauserRegistrySet", logs: logs, sub: sub}, nil
}

// WatchPauserRegistrySet is a free log subscription operation binding the contract event 0x6e9fcd539896fca60e8b0f01dd580233e48a6b0f7df013b89ba7f565869acdb6.
//
// Solidity: event PauserRegistrySet(address pauserRegistry, address newPauserRegistry)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchPauserRegistrySet(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerPauserRegistrySet) (event.Subscription, error) {

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "PauserRegistrySet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerPauserRegistrySet)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "PauserRegistrySet", log); err != nil {
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

// ParsePauserRegistrySet is a log parse operation binding the contract event 0x6e9fcd539896fca60e8b0f01dd580233e48a6b0f7df013b89ba7f565869acdb6.
//
// Solidity: event PauserRegistrySet(address pauserRegistry, address newPauserRegistry)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParsePauserRegistrySet(log types.Log) (*ContractTriggerXServiceManagerPauserRegistrySet, error) {
	event := new(ContractTriggerXServiceManagerPauserRegistrySet)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "PauserRegistrySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerQuorumManagerSetIterator is returned from FilterQuorumManagerSet and is used to iterate over the raw logs and unpacked data for QuorumManagerSet events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerQuorumManagerSetIterator struct {
	Event *ContractTriggerXServiceManagerQuorumManagerSet // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerQuorumManagerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerQuorumManagerSet)
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
		it.Event = new(ContractTriggerXServiceManagerQuorumManagerSet)
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
func (it *ContractTriggerXServiceManagerQuorumManagerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerQuorumManagerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerQuorumManagerSet represents a QuorumManagerSet event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerQuorumManagerSet struct {
	OldQuorumManager common.Address
	NewQuorumManager common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterQuorumManagerSet is a free log retrieval operation binding the contract event 0x5d68f69aebb8a1d75f8b2a87dddd6e8db842f9fa3bb63745ea7f49ea5bdf03a9.
//
// Solidity: event QuorumManagerSet(address indexed oldQuorumManager, address indexed newQuorumManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterQuorumManagerSet(opts *bind.FilterOpts, oldQuorumManager []common.Address, newQuorumManager []common.Address) (*ContractTriggerXServiceManagerQuorumManagerSetIterator, error) {

	var oldQuorumManagerRule []interface{}
	for _, oldQuorumManagerItem := range oldQuorumManager {
		oldQuorumManagerRule = append(oldQuorumManagerRule, oldQuorumManagerItem)
	}
	var newQuorumManagerRule []interface{}
	for _, newQuorumManagerItem := range newQuorumManager {
		newQuorumManagerRule = append(newQuorumManagerRule, newQuorumManagerItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "QuorumManagerSet", oldQuorumManagerRule, newQuorumManagerRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerQuorumManagerSetIterator{contract: _ContractTriggerXServiceManager.contract, event: "QuorumManagerSet", logs: logs, sub: sub}, nil
}

// WatchQuorumManagerSet is a free log subscription operation binding the contract event 0x5d68f69aebb8a1d75f8b2a87dddd6e8db842f9fa3bb63745ea7f49ea5bdf03a9.
//
// Solidity: event QuorumManagerSet(address indexed oldQuorumManager, address indexed newQuorumManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchQuorumManagerSet(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerQuorumManagerSet, oldQuorumManager []common.Address, newQuorumManager []common.Address) (event.Subscription, error) {

	var oldQuorumManagerRule []interface{}
	for _, oldQuorumManagerItem := range oldQuorumManager {
		oldQuorumManagerRule = append(oldQuorumManagerRule, oldQuorumManagerItem)
	}
	var newQuorumManagerRule []interface{}
	for _, newQuorumManagerItem := range newQuorumManager {
		newQuorumManagerRule = append(newQuorumManagerRule, newQuorumManagerItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "QuorumManagerSet", oldQuorumManagerRule, newQuorumManagerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerQuorumManagerSet)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "QuorumManagerSet", log); err != nil {
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

// ParseQuorumManagerSet is a log parse operation binding the contract event 0x5d68f69aebb8a1d75f8b2a87dddd6e8db842f9fa3bb63745ea7f49ea5bdf03a9.
//
// Solidity: event QuorumManagerSet(address indexed oldQuorumManager, address indexed newQuorumManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseQuorumManagerSet(log types.Log) (*ContractTriggerXServiceManagerQuorumManagerSet, error) {
	event := new(ContractTriggerXServiceManagerQuorumManagerSet)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "QuorumManagerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerRewardsInitiatorUpdatedIterator is returned from FilterRewardsInitiatorUpdated and is used to iterate over the raw logs and unpacked data for RewardsInitiatorUpdated events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerRewardsInitiatorUpdatedIterator struct {
	Event *ContractTriggerXServiceManagerRewardsInitiatorUpdated // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerRewardsInitiatorUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerRewardsInitiatorUpdated)
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
		it.Event = new(ContractTriggerXServiceManagerRewardsInitiatorUpdated)
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
func (it *ContractTriggerXServiceManagerRewardsInitiatorUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerRewardsInitiatorUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerRewardsInitiatorUpdated represents a RewardsInitiatorUpdated event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerRewardsInitiatorUpdated struct {
	PrevRewardsInitiator common.Address
	NewRewardsInitiator  common.Address
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterRewardsInitiatorUpdated is a free log retrieval operation binding the contract event 0xe11cddf1816a43318ca175bbc52cd0185436e9cbead7c83acc54a73e461717e3.
//
// Solidity: event RewardsInitiatorUpdated(address prevRewardsInitiator, address newRewardsInitiator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterRewardsInitiatorUpdated(opts *bind.FilterOpts) (*ContractTriggerXServiceManagerRewardsInitiatorUpdatedIterator, error) {

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "RewardsInitiatorUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerRewardsInitiatorUpdatedIterator{contract: _ContractTriggerXServiceManager.contract, event: "RewardsInitiatorUpdated", logs: logs, sub: sub}, nil
}

// WatchRewardsInitiatorUpdated is a free log subscription operation binding the contract event 0xe11cddf1816a43318ca175bbc52cd0185436e9cbead7c83acc54a73e461717e3.
//
// Solidity: event RewardsInitiatorUpdated(address prevRewardsInitiator, address newRewardsInitiator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchRewardsInitiatorUpdated(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerRewardsInitiatorUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "RewardsInitiatorUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerRewardsInitiatorUpdated)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "RewardsInitiatorUpdated", log); err != nil {
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

// ParseRewardsInitiatorUpdated is a log parse operation binding the contract event 0xe11cddf1816a43318ca175bbc52cd0185436e9cbead7c83acc54a73e461717e3.
//
// Solidity: event RewardsInitiatorUpdated(address prevRewardsInitiator, address newRewardsInitiator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseRewardsInitiatorUpdated(log types.Log) (*ContractTriggerXServiceManagerRewardsInitiatorUpdated, error) {
	event := new(ContractTriggerXServiceManagerRewardsInitiatorUpdated)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "RewardsInitiatorUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerStaleStakesForbiddenUpdateIterator is returned from FilterStaleStakesForbiddenUpdate and is used to iterate over the raw logs and unpacked data for StaleStakesForbiddenUpdate events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerStaleStakesForbiddenUpdateIterator struct {
	Event *ContractTriggerXServiceManagerStaleStakesForbiddenUpdate // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerStaleStakesForbiddenUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerStaleStakesForbiddenUpdate)
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
		it.Event = new(ContractTriggerXServiceManagerStaleStakesForbiddenUpdate)
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
func (it *ContractTriggerXServiceManagerStaleStakesForbiddenUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerStaleStakesForbiddenUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerStaleStakesForbiddenUpdate represents a StaleStakesForbiddenUpdate event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerStaleStakesForbiddenUpdate struct {
	Value bool
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterStaleStakesForbiddenUpdate is a free log retrieval operation binding the contract event 0x40e4ed880a29e0f6ddce307457fb75cddf4feef7d3ecb0301bfdf4976a0e2dfc.
//
// Solidity: event StaleStakesForbiddenUpdate(bool value)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterStaleStakesForbiddenUpdate(opts *bind.FilterOpts) (*ContractTriggerXServiceManagerStaleStakesForbiddenUpdateIterator, error) {

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "StaleStakesForbiddenUpdate")
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerStaleStakesForbiddenUpdateIterator{contract: _ContractTriggerXServiceManager.contract, event: "StaleStakesForbiddenUpdate", logs: logs, sub: sub}, nil
}

// WatchStaleStakesForbiddenUpdate is a free log subscription operation binding the contract event 0x40e4ed880a29e0f6ddce307457fb75cddf4feef7d3ecb0301bfdf4976a0e2dfc.
//
// Solidity: event StaleStakesForbiddenUpdate(bool value)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchStaleStakesForbiddenUpdate(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerStaleStakesForbiddenUpdate) (event.Subscription, error) {

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "StaleStakesForbiddenUpdate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerStaleStakesForbiddenUpdate)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "StaleStakesForbiddenUpdate", log); err != nil {
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

// ParseStaleStakesForbiddenUpdate is a log parse operation binding the contract event 0x40e4ed880a29e0f6ddce307457fb75cddf4feef7d3ecb0301bfdf4976a0e2dfc.
//
// Solidity: event StaleStakesForbiddenUpdate(bool value)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseStaleStakesForbiddenUpdate(log types.Log) (*ContractTriggerXServiceManagerStaleStakesForbiddenUpdate, error) {
	event := new(ContractTriggerXServiceManagerStaleStakesForbiddenUpdate)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "StaleStakesForbiddenUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerTaskManagerContractUpdatedIterator is returned from FilterTaskManagerContractUpdated and is used to iterate over the raw logs and unpacked data for TaskManagerContractUpdated events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerTaskManagerContractUpdatedIterator struct {
	Event *ContractTriggerXServiceManagerTaskManagerContractUpdated // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerTaskManagerContractUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerTaskManagerContractUpdated)
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
		it.Event = new(ContractTriggerXServiceManagerTaskManagerContractUpdated)
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
func (it *ContractTriggerXServiceManagerTaskManagerContractUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerTaskManagerContractUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerTaskManagerContractUpdated represents a TaskManagerContractUpdated event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerTaskManagerContractUpdated struct {
	OldTaskManager common.Address
	NewTaskManager common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterTaskManagerContractUpdated is a free log retrieval operation binding the contract event 0xecc36ab0616c35e413d90730ada58214cbe0188c5f26632fbb32451c9a135784.
//
// Solidity: event TaskManagerContractUpdated(address indexed oldTaskManager, address indexed newTaskManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterTaskManagerContractUpdated(opts *bind.FilterOpts, oldTaskManager []common.Address, newTaskManager []common.Address) (*ContractTriggerXServiceManagerTaskManagerContractUpdatedIterator, error) {

	var oldTaskManagerRule []interface{}
	for _, oldTaskManagerItem := range oldTaskManager {
		oldTaskManagerRule = append(oldTaskManagerRule, oldTaskManagerItem)
	}
	var newTaskManagerRule []interface{}
	for _, newTaskManagerItem := range newTaskManager {
		newTaskManagerRule = append(newTaskManagerRule, newTaskManagerItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "TaskManagerContractUpdated", oldTaskManagerRule, newTaskManagerRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerTaskManagerContractUpdatedIterator{contract: _ContractTriggerXServiceManager.contract, event: "TaskManagerContractUpdated", logs: logs, sub: sub}, nil
}

// WatchTaskManagerContractUpdated is a free log subscription operation binding the contract event 0xecc36ab0616c35e413d90730ada58214cbe0188c5f26632fbb32451c9a135784.
//
// Solidity: event TaskManagerContractUpdated(address indexed oldTaskManager, address indexed newTaskManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchTaskManagerContractUpdated(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerTaskManagerContractUpdated, oldTaskManager []common.Address, newTaskManager []common.Address) (event.Subscription, error) {

	var oldTaskManagerRule []interface{}
	for _, oldTaskManagerItem := range oldTaskManager {
		oldTaskManagerRule = append(oldTaskManagerRule, oldTaskManagerItem)
	}
	var newTaskManagerRule []interface{}
	for _, newTaskManagerItem := range newTaskManager {
		newTaskManagerRule = append(newTaskManagerRule, newTaskManagerItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "TaskManagerContractUpdated", oldTaskManagerRule, newTaskManagerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerTaskManagerContractUpdated)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "TaskManagerContractUpdated", log); err != nil {
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

// ParseTaskManagerContractUpdated is a log parse operation binding the contract event 0xecc36ab0616c35e413d90730ada58214cbe0188c5f26632fbb32451c9a135784.
//
// Solidity: event TaskManagerContractUpdated(address indexed oldTaskManager, address indexed newTaskManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseTaskManagerContractUpdated(log types.Log) (*ContractTriggerXServiceManagerTaskManagerContractUpdated, error) {
	event := new(ContractTriggerXServiceManagerTaskManagerContractUpdated)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "TaskManagerContractUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerTaskManagerSetIterator is returned from FilterTaskManagerSet and is used to iterate over the raw logs and unpacked data for TaskManagerSet events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerTaskManagerSetIterator struct {
	Event *ContractTriggerXServiceManagerTaskManagerSet // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerTaskManagerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerTaskManagerSet)
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
		it.Event = new(ContractTriggerXServiceManagerTaskManagerSet)
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
func (it *ContractTriggerXServiceManagerTaskManagerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerTaskManagerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerTaskManagerSet represents a TaskManagerSet event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerTaskManagerSet struct {
	OldTaskManager common.Address
	NewTaskManager common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterTaskManagerSet is a free log retrieval operation binding the contract event 0xdae652729007c1e5ea13dfe9e126a1b72358db8db69c68b96f316dcea59c1c5f.
//
// Solidity: event TaskManagerSet(address indexed oldTaskManager, address indexed newTaskManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterTaskManagerSet(opts *bind.FilterOpts, oldTaskManager []common.Address, newTaskManager []common.Address) (*ContractTriggerXServiceManagerTaskManagerSetIterator, error) {

	var oldTaskManagerRule []interface{}
	for _, oldTaskManagerItem := range oldTaskManager {
		oldTaskManagerRule = append(oldTaskManagerRule, oldTaskManagerItem)
	}
	var newTaskManagerRule []interface{}
	for _, newTaskManagerItem := range newTaskManager {
		newTaskManagerRule = append(newTaskManagerRule, newTaskManagerItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "TaskManagerSet", oldTaskManagerRule, newTaskManagerRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerTaskManagerSetIterator{contract: _ContractTriggerXServiceManager.contract, event: "TaskManagerSet", logs: logs, sub: sub}, nil
}

// WatchTaskManagerSet is a free log subscription operation binding the contract event 0xdae652729007c1e5ea13dfe9e126a1b72358db8db69c68b96f316dcea59c1c5f.
//
// Solidity: event TaskManagerSet(address indexed oldTaskManager, address indexed newTaskManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchTaskManagerSet(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerTaskManagerSet, oldTaskManager []common.Address, newTaskManager []common.Address) (event.Subscription, error) {

	var oldTaskManagerRule []interface{}
	for _, oldTaskManagerItem := range oldTaskManager {
		oldTaskManagerRule = append(oldTaskManagerRule, oldTaskManagerItem)
	}
	var newTaskManagerRule []interface{}
	for _, newTaskManagerItem := range newTaskManager {
		newTaskManagerRule = append(newTaskManagerRule, newTaskManagerItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "TaskManagerSet", oldTaskManagerRule, newTaskManagerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerTaskManagerSet)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "TaskManagerSet", log); err != nil {
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

// ParseTaskManagerSet is a log parse operation binding the contract event 0xdae652729007c1e5ea13dfe9e126a1b72358db8db69c68b96f316dcea59c1c5f.
//
// Solidity: event TaskManagerSet(address indexed oldTaskManager, address indexed newTaskManager)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseTaskManagerSet(log types.Log) (*ContractTriggerXServiceManagerTaskManagerSet, error) {
	event := new(ContractTriggerXServiceManagerTaskManagerSet)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "TaskManagerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerTaskValidatorSetIterator is returned from FilterTaskValidatorSet and is used to iterate over the raw logs and unpacked data for TaskValidatorSet events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerTaskValidatorSetIterator struct {
	Event *ContractTriggerXServiceManagerTaskValidatorSet // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerTaskValidatorSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerTaskValidatorSet)
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
		it.Event = new(ContractTriggerXServiceManagerTaskValidatorSet)
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
func (it *ContractTriggerXServiceManagerTaskValidatorSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerTaskValidatorSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerTaskValidatorSet represents a TaskValidatorSet event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerTaskValidatorSet struct {
	OldTaskValidator common.Address
	NewTaskValidator common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterTaskValidatorSet is a free log retrieval operation binding the contract event 0x86cebbccc5bbc9ac544f7e5361e6c5e4acf14f39d85c855032a815125cb97fd8.
//
// Solidity: event TaskValidatorSet(address indexed oldTaskValidator, address indexed newTaskValidator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterTaskValidatorSet(opts *bind.FilterOpts, oldTaskValidator []common.Address, newTaskValidator []common.Address) (*ContractTriggerXServiceManagerTaskValidatorSetIterator, error) {

	var oldTaskValidatorRule []interface{}
	for _, oldTaskValidatorItem := range oldTaskValidator {
		oldTaskValidatorRule = append(oldTaskValidatorRule, oldTaskValidatorItem)
	}
	var newTaskValidatorRule []interface{}
	for _, newTaskValidatorItem := range newTaskValidator {
		newTaskValidatorRule = append(newTaskValidatorRule, newTaskValidatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "TaskValidatorSet", oldTaskValidatorRule, newTaskValidatorRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerTaskValidatorSetIterator{contract: _ContractTriggerXServiceManager.contract, event: "TaskValidatorSet", logs: logs, sub: sub}, nil
}

// WatchTaskValidatorSet is a free log subscription operation binding the contract event 0x86cebbccc5bbc9ac544f7e5361e6c5e4acf14f39d85c855032a815125cb97fd8.
//
// Solidity: event TaskValidatorSet(address indexed oldTaskValidator, address indexed newTaskValidator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchTaskValidatorSet(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerTaskValidatorSet, oldTaskValidator []common.Address, newTaskValidator []common.Address) (event.Subscription, error) {

	var oldTaskValidatorRule []interface{}
	for _, oldTaskValidatorItem := range oldTaskValidator {
		oldTaskValidatorRule = append(oldTaskValidatorRule, oldTaskValidatorItem)
	}
	var newTaskValidatorRule []interface{}
	for _, newTaskValidatorItem := range newTaskValidator {
		newTaskValidatorRule = append(newTaskValidatorRule, newTaskValidatorItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "TaskValidatorSet", oldTaskValidatorRule, newTaskValidatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerTaskValidatorSet)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "TaskValidatorSet", log); err != nil {
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

// ParseTaskValidatorSet is a log parse operation binding the contract event 0x86cebbccc5bbc9ac544f7e5361e6c5e4acf14f39d85c855032a815125cb97fd8.
//
// Solidity: event TaskValidatorSet(address indexed oldTaskValidator, address indexed newTaskValidator)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseTaskValidatorSet(log types.Log) (*ContractTriggerXServiceManagerTaskValidatorSet, error) {
	event := new(ContractTriggerXServiceManagerTaskValidatorSet)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "TaskValidatorSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractTriggerXServiceManagerUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerUnpausedIterator struct {
	Event *ContractTriggerXServiceManagerUnpaused // Event containing the contract specifics and raw log

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
func (it *ContractTriggerXServiceManagerUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTriggerXServiceManagerUnpaused)
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
		it.Event = new(ContractTriggerXServiceManagerUnpaused)
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
func (it *ContractTriggerXServiceManagerUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTriggerXServiceManagerUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTriggerXServiceManagerUnpaused represents a Unpaused event raised by the ContractTriggerXServiceManager contract.
type ContractTriggerXServiceManagerUnpaused struct {
	Account         common.Address
	NewPausedStatus *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x3582d1828e26bf56bd801502bc021ac0bc8afb57c826e4986b45593c8fad389c.
//
// Solidity: event Unpaused(address indexed account, uint256 newPausedStatus)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) FilterUnpaused(opts *bind.FilterOpts, account []common.Address) (*ContractTriggerXServiceManagerUnpausedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.FilterLogs(opts, "Unpaused", accountRule)
	if err != nil {
		return nil, err
	}
	return &ContractTriggerXServiceManagerUnpausedIterator{contract: _ContractTriggerXServiceManager.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x3582d1828e26bf56bd801502bc021ac0bc8afb57c826e4986b45593c8fad389c.
//
// Solidity: event Unpaused(address indexed account, uint256 newPausedStatus)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *ContractTriggerXServiceManagerUnpaused, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ContractTriggerXServiceManager.contract.WatchLogs(opts, "Unpaused", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTriggerXServiceManagerUnpaused)
				if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "Unpaused", log); err != nil {
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

// ParseUnpaused is a log parse operation binding the contract event 0x3582d1828e26bf56bd801502bc021ac0bc8afb57c826e4986b45593c8fad389c.
//
// Solidity: event Unpaused(address indexed account, uint256 newPausedStatus)
func (_ContractTriggerXServiceManager *ContractTriggerXServiceManagerFilterer) ParseUnpaused(log types.Log) (*ContractTriggerXServiceManagerUnpaused, error) {
	event := new(ContractTriggerXServiceManagerUnpaused)
	if err := _ContractTriggerXServiceManager.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
