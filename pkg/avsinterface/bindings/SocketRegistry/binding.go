// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractSocketRegistry

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

// ContractSocketRegistryMetaData contains all meta data concerning the ContractSocketRegistry contract.
var ContractSocketRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIRegistryCoordinator\",\"name\":\"_registryCoordinator\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_operatorId\",\"type\":\"bytes32\"}],\"name\":\"getOperatorSocket\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"_operatorIds\",\"type\":\"bytes32[]\"},{\"internalType\":\"string[]\",\"name\":\"_sockets\",\"type\":\"string[]\"}],\"name\":\"migrateOperatorSockets\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"operatorIdToSocket\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registryCoordinator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_operatorId\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"_socket\",\"type\":\"string\"}],\"name\":\"setOperatorSocket\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052600436101561001257600080fd5b60003560e01c806310bea0d71461026a578063639b9957146102eb5780636d14a987146102a6578063af65fdfc1461026a5763f043367e1461005357600080fd5b346102655760403660031901126102655760243567ffffffffffffffff811161026557610084903690600401610711565b7f00000000000000000000000013a05d12b8061f8f12beca62a42b9815310214396001600160a01b031633036101e45760043560005260006020526040600020815167ffffffffffffffff81116101ce576100df8254610768565b601f8111610186575b50602092601f8211600114610126579281929360009261011b575b5050600019600383901b1c191660019190911b179055005b015190503880610103565b601f1982169383600052806000209160005b86811061016e5750836001959610610155575b505050811b019055005b015160001960f88460031b161c1916905538808061014b565b91926020600181928685015181550194019201610138565b826000526020600020601f830160051c810191602084106101c4575b601f0160051c01905b8181106101b857506100e8565b600081556001016101ab565b90915081906101a2565b634e487b7160e01b600052604160045260246000fd5b60405162461bcd60e51b815260206004820152604d60248201527f536f636b657452656769737472792e6f6e6c795265676973747279436f6f726460448201527f696e61746f723a2063616c6c6572206973206e6f74207468652052656769737460648201526c393ca1b7b7b93234b730ba37b960991b608482015260a490fd5b600080fd5b346102655760203660031901126102655760043560005260006020526102a261029660406000206107a2565b6040519182918261068e565b0390f35b34610265576000366003190112610265576040517f00000000000000000000000013a05d12b8061f8f12beca62a42b9815310214396001600160a01b03168152602090f35b346102655760403660031901126102655760043567ffffffffffffffff8111610265573660238201121561026557806004013590610328826106f9565b9161033660405193846106d7565b8083526024602084019160051b8301019136831161026557602401905b82821061067e578360243567ffffffffffffffff81116102655736602382011215610265578060040135610386816106f9565b9161039460405193846106d7565b8183526024602084019260051b820101903682116102655760248101925b82841061064e57604051638da5cb5b60e01b8152869086906020816004817f00000000000000000000000013a05d12b8061f8f12beca62a42b9815310214396001600160a01b03165afa908115610642576000916105f9575b506001600160a01b0316330361056857906000915b8151831015610566576104338382610846565b519261043f8184610846565b5160005260006020526040600020845167ffffffffffffffff81116101ce576104688254610768565b601f811161051e575b506020601f82116001146104b557819060019596976000926104aa575b5050600019600383901b1c191690841b1790555b019190610420565b01519050878061048e565b601f1982169683600052816000209760005b81811061050657509160019697989184889594106104ed575b505050811b0190556104a2565b015160001960f88460031b161c191690558780806104e0565b92986020600181928c8601518155019a0193016104c7565b826000526020600020601f830160051c8101916020841061055c575b601f0160051c01905b8181106105505750610471565b60008155600101610543565b909150819061053a565b005b60405162461bcd60e51b815260206004820152605760248201527f536f636b657452656769737472792e6f6e6c79436f6f7264696e61746f724f7760448201527f6e65723a2063616c6c6572206973206e6f7420746865206f776e6572206f662060648201527f746865207265676973747279436f6f7264696e61746f72000000000000000000608482015260a490fd5b6020813d60201161063a575b81610612602093836106d7565b810103126106365751906001600160a01b038216820361063357508361040b565b80fd5b5080fd5b3d9150610605565b6040513d6000823e3d90fd5b833567ffffffffffffffff811161026557602091610673839260243691870101610711565b8152019301926103b2565b8135815260209182019101610353565b91909160208152825180602083015260005b8181106106c1575060409293506000838284010152601f8019910116010190565b80602080928701015160408286010152016106a0565b90601f8019910116810190811067ffffffffffffffff8211176101ce57604052565b67ffffffffffffffff81116101ce5760051b60200190565b81601f820112156102655780359067ffffffffffffffff82116101ce5760405192610746601f8401601f1916602001856106d7565b8284526020838301011161026557816000926020809301838601378301015290565b90600182811c92168015610798575b602083101461078257565b634e487b7160e01b600052602260045260246000fd5b91607f1691610777565b90604051918260008254926107b684610768565b808452936001811690811561082457506001146107dd575b506107db925003836106d7565b565b90506000929192526020600020906000915b8183106108085750509060206107db92820101386107ce565b60209193508060019154838589010152019101909184926107ef565b9050602092506107db94915060ff191682840152151560051b820101386107ce565b805182101561085a5760209160051b010190565b634e487b7160e01b600052603260045260246000fdfea26469706673582212203f8170f3a74b453cded6b82a6f9c2270c0a6fa919f00edad7dc66e0dcc5c1bf464736f6c634300081a0033",
}

// ContractSocketRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractSocketRegistryMetaData.ABI instead.
var ContractSocketRegistryABI = ContractSocketRegistryMetaData.ABI

// ContractSocketRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractSocketRegistryMetaData.Bin instead.
var ContractSocketRegistryBin = ContractSocketRegistryMetaData.Bin

// DeployContractSocketRegistry deploys a new Ethereum contract, binding an instance of ContractSocketRegistry to it.
func DeployContractSocketRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, _registryCoordinator common.Address) (common.Address, *types.Transaction, *ContractSocketRegistry, error) {
	parsed, err := ContractSocketRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractSocketRegistryBin), backend, _registryCoordinator)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractSocketRegistry{ContractSocketRegistryCaller: ContractSocketRegistryCaller{contract: contract}, ContractSocketRegistryTransactor: ContractSocketRegistryTransactor{contract: contract}, ContractSocketRegistryFilterer: ContractSocketRegistryFilterer{contract: contract}}, nil
}

// ContractSocketRegistryMethods is an auto generated interface around an Ethereum contract.
type ContractSocketRegistryMethods interface {
	ContractSocketRegistryCalls
	ContractSocketRegistryTransacts
	ContractSocketRegistryFilters
}

// ContractSocketRegistryCalls is an auto generated interface that defines the call methods available for an Ethereum contract.
type ContractSocketRegistryCalls interface {
	GetOperatorSocket(opts *bind.CallOpts, _operatorId [32]byte) (string, error)

	OperatorIdToSocket(opts *bind.CallOpts, arg0 [32]byte) (string, error)

	RegistryCoordinator(opts *bind.CallOpts) (common.Address, error)
}

// ContractSocketRegistryTransacts is an auto generated interface that defines the transact methods available for an Ethereum contract.
type ContractSocketRegistryTransacts interface {
	MigrateOperatorSockets(opts *bind.TransactOpts, _operatorIds [][32]byte, _sockets []string) (*types.Transaction, error)

	SetOperatorSocket(opts *bind.TransactOpts, _operatorId [32]byte, _socket string) (*types.Transaction, error)
}

// ContractSocketRegistryFilterer is an auto generated interface that defines the log filtering methods available for an Ethereum contract.
type ContractSocketRegistryFilters interface {
}

// ContractSocketRegistry is an auto generated Go binding around an Ethereum contract.
type ContractSocketRegistry struct {
	ContractSocketRegistryCaller     // Read-only binding to the contract
	ContractSocketRegistryTransactor // Write-only binding to the contract
	ContractSocketRegistryFilterer   // Log filterer for contract events
}

// ContractSocketRegistry implements the ContractSocketRegistryMethods interface.
var _ ContractSocketRegistryMethods = (*ContractSocketRegistry)(nil)

// ContractSocketRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractSocketRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSocketRegistryCaller implements the ContractSocketRegistryCalls interface.
var _ ContractSocketRegistryCalls = (*ContractSocketRegistryCaller)(nil)

// ContractSocketRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractSocketRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSocketRegistryTransactor implements the ContractSocketRegistryTransacts interface.
var _ ContractSocketRegistryTransacts = (*ContractSocketRegistryTransactor)(nil)

// ContractSocketRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractSocketRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSocketRegistryFilterer implements the ContractSocketRegistryFilters interface.
var _ ContractSocketRegistryFilters = (*ContractSocketRegistryFilterer)(nil)

// ContractSocketRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSocketRegistrySession struct {
	Contract     *ContractSocketRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ContractSocketRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractSocketRegistryCallerSession struct {
	Contract *ContractSocketRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// ContractSocketRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractSocketRegistryTransactorSession struct {
	Contract     *ContractSocketRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// ContractSocketRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractSocketRegistryRaw struct {
	Contract *ContractSocketRegistry // Generic contract binding to access the raw methods on
}

// ContractSocketRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractSocketRegistryCallerRaw struct {
	Contract *ContractSocketRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ContractSocketRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractSocketRegistryTransactorRaw struct {
	Contract *ContractSocketRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractSocketRegistry creates a new instance of ContractSocketRegistry, bound to a specific deployed contract.
func NewContractSocketRegistry(address common.Address, backend bind.ContractBackend) (*ContractSocketRegistry, error) {
	contract, err := bindContractSocketRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractSocketRegistry{ContractSocketRegistryCaller: ContractSocketRegistryCaller{contract: contract}, ContractSocketRegistryTransactor: ContractSocketRegistryTransactor{contract: contract}, ContractSocketRegistryFilterer: ContractSocketRegistryFilterer{contract: contract}}, nil
}

// NewContractSocketRegistryCaller creates a new read-only instance of ContractSocketRegistry, bound to a specific deployed contract.
func NewContractSocketRegistryCaller(address common.Address, caller bind.ContractCaller) (*ContractSocketRegistryCaller, error) {
	contract, err := bindContractSocketRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractSocketRegistryCaller{contract: contract}, nil
}

// NewContractSocketRegistryTransactor creates a new write-only instance of ContractSocketRegistry, bound to a specific deployed contract.
func NewContractSocketRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractSocketRegistryTransactor, error) {
	contract, err := bindContractSocketRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractSocketRegistryTransactor{contract: contract}, nil
}

// NewContractSocketRegistryFilterer creates a new log filterer instance of ContractSocketRegistry, bound to a specific deployed contract.
func NewContractSocketRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractSocketRegistryFilterer, error) {
	contract, err := bindContractSocketRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractSocketRegistryFilterer{contract: contract}, nil
}

// bindContractSocketRegistry binds a generic wrapper to an already deployed contract.
func bindContractSocketRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractSocketRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractSocketRegistry *ContractSocketRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractSocketRegistry.Contract.ContractSocketRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractSocketRegistry *ContractSocketRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractSocketRegistry.Contract.ContractSocketRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractSocketRegistry *ContractSocketRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractSocketRegistry.Contract.ContractSocketRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractSocketRegistry *ContractSocketRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractSocketRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractSocketRegistry *ContractSocketRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractSocketRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractSocketRegistry *ContractSocketRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractSocketRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x10bea0d7.
//
// Solidity: function getOperatorSocket(bytes32 _operatorId) view returns(string)
func (_ContractSocketRegistry *ContractSocketRegistryCaller) GetOperatorSocket(opts *bind.CallOpts, _operatorId [32]byte) (string, error) {
	var out []interface{}
	err := _ContractSocketRegistry.contract.Call(opts, &out, "getOperatorSocket", _operatorId)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x10bea0d7.
//
// Solidity: function getOperatorSocket(bytes32 _operatorId) view returns(string)
func (_ContractSocketRegistry *ContractSocketRegistrySession) GetOperatorSocket(_operatorId [32]byte) (string, error) {
	return _ContractSocketRegistry.Contract.GetOperatorSocket(&_ContractSocketRegistry.CallOpts, _operatorId)
}

// GetOperatorSocket is a free data retrieval call binding the contract method 0x10bea0d7.
//
// Solidity: function getOperatorSocket(bytes32 _operatorId) view returns(string)
func (_ContractSocketRegistry *ContractSocketRegistryCallerSession) GetOperatorSocket(_operatorId [32]byte) (string, error) {
	return _ContractSocketRegistry.Contract.GetOperatorSocket(&_ContractSocketRegistry.CallOpts, _operatorId)
}

// OperatorIdToSocket is a free data retrieval call binding the contract method 0xaf65fdfc.
//
// Solidity: function operatorIdToSocket(bytes32 ) view returns(string)
func (_ContractSocketRegistry *ContractSocketRegistryCaller) OperatorIdToSocket(opts *bind.CallOpts, arg0 [32]byte) (string, error) {
	var out []interface{}
	err := _ContractSocketRegistry.contract.Call(opts, &out, "operatorIdToSocket", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// OperatorIdToSocket is a free data retrieval call binding the contract method 0xaf65fdfc.
//
// Solidity: function operatorIdToSocket(bytes32 ) view returns(string)
func (_ContractSocketRegistry *ContractSocketRegistrySession) OperatorIdToSocket(arg0 [32]byte) (string, error) {
	return _ContractSocketRegistry.Contract.OperatorIdToSocket(&_ContractSocketRegistry.CallOpts, arg0)
}

// OperatorIdToSocket is a free data retrieval call binding the contract method 0xaf65fdfc.
//
// Solidity: function operatorIdToSocket(bytes32 ) view returns(string)
func (_ContractSocketRegistry *ContractSocketRegistryCallerSession) OperatorIdToSocket(arg0 [32]byte) (string, error) {
	return _ContractSocketRegistry.Contract.OperatorIdToSocket(&_ContractSocketRegistry.CallOpts, arg0)
}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractSocketRegistry *ContractSocketRegistryCaller) RegistryCoordinator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractSocketRegistry.contract.Call(opts, &out, "registryCoordinator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractSocketRegistry *ContractSocketRegistrySession) RegistryCoordinator() (common.Address, error) {
	return _ContractSocketRegistry.Contract.RegistryCoordinator(&_ContractSocketRegistry.CallOpts)
}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractSocketRegistry *ContractSocketRegistryCallerSession) RegistryCoordinator() (common.Address, error) {
	return _ContractSocketRegistry.Contract.RegistryCoordinator(&_ContractSocketRegistry.CallOpts)
}

// MigrateOperatorSockets is a paid mutator transaction binding the contract method 0x639b9957.
//
// Solidity: function migrateOperatorSockets(bytes32[] _operatorIds, string[] _sockets) returns()
func (_ContractSocketRegistry *ContractSocketRegistryTransactor) MigrateOperatorSockets(opts *bind.TransactOpts, _operatorIds [][32]byte, _sockets []string) (*types.Transaction, error) {
	return _ContractSocketRegistry.contract.Transact(opts, "migrateOperatorSockets", _operatorIds, _sockets)
}

// MigrateOperatorSockets is a paid mutator transaction binding the contract method 0x639b9957.
//
// Solidity: function migrateOperatorSockets(bytes32[] _operatorIds, string[] _sockets) returns()
func (_ContractSocketRegistry *ContractSocketRegistrySession) MigrateOperatorSockets(_operatorIds [][32]byte, _sockets []string) (*types.Transaction, error) {
	return _ContractSocketRegistry.Contract.MigrateOperatorSockets(&_ContractSocketRegistry.TransactOpts, _operatorIds, _sockets)
}

// MigrateOperatorSockets is a paid mutator transaction binding the contract method 0x639b9957.
//
// Solidity: function migrateOperatorSockets(bytes32[] _operatorIds, string[] _sockets) returns()
func (_ContractSocketRegistry *ContractSocketRegistryTransactorSession) MigrateOperatorSockets(_operatorIds [][32]byte, _sockets []string) (*types.Transaction, error) {
	return _ContractSocketRegistry.Contract.MigrateOperatorSockets(&_ContractSocketRegistry.TransactOpts, _operatorIds, _sockets)
}

// SetOperatorSocket is a paid mutator transaction binding the contract method 0xf043367e.
//
// Solidity: function setOperatorSocket(bytes32 _operatorId, string _socket) returns()
func (_ContractSocketRegistry *ContractSocketRegistryTransactor) SetOperatorSocket(opts *bind.TransactOpts, _operatorId [32]byte, _socket string) (*types.Transaction, error) {
	return _ContractSocketRegistry.contract.Transact(opts, "setOperatorSocket", _operatorId, _socket)
}

// SetOperatorSocket is a paid mutator transaction binding the contract method 0xf043367e.
//
// Solidity: function setOperatorSocket(bytes32 _operatorId, string _socket) returns()
func (_ContractSocketRegistry *ContractSocketRegistrySession) SetOperatorSocket(_operatorId [32]byte, _socket string) (*types.Transaction, error) {
	return _ContractSocketRegistry.Contract.SetOperatorSocket(&_ContractSocketRegistry.TransactOpts, _operatorId, _socket)
}

// SetOperatorSocket is a paid mutator transaction binding the contract method 0xf043367e.
//
// Solidity: function setOperatorSocket(bytes32 _operatorId, string _socket) returns()
func (_ContractSocketRegistry *ContractSocketRegistryTransactorSession) SetOperatorSocket(_operatorId [32]byte, _socket string) (*types.Transaction, error) {
	return _ContractSocketRegistry.Contract.SetOperatorSocket(&_ContractSocketRegistry.TransactOpts, _operatorId, _socket)
}
