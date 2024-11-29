// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractPauserRegistry

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

// ContractPauserRegistryMetaData contains all meta data concerning the ContractPauserRegistry contract.
var ContractPauserRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_pausers\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"_unpauser\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"pauser\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"canPause\",\"type\":\"bool\"}],\"name\":\"PauserStatusChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousUnpauser\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newUnpauser\",\"type\":\"address\"}],\"name\":\"UnpauserChanged\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"isPauser\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newPauser\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"canPause\",\"type\":\"bool\"}],\"name\":\"setIsPauser\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newUnpauser\",\"type\":\"address\"}],\"name\":\"setUnpauser\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unpauser\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608080604052600436101561001357600080fd5b60003560e01c90816346fbf68e1461024b575080638568520614610158578063ce548428146100795763eab66d7a1461004b57600080fd5b34610074576000366003190112610074576001546040516001600160a01b039091168152602090f35b600080fd5b3461007457602036600319011261007457610092610286565b6001546001600160a01b038116916100ab33841461029c565b6001600160a01b03169182156100fb5760407f06b4167a2528887a1e97a366eefe8549bfbf1ea3e6ac81cb2564a934d20e8892918151908152846020820152a16001600160a01b03191617600155005b60405162461bcd60e51b815260206004820152602f60248201527f50617573657252656769737472792e5f736574556e7061757365723a207a657260448201526e1bc81859191c995cdcc81a5b9c1d5d608a1b6064820152608490fd5b3461007457604036600319011261007457610171610286565b602435908115158092036100745761019460018060a01b0360015416331461029c565b6001600160a01b03169081156101f057816040917f65d3a1fd4c13f05cba164f80d03ce90fb4b5e21946bfc3ab7dbd434c2d0b91529360005260006020528260002060ff1981541660ff831617905582519182526020820152a1005b60405162461bcd60e51b815260206004820152602d60248201527f50617573657252656769737472792e5f7365745061757365723a207a65726f2060448201526c1859191c995cdcc81a5b9c1d5d609a1b6064820152608490fd5b34610074576020366003190112610074576020906001600160a01b0361026f610286565b166000526000825260ff6040600020541615158152f35b600435906001600160a01b038216820361007457565b156102a357565b60405162461bcd60e51b815260206004820152602a60248201527f6d73672e73656e646572206973206e6f74207065726d697373696f6e6564206160448201526939903ab73830bab9b2b960b11b6064820152608490fdfea264697066735822122032dca1f4b9e12a0666cb999cb7f2953dee680ee8687a6b57ae5b7555b3b8904164736f6c634300081a0033",
}

// ContractPauserRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractPauserRegistryMetaData.ABI instead.
var ContractPauserRegistryABI = ContractPauserRegistryMetaData.ABI

// ContractPauserRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractPauserRegistryMetaData.Bin instead.
var ContractPauserRegistryBin = ContractPauserRegistryMetaData.Bin

// DeployContractPauserRegistry deploys a new Ethereum contract, binding an instance of ContractPauserRegistry to it.
func DeployContractPauserRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, _pausers []common.Address, _unpauser common.Address) (common.Address, *types.Transaction, *ContractPauserRegistry, error) {
	parsed, err := ContractPauserRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractPauserRegistryBin), backend, _pausers, _unpauser)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractPauserRegistry{ContractPauserRegistryCaller: ContractPauserRegistryCaller{contract: contract}, ContractPauserRegistryTransactor: ContractPauserRegistryTransactor{contract: contract}, ContractPauserRegistryFilterer: ContractPauserRegistryFilterer{contract: contract}}, nil
}

// ContractPauserRegistryMethods is an auto generated interface around an Ethereum contract.
type ContractPauserRegistryMethods interface {
	ContractPauserRegistryCalls
	ContractPauserRegistryTransacts
	ContractPauserRegistryFilters
}

// ContractPauserRegistryCalls is an auto generated interface that defines the call methods available for an Ethereum contract.
type ContractPauserRegistryCalls interface {
	IsPauser(opts *bind.CallOpts, arg0 common.Address) (bool, error)

	Unpauser(opts *bind.CallOpts) (common.Address, error)
}

// ContractPauserRegistryTransacts is an auto generated interface that defines the transact methods available for an Ethereum contract.
type ContractPauserRegistryTransacts interface {
	SetIsPauser(opts *bind.TransactOpts, newPauser common.Address, canPause bool) (*types.Transaction, error)

	SetUnpauser(opts *bind.TransactOpts, newUnpauser common.Address) (*types.Transaction, error)
}

// ContractPauserRegistryFilterer is an auto generated interface that defines the log filtering methods available for an Ethereum contract.
type ContractPauserRegistryFilters interface {
	FilterPauserStatusChanged(opts *bind.FilterOpts) (*ContractPauserRegistryPauserStatusChangedIterator, error)
	WatchPauserStatusChanged(opts *bind.WatchOpts, sink chan<- *ContractPauserRegistryPauserStatusChanged) (event.Subscription, error)
	ParsePauserStatusChanged(log types.Log) (*ContractPauserRegistryPauserStatusChanged, error)

	FilterUnpauserChanged(opts *bind.FilterOpts) (*ContractPauserRegistryUnpauserChangedIterator, error)
	WatchUnpauserChanged(opts *bind.WatchOpts, sink chan<- *ContractPauserRegistryUnpauserChanged) (event.Subscription, error)
	ParseUnpauserChanged(log types.Log) (*ContractPauserRegistryUnpauserChanged, error)
}

// ContractPauserRegistry is an auto generated Go binding around an Ethereum contract.
type ContractPauserRegistry struct {
	ContractPauserRegistryCaller     // Read-only binding to the contract
	ContractPauserRegistryTransactor // Write-only binding to the contract
	ContractPauserRegistryFilterer   // Log filterer for contract events
}

// ContractPauserRegistry implements the ContractPauserRegistryMethods interface.
var _ ContractPauserRegistryMethods = (*ContractPauserRegistry)(nil)

// ContractPauserRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractPauserRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractPauserRegistryCaller implements the ContractPauserRegistryCalls interface.
var _ ContractPauserRegistryCalls = (*ContractPauserRegistryCaller)(nil)

// ContractPauserRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractPauserRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractPauserRegistryTransactor implements the ContractPauserRegistryTransacts interface.
var _ ContractPauserRegistryTransacts = (*ContractPauserRegistryTransactor)(nil)

// ContractPauserRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractPauserRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractPauserRegistryFilterer implements the ContractPauserRegistryFilters interface.
var _ ContractPauserRegistryFilters = (*ContractPauserRegistryFilterer)(nil)

// ContractPauserRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractPauserRegistrySession struct {
	Contract     *ContractPauserRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ContractPauserRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractPauserRegistryCallerSession struct {
	Contract *ContractPauserRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// ContractPauserRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractPauserRegistryTransactorSession struct {
	Contract     *ContractPauserRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// ContractPauserRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractPauserRegistryRaw struct {
	Contract *ContractPauserRegistry // Generic contract binding to access the raw methods on
}

// ContractPauserRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractPauserRegistryCallerRaw struct {
	Contract *ContractPauserRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ContractPauserRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractPauserRegistryTransactorRaw struct {
	Contract *ContractPauserRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractPauserRegistry creates a new instance of ContractPauserRegistry, bound to a specific deployed contract.
func NewContractPauserRegistry(address common.Address, backend bind.ContractBackend) (*ContractPauserRegistry, error) {
	contract, err := bindContractPauserRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractPauserRegistry{ContractPauserRegistryCaller: ContractPauserRegistryCaller{contract: contract}, ContractPauserRegistryTransactor: ContractPauserRegistryTransactor{contract: contract}, ContractPauserRegistryFilterer: ContractPauserRegistryFilterer{contract: contract}}, nil
}

// NewContractPauserRegistryCaller creates a new read-only instance of ContractPauserRegistry, bound to a specific deployed contract.
func NewContractPauserRegistryCaller(address common.Address, caller bind.ContractCaller) (*ContractPauserRegistryCaller, error) {
	contract, err := bindContractPauserRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractPauserRegistryCaller{contract: contract}, nil
}

// NewContractPauserRegistryTransactor creates a new write-only instance of ContractPauserRegistry, bound to a specific deployed contract.
func NewContractPauserRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractPauserRegistryTransactor, error) {
	contract, err := bindContractPauserRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractPauserRegistryTransactor{contract: contract}, nil
}

// NewContractPauserRegistryFilterer creates a new log filterer instance of ContractPauserRegistry, bound to a specific deployed contract.
func NewContractPauserRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractPauserRegistryFilterer, error) {
	contract, err := bindContractPauserRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractPauserRegistryFilterer{contract: contract}, nil
}

// bindContractPauserRegistry binds a generic wrapper to an already deployed contract.
func bindContractPauserRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractPauserRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractPauserRegistry *ContractPauserRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractPauserRegistry.Contract.ContractPauserRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractPauserRegistry *ContractPauserRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractPauserRegistry.Contract.ContractPauserRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractPauserRegistry *ContractPauserRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractPauserRegistry.Contract.ContractPauserRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractPauserRegistry *ContractPauserRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractPauserRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractPauserRegistry *ContractPauserRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractPauserRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractPauserRegistry *ContractPauserRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractPauserRegistry.Contract.contract.Transact(opts, method, params...)
}

// IsPauser is a free data retrieval call binding the contract method 0x46fbf68e.
//
// Solidity: function isPauser(address ) view returns(bool)
func (_ContractPauserRegistry *ContractPauserRegistryCaller) IsPauser(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _ContractPauserRegistry.contract.Call(opts, &out, "isPauser", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsPauser is a free data retrieval call binding the contract method 0x46fbf68e.
//
// Solidity: function isPauser(address ) view returns(bool)
func (_ContractPauserRegistry *ContractPauserRegistrySession) IsPauser(arg0 common.Address) (bool, error) {
	return _ContractPauserRegistry.Contract.IsPauser(&_ContractPauserRegistry.CallOpts, arg0)
}

// IsPauser is a free data retrieval call binding the contract method 0x46fbf68e.
//
// Solidity: function isPauser(address ) view returns(bool)
func (_ContractPauserRegistry *ContractPauserRegistryCallerSession) IsPauser(arg0 common.Address) (bool, error) {
	return _ContractPauserRegistry.Contract.IsPauser(&_ContractPauserRegistry.CallOpts, arg0)
}

// Unpauser is a free data retrieval call binding the contract method 0xeab66d7a.
//
// Solidity: function unpauser() view returns(address)
func (_ContractPauserRegistry *ContractPauserRegistryCaller) Unpauser(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractPauserRegistry.contract.Call(opts, &out, "unpauser")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Unpauser is a free data retrieval call binding the contract method 0xeab66d7a.
//
// Solidity: function unpauser() view returns(address)
func (_ContractPauserRegistry *ContractPauserRegistrySession) Unpauser() (common.Address, error) {
	return _ContractPauserRegistry.Contract.Unpauser(&_ContractPauserRegistry.CallOpts)
}

// Unpauser is a free data retrieval call binding the contract method 0xeab66d7a.
//
// Solidity: function unpauser() view returns(address)
func (_ContractPauserRegistry *ContractPauserRegistryCallerSession) Unpauser() (common.Address, error) {
	return _ContractPauserRegistry.Contract.Unpauser(&_ContractPauserRegistry.CallOpts)
}

// SetIsPauser is a paid mutator transaction binding the contract method 0x85685206.
//
// Solidity: function setIsPauser(address newPauser, bool canPause) returns()
func (_ContractPauserRegistry *ContractPauserRegistryTransactor) SetIsPauser(opts *bind.TransactOpts, newPauser common.Address, canPause bool) (*types.Transaction, error) {
	return _ContractPauserRegistry.contract.Transact(opts, "setIsPauser", newPauser, canPause)
}

// SetIsPauser is a paid mutator transaction binding the contract method 0x85685206.
//
// Solidity: function setIsPauser(address newPauser, bool canPause) returns()
func (_ContractPauserRegistry *ContractPauserRegistrySession) SetIsPauser(newPauser common.Address, canPause bool) (*types.Transaction, error) {
	return _ContractPauserRegistry.Contract.SetIsPauser(&_ContractPauserRegistry.TransactOpts, newPauser, canPause)
}

// SetIsPauser is a paid mutator transaction binding the contract method 0x85685206.
//
// Solidity: function setIsPauser(address newPauser, bool canPause) returns()
func (_ContractPauserRegistry *ContractPauserRegistryTransactorSession) SetIsPauser(newPauser common.Address, canPause bool) (*types.Transaction, error) {
	return _ContractPauserRegistry.Contract.SetIsPauser(&_ContractPauserRegistry.TransactOpts, newPauser, canPause)
}

// SetUnpauser is a paid mutator transaction binding the contract method 0xce548428.
//
// Solidity: function setUnpauser(address newUnpauser) returns()
func (_ContractPauserRegistry *ContractPauserRegistryTransactor) SetUnpauser(opts *bind.TransactOpts, newUnpauser common.Address) (*types.Transaction, error) {
	return _ContractPauserRegistry.contract.Transact(opts, "setUnpauser", newUnpauser)
}

// SetUnpauser is a paid mutator transaction binding the contract method 0xce548428.
//
// Solidity: function setUnpauser(address newUnpauser) returns()
func (_ContractPauserRegistry *ContractPauserRegistrySession) SetUnpauser(newUnpauser common.Address) (*types.Transaction, error) {
	return _ContractPauserRegistry.Contract.SetUnpauser(&_ContractPauserRegistry.TransactOpts, newUnpauser)
}

// SetUnpauser is a paid mutator transaction binding the contract method 0xce548428.
//
// Solidity: function setUnpauser(address newUnpauser) returns()
func (_ContractPauserRegistry *ContractPauserRegistryTransactorSession) SetUnpauser(newUnpauser common.Address) (*types.Transaction, error) {
	return _ContractPauserRegistry.Contract.SetUnpauser(&_ContractPauserRegistry.TransactOpts, newUnpauser)
}

// ContractPauserRegistryPauserStatusChangedIterator is returned from FilterPauserStatusChanged and is used to iterate over the raw logs and unpacked data for PauserStatusChanged events raised by the ContractPauserRegistry contract.
type ContractPauserRegistryPauserStatusChangedIterator struct {
	Event *ContractPauserRegistryPauserStatusChanged // Event containing the contract specifics and raw log

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
func (it *ContractPauserRegistryPauserStatusChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractPauserRegistryPauserStatusChanged)
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
		it.Event = new(ContractPauserRegistryPauserStatusChanged)
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
func (it *ContractPauserRegistryPauserStatusChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractPauserRegistryPauserStatusChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractPauserRegistryPauserStatusChanged represents a PauserStatusChanged event raised by the ContractPauserRegistry contract.
type ContractPauserRegistryPauserStatusChanged struct {
	Pauser   common.Address
	CanPause bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterPauserStatusChanged is a free log retrieval operation binding the contract event 0x65d3a1fd4c13f05cba164f80d03ce90fb4b5e21946bfc3ab7dbd434c2d0b9152.
//
// Solidity: event PauserStatusChanged(address pauser, bool canPause)
func (_ContractPauserRegistry *ContractPauserRegistryFilterer) FilterPauserStatusChanged(opts *bind.FilterOpts) (*ContractPauserRegistryPauserStatusChangedIterator, error) {

	logs, sub, err := _ContractPauserRegistry.contract.FilterLogs(opts, "PauserStatusChanged")
	if err != nil {
		return nil, err
	}
	return &ContractPauserRegistryPauserStatusChangedIterator{contract: _ContractPauserRegistry.contract, event: "PauserStatusChanged", logs: logs, sub: sub}, nil
}

// WatchPauserStatusChanged is a free log subscription operation binding the contract event 0x65d3a1fd4c13f05cba164f80d03ce90fb4b5e21946bfc3ab7dbd434c2d0b9152.
//
// Solidity: event PauserStatusChanged(address pauser, bool canPause)
func (_ContractPauserRegistry *ContractPauserRegistryFilterer) WatchPauserStatusChanged(opts *bind.WatchOpts, sink chan<- *ContractPauserRegistryPauserStatusChanged) (event.Subscription, error) {

	logs, sub, err := _ContractPauserRegistry.contract.WatchLogs(opts, "PauserStatusChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractPauserRegistryPauserStatusChanged)
				if err := _ContractPauserRegistry.contract.UnpackLog(event, "PauserStatusChanged", log); err != nil {
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

// ParsePauserStatusChanged is a log parse operation binding the contract event 0x65d3a1fd4c13f05cba164f80d03ce90fb4b5e21946bfc3ab7dbd434c2d0b9152.
//
// Solidity: event PauserStatusChanged(address pauser, bool canPause)
func (_ContractPauserRegistry *ContractPauserRegistryFilterer) ParsePauserStatusChanged(log types.Log) (*ContractPauserRegistryPauserStatusChanged, error) {
	event := new(ContractPauserRegistryPauserStatusChanged)
	if err := _ContractPauserRegistry.contract.UnpackLog(event, "PauserStatusChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractPauserRegistryUnpauserChangedIterator is returned from FilterUnpauserChanged and is used to iterate over the raw logs and unpacked data for UnpauserChanged events raised by the ContractPauserRegistry contract.
type ContractPauserRegistryUnpauserChangedIterator struct {
	Event *ContractPauserRegistryUnpauserChanged // Event containing the contract specifics and raw log

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
func (it *ContractPauserRegistryUnpauserChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractPauserRegistryUnpauserChanged)
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
		it.Event = new(ContractPauserRegistryUnpauserChanged)
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
func (it *ContractPauserRegistryUnpauserChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractPauserRegistryUnpauserChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractPauserRegistryUnpauserChanged represents a UnpauserChanged event raised by the ContractPauserRegistry contract.
type ContractPauserRegistryUnpauserChanged struct {
	PreviousUnpauser common.Address
	NewUnpauser      common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterUnpauserChanged is a free log retrieval operation binding the contract event 0x06b4167a2528887a1e97a366eefe8549bfbf1ea3e6ac81cb2564a934d20e8892.
//
// Solidity: event UnpauserChanged(address previousUnpauser, address newUnpauser)
func (_ContractPauserRegistry *ContractPauserRegistryFilterer) FilterUnpauserChanged(opts *bind.FilterOpts) (*ContractPauserRegistryUnpauserChangedIterator, error) {

	logs, sub, err := _ContractPauserRegistry.contract.FilterLogs(opts, "UnpauserChanged")
	if err != nil {
		return nil, err
	}
	return &ContractPauserRegistryUnpauserChangedIterator{contract: _ContractPauserRegistry.contract, event: "UnpauserChanged", logs: logs, sub: sub}, nil
}

// WatchUnpauserChanged is a free log subscription operation binding the contract event 0x06b4167a2528887a1e97a366eefe8549bfbf1ea3e6ac81cb2564a934d20e8892.
//
// Solidity: event UnpauserChanged(address previousUnpauser, address newUnpauser)
func (_ContractPauserRegistry *ContractPauserRegistryFilterer) WatchUnpauserChanged(opts *bind.WatchOpts, sink chan<- *ContractPauserRegistryUnpauserChanged) (event.Subscription, error) {

	logs, sub, err := _ContractPauserRegistry.contract.WatchLogs(opts, "UnpauserChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractPauserRegistryUnpauserChanged)
				if err := _ContractPauserRegistry.contract.UnpackLog(event, "UnpauserChanged", log); err != nil {
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

// ParseUnpauserChanged is a log parse operation binding the contract event 0x06b4167a2528887a1e97a366eefe8549bfbf1ea3e6ac81cb2564a934d20e8892.
//
// Solidity: event UnpauserChanged(address previousUnpauser, address newUnpauser)
func (_ContractPauserRegistry *ContractPauserRegistryFilterer) ParseUnpauserChanged(log types.Log) (*ContractPauserRegistryUnpauserChanged, error) {
	event := new(ContractPauserRegistryUnpauserChanged)
	if err := _ContractPauserRegistry.contract.UnpackLog(event, "UnpauserChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
