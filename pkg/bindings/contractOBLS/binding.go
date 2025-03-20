// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractOBLS

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

// IOBLSOperatorVotingPower is an auto generated low-level Go binding around an user-defined struct.
type IOBLSOperatorVotingPower struct {
	OperatorId  *big.Int
	VotingPower *big.Int
}

// OBLSMetaData contains all meta data concerning the OBLS contract.
var OBLSMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decreaseBatchOperatorVotingPower\",\"inputs\":[{\"name\":\"_operatorsVotingPower\",\"type\":\"tuple[]\",\"internalType\":\"structIOBLS.OperatorVotingPower[]\",\"components\":[{\"name\":\"operatorId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"decreaseOperatorVotingPower\",\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"decreaseOperatorVotingPowerPerTaskDefinition\",\"inputs\":[{\"name\":\"_taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"_votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getOblsManager\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashToPoint\",\"inputs\":[{\"name\":\"domain\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"increaseBatchOperatorVotingPower\",\"inputs\":[{\"name\":\"_operatorsVotingPower\",\"type\":\"tuple[]\",\"internalType\":\"structIOBLS.OperatorVotingPower[]\",\"components\":[{\"name\":\"operatorId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"increaseOperatorVotingPower\",\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"increaseOperatorVotingPowerPerTaskDefinition\",\"inputs\":[{\"name\":\"_taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"_votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isActive\",\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"modifyOperatorActiveStatus\",\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_isActive\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"modifyOperatorBlsKey\",\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_blsKey\",\"type\":\"uint256[4]\",\"internalType\":\"uint256[4]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerOperator\",\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_votingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_blsKey\",\"type\":\"uint256[4]\",\"internalType\":\"uint256[4]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOblsManager\",\"inputs\":[{\"name\":\"_oblsManager\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOblsSharesSyncer\",\"inputs\":[{\"name\":\"_oblsSharesSyncer\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTotalVotingPowerPerRestrictedTaskDefinition\",\"inputs\":[{\"name\":\"_taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"_minimumVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_restrictedOperatorIndexes\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTotalVotingPowerPerTaskDefinition\",\"inputs\":[{\"name\":\"_taskDefinitionId\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"_numOfTotalOperators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_minimumVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalVotingPower\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalVotingPowerPerTaskDefinition\",\"inputs\":[{\"name\":\"_id\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unRegisterOperator\",\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"verifyAuthSignature\",\"inputs\":[{\"name\":\"_signature\",\"type\":\"tuple\",\"internalType\":\"structBLSAuthLibrary.Signature\",\"components\":[{\"name\":\"signature\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"}]},{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_contract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_blsKey\",\"type\":\"uint256[4]\",\"internalType\":\"uint256[4]\"}],\"outputs\":[],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifySignature\",\"inputs\":[{\"name\":\"_message\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"_signature\",\"type\":\"uint256[2]\",\"internalType\":\"uint256[2]\"},{\"name\":\"_indexes\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"},{\"name\":\"_requiredVotingPower\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_minimumVotingPowerPerTaskDefinition\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"votingPower\",\"inputs\":[{\"name\":\"_index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetOBLSManager\",\"inputs\":[{\"name\":\"newOBLSManager\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SharesSyncerModified\",\"inputs\":[{\"name\":\"syncer\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"BNAddCallFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"BadFTMappingImplementation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InactiveOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InsufficientVotingPower\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAuthSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidFieldElement\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOBLSSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidOperatorIndexes\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidRequiredVotingPower\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ModularInverseError\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotOBLSManager\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotOBLSManagerOrShareSyncer\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorDoesNotHaveMinimumVotingPower\",\"inputs\":[{\"name\":\"_operatorIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"PointNotOnCurve\",\"inputs\":[]}]",
}

// OBLSABI is the input ABI used to generate the binding from.
// Deprecated: Use OBLSMetaData.ABI instead.
var OBLSABI = OBLSMetaData.ABI

// OBLS is an auto generated Go binding around an Ethereum contract.
type OBLS struct {
	OBLSCaller     // Read-only binding to the contract
	OBLSTransactor // Write-only binding to the contract
	OBLSFilterer   // Log filterer for contract events
}

// OBLSCaller is an auto generated read-only Go binding around an Ethereum contract.
type OBLSCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OBLSTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OBLSTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OBLSFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OBLSFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OBLSSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OBLSSession struct {
	Contract     *OBLS             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OBLSCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OBLSCallerSession struct {
	Contract *OBLSCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// OBLSTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OBLSTransactorSession struct {
	Contract     *OBLSTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OBLSRaw is an auto generated low-level Go binding around an Ethereum contract.
type OBLSRaw struct {
	Contract *OBLS // Generic contract binding to access the raw methods on
}

// OBLSCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OBLSCallerRaw struct {
	Contract *OBLSCaller // Generic read-only contract binding to access the raw methods on
}

// OBLSTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OBLSTransactorRaw struct {
	Contract *OBLSTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOBLS creates a new instance of OBLS, bound to a specific deployed contract.
func NewOBLS(address common.Address, backend bind.ContractBackend) (*OBLS, error) {
	contract, err := bindOBLS(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OBLS{OBLSCaller: OBLSCaller{contract: contract}, OBLSTransactor: OBLSTransactor{contract: contract}, OBLSFilterer: OBLSFilterer{contract: contract}}, nil
}

// NewOBLSCaller creates a new read-only instance of OBLS, bound to a specific deployed contract.
func NewOBLSCaller(address common.Address, caller bind.ContractCaller) (*OBLSCaller, error) {
	contract, err := bindOBLS(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OBLSCaller{contract: contract}, nil
}

// NewOBLSTransactor creates a new write-only instance of OBLS, bound to a specific deployed contract.
func NewOBLSTransactor(address common.Address, transactor bind.ContractTransactor) (*OBLSTransactor, error) {
	contract, err := bindOBLS(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OBLSTransactor{contract: contract}, nil
}

// NewOBLSFilterer creates a new log filterer instance of OBLS, bound to a specific deployed contract.
func NewOBLSFilterer(address common.Address, filterer bind.ContractFilterer) (*OBLSFilterer, error) {
	contract, err := bindOBLS(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OBLSFilterer{contract: contract}, nil
}

// bindOBLS binds a generic wrapper to an already deployed contract.
func bindOBLS(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OBLSMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OBLS *OBLSRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OBLS.Contract.OBLSCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OBLS *OBLSRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OBLS.Contract.OBLSTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OBLS *OBLSRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OBLS.Contract.OBLSTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OBLS *OBLSCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OBLS.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OBLS *OBLSTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OBLS.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OBLS *OBLSTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OBLS.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_OBLS *OBLSCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_OBLS *OBLSSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _OBLS.Contract.DEFAULTADMINROLE(&_OBLS.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_OBLS *OBLSCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _OBLS.Contract.DEFAULTADMINROLE(&_OBLS.CallOpts)
}

// GetOblsManager is a free data retrieval call binding the contract method 0xb3dfe1e8.
//
// Solidity: function getOblsManager() view returns(address)
func (_OBLS *OBLSCaller) GetOblsManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "getOblsManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOblsManager is a free data retrieval call binding the contract method 0xb3dfe1e8.
//
// Solidity: function getOblsManager() view returns(address)
func (_OBLS *OBLSSession) GetOblsManager() (common.Address, error) {
	return _OBLS.Contract.GetOblsManager(&_OBLS.CallOpts)
}

// GetOblsManager is a free data retrieval call binding the contract method 0xb3dfe1e8.
//
// Solidity: function getOblsManager() view returns(address)
func (_OBLS *OBLSCallerSession) GetOblsManager() (common.Address, error) {
	return _OBLS.Contract.GetOblsManager(&_OBLS.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_OBLS *OBLSCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_OBLS *OBLSSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _OBLS.Contract.GetRoleAdmin(&_OBLS.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_OBLS *OBLSCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _OBLS.Contract.GetRoleAdmin(&_OBLS.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_OBLS *OBLSCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_OBLS *OBLSSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _OBLS.Contract.HasRole(&_OBLS.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_OBLS *OBLSCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _OBLS.Contract.HasRole(&_OBLS.CallOpts, role, account)
}

// HashToPoint is a free data retrieval call binding the contract method 0xa850a909.
//
// Solidity: function hashToPoint(bytes32 domain, bytes message) view returns(uint256[2])
func (_OBLS *OBLSCaller) HashToPoint(opts *bind.CallOpts, domain [32]byte, message []byte) ([2]*big.Int, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "hashToPoint", domain, message)

	if err != nil {
		return *new([2]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([2]*big.Int)).(*[2]*big.Int)

	return out0, err

}

// HashToPoint is a free data retrieval call binding the contract method 0xa850a909.
//
// Solidity: function hashToPoint(bytes32 domain, bytes message) view returns(uint256[2])
func (_OBLS *OBLSSession) HashToPoint(domain [32]byte, message []byte) ([2]*big.Int, error) {
	return _OBLS.Contract.HashToPoint(&_OBLS.CallOpts, domain, message)
}

// HashToPoint is a free data retrieval call binding the contract method 0xa850a909.
//
// Solidity: function hashToPoint(bytes32 domain, bytes message) view returns(uint256[2])
func (_OBLS *OBLSCallerSession) HashToPoint(domain [32]byte, message []byte) ([2]*big.Int, error) {
	return _OBLS.Contract.HashToPoint(&_OBLS.CallOpts, domain, message)
}

// IsActive is a free data retrieval call binding the contract method 0x82afd23b.
//
// Solidity: function isActive(uint256 _index) view returns(bool)
func (_OBLS *OBLSCaller) IsActive(opts *bind.CallOpts, _index *big.Int) (bool, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "isActive", _index)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsActive is a free data retrieval call binding the contract method 0x82afd23b.
//
// Solidity: function isActive(uint256 _index) view returns(bool)
func (_OBLS *OBLSSession) IsActive(_index *big.Int) (bool, error) {
	return _OBLS.Contract.IsActive(&_OBLS.CallOpts, _index)
}

// IsActive is a free data retrieval call binding the contract method 0x82afd23b.
//
// Solidity: function isActive(uint256 _index) view returns(bool)
func (_OBLS *OBLSCallerSession) IsActive(_index *big.Int) (bool, error) {
	return _OBLS.Contract.IsActive(&_OBLS.CallOpts, _index)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_OBLS *OBLSCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_OBLS *OBLSSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _OBLS.Contract.SupportsInterface(&_OBLS.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_OBLS *OBLSCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _OBLS.Contract.SupportsInterface(&_OBLS.CallOpts, interfaceId)
}

// TotalVotingPower is a free data retrieval call binding the contract method 0x671b3793.
//
// Solidity: function totalVotingPower() view returns(uint256)
func (_OBLS *OBLSCaller) TotalVotingPower(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "totalVotingPower")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalVotingPower is a free data retrieval call binding the contract method 0x671b3793.
//
// Solidity: function totalVotingPower() view returns(uint256)
func (_OBLS *OBLSSession) TotalVotingPower() (*big.Int, error) {
	return _OBLS.Contract.TotalVotingPower(&_OBLS.CallOpts)
}

// TotalVotingPower is a free data retrieval call binding the contract method 0x671b3793.
//
// Solidity: function totalVotingPower() view returns(uint256)
func (_OBLS *OBLSCallerSession) TotalVotingPower() (*big.Int, error) {
	return _OBLS.Contract.TotalVotingPower(&_OBLS.CallOpts)
}

// TotalVotingPowerPerTaskDefinition is a free data retrieval call binding the contract method 0x255da331.
//
// Solidity: function totalVotingPowerPerTaskDefinition(uint256 _id) view returns(uint256)
func (_OBLS *OBLSCaller) TotalVotingPowerPerTaskDefinition(opts *bind.CallOpts, _id *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "totalVotingPowerPerTaskDefinition", _id)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalVotingPowerPerTaskDefinition is a free data retrieval call binding the contract method 0x255da331.
//
// Solidity: function totalVotingPowerPerTaskDefinition(uint256 _id) view returns(uint256)
func (_OBLS *OBLSSession) TotalVotingPowerPerTaskDefinition(_id *big.Int) (*big.Int, error) {
	return _OBLS.Contract.TotalVotingPowerPerTaskDefinition(&_OBLS.CallOpts, _id)
}

// TotalVotingPowerPerTaskDefinition is a free data retrieval call binding the contract method 0x255da331.
//
// Solidity: function totalVotingPowerPerTaskDefinition(uint256 _id) view returns(uint256)
func (_OBLS *OBLSCallerSession) TotalVotingPowerPerTaskDefinition(_id *big.Int) (*big.Int, error) {
	return _OBLS.Contract.TotalVotingPowerPerTaskDefinition(&_OBLS.CallOpts, _id)
}

// VerifyAuthSignature is a free data retrieval call binding the contract method 0x8ebaaa3e.
//
// Solidity: function verifyAuthSignature((uint256[2]) _signature, address _operator, address _contract, uint256[4] _blsKey) view returns()
func (_OBLS *OBLSCaller) VerifyAuthSignature(opts *bind.CallOpts, _signature BLSAuthLibrarySignature, _operator common.Address, _contract common.Address, _blsKey [4]*big.Int) error {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "verifyAuthSignature", _signature, _operator, _contract, _blsKey)

	if err != nil {
		return err
	}

	return err

}

// VerifyAuthSignature is a free data retrieval call binding the contract method 0x8ebaaa3e.
//
// Solidity: function verifyAuthSignature((uint256[2]) _signature, address _operator, address _contract, uint256[4] _blsKey) view returns()
func (_OBLS *OBLSSession) VerifyAuthSignature(_signature BLSAuthLibrarySignature, _operator common.Address, _contract common.Address, _blsKey [4]*big.Int) error {
	return _OBLS.Contract.VerifyAuthSignature(&_OBLS.CallOpts, _signature, _operator, _contract, _blsKey)
}

// VerifyAuthSignature is a free data retrieval call binding the contract method 0x8ebaaa3e.
//
// Solidity: function verifyAuthSignature((uint256[2]) _signature, address _operator, address _contract, uint256[4] _blsKey) view returns()
func (_OBLS *OBLSCallerSession) VerifyAuthSignature(_signature BLSAuthLibrarySignature, _operator common.Address, _contract common.Address, _blsKey [4]*big.Int) error {
	return _OBLS.Contract.VerifyAuthSignature(&_OBLS.CallOpts, _signature, _operator, _contract, _blsKey)
}

// VerifySignature is a free data retrieval call binding the contract method 0x65c46475.
//
// Solidity: function verifySignature(uint256[2] _message, uint256[2] _signature, uint256[] _indexes, uint256 _requiredVotingPower, uint256 _minimumVotingPowerPerTaskDefinition) view returns()
func (_OBLS *OBLSCaller) VerifySignature(opts *bind.CallOpts, _message [2]*big.Int, _signature [2]*big.Int, _indexes []*big.Int, _requiredVotingPower *big.Int, _minimumVotingPowerPerTaskDefinition *big.Int) error {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "verifySignature", _message, _signature, _indexes, _requiredVotingPower, _minimumVotingPowerPerTaskDefinition)

	if err != nil {
		return err
	}

	return err

}

// VerifySignature is a free data retrieval call binding the contract method 0x65c46475.
//
// Solidity: function verifySignature(uint256[2] _message, uint256[2] _signature, uint256[] _indexes, uint256 _requiredVotingPower, uint256 _minimumVotingPowerPerTaskDefinition) view returns()
func (_OBLS *OBLSSession) VerifySignature(_message [2]*big.Int, _signature [2]*big.Int, _indexes []*big.Int, _requiredVotingPower *big.Int, _minimumVotingPowerPerTaskDefinition *big.Int) error {
	return _OBLS.Contract.VerifySignature(&_OBLS.CallOpts, _message, _signature, _indexes, _requiredVotingPower, _minimumVotingPowerPerTaskDefinition)
}

// VerifySignature is a free data retrieval call binding the contract method 0x65c46475.
//
// Solidity: function verifySignature(uint256[2] _message, uint256[2] _signature, uint256[] _indexes, uint256 _requiredVotingPower, uint256 _minimumVotingPowerPerTaskDefinition) view returns()
func (_OBLS *OBLSCallerSession) VerifySignature(_message [2]*big.Int, _signature [2]*big.Int, _indexes []*big.Int, _requiredVotingPower *big.Int, _minimumVotingPowerPerTaskDefinition *big.Int) error {
	return _OBLS.Contract.VerifySignature(&_OBLS.CallOpts, _message, _signature, _indexes, _requiredVotingPower, _minimumVotingPowerPerTaskDefinition)
}

// VotingPower is a free data retrieval call binding the contract method 0x72c4a927.
//
// Solidity: function votingPower(uint256 _index) view returns(uint256)
func (_OBLS *OBLSCaller) VotingPower(opts *bind.CallOpts, _index *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _OBLS.contract.Call(opts, &out, "votingPower", _index)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VotingPower is a free data retrieval call binding the contract method 0x72c4a927.
//
// Solidity: function votingPower(uint256 _index) view returns(uint256)
func (_OBLS *OBLSSession) VotingPower(_index *big.Int) (*big.Int, error) {
	return _OBLS.Contract.VotingPower(&_OBLS.CallOpts, _index)
}

// VotingPower is a free data retrieval call binding the contract method 0x72c4a927.
//
// Solidity: function votingPower(uint256 _index) view returns(uint256)
func (_OBLS *OBLSCallerSession) VotingPower(_index *big.Int) (*big.Int, error) {
	return _OBLS.Contract.VotingPower(&_OBLS.CallOpts, _index)
}

// DecreaseBatchOperatorVotingPower is a paid mutator transaction binding the contract method 0x6533df08.
//
// Solidity: function decreaseBatchOperatorVotingPower((uint256,uint256)[] _operatorsVotingPower) returns()
func (_OBLS *OBLSTransactor) DecreaseBatchOperatorVotingPower(opts *bind.TransactOpts, _operatorsVotingPower []IOBLSOperatorVotingPower) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "decreaseBatchOperatorVotingPower", _operatorsVotingPower)
}

// DecreaseBatchOperatorVotingPower is a paid mutator transaction binding the contract method 0x6533df08.
//
// Solidity: function decreaseBatchOperatorVotingPower((uint256,uint256)[] _operatorsVotingPower) returns()
func (_OBLS *OBLSSession) DecreaseBatchOperatorVotingPower(_operatorsVotingPower []IOBLSOperatorVotingPower) (*types.Transaction, error) {
	return _OBLS.Contract.DecreaseBatchOperatorVotingPower(&_OBLS.TransactOpts, _operatorsVotingPower)
}

// DecreaseBatchOperatorVotingPower is a paid mutator transaction binding the contract method 0x6533df08.
//
// Solidity: function decreaseBatchOperatorVotingPower((uint256,uint256)[] _operatorsVotingPower) returns()
func (_OBLS *OBLSTransactorSession) DecreaseBatchOperatorVotingPower(_operatorsVotingPower []IOBLSOperatorVotingPower) (*types.Transaction, error) {
	return _OBLS.Contract.DecreaseBatchOperatorVotingPower(&_OBLS.TransactOpts, _operatorsVotingPower)
}

// DecreaseOperatorVotingPower is a paid mutator transaction binding the contract method 0x20f527ad.
//
// Solidity: function decreaseOperatorVotingPower(uint256 _index, uint256 _votingPower) returns()
func (_OBLS *OBLSTransactor) DecreaseOperatorVotingPower(opts *bind.TransactOpts, _index *big.Int, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "decreaseOperatorVotingPower", _index, _votingPower)
}

// DecreaseOperatorVotingPower is a paid mutator transaction binding the contract method 0x20f527ad.
//
// Solidity: function decreaseOperatorVotingPower(uint256 _index, uint256 _votingPower) returns()
func (_OBLS *OBLSSession) DecreaseOperatorVotingPower(_index *big.Int, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.DecreaseOperatorVotingPower(&_OBLS.TransactOpts, _index, _votingPower)
}

// DecreaseOperatorVotingPower is a paid mutator transaction binding the contract method 0x20f527ad.
//
// Solidity: function decreaseOperatorVotingPower(uint256 _index, uint256 _votingPower) returns()
func (_OBLS *OBLSTransactorSession) DecreaseOperatorVotingPower(_index *big.Int, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.DecreaseOperatorVotingPower(&_OBLS.TransactOpts, _index, _votingPower)
}

// DecreaseOperatorVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0xff1e2c28.
//
// Solidity: function decreaseOperatorVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _votingPower) returns()
func (_OBLS *OBLSTransactor) DecreaseOperatorVotingPowerPerTaskDefinition(opts *bind.TransactOpts, _taskDefinitionId uint16, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "decreaseOperatorVotingPowerPerTaskDefinition", _taskDefinitionId, _votingPower)
}

// DecreaseOperatorVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0xff1e2c28.
//
// Solidity: function decreaseOperatorVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _votingPower) returns()
func (_OBLS *OBLSSession) DecreaseOperatorVotingPowerPerTaskDefinition(_taskDefinitionId uint16, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.DecreaseOperatorVotingPowerPerTaskDefinition(&_OBLS.TransactOpts, _taskDefinitionId, _votingPower)
}

// DecreaseOperatorVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0xff1e2c28.
//
// Solidity: function decreaseOperatorVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _votingPower) returns()
func (_OBLS *OBLSTransactorSession) DecreaseOperatorVotingPowerPerTaskDefinition(_taskDefinitionId uint16, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.DecreaseOperatorVotingPowerPerTaskDefinition(&_OBLS.TransactOpts, _taskDefinitionId, _votingPower)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_OBLS *OBLSTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_OBLS *OBLSSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.GrantRole(&_OBLS.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_OBLS *OBLSTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.GrantRole(&_OBLS.TransactOpts, role, account)
}

// IncreaseBatchOperatorVotingPower is a paid mutator transaction binding the contract method 0x63f903cd.
//
// Solidity: function increaseBatchOperatorVotingPower((uint256,uint256)[] _operatorsVotingPower) returns()
func (_OBLS *OBLSTransactor) IncreaseBatchOperatorVotingPower(opts *bind.TransactOpts, _operatorsVotingPower []IOBLSOperatorVotingPower) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "increaseBatchOperatorVotingPower", _operatorsVotingPower)
}

// IncreaseBatchOperatorVotingPower is a paid mutator transaction binding the contract method 0x63f903cd.
//
// Solidity: function increaseBatchOperatorVotingPower((uint256,uint256)[] _operatorsVotingPower) returns()
func (_OBLS *OBLSSession) IncreaseBatchOperatorVotingPower(_operatorsVotingPower []IOBLSOperatorVotingPower) (*types.Transaction, error) {
	return _OBLS.Contract.IncreaseBatchOperatorVotingPower(&_OBLS.TransactOpts, _operatorsVotingPower)
}

// IncreaseBatchOperatorVotingPower is a paid mutator transaction binding the contract method 0x63f903cd.
//
// Solidity: function increaseBatchOperatorVotingPower((uint256,uint256)[] _operatorsVotingPower) returns()
func (_OBLS *OBLSTransactorSession) IncreaseBatchOperatorVotingPower(_operatorsVotingPower []IOBLSOperatorVotingPower) (*types.Transaction, error) {
	return _OBLS.Contract.IncreaseBatchOperatorVotingPower(&_OBLS.TransactOpts, _operatorsVotingPower)
}

// IncreaseOperatorVotingPower is a paid mutator transaction binding the contract method 0xd66f643d.
//
// Solidity: function increaseOperatorVotingPower(uint256 _index, uint256 _votingPower) returns()
func (_OBLS *OBLSTransactor) IncreaseOperatorVotingPower(opts *bind.TransactOpts, _index *big.Int, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "increaseOperatorVotingPower", _index, _votingPower)
}

// IncreaseOperatorVotingPower is a paid mutator transaction binding the contract method 0xd66f643d.
//
// Solidity: function increaseOperatorVotingPower(uint256 _index, uint256 _votingPower) returns()
func (_OBLS *OBLSSession) IncreaseOperatorVotingPower(_index *big.Int, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.IncreaseOperatorVotingPower(&_OBLS.TransactOpts, _index, _votingPower)
}

// IncreaseOperatorVotingPower is a paid mutator transaction binding the contract method 0xd66f643d.
//
// Solidity: function increaseOperatorVotingPower(uint256 _index, uint256 _votingPower) returns()
func (_OBLS *OBLSTransactorSession) IncreaseOperatorVotingPower(_index *big.Int, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.IncreaseOperatorVotingPower(&_OBLS.TransactOpts, _index, _votingPower)
}

// IncreaseOperatorVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0x93f438be.
//
// Solidity: function increaseOperatorVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _votingPower) returns()
func (_OBLS *OBLSTransactor) IncreaseOperatorVotingPowerPerTaskDefinition(opts *bind.TransactOpts, _taskDefinitionId uint16, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "increaseOperatorVotingPowerPerTaskDefinition", _taskDefinitionId, _votingPower)
}

// IncreaseOperatorVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0x93f438be.
//
// Solidity: function increaseOperatorVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _votingPower) returns()
func (_OBLS *OBLSSession) IncreaseOperatorVotingPowerPerTaskDefinition(_taskDefinitionId uint16, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.IncreaseOperatorVotingPowerPerTaskDefinition(&_OBLS.TransactOpts, _taskDefinitionId, _votingPower)
}

// IncreaseOperatorVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0x93f438be.
//
// Solidity: function increaseOperatorVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _votingPower) returns()
func (_OBLS *OBLSTransactorSession) IncreaseOperatorVotingPowerPerTaskDefinition(_taskDefinitionId uint16, _votingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.IncreaseOperatorVotingPowerPerTaskDefinition(&_OBLS.TransactOpts, _taskDefinitionId, _votingPower)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_OBLS *OBLSTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_OBLS *OBLSSession) Initialize() (*types.Transaction, error) {
	return _OBLS.Contract.Initialize(&_OBLS.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_OBLS *OBLSTransactorSession) Initialize() (*types.Transaction, error) {
	return _OBLS.Contract.Initialize(&_OBLS.TransactOpts)
}

// ModifyOperatorActiveStatus is a paid mutator transaction binding the contract method 0x67ad3232.
//
// Solidity: function modifyOperatorActiveStatus(uint256 _index, bool _isActive) returns()
func (_OBLS *OBLSTransactor) ModifyOperatorActiveStatus(opts *bind.TransactOpts, _index *big.Int, _isActive bool) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "modifyOperatorActiveStatus", _index, _isActive)
}

// ModifyOperatorActiveStatus is a paid mutator transaction binding the contract method 0x67ad3232.
//
// Solidity: function modifyOperatorActiveStatus(uint256 _index, bool _isActive) returns()
func (_OBLS *OBLSSession) ModifyOperatorActiveStatus(_index *big.Int, _isActive bool) (*types.Transaction, error) {
	return _OBLS.Contract.ModifyOperatorActiveStatus(&_OBLS.TransactOpts, _index, _isActive)
}

// ModifyOperatorActiveStatus is a paid mutator transaction binding the contract method 0x67ad3232.
//
// Solidity: function modifyOperatorActiveStatus(uint256 _index, bool _isActive) returns()
func (_OBLS *OBLSTransactorSession) ModifyOperatorActiveStatus(_index *big.Int, _isActive bool) (*types.Transaction, error) {
	return _OBLS.Contract.ModifyOperatorActiveStatus(&_OBLS.TransactOpts, _index, _isActive)
}

// ModifyOperatorBlsKey is a paid mutator transaction binding the contract method 0x86d897a7.
//
// Solidity: function modifyOperatorBlsKey(uint256 _index, uint256[4] _blsKey) returns()
func (_OBLS *OBLSTransactor) ModifyOperatorBlsKey(opts *bind.TransactOpts, _index *big.Int, _blsKey [4]*big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "modifyOperatorBlsKey", _index, _blsKey)
}

// ModifyOperatorBlsKey is a paid mutator transaction binding the contract method 0x86d897a7.
//
// Solidity: function modifyOperatorBlsKey(uint256 _index, uint256[4] _blsKey) returns()
func (_OBLS *OBLSSession) ModifyOperatorBlsKey(_index *big.Int, _blsKey [4]*big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.ModifyOperatorBlsKey(&_OBLS.TransactOpts, _index, _blsKey)
}

// ModifyOperatorBlsKey is a paid mutator transaction binding the contract method 0x86d897a7.
//
// Solidity: function modifyOperatorBlsKey(uint256 _index, uint256[4] _blsKey) returns()
func (_OBLS *OBLSTransactorSession) ModifyOperatorBlsKey(_index *big.Int, _blsKey [4]*big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.ModifyOperatorBlsKey(&_OBLS.TransactOpts, _index, _blsKey)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x891a80bb.
//
// Solidity: function registerOperator(uint256 _index, uint256 _votingPower, uint256[4] _blsKey) returns()
func (_OBLS *OBLSTransactor) RegisterOperator(opts *bind.TransactOpts, _index *big.Int, _votingPower *big.Int, _blsKey [4]*big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "registerOperator", _index, _votingPower, _blsKey)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x891a80bb.
//
// Solidity: function registerOperator(uint256 _index, uint256 _votingPower, uint256[4] _blsKey) returns()
func (_OBLS *OBLSSession) RegisterOperator(_index *big.Int, _votingPower *big.Int, _blsKey [4]*big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.RegisterOperator(&_OBLS.TransactOpts, _index, _votingPower, _blsKey)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x891a80bb.
//
// Solidity: function registerOperator(uint256 _index, uint256 _votingPower, uint256[4] _blsKey) returns()
func (_OBLS *OBLSTransactorSession) RegisterOperator(_index *big.Int, _votingPower *big.Int, _blsKey [4]*big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.RegisterOperator(&_OBLS.TransactOpts, _index, _votingPower, _blsKey)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_OBLS *OBLSTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_OBLS *OBLSSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.RenounceRole(&_OBLS.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_OBLS *OBLSTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.RenounceRole(&_OBLS.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_OBLS *OBLSTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_OBLS *OBLSSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.RevokeRole(&_OBLS.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_OBLS *OBLSTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.RevokeRole(&_OBLS.TransactOpts, role, account)
}

// SetOblsManager is a paid mutator transaction binding the contract method 0xd9835b61.
//
// Solidity: function setOblsManager(address _oblsManager) returns()
func (_OBLS *OBLSTransactor) SetOblsManager(opts *bind.TransactOpts, _oblsManager common.Address) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "setOblsManager", _oblsManager)
}

// SetOblsManager is a paid mutator transaction binding the contract method 0xd9835b61.
//
// Solidity: function setOblsManager(address _oblsManager) returns()
func (_OBLS *OBLSSession) SetOblsManager(_oblsManager common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.SetOblsManager(&_OBLS.TransactOpts, _oblsManager)
}

// SetOblsManager is a paid mutator transaction binding the contract method 0xd9835b61.
//
// Solidity: function setOblsManager(address _oblsManager) returns()
func (_OBLS *OBLSTransactorSession) SetOblsManager(_oblsManager common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.SetOblsManager(&_OBLS.TransactOpts, _oblsManager)
}

// SetOblsSharesSyncer is a paid mutator transaction binding the contract method 0x1164224e.
//
// Solidity: function setOblsSharesSyncer(address _oblsSharesSyncer) returns()
func (_OBLS *OBLSTransactor) SetOblsSharesSyncer(opts *bind.TransactOpts, _oblsSharesSyncer common.Address) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "setOblsSharesSyncer", _oblsSharesSyncer)
}

// SetOblsSharesSyncer is a paid mutator transaction binding the contract method 0x1164224e.
//
// Solidity: function setOblsSharesSyncer(address _oblsSharesSyncer) returns()
func (_OBLS *OBLSSession) SetOblsSharesSyncer(_oblsSharesSyncer common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.SetOblsSharesSyncer(&_OBLS.TransactOpts, _oblsSharesSyncer)
}

// SetOblsSharesSyncer is a paid mutator transaction binding the contract method 0x1164224e.
//
// Solidity: function setOblsSharesSyncer(address _oblsSharesSyncer) returns()
func (_OBLS *OBLSTransactorSession) SetOblsSharesSyncer(_oblsSharesSyncer common.Address) (*types.Transaction, error) {
	return _OBLS.Contract.SetOblsSharesSyncer(&_OBLS.TransactOpts, _oblsSharesSyncer)
}

// SetTotalVotingPowerPerRestrictedTaskDefinition is a paid mutator transaction binding the contract method 0xca87bf8f.
//
// Solidity: function setTotalVotingPowerPerRestrictedTaskDefinition(uint16 _taskDefinitionId, uint256 _minimumVotingPower, uint256[] _restrictedOperatorIndexes) returns()
func (_OBLS *OBLSTransactor) SetTotalVotingPowerPerRestrictedTaskDefinition(opts *bind.TransactOpts, _taskDefinitionId uint16, _minimumVotingPower *big.Int, _restrictedOperatorIndexes []*big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "setTotalVotingPowerPerRestrictedTaskDefinition", _taskDefinitionId, _minimumVotingPower, _restrictedOperatorIndexes)
}

// SetTotalVotingPowerPerRestrictedTaskDefinition is a paid mutator transaction binding the contract method 0xca87bf8f.
//
// Solidity: function setTotalVotingPowerPerRestrictedTaskDefinition(uint16 _taskDefinitionId, uint256 _minimumVotingPower, uint256[] _restrictedOperatorIndexes) returns()
func (_OBLS *OBLSSession) SetTotalVotingPowerPerRestrictedTaskDefinition(_taskDefinitionId uint16, _minimumVotingPower *big.Int, _restrictedOperatorIndexes []*big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.SetTotalVotingPowerPerRestrictedTaskDefinition(&_OBLS.TransactOpts, _taskDefinitionId, _minimumVotingPower, _restrictedOperatorIndexes)
}

// SetTotalVotingPowerPerRestrictedTaskDefinition is a paid mutator transaction binding the contract method 0xca87bf8f.
//
// Solidity: function setTotalVotingPowerPerRestrictedTaskDefinition(uint16 _taskDefinitionId, uint256 _minimumVotingPower, uint256[] _restrictedOperatorIndexes) returns()
func (_OBLS *OBLSTransactorSession) SetTotalVotingPowerPerRestrictedTaskDefinition(_taskDefinitionId uint16, _minimumVotingPower *big.Int, _restrictedOperatorIndexes []*big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.SetTotalVotingPowerPerRestrictedTaskDefinition(&_OBLS.TransactOpts, _taskDefinitionId, _minimumVotingPower, _restrictedOperatorIndexes)
}

// SetTotalVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0xe010f957.
//
// Solidity: function setTotalVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _numOfTotalOperators, uint256 _minimumVotingPower) returns()
func (_OBLS *OBLSTransactor) SetTotalVotingPowerPerTaskDefinition(opts *bind.TransactOpts, _taskDefinitionId uint16, _numOfTotalOperators *big.Int, _minimumVotingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "setTotalVotingPowerPerTaskDefinition", _taskDefinitionId, _numOfTotalOperators, _minimumVotingPower)
}

// SetTotalVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0xe010f957.
//
// Solidity: function setTotalVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _numOfTotalOperators, uint256 _minimumVotingPower) returns()
func (_OBLS *OBLSSession) SetTotalVotingPowerPerTaskDefinition(_taskDefinitionId uint16, _numOfTotalOperators *big.Int, _minimumVotingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.SetTotalVotingPowerPerTaskDefinition(&_OBLS.TransactOpts, _taskDefinitionId, _numOfTotalOperators, _minimumVotingPower)
}

// SetTotalVotingPowerPerTaskDefinition is a paid mutator transaction binding the contract method 0xe010f957.
//
// Solidity: function setTotalVotingPowerPerTaskDefinition(uint16 _taskDefinitionId, uint256 _numOfTotalOperators, uint256 _minimumVotingPower) returns()
func (_OBLS *OBLSTransactorSession) SetTotalVotingPowerPerTaskDefinition(_taskDefinitionId uint16, _numOfTotalOperators *big.Int, _minimumVotingPower *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.SetTotalVotingPowerPerTaskDefinition(&_OBLS.TransactOpts, _taskDefinitionId, _numOfTotalOperators, _minimumVotingPower)
}

// UnRegisterOperator is a paid mutator transaction binding the contract method 0xcf33bccc.
//
// Solidity: function unRegisterOperator(uint256 _index) returns()
func (_OBLS *OBLSTransactor) UnRegisterOperator(opts *bind.TransactOpts, _index *big.Int) (*types.Transaction, error) {
	return _OBLS.contract.Transact(opts, "unRegisterOperator", _index)
}

// UnRegisterOperator is a paid mutator transaction binding the contract method 0xcf33bccc.
//
// Solidity: function unRegisterOperator(uint256 _index) returns()
func (_OBLS *OBLSSession) UnRegisterOperator(_index *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.UnRegisterOperator(&_OBLS.TransactOpts, _index)
}

// UnRegisterOperator is a paid mutator transaction binding the contract method 0xcf33bccc.
//
// Solidity: function unRegisterOperator(uint256 _index) returns()
func (_OBLS *OBLSTransactorSession) UnRegisterOperator(_index *big.Int) (*types.Transaction, error) {
	return _OBLS.Contract.UnRegisterOperator(&_OBLS.TransactOpts, _index)
}

// OBLSInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the OBLS contract.
type OBLSInitializedIterator struct {
	Event *OBLSInitialized // Event containing the contract specifics and raw log

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
func (it *OBLSInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OBLSInitialized)
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
		it.Event = new(OBLSInitialized)
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
func (it *OBLSInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OBLSInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OBLSInitialized represents a Initialized event raised by the OBLS contract.
type OBLSInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_OBLS *OBLSFilterer) FilterInitialized(opts *bind.FilterOpts) (*OBLSInitializedIterator, error) {

	logs, sub, err := _OBLS.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &OBLSInitializedIterator{contract: _OBLS.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_OBLS *OBLSFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *OBLSInitialized) (event.Subscription, error) {

	logs, sub, err := _OBLS.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OBLSInitialized)
				if err := _OBLS.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_OBLS *OBLSFilterer) ParseInitialized(log types.Log) (*OBLSInitialized, error) {
	event := new(OBLSInitialized)
	if err := _OBLS.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OBLSRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the OBLS contract.
type OBLSRoleAdminChangedIterator struct {
	Event *OBLSRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *OBLSRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OBLSRoleAdminChanged)
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
		it.Event = new(OBLSRoleAdminChanged)
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
func (it *OBLSRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OBLSRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OBLSRoleAdminChanged represents a RoleAdminChanged event raised by the OBLS contract.
type OBLSRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_OBLS *OBLSFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*OBLSRoleAdminChangedIterator, error) {

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

	logs, sub, err := _OBLS.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &OBLSRoleAdminChangedIterator{contract: _OBLS.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_OBLS *OBLSFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *OBLSRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _OBLS.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OBLSRoleAdminChanged)
				if err := _OBLS.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_OBLS *OBLSFilterer) ParseRoleAdminChanged(log types.Log) (*OBLSRoleAdminChanged, error) {
	event := new(OBLSRoleAdminChanged)
	if err := _OBLS.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OBLSRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the OBLS contract.
type OBLSRoleGrantedIterator struct {
	Event *OBLSRoleGranted // Event containing the contract specifics and raw log

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
func (it *OBLSRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OBLSRoleGranted)
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
		it.Event = new(OBLSRoleGranted)
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
func (it *OBLSRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OBLSRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OBLSRoleGranted represents a RoleGranted event raised by the OBLS contract.
type OBLSRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_OBLS *OBLSFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*OBLSRoleGrantedIterator, error) {

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

	logs, sub, err := _OBLS.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &OBLSRoleGrantedIterator{contract: _OBLS.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_OBLS *OBLSFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *OBLSRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _OBLS.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OBLSRoleGranted)
				if err := _OBLS.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_OBLS *OBLSFilterer) ParseRoleGranted(log types.Log) (*OBLSRoleGranted, error) {
	event := new(OBLSRoleGranted)
	if err := _OBLS.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OBLSRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the OBLS contract.
type OBLSRoleRevokedIterator struct {
	Event *OBLSRoleRevoked // Event containing the contract specifics and raw log

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
func (it *OBLSRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OBLSRoleRevoked)
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
		it.Event = new(OBLSRoleRevoked)
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
func (it *OBLSRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OBLSRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OBLSRoleRevoked represents a RoleRevoked event raised by the OBLS contract.
type OBLSRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_OBLS *OBLSFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*OBLSRoleRevokedIterator, error) {

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

	logs, sub, err := _OBLS.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &OBLSRoleRevokedIterator{contract: _OBLS.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_OBLS *OBLSFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *OBLSRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _OBLS.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OBLSRoleRevoked)
				if err := _OBLS.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_OBLS *OBLSFilterer) ParseRoleRevoked(log types.Log) (*OBLSRoleRevoked, error) {
	event := new(OBLSRoleRevoked)
	if err := _OBLS.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OBLSSetOBLSManagerIterator is returned from FilterSetOBLSManager and is used to iterate over the raw logs and unpacked data for SetOBLSManager events raised by the OBLS contract.
type OBLSSetOBLSManagerIterator struct {
	Event *OBLSSetOBLSManager // Event containing the contract specifics and raw log

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
func (it *OBLSSetOBLSManagerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OBLSSetOBLSManager)
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
		it.Event = new(OBLSSetOBLSManager)
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
func (it *OBLSSetOBLSManagerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OBLSSetOBLSManagerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OBLSSetOBLSManager represents a SetOBLSManager event raised by the OBLS contract.
type OBLSSetOBLSManager struct {
	NewOBLSManager common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSetOBLSManager is a free log retrieval operation binding the contract event 0xff7dfc53b4a07266bc3bbaabbdb9992fc67be75384df0de05509a5cfdae75101.
//
// Solidity: event SetOBLSManager(address newOBLSManager)
func (_OBLS *OBLSFilterer) FilterSetOBLSManager(opts *bind.FilterOpts) (*OBLSSetOBLSManagerIterator, error) {

	logs, sub, err := _OBLS.contract.FilterLogs(opts, "SetOBLSManager")
	if err != nil {
		return nil, err
	}
	return &OBLSSetOBLSManagerIterator{contract: _OBLS.contract, event: "SetOBLSManager", logs: logs, sub: sub}, nil
}

// WatchSetOBLSManager is a free log subscription operation binding the contract event 0xff7dfc53b4a07266bc3bbaabbdb9992fc67be75384df0de05509a5cfdae75101.
//
// Solidity: event SetOBLSManager(address newOBLSManager)
func (_OBLS *OBLSFilterer) WatchSetOBLSManager(opts *bind.WatchOpts, sink chan<- *OBLSSetOBLSManager) (event.Subscription, error) {

	logs, sub, err := _OBLS.contract.WatchLogs(opts, "SetOBLSManager")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OBLSSetOBLSManager)
				if err := _OBLS.contract.UnpackLog(event, "SetOBLSManager", log); err != nil {
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

// ParseSetOBLSManager is a log parse operation binding the contract event 0xff7dfc53b4a07266bc3bbaabbdb9992fc67be75384df0de05509a5cfdae75101.
//
// Solidity: event SetOBLSManager(address newOBLSManager)
func (_OBLS *OBLSFilterer) ParseSetOBLSManager(log types.Log) (*OBLSSetOBLSManager, error) {
	event := new(OBLSSetOBLSManager)
	if err := _OBLS.contract.UnpackLog(event, "SetOBLSManager", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OBLSSharesSyncerModifiedIterator is returned from FilterSharesSyncerModified and is used to iterate over the raw logs and unpacked data for SharesSyncerModified events raised by the OBLS contract.
type OBLSSharesSyncerModifiedIterator struct {
	Event *OBLSSharesSyncerModified // Event containing the contract specifics and raw log

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
func (it *OBLSSharesSyncerModifiedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OBLSSharesSyncerModified)
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
		it.Event = new(OBLSSharesSyncerModified)
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
func (it *OBLSSharesSyncerModifiedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OBLSSharesSyncerModifiedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OBLSSharesSyncerModified represents a SharesSyncerModified event raised by the OBLS contract.
type OBLSSharesSyncerModified struct {
	Syncer common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSharesSyncerModified is a free log retrieval operation binding the contract event 0x797f7e8ad76b743fbff71cb58d982a1bd35e4df14315d7687b187c1c5c257e4d.
//
// Solidity: event SharesSyncerModified(address syncer)
func (_OBLS *OBLSFilterer) FilterSharesSyncerModified(opts *bind.FilterOpts) (*OBLSSharesSyncerModifiedIterator, error) {

	logs, sub, err := _OBLS.contract.FilterLogs(opts, "SharesSyncerModified")
	if err != nil {
		return nil, err
	}
	return &OBLSSharesSyncerModifiedIterator{contract: _OBLS.contract, event: "SharesSyncerModified", logs: logs, sub: sub}, nil
}

// WatchSharesSyncerModified is a free log subscription operation binding the contract event 0x797f7e8ad76b743fbff71cb58d982a1bd35e4df14315d7687b187c1c5c257e4d.
//
// Solidity: event SharesSyncerModified(address syncer)
func (_OBLS *OBLSFilterer) WatchSharesSyncerModified(opts *bind.WatchOpts, sink chan<- *OBLSSharesSyncerModified) (event.Subscription, error) {

	logs, sub, err := _OBLS.contract.WatchLogs(opts, "SharesSyncerModified")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OBLSSharesSyncerModified)
				if err := _OBLS.contract.UnpackLog(event, "SharesSyncerModified", log); err != nil {
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

// ParseSharesSyncerModified is a log parse operation binding the contract event 0x797f7e8ad76b743fbff71cb58d982a1bd35e4df14315d7687b187c1c5c257e4d.
//
// Solidity: event SharesSyncerModified(address syncer)
func (_OBLS *OBLSFilterer) ParseSharesSyncerModified(log types.Log) (*OBLSSharesSyncerModified, error) {
	event := new(OBLSSharesSyncerModified)
	if err := _OBLS.contract.UnpackLog(event, "SharesSyncerModified", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
