// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractProxyAdmin

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

// ContractProxyAdminMetaData contains all meta data concerning the ContractProxyAdmin contract.
var ContractProxyAdminMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"contractTransparentUpgradeableProxy\",\"name\":\"proxy\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"changeProxyAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractTransparentUpgradeableProxy\",\"name\":\"proxy\",\"type\":\"address\"}],\"name\":\"getProxyAdmin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractTransparentUpgradeableProxy\",\"name\":\"proxy\",\"type\":\"address\"}],\"name\":\"getProxyImplementation\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractTransparentUpgradeableProxy\",\"name\":\"proxy\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"upgrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractTransparentUpgradeableProxy\",\"name\":\"proxy\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
	Bin: "0x608080604052600436101561001357600080fd5b60009081803560e01c918263204e1c7a1461046d5750508063715018a6146104135780637eff275e146103825780638da5cb5b1461035b5780639623609d1461024957806399a88ec4146101b4578063f2fde38b146100ee5763f3b7dead1461007b57600080fd5b346100eb5760203660031901126100eb57808060046001600160a01b036100a06104ad565b6040516303e1469160e61b815291165afa6100b9610532565b90156100e75780516020916001600160a01b03916100de919081018401908401610562565b16604051908152f35b5080fd5b80fd5b50346100eb5760203660031901126100eb576101086104ad565b610110610581565b6001600160a01b031680156101605781546001600160a01b03198116821783556001600160a01b03167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e08380a380f35b60405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b6064820152608490fd5b50346100eb5760403660031901126100eb57806101cf6104ad565b6101d76104c8565b906101e0610581565b6001600160a01b031690813b1561024557604051631b2ce7f360e11b81526001600160a01b0390911660048201529082908290602490829084905af1801561023a576102295750f35b81610233916104de565b6100eb5780f35b6040513d84823e3d90fd5b5050fd5b5060603660031901126100eb5761025e6104ad565b906102676104c8565b6044359267ffffffffffffffff8411610353573660238501121561035357836004013561029381610516565b946102a160405196876104de565b81865236602483830101116103575781859260246020930183890137860101526102c9610581565b6001600160a01b0316803b15610353576040805163278f794360e11b81526001600160a01b0390931660048401526024830152835160448301819052835b81811061033d57848085818187816064818a86838284010152601f8019910116810103019134905af1801561023a576102295750f35b8060208092880101516064828701015201610307565b8280fd5b8480fd5b50346100eb57806003193601126100eb57546040516001600160a01b039091168152602090f35b50346100eb5760403660031901126100eb578061039d6104ad565b6103a56104c8565b906103ae610581565b6001600160a01b031690813b15610245576040516308f2839760e41b81526001600160a01b039091166004820152919081908390602490829084905af18015610406576103f85780f35b610401916104de565b388180f35b50604051903d90823e3d90fd5b50346100eb57806003193601126100eb5761042c610581565b80546001600160a01b03198116825581906001600160a01b03167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e08280a380f35b8190346104aa5760203660031901126104aa5781906004906001600160a01b036104956104ad565b635c60da1b60e01b8352165afa6100b9610532565b50fd5b600435906001600160a01b03821682036104c357565b600080fd5b602435906001600160a01b03821682036104c357565b90601f8019910116810190811067ffffffffffffffff82111761050057604052565b634e487b7160e01b600052604160045260246000fd5b67ffffffffffffffff811161050057601f01601f191660200190565b3d1561055d573d9061054382610516565b9161055160405193846104de565b82523d6000602084013e565b606090565b908160209103126104c357516001600160a01b03811681036104c35790565b6000546001600160a01b0316330361059557565b606460405162461bcd60e51b815260206004820152602060248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152fdfea264697066735822122079fcf01cf4da43dfa66c3275c8045443612f453afeb6d2cdd87f95bf222b11a964736f6c634300081a0033",
}

// ContractProxyAdminABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractProxyAdminMetaData.ABI instead.
var ContractProxyAdminABI = ContractProxyAdminMetaData.ABI

// ContractProxyAdminBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractProxyAdminMetaData.Bin instead.
var ContractProxyAdminBin = ContractProxyAdminMetaData.Bin

// DeployContractProxyAdmin deploys a new Ethereum contract, binding an instance of ContractProxyAdmin to it.
func DeployContractProxyAdmin(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ContractProxyAdmin, error) {
	parsed, err := ContractProxyAdminMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractProxyAdminBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractProxyAdmin{ContractProxyAdminCaller: ContractProxyAdminCaller{contract: contract}, ContractProxyAdminTransactor: ContractProxyAdminTransactor{contract: contract}, ContractProxyAdminFilterer: ContractProxyAdminFilterer{contract: contract}}, nil
}

// ContractProxyAdminMethods is an auto generated interface around an Ethereum contract.
type ContractProxyAdminMethods interface {
	ContractProxyAdminCalls
	ContractProxyAdminTransacts
	ContractProxyAdminFilters
}

// ContractProxyAdminCalls is an auto generated interface that defines the call methods available for an Ethereum contract.
type ContractProxyAdminCalls interface {
	GetProxyAdmin(opts *bind.CallOpts, proxy common.Address) (common.Address, error)

	GetProxyImplementation(opts *bind.CallOpts, proxy common.Address) (common.Address, error)

	Owner(opts *bind.CallOpts) (common.Address, error)
}

// ContractProxyAdminTransacts is an auto generated interface that defines the transact methods available for an Ethereum contract.
type ContractProxyAdminTransacts interface {
	ChangeProxyAdmin(opts *bind.TransactOpts, proxy common.Address, newAdmin common.Address) (*types.Transaction, error)

	RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error)

	TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error)

	Upgrade(opts *bind.TransactOpts, proxy common.Address, implementation common.Address) (*types.Transaction, error)

	UpgradeAndCall(opts *bind.TransactOpts, proxy common.Address, implementation common.Address, data []byte) (*types.Transaction, error)
}

// ContractProxyAdminFilterer is an auto generated interface that defines the log filtering methods available for an Ethereum contract.
type ContractProxyAdminFilters interface {
	FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractProxyAdminOwnershipTransferredIterator, error)
	WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractProxyAdminOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error)
	ParseOwnershipTransferred(log types.Log) (*ContractProxyAdminOwnershipTransferred, error)
}

// ContractProxyAdmin is an auto generated Go binding around an Ethereum contract.
type ContractProxyAdmin struct {
	ContractProxyAdminCaller     // Read-only binding to the contract
	ContractProxyAdminTransactor // Write-only binding to the contract
	ContractProxyAdminFilterer   // Log filterer for contract events
}

// ContractProxyAdmin implements the ContractProxyAdminMethods interface.
var _ ContractProxyAdminMethods = (*ContractProxyAdmin)(nil)

// ContractProxyAdminCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractProxyAdminCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractProxyAdminCaller implements the ContractProxyAdminCalls interface.
var _ ContractProxyAdminCalls = (*ContractProxyAdminCaller)(nil)

// ContractProxyAdminTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractProxyAdminTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractProxyAdminTransactor implements the ContractProxyAdminTransacts interface.
var _ ContractProxyAdminTransacts = (*ContractProxyAdminTransactor)(nil)

// ContractProxyAdminFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractProxyAdminFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractProxyAdminFilterer implements the ContractProxyAdminFilters interface.
var _ ContractProxyAdminFilters = (*ContractProxyAdminFilterer)(nil)

// ContractProxyAdminSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractProxyAdminSession struct {
	Contract     *ContractProxyAdmin // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractProxyAdminCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractProxyAdminCallerSession struct {
	Contract *ContractProxyAdminCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// ContractProxyAdminTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractProxyAdminTransactorSession struct {
	Contract     *ContractProxyAdminTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// ContractProxyAdminRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractProxyAdminRaw struct {
	Contract *ContractProxyAdmin // Generic contract binding to access the raw methods on
}

// ContractProxyAdminCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractProxyAdminCallerRaw struct {
	Contract *ContractProxyAdminCaller // Generic read-only contract binding to access the raw methods on
}

// ContractProxyAdminTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractProxyAdminTransactorRaw struct {
	Contract *ContractProxyAdminTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractProxyAdmin creates a new instance of ContractProxyAdmin, bound to a specific deployed contract.
func NewContractProxyAdmin(address common.Address, backend bind.ContractBackend) (*ContractProxyAdmin, error) {
	contract, err := bindContractProxyAdmin(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractProxyAdmin{ContractProxyAdminCaller: ContractProxyAdminCaller{contract: contract}, ContractProxyAdminTransactor: ContractProxyAdminTransactor{contract: contract}, ContractProxyAdminFilterer: ContractProxyAdminFilterer{contract: contract}}, nil
}

// NewContractProxyAdminCaller creates a new read-only instance of ContractProxyAdmin, bound to a specific deployed contract.
func NewContractProxyAdminCaller(address common.Address, caller bind.ContractCaller) (*ContractProxyAdminCaller, error) {
	contract, err := bindContractProxyAdmin(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractProxyAdminCaller{contract: contract}, nil
}

// NewContractProxyAdminTransactor creates a new write-only instance of ContractProxyAdmin, bound to a specific deployed contract.
func NewContractProxyAdminTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractProxyAdminTransactor, error) {
	contract, err := bindContractProxyAdmin(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractProxyAdminTransactor{contract: contract}, nil
}

// NewContractProxyAdminFilterer creates a new log filterer instance of ContractProxyAdmin, bound to a specific deployed contract.
func NewContractProxyAdminFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractProxyAdminFilterer, error) {
	contract, err := bindContractProxyAdmin(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractProxyAdminFilterer{contract: contract}, nil
}

// bindContractProxyAdmin binds a generic wrapper to an already deployed contract.
func bindContractProxyAdmin(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractProxyAdminMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractProxyAdmin *ContractProxyAdminRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractProxyAdmin.Contract.ContractProxyAdminCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractProxyAdmin *ContractProxyAdminRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.ContractProxyAdminTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractProxyAdmin *ContractProxyAdminRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.ContractProxyAdminTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractProxyAdmin *ContractProxyAdminCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractProxyAdmin.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractProxyAdmin *ContractProxyAdminTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractProxyAdmin *ContractProxyAdminTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.contract.Transact(opts, method, params...)
}

// GetProxyAdmin is a free data retrieval call binding the contract method 0xf3b7dead.
//
// Solidity: function getProxyAdmin(address proxy) view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminCaller) GetProxyAdmin(opts *bind.CallOpts, proxy common.Address) (common.Address, error) {
	var out []interface{}
	err := _ContractProxyAdmin.contract.Call(opts, &out, "getProxyAdmin", proxy)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetProxyAdmin is a free data retrieval call binding the contract method 0xf3b7dead.
//
// Solidity: function getProxyAdmin(address proxy) view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminSession) GetProxyAdmin(proxy common.Address) (common.Address, error) {
	return _ContractProxyAdmin.Contract.GetProxyAdmin(&_ContractProxyAdmin.CallOpts, proxy)
}

// GetProxyAdmin is a free data retrieval call binding the contract method 0xf3b7dead.
//
// Solidity: function getProxyAdmin(address proxy) view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminCallerSession) GetProxyAdmin(proxy common.Address) (common.Address, error) {
	return _ContractProxyAdmin.Contract.GetProxyAdmin(&_ContractProxyAdmin.CallOpts, proxy)
}

// GetProxyImplementation is a free data retrieval call binding the contract method 0x204e1c7a.
//
// Solidity: function getProxyImplementation(address proxy) view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminCaller) GetProxyImplementation(opts *bind.CallOpts, proxy common.Address) (common.Address, error) {
	var out []interface{}
	err := _ContractProxyAdmin.contract.Call(opts, &out, "getProxyImplementation", proxy)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetProxyImplementation is a free data retrieval call binding the contract method 0x204e1c7a.
//
// Solidity: function getProxyImplementation(address proxy) view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminSession) GetProxyImplementation(proxy common.Address) (common.Address, error) {
	return _ContractProxyAdmin.Contract.GetProxyImplementation(&_ContractProxyAdmin.CallOpts, proxy)
}

// GetProxyImplementation is a free data retrieval call binding the contract method 0x204e1c7a.
//
// Solidity: function getProxyImplementation(address proxy) view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminCallerSession) GetProxyImplementation(proxy common.Address) (common.Address, error) {
	return _ContractProxyAdmin.Contract.GetProxyImplementation(&_ContractProxyAdmin.CallOpts, proxy)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractProxyAdmin.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminSession) Owner() (common.Address, error) {
	return _ContractProxyAdmin.Contract.Owner(&_ContractProxyAdmin.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractProxyAdmin *ContractProxyAdminCallerSession) Owner() (common.Address, error) {
	return _ContractProxyAdmin.Contract.Owner(&_ContractProxyAdmin.CallOpts)
}

// ChangeProxyAdmin is a paid mutator transaction binding the contract method 0x7eff275e.
//
// Solidity: function changeProxyAdmin(address proxy, address newAdmin) returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactor) ChangeProxyAdmin(opts *bind.TransactOpts, proxy common.Address, newAdmin common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.contract.Transact(opts, "changeProxyAdmin", proxy, newAdmin)
}

// ChangeProxyAdmin is a paid mutator transaction binding the contract method 0x7eff275e.
//
// Solidity: function changeProxyAdmin(address proxy, address newAdmin) returns()
func (_ContractProxyAdmin *ContractProxyAdminSession) ChangeProxyAdmin(proxy common.Address, newAdmin common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.ChangeProxyAdmin(&_ContractProxyAdmin.TransactOpts, proxy, newAdmin)
}

// ChangeProxyAdmin is a paid mutator transaction binding the contract method 0x7eff275e.
//
// Solidity: function changeProxyAdmin(address proxy, address newAdmin) returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactorSession) ChangeProxyAdmin(proxy common.Address, newAdmin common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.ChangeProxyAdmin(&_ContractProxyAdmin.TransactOpts, proxy, newAdmin)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractProxyAdmin.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractProxyAdmin *ContractProxyAdminSession) RenounceOwnership() (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.RenounceOwnership(&_ContractProxyAdmin.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.RenounceOwnership(&_ContractProxyAdmin.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractProxyAdmin *ContractProxyAdminSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.TransferOwnership(&_ContractProxyAdmin.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.TransferOwnership(&_ContractProxyAdmin.TransactOpts, newOwner)
}

// Upgrade is a paid mutator transaction binding the contract method 0x99a88ec4.
//
// Solidity: function upgrade(address proxy, address implementation) returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactor) Upgrade(opts *bind.TransactOpts, proxy common.Address, implementation common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.contract.Transact(opts, "upgrade", proxy, implementation)
}

// Upgrade is a paid mutator transaction binding the contract method 0x99a88ec4.
//
// Solidity: function upgrade(address proxy, address implementation) returns()
func (_ContractProxyAdmin *ContractProxyAdminSession) Upgrade(proxy common.Address, implementation common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.Upgrade(&_ContractProxyAdmin.TransactOpts, proxy, implementation)
}

// Upgrade is a paid mutator transaction binding the contract method 0x99a88ec4.
//
// Solidity: function upgrade(address proxy, address implementation) returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactorSession) Upgrade(proxy common.Address, implementation common.Address) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.Upgrade(&_ContractProxyAdmin.TransactOpts, proxy, implementation)
}

// UpgradeAndCall is a paid mutator transaction binding the contract method 0x9623609d.
//
// Solidity: function upgradeAndCall(address proxy, address implementation, bytes data) payable returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactor) UpgradeAndCall(opts *bind.TransactOpts, proxy common.Address, implementation common.Address, data []byte) (*types.Transaction, error) {
	return _ContractProxyAdmin.contract.Transact(opts, "upgradeAndCall", proxy, implementation, data)
}

// UpgradeAndCall is a paid mutator transaction binding the contract method 0x9623609d.
//
// Solidity: function upgradeAndCall(address proxy, address implementation, bytes data) payable returns()
func (_ContractProxyAdmin *ContractProxyAdminSession) UpgradeAndCall(proxy common.Address, implementation common.Address, data []byte) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.UpgradeAndCall(&_ContractProxyAdmin.TransactOpts, proxy, implementation, data)
}

// UpgradeAndCall is a paid mutator transaction binding the contract method 0x9623609d.
//
// Solidity: function upgradeAndCall(address proxy, address implementation, bytes data) payable returns()
func (_ContractProxyAdmin *ContractProxyAdminTransactorSession) UpgradeAndCall(proxy common.Address, implementation common.Address, data []byte) (*types.Transaction, error) {
	return _ContractProxyAdmin.Contract.UpgradeAndCall(&_ContractProxyAdmin.TransactOpts, proxy, implementation, data)
}

// ContractProxyAdminOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ContractProxyAdmin contract.
type ContractProxyAdminOwnershipTransferredIterator struct {
	Event *ContractProxyAdminOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ContractProxyAdminOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractProxyAdminOwnershipTransferred)
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
		it.Event = new(ContractProxyAdminOwnershipTransferred)
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
func (it *ContractProxyAdminOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractProxyAdminOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractProxyAdminOwnershipTransferred represents a OwnershipTransferred event raised by the ContractProxyAdmin contract.
type ContractProxyAdminOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractProxyAdmin *ContractProxyAdminFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractProxyAdminOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ContractProxyAdmin.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ContractProxyAdminOwnershipTransferredIterator{contract: _ContractProxyAdmin.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractProxyAdmin *ContractProxyAdminFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractProxyAdminOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ContractProxyAdmin.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractProxyAdminOwnershipTransferred)
				if err := _ContractProxyAdmin.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ContractProxyAdmin *ContractProxyAdminFilterer) ParseOwnershipTransferred(log types.Log) (*ContractProxyAdminOwnershipTransferred, error) {
	event := new(ContractProxyAdminOwnershipTransferred)
	if err := _ContractProxyAdmin.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
