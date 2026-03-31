// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

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

// BatchAuthenticatorMetaData contains all meta data concerning the BatchAuthenticator contract.
var BatchAuthenticatorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"GUARDIAN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"activeIsTee\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"guardian\",\"type\":\"address\"}],\"name\":\"addGuardian\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_signature\",\"type\":\"bytes\"}],\"name\":\"authenticateBatchInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"espressoTEEVerifier\",\"outputs\":[{\"internalType\":\"contractIEspressoTEEVerifier\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getGuardians\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getRoleMember\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleMemberCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleMembers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"guardianCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initVersion\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIEspressoTEEVerifier\",\"name\":\"_espressoTEEVerifier\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_teeBatcher\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_systemConfig\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"isGuardian\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nitroValidator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxyAdmin\",\"outputs\":[{\"internalType\":\"contractIProxyAdmin\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxyAdminOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"attestationTbs\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"registerSigner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"guardian\",\"type\":\"address\"}],\"name\":\"removeGuardian\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"callerConfirmation\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_newTeeBatcher\",\"type\":\"address\"}],\"name\":\"setTeeBatcher\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"switchBatcher\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"systemConfig\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"teeBatcher\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"}],\"name\":\"BatchInfoAuthenticated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"activeIsTee\",\"type\":\"bool\"}],\"name\":\"BatcherSwitched\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"guardian\",\"type\":\"address\"}],\"name\":\"GuardianAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"guardian\",\"type\":\"address\"}],\"name\":\"GuardianRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"SignerRegistrationInitiated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldTeeBatcher\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newTeeBatcher\",\"type\":\"address\"}],\"name\":\"TeeBatcherUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"AccessControlBadConfirmation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"neededRole\",\"type\":\"bytes32\"}],\"name\":\"AccessControlUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"}],\"name\":\"InvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidGuardianAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"NotGuardian\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"NotGuardianOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotProxyAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotProxyAdminOrProxyAdminOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotProxyAdminOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotResolvedDelegateProxy\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotSharedProxyAdminOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_ProxyAdminNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ReinitializableBase_ZeroInitVersion\",\"type\":\"error\"}]",
}

// BatchAuthenticatorABI is the input ABI used to generate the binding from.
// Deprecated: Use BatchAuthenticatorMetaData.ABI instead.
var BatchAuthenticatorABI = BatchAuthenticatorMetaData.ABI

// BatchAuthenticator is an auto generated Go binding around an Ethereum contract.
type BatchAuthenticator struct {
	BatchAuthenticatorCaller     // Read-only binding to the contract
	BatchAuthenticatorTransactor // Write-only binding to the contract
	BatchAuthenticatorFilterer   // Log filterer for contract events
}

// BatchAuthenticatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type BatchAuthenticatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchAuthenticatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BatchAuthenticatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchAuthenticatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BatchAuthenticatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BatchAuthenticatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BatchAuthenticatorSession struct {
	Contract     *BatchAuthenticator // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// BatchAuthenticatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BatchAuthenticatorCallerSession struct {
	Contract *BatchAuthenticatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// BatchAuthenticatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BatchAuthenticatorTransactorSession struct {
	Contract     *BatchAuthenticatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// BatchAuthenticatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type BatchAuthenticatorRaw struct {
	Contract *BatchAuthenticator // Generic contract binding to access the raw methods on
}

// BatchAuthenticatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BatchAuthenticatorCallerRaw struct {
	Contract *BatchAuthenticatorCaller // Generic read-only contract binding to access the raw methods on
}

// BatchAuthenticatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BatchAuthenticatorTransactorRaw struct {
	Contract *BatchAuthenticatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBatchAuthenticator creates a new instance of BatchAuthenticator, bound to a specific deployed contract.
func NewBatchAuthenticator(address common.Address, backend bind.ContractBackend) (*BatchAuthenticator, error) {
	contract, err := bindBatchAuthenticator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticator{BatchAuthenticatorCaller: BatchAuthenticatorCaller{contract: contract}, BatchAuthenticatorTransactor: BatchAuthenticatorTransactor{contract: contract}, BatchAuthenticatorFilterer: BatchAuthenticatorFilterer{contract: contract}}, nil
}

// NewBatchAuthenticatorCaller creates a new read-only instance of BatchAuthenticator, bound to a specific deployed contract.
func NewBatchAuthenticatorCaller(address common.Address, caller bind.ContractCaller) (*BatchAuthenticatorCaller, error) {
	contract, err := bindBatchAuthenticator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorCaller{contract: contract}, nil
}

// NewBatchAuthenticatorTransactor creates a new write-only instance of BatchAuthenticator, bound to a specific deployed contract.
func NewBatchAuthenticatorTransactor(address common.Address, transactor bind.ContractTransactor) (*BatchAuthenticatorTransactor, error) {
	contract, err := bindBatchAuthenticator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorTransactor{contract: contract}, nil
}

// NewBatchAuthenticatorFilterer creates a new log filterer instance of BatchAuthenticator, bound to a specific deployed contract.
func NewBatchAuthenticatorFilterer(address common.Address, filterer bind.ContractFilterer) (*BatchAuthenticatorFilterer, error) {
	contract, err := bindBatchAuthenticator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorFilterer{contract: contract}, nil
}

// bindBatchAuthenticator binds a generic wrapper to an already deployed contract.
func bindBatchAuthenticator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BatchAuthenticatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BatchAuthenticator *BatchAuthenticatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BatchAuthenticator.Contract.BatchAuthenticatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BatchAuthenticator *BatchAuthenticatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.BatchAuthenticatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BatchAuthenticator *BatchAuthenticatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.BatchAuthenticatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BatchAuthenticator *BatchAuthenticatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BatchAuthenticator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BatchAuthenticator *BatchAuthenticatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BatchAuthenticator *BatchAuthenticatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BatchAuthenticator.Contract.DEFAULTADMINROLE(&_BatchAuthenticator.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BatchAuthenticator.Contract.DEFAULTADMINROLE(&_BatchAuthenticator.CallOpts)
}

// GUARDIANROLE is a free data retrieval call binding the contract method 0x24ea54f4.
//
// Solidity: function GUARDIAN_ROLE() view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorCaller) GUARDIANROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "GUARDIAN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GUARDIANROLE is a free data retrieval call binding the contract method 0x24ea54f4.
//
// Solidity: function GUARDIAN_ROLE() view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorSession) GUARDIANROLE() ([32]byte, error) {
	return _BatchAuthenticator.Contract.GUARDIANROLE(&_BatchAuthenticator.CallOpts)
}

// GUARDIANROLE is a free data retrieval call binding the contract method 0x24ea54f4.
//
// Solidity: function GUARDIAN_ROLE() view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) GUARDIANROLE() ([32]byte, error) {
	return _BatchAuthenticator.Contract.GUARDIANROLE(&_BatchAuthenticator.CallOpts)
}

// ActiveIsTee is a free data retrieval call binding the contract method 0x7877a9ed.
//
// Solidity: function activeIsTee() view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorCaller) ActiveIsTee(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "activeIsTee")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ActiveIsTee is a free data retrieval call binding the contract method 0x7877a9ed.
//
// Solidity: function activeIsTee() view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorSession) ActiveIsTee() (bool, error) {
	return _BatchAuthenticator.Contract.ActiveIsTee(&_BatchAuthenticator.CallOpts)
}

// ActiveIsTee is a free data retrieval call binding the contract method 0x7877a9ed.
//
// Solidity: function activeIsTee() view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) ActiveIsTee() (bool, error) {
	return _BatchAuthenticator.Contract.ActiveIsTee(&_BatchAuthenticator.CallOpts)
}

// EspressoTEEVerifier is a free data retrieval call binding the contract method 0xfa14fe6d.
//
// Solidity: function espressoTEEVerifier() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) EspressoTEEVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "espressoTEEVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EspressoTEEVerifier is a free data retrieval call binding the contract method 0xfa14fe6d.
//
// Solidity: function espressoTEEVerifier() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) EspressoTEEVerifier() (common.Address, error) {
	return _BatchAuthenticator.Contract.EspressoTEEVerifier(&_BatchAuthenticator.CallOpts)
}

// EspressoTEEVerifier is a free data retrieval call binding the contract method 0xfa14fe6d.
//
// Solidity: function espressoTEEVerifier() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) EspressoTEEVerifier() (common.Address, error) {
	return _BatchAuthenticator.Contract.EspressoTEEVerifier(&_BatchAuthenticator.CallOpts)
}

// GetGuardians is a free data retrieval call binding the contract method 0x0665f04b.
//
// Solidity: function getGuardians() view returns(address[])
func (_BatchAuthenticator *BatchAuthenticatorCaller) GetGuardians(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "getGuardians")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetGuardians is a free data retrieval call binding the contract method 0x0665f04b.
//
// Solidity: function getGuardians() view returns(address[])
func (_BatchAuthenticator *BatchAuthenticatorSession) GetGuardians() ([]common.Address, error) {
	return _BatchAuthenticator.Contract.GetGuardians(&_BatchAuthenticator.CallOpts)
}

// GetGuardians is a free data retrieval call binding the contract method 0x0665f04b.
//
// Solidity: function getGuardians() view returns(address[])
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) GetGuardians() ([]common.Address, error) {
	return _BatchAuthenticator.Contract.GetGuardians(&_BatchAuthenticator.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BatchAuthenticator.Contract.GetRoleAdmin(&_BatchAuthenticator.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BatchAuthenticator.Contract.GetRoleAdmin(&_BatchAuthenticator.CallOpts, role)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) GetRoleMember(opts *bind.CallOpts, role [32]byte, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "getRoleMember", role, index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _BatchAuthenticator.Contract.GetRoleMember(&_BatchAuthenticator.CallOpts, role, index)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _BatchAuthenticator.Contract.GetRoleMember(&_BatchAuthenticator.CallOpts, role, index)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_BatchAuthenticator *BatchAuthenticatorCaller) GetRoleMemberCount(opts *bind.CallOpts, role [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "getRoleMemberCount", role)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_BatchAuthenticator *BatchAuthenticatorSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _BatchAuthenticator.Contract.GetRoleMemberCount(&_BatchAuthenticator.CallOpts, role)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _BatchAuthenticator.Contract.GetRoleMemberCount(&_BatchAuthenticator.CallOpts, role)
}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_BatchAuthenticator *BatchAuthenticatorCaller) GetRoleMembers(opts *bind.CallOpts, role [32]byte) ([]common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "getRoleMembers", role)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_BatchAuthenticator *BatchAuthenticatorSession) GetRoleMembers(role [32]byte) ([]common.Address, error) {
	return _BatchAuthenticator.Contract.GetRoleMembers(&_BatchAuthenticator.CallOpts, role)
}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) GetRoleMembers(role [32]byte) ([]common.Address, error) {
	return _BatchAuthenticator.Contract.GetRoleMembers(&_BatchAuthenticator.CallOpts, role)
}

// GuardianCount is a free data retrieval call binding the contract method 0x54387ad7.
//
// Solidity: function guardianCount() view returns(uint256)
func (_BatchAuthenticator *BatchAuthenticatorCaller) GuardianCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "guardianCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GuardianCount is a free data retrieval call binding the contract method 0x54387ad7.
//
// Solidity: function guardianCount() view returns(uint256)
func (_BatchAuthenticator *BatchAuthenticatorSession) GuardianCount() (*big.Int, error) {
	return _BatchAuthenticator.Contract.GuardianCount(&_BatchAuthenticator.CallOpts)
}

// GuardianCount is a free data retrieval call binding the contract method 0x54387ad7.
//
// Solidity: function guardianCount() view returns(uint256)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) GuardianCount() (*big.Int, error) {
	return _BatchAuthenticator.Contract.GuardianCount(&_BatchAuthenticator.CallOpts)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BatchAuthenticator.Contract.HasRole(&_BatchAuthenticator.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BatchAuthenticator.Contract.HasRole(&_BatchAuthenticator.CallOpts, role, account)
}

// InitVersion is a free data retrieval call binding the contract method 0x38d38c97.
//
// Solidity: function initVersion() view returns(uint8)
func (_BatchAuthenticator *BatchAuthenticatorCaller) InitVersion(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "initVersion")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// InitVersion is a free data retrieval call binding the contract method 0x38d38c97.
//
// Solidity: function initVersion() view returns(uint8)
func (_BatchAuthenticator *BatchAuthenticatorSession) InitVersion() (uint8, error) {
	return _BatchAuthenticator.Contract.InitVersion(&_BatchAuthenticator.CallOpts)
}

// InitVersion is a free data retrieval call binding the contract method 0x38d38c97.
//
// Solidity: function initVersion() view returns(uint8)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) InitVersion() (uint8, error) {
	return _BatchAuthenticator.Contract.InitVersion(&_BatchAuthenticator.CallOpts)
}

// IsGuardian is a free data retrieval call binding the contract method 0x0c68ba21.
//
// Solidity: function isGuardian(address account) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorCaller) IsGuardian(opts *bind.CallOpts, account common.Address) (bool, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "isGuardian", account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGuardian is a free data retrieval call binding the contract method 0x0c68ba21.
//
// Solidity: function isGuardian(address account) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorSession) IsGuardian(account common.Address) (bool, error) {
	return _BatchAuthenticator.Contract.IsGuardian(&_BatchAuthenticator.CallOpts, account)
}

// IsGuardian is a free data retrieval call binding the contract method 0x0c68ba21.
//
// Solidity: function isGuardian(address account) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) IsGuardian(account common.Address) (bool, error) {
	return _BatchAuthenticator.Contract.IsGuardian(&_BatchAuthenticator.CallOpts, account)
}

// NitroValidator is a free data retrieval call binding the contract method 0x1b076a4c.
//
// Solidity: function nitroValidator() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) NitroValidator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "nitroValidator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NitroValidator is a free data retrieval call binding the contract method 0x1b076a4c.
//
// Solidity: function nitroValidator() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) NitroValidator() (common.Address, error) {
	return _BatchAuthenticator.Contract.NitroValidator(&_BatchAuthenticator.CallOpts)
}

// NitroValidator is a free data retrieval call binding the contract method 0x1b076a4c.
//
// Solidity: function nitroValidator() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) NitroValidator() (common.Address, error) {
	return _BatchAuthenticator.Contract.NitroValidator(&_BatchAuthenticator.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) Owner() (common.Address, error) {
	return _BatchAuthenticator.Contract.Owner(&_BatchAuthenticator.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) Owner() (common.Address, error) {
	return _BatchAuthenticator.Contract.Owner(&_BatchAuthenticator.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) PendingOwner() (common.Address, error) {
	return _BatchAuthenticator.Contract.PendingOwner(&_BatchAuthenticator.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) PendingOwner() (common.Address, error) {
	return _BatchAuthenticator.Contract.PendingOwner(&_BatchAuthenticator.CallOpts)
}

// ProxyAdmin is a free data retrieval call binding the contract method 0x3e47158c.
//
// Solidity: function proxyAdmin() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) ProxyAdmin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "proxyAdmin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ProxyAdmin is a free data retrieval call binding the contract method 0x3e47158c.
//
// Solidity: function proxyAdmin() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) ProxyAdmin() (common.Address, error) {
	return _BatchAuthenticator.Contract.ProxyAdmin(&_BatchAuthenticator.CallOpts)
}

// ProxyAdmin is a free data retrieval call binding the contract method 0x3e47158c.
//
// Solidity: function proxyAdmin() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) ProxyAdmin() (common.Address, error) {
	return _BatchAuthenticator.Contract.ProxyAdmin(&_BatchAuthenticator.CallOpts)
}

// ProxyAdminOwner is a free data retrieval call binding the contract method 0xdad544e0.
//
// Solidity: function proxyAdminOwner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) ProxyAdminOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "proxyAdminOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ProxyAdminOwner is a free data retrieval call binding the contract method 0xdad544e0.
//
// Solidity: function proxyAdminOwner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) ProxyAdminOwner() (common.Address, error) {
	return _BatchAuthenticator.Contract.ProxyAdminOwner(&_BatchAuthenticator.CallOpts)
}

// ProxyAdminOwner is a free data retrieval call binding the contract method 0xdad544e0.
//
// Solidity: function proxyAdminOwner() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) ProxyAdminOwner() (common.Address, error) {
	return _BatchAuthenticator.Contract.ProxyAdminOwner(&_BatchAuthenticator.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BatchAuthenticator.Contract.SupportsInterface(&_BatchAuthenticator.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BatchAuthenticator.Contract.SupportsInterface(&_BatchAuthenticator.CallOpts, interfaceId)
}

// SystemConfig is a free data retrieval call binding the contract method 0x33d7e2bd.
//
// Solidity: function systemConfig() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) SystemConfig(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "systemConfig")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SystemConfig is a free data retrieval call binding the contract method 0x33d7e2bd.
//
// Solidity: function systemConfig() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) SystemConfig() (common.Address, error) {
	return _BatchAuthenticator.Contract.SystemConfig(&_BatchAuthenticator.CallOpts)
}

// SystemConfig is a free data retrieval call binding the contract method 0x33d7e2bd.
//
// Solidity: function systemConfig() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) SystemConfig() (common.Address, error) {
	return _BatchAuthenticator.Contract.SystemConfig(&_BatchAuthenticator.CallOpts)
}

// TeeBatcher is a free data retrieval call binding the contract method 0xd909ba7c.
//
// Solidity: function teeBatcher() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCaller) TeeBatcher(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "teeBatcher")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TeeBatcher is a free data retrieval call binding the contract method 0xd909ba7c.
//
// Solidity: function teeBatcher() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorSession) TeeBatcher() (common.Address, error) {
	return _BatchAuthenticator.Contract.TeeBatcher(&_BatchAuthenticator.CallOpts)
}

// TeeBatcher is a free data retrieval call binding the contract method 0xd909ba7c.
//
// Solidity: function teeBatcher() view returns(address)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) TeeBatcher() (common.Address, error) {
	return _BatchAuthenticator.Contract.TeeBatcher(&_BatchAuthenticator.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_BatchAuthenticator *BatchAuthenticatorCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BatchAuthenticator.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_BatchAuthenticator *BatchAuthenticatorSession) Version() (string, error) {
	return _BatchAuthenticator.Contract.Version(&_BatchAuthenticator.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_BatchAuthenticator *BatchAuthenticatorCallerSession) Version() (string, error) {
	return _BatchAuthenticator.Contract.Version(&_BatchAuthenticator.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.AcceptOwnership(&_BatchAuthenticator.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.AcceptOwnership(&_BatchAuthenticator.TransactOpts)
}

// AddGuardian is a paid mutator transaction binding the contract method 0xa526d83b.
//
// Solidity: function addGuardian(address guardian) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) AddGuardian(opts *bind.TransactOpts, guardian common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "addGuardian", guardian)
}

// AddGuardian is a paid mutator transaction binding the contract method 0xa526d83b.
//
// Solidity: function addGuardian(address guardian) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) AddGuardian(guardian common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.AddGuardian(&_BatchAuthenticator.TransactOpts, guardian)
}

// AddGuardian is a paid mutator transaction binding the contract method 0xa526d83b.
//
// Solidity: function addGuardian(address guardian) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) AddGuardian(guardian common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.AddGuardian(&_BatchAuthenticator.TransactOpts, guardian)
}

// AuthenticateBatchInfo is a paid mutator transaction binding the contract method 0xfc619e41.
//
// Solidity: function authenticateBatchInfo(bytes32 commitment, bytes _signature) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) AuthenticateBatchInfo(opts *bind.TransactOpts, commitment [32]byte, _signature []byte) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "authenticateBatchInfo", commitment, _signature)
}

// AuthenticateBatchInfo is a paid mutator transaction binding the contract method 0xfc619e41.
//
// Solidity: function authenticateBatchInfo(bytes32 commitment, bytes _signature) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) AuthenticateBatchInfo(commitment [32]byte, _signature []byte) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.AuthenticateBatchInfo(&_BatchAuthenticator.TransactOpts, commitment, _signature)
}

// AuthenticateBatchInfo is a paid mutator transaction binding the contract method 0xfc619e41.
//
// Solidity: function authenticateBatchInfo(bytes32 commitment, bytes _signature) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) AuthenticateBatchInfo(commitment [32]byte, _signature []byte) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.AuthenticateBatchInfo(&_BatchAuthenticator.TransactOpts, commitment, _signature)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.GrantRole(&_BatchAuthenticator.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.GrantRole(&_BatchAuthenticator.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0xf8c8765e.
//
// Solidity: function initialize(address _espressoTEEVerifier, address _teeBatcher, address _systemConfig, address _owner) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) Initialize(opts *bind.TransactOpts, _espressoTEEVerifier common.Address, _teeBatcher common.Address, _systemConfig common.Address, _owner common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "initialize", _espressoTEEVerifier, _teeBatcher, _systemConfig, _owner)
}

// Initialize is a paid mutator transaction binding the contract method 0xf8c8765e.
//
// Solidity: function initialize(address _espressoTEEVerifier, address _teeBatcher, address _systemConfig, address _owner) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) Initialize(_espressoTEEVerifier common.Address, _teeBatcher common.Address, _systemConfig common.Address, _owner common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.Initialize(&_BatchAuthenticator.TransactOpts, _espressoTEEVerifier, _teeBatcher, _systemConfig, _owner)
}

// Initialize is a paid mutator transaction binding the contract method 0xf8c8765e.
//
// Solidity: function initialize(address _espressoTEEVerifier, address _teeBatcher, address _systemConfig, address _owner) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) Initialize(_espressoTEEVerifier common.Address, _teeBatcher common.Address, _systemConfig common.Address, _owner common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.Initialize(&_BatchAuthenticator.TransactOpts, _espressoTEEVerifier, _teeBatcher, _systemConfig, _owner)
}

// RegisterSigner is a paid mutator transaction binding the contract method 0xba58e82a.
//
// Solidity: function registerSigner(bytes attestationTbs, bytes signature) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) RegisterSigner(opts *bind.TransactOpts, attestationTbs []byte, signature []byte) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "registerSigner", attestationTbs, signature)
}

// RegisterSigner is a paid mutator transaction binding the contract method 0xba58e82a.
//
// Solidity: function registerSigner(bytes attestationTbs, bytes signature) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) RegisterSigner(attestationTbs []byte, signature []byte) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RegisterSigner(&_BatchAuthenticator.TransactOpts, attestationTbs, signature)
}

// RegisterSigner is a paid mutator transaction binding the contract method 0xba58e82a.
//
// Solidity: function registerSigner(bytes attestationTbs, bytes signature) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) RegisterSigner(attestationTbs []byte, signature []byte) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RegisterSigner(&_BatchAuthenticator.TransactOpts, attestationTbs, signature)
}

// RemoveGuardian is a paid mutator transaction binding the contract method 0x71404156.
//
// Solidity: function removeGuardian(address guardian) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) RemoveGuardian(opts *bind.TransactOpts, guardian common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "removeGuardian", guardian)
}

// RemoveGuardian is a paid mutator transaction binding the contract method 0x71404156.
//
// Solidity: function removeGuardian(address guardian) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) RemoveGuardian(guardian common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RemoveGuardian(&_BatchAuthenticator.TransactOpts, guardian)
}

// RemoveGuardian is a paid mutator transaction binding the contract method 0x71404156.
//
// Solidity: function removeGuardian(address guardian) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) RemoveGuardian(guardian common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RemoveGuardian(&_BatchAuthenticator.TransactOpts, guardian)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RenounceOwnership(&_BatchAuthenticator.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RenounceOwnership(&_BatchAuthenticator.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RenounceRole(&_BatchAuthenticator.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RenounceRole(&_BatchAuthenticator.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RevokeRole(&_BatchAuthenticator.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.RevokeRole(&_BatchAuthenticator.TransactOpts, role, account)
}

// SetTeeBatcher is a paid mutator transaction binding the contract method 0x6f7eda47.
//
// Solidity: function setTeeBatcher(address _newTeeBatcher) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) SetTeeBatcher(opts *bind.TransactOpts, _newTeeBatcher common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "setTeeBatcher", _newTeeBatcher)
}

// SetTeeBatcher is a paid mutator transaction binding the contract method 0x6f7eda47.
//
// Solidity: function setTeeBatcher(address _newTeeBatcher) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) SetTeeBatcher(_newTeeBatcher common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.SetTeeBatcher(&_BatchAuthenticator.TransactOpts, _newTeeBatcher)
}

// SetTeeBatcher is a paid mutator transaction binding the contract method 0x6f7eda47.
//
// Solidity: function setTeeBatcher(address _newTeeBatcher) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) SetTeeBatcher(_newTeeBatcher common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.SetTeeBatcher(&_BatchAuthenticator.TransactOpts, _newTeeBatcher)
}

// SwitchBatcher is a paid mutator transaction binding the contract method 0xbc347f47.
//
// Solidity: function switchBatcher() returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) SwitchBatcher(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "switchBatcher")
}

// SwitchBatcher is a paid mutator transaction binding the contract method 0xbc347f47.
//
// Solidity: function switchBatcher() returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) SwitchBatcher() (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.SwitchBatcher(&_BatchAuthenticator.TransactOpts)
}

// SwitchBatcher is a paid mutator transaction binding the contract method 0xbc347f47.
//
// Solidity: function switchBatcher() returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) SwitchBatcher() (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.SwitchBatcher(&_BatchAuthenticator.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchAuthenticator *BatchAuthenticatorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.TransferOwnership(&_BatchAuthenticator.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BatchAuthenticator *BatchAuthenticatorTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BatchAuthenticator.Contract.TransferOwnership(&_BatchAuthenticator.TransactOpts, newOwner)
}

// BatchAuthenticatorBatchInfoAuthenticatedIterator is returned from FilterBatchInfoAuthenticated and is used to iterate over the raw logs and unpacked data for BatchInfoAuthenticated events raised by the BatchAuthenticator contract.
type BatchAuthenticatorBatchInfoAuthenticatedIterator struct {
	Event *BatchAuthenticatorBatchInfoAuthenticated // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorBatchInfoAuthenticatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorBatchInfoAuthenticated)
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
		it.Event = new(BatchAuthenticatorBatchInfoAuthenticated)
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
func (it *BatchAuthenticatorBatchInfoAuthenticatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorBatchInfoAuthenticatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorBatchInfoAuthenticated represents a BatchInfoAuthenticated event raised by the BatchAuthenticator contract.
type BatchAuthenticatorBatchInfoAuthenticated struct {
	Commitment [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterBatchInfoAuthenticated is a free log retrieval operation binding the contract event 0xee0d07d204d979d28885955e59a46f754c4db7378b7df1a95123525aac6e3f80.
//
// Solidity: event BatchInfoAuthenticated(bytes32 indexed commitment)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterBatchInfoAuthenticated(opts *bind.FilterOpts, commitment [][32]byte) (*BatchAuthenticatorBatchInfoAuthenticatedIterator, error) {

	var commitmentRule []interface{}
	for _, commitmentItem := range commitment {
		commitmentRule = append(commitmentRule, commitmentItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "BatchInfoAuthenticated", commitmentRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorBatchInfoAuthenticatedIterator{contract: _BatchAuthenticator.contract, event: "BatchInfoAuthenticated", logs: logs, sub: sub}, nil
}

// WatchBatchInfoAuthenticated is a free log subscription operation binding the contract event 0xee0d07d204d979d28885955e59a46f754c4db7378b7df1a95123525aac6e3f80.
//
// Solidity: event BatchInfoAuthenticated(bytes32 indexed commitment)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchBatchInfoAuthenticated(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorBatchInfoAuthenticated, commitment [][32]byte) (event.Subscription, error) {

	var commitmentRule []interface{}
	for _, commitmentItem := range commitment {
		commitmentRule = append(commitmentRule, commitmentItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "BatchInfoAuthenticated", commitmentRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorBatchInfoAuthenticated)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "BatchInfoAuthenticated", log); err != nil {
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

// ParseBatchInfoAuthenticated is a log parse operation binding the contract event 0xee0d07d204d979d28885955e59a46f754c4db7378b7df1a95123525aac6e3f80.
//
// Solidity: event BatchInfoAuthenticated(bytes32 indexed commitment)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseBatchInfoAuthenticated(log types.Log) (*BatchAuthenticatorBatchInfoAuthenticated, error) {
	event := new(BatchAuthenticatorBatchInfoAuthenticated)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "BatchInfoAuthenticated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorBatcherSwitchedIterator is returned from FilterBatcherSwitched and is used to iterate over the raw logs and unpacked data for BatcherSwitched events raised by the BatchAuthenticator contract.
type BatchAuthenticatorBatcherSwitchedIterator struct {
	Event *BatchAuthenticatorBatcherSwitched // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorBatcherSwitchedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorBatcherSwitched)
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
		it.Event = new(BatchAuthenticatorBatcherSwitched)
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
func (it *BatchAuthenticatorBatcherSwitchedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorBatcherSwitchedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorBatcherSwitched represents a BatcherSwitched event raised by the BatchAuthenticator contract.
type BatchAuthenticatorBatcherSwitched struct {
	ActiveIsTee bool
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterBatcherSwitched is a free log retrieval operation binding the contract event 0xb957d7fc29e5974594db2f2e132076d52f42c0734eae05fd5ea080d1ba175ad3.
//
// Solidity: event BatcherSwitched(bool indexed activeIsTee)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterBatcherSwitched(opts *bind.FilterOpts, activeIsTee []bool) (*BatchAuthenticatorBatcherSwitchedIterator, error) {

	var activeIsTeeRule []interface{}
	for _, activeIsTeeItem := range activeIsTee {
		activeIsTeeRule = append(activeIsTeeRule, activeIsTeeItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "BatcherSwitched", activeIsTeeRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorBatcherSwitchedIterator{contract: _BatchAuthenticator.contract, event: "BatcherSwitched", logs: logs, sub: sub}, nil
}

// WatchBatcherSwitched is a free log subscription operation binding the contract event 0xb957d7fc29e5974594db2f2e132076d52f42c0734eae05fd5ea080d1ba175ad3.
//
// Solidity: event BatcherSwitched(bool indexed activeIsTee)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchBatcherSwitched(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorBatcherSwitched, activeIsTee []bool) (event.Subscription, error) {

	var activeIsTeeRule []interface{}
	for _, activeIsTeeItem := range activeIsTee {
		activeIsTeeRule = append(activeIsTeeRule, activeIsTeeItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "BatcherSwitched", activeIsTeeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorBatcherSwitched)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "BatcherSwitched", log); err != nil {
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

// ParseBatcherSwitched is a log parse operation binding the contract event 0xb957d7fc29e5974594db2f2e132076d52f42c0734eae05fd5ea080d1ba175ad3.
//
// Solidity: event BatcherSwitched(bool indexed activeIsTee)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseBatcherSwitched(log types.Log) (*BatchAuthenticatorBatcherSwitched, error) {
	event := new(BatchAuthenticatorBatcherSwitched)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "BatcherSwitched", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorGuardianAddedIterator is returned from FilterGuardianAdded and is used to iterate over the raw logs and unpacked data for GuardianAdded events raised by the BatchAuthenticator contract.
type BatchAuthenticatorGuardianAddedIterator struct {
	Event *BatchAuthenticatorGuardianAdded // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorGuardianAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorGuardianAdded)
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
		it.Event = new(BatchAuthenticatorGuardianAdded)
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
func (it *BatchAuthenticatorGuardianAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorGuardianAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorGuardianAdded represents a GuardianAdded event raised by the BatchAuthenticator contract.
type BatchAuthenticatorGuardianAdded struct {
	Guardian common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterGuardianAdded is a free log retrieval operation binding the contract event 0x038596bb31e2e7d3d9f184d4c98b310103f6d7f5830e5eec32bffe6f1728f969.
//
// Solidity: event GuardianAdded(address indexed guardian)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterGuardianAdded(opts *bind.FilterOpts, guardian []common.Address) (*BatchAuthenticatorGuardianAddedIterator, error) {

	var guardianRule []interface{}
	for _, guardianItem := range guardian {
		guardianRule = append(guardianRule, guardianItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "GuardianAdded", guardianRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorGuardianAddedIterator{contract: _BatchAuthenticator.contract, event: "GuardianAdded", logs: logs, sub: sub}, nil
}

// WatchGuardianAdded is a free log subscription operation binding the contract event 0x038596bb31e2e7d3d9f184d4c98b310103f6d7f5830e5eec32bffe6f1728f969.
//
// Solidity: event GuardianAdded(address indexed guardian)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchGuardianAdded(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorGuardianAdded, guardian []common.Address) (event.Subscription, error) {

	var guardianRule []interface{}
	for _, guardianItem := range guardian {
		guardianRule = append(guardianRule, guardianItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "GuardianAdded", guardianRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorGuardianAdded)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "GuardianAdded", log); err != nil {
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

// ParseGuardianAdded is a log parse operation binding the contract event 0x038596bb31e2e7d3d9f184d4c98b310103f6d7f5830e5eec32bffe6f1728f969.
//
// Solidity: event GuardianAdded(address indexed guardian)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseGuardianAdded(log types.Log) (*BatchAuthenticatorGuardianAdded, error) {
	event := new(BatchAuthenticatorGuardianAdded)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "GuardianAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorGuardianRemovedIterator is returned from FilterGuardianRemoved and is used to iterate over the raw logs and unpacked data for GuardianRemoved events raised by the BatchAuthenticator contract.
type BatchAuthenticatorGuardianRemovedIterator struct {
	Event *BatchAuthenticatorGuardianRemoved // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorGuardianRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorGuardianRemoved)
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
		it.Event = new(BatchAuthenticatorGuardianRemoved)
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
func (it *BatchAuthenticatorGuardianRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorGuardianRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorGuardianRemoved represents a GuardianRemoved event raised by the BatchAuthenticator contract.
type BatchAuthenticatorGuardianRemoved struct {
	Guardian common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterGuardianRemoved is a free log retrieval operation binding the contract event 0xb8107d0c6b40be480ce3172ee66ba6d64b71f6b1685a851340036e6e2e3e3c52.
//
// Solidity: event GuardianRemoved(address indexed guardian)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterGuardianRemoved(opts *bind.FilterOpts, guardian []common.Address) (*BatchAuthenticatorGuardianRemovedIterator, error) {

	var guardianRule []interface{}
	for _, guardianItem := range guardian {
		guardianRule = append(guardianRule, guardianItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "GuardianRemoved", guardianRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorGuardianRemovedIterator{contract: _BatchAuthenticator.contract, event: "GuardianRemoved", logs: logs, sub: sub}, nil
}

// WatchGuardianRemoved is a free log subscription operation binding the contract event 0xb8107d0c6b40be480ce3172ee66ba6d64b71f6b1685a851340036e6e2e3e3c52.
//
// Solidity: event GuardianRemoved(address indexed guardian)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchGuardianRemoved(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorGuardianRemoved, guardian []common.Address) (event.Subscription, error) {

	var guardianRule []interface{}
	for _, guardianItem := range guardian {
		guardianRule = append(guardianRule, guardianItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "GuardianRemoved", guardianRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorGuardianRemoved)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "GuardianRemoved", log); err != nil {
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

// ParseGuardianRemoved is a log parse operation binding the contract event 0xb8107d0c6b40be480ce3172ee66ba6d64b71f6b1685a851340036e6e2e3e3c52.
//
// Solidity: event GuardianRemoved(address indexed guardian)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseGuardianRemoved(log types.Log) (*BatchAuthenticatorGuardianRemoved, error) {
	event := new(BatchAuthenticatorGuardianRemoved)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "GuardianRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the BatchAuthenticator contract.
type BatchAuthenticatorInitializedIterator struct {
	Event *BatchAuthenticatorInitialized // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorInitialized)
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
		it.Event = new(BatchAuthenticatorInitialized)
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
func (it *BatchAuthenticatorInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorInitialized represents a Initialized event raised by the BatchAuthenticator contract.
type BatchAuthenticatorInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterInitialized(opts *bind.FilterOpts) (*BatchAuthenticatorInitializedIterator, error) {

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorInitializedIterator{contract: _BatchAuthenticator.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorInitialized) (event.Subscription, error) {

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorInitialized)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseInitialized(log types.Log) (*BatchAuthenticatorInitialized, error) {
	event := new(BatchAuthenticatorInitialized)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BatchAuthenticator contract.
type BatchAuthenticatorOwnershipTransferStartedIterator struct {
	Event *BatchAuthenticatorOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorOwnershipTransferStarted)
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
		it.Event = new(BatchAuthenticatorOwnershipTransferStarted)
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
func (it *BatchAuthenticatorOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BatchAuthenticator contract.
type BatchAuthenticatorOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BatchAuthenticatorOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorOwnershipTransferStartedIterator{contract: _BatchAuthenticator.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorOwnershipTransferStarted)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseOwnershipTransferStarted(log types.Log) (*BatchAuthenticatorOwnershipTransferStarted, error) {
	event := new(BatchAuthenticatorOwnershipTransferStarted)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BatchAuthenticator contract.
type BatchAuthenticatorOwnershipTransferredIterator struct {
	Event *BatchAuthenticatorOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorOwnershipTransferred)
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
		it.Event = new(BatchAuthenticatorOwnershipTransferred)
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
func (it *BatchAuthenticatorOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorOwnershipTransferred represents a OwnershipTransferred event raised by the BatchAuthenticator contract.
type BatchAuthenticatorOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BatchAuthenticatorOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorOwnershipTransferredIterator{contract: _BatchAuthenticator.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorOwnershipTransferred)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseOwnershipTransferred(log types.Log) (*BatchAuthenticatorOwnershipTransferred, error) {
	event := new(BatchAuthenticatorOwnershipTransferred)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the BatchAuthenticator contract.
type BatchAuthenticatorRoleAdminChangedIterator struct {
	Event *BatchAuthenticatorRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorRoleAdminChanged)
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
		it.Event = new(BatchAuthenticatorRoleAdminChanged)
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
func (it *BatchAuthenticatorRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorRoleAdminChanged represents a RoleAdminChanged event raised by the BatchAuthenticator contract.
type BatchAuthenticatorRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*BatchAuthenticatorRoleAdminChangedIterator, error) {

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

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorRoleAdminChangedIterator{contract: _BatchAuthenticator.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorRoleAdminChanged)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseRoleAdminChanged(log types.Log) (*BatchAuthenticatorRoleAdminChanged, error) {
	event := new(BatchAuthenticatorRoleAdminChanged)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the BatchAuthenticator contract.
type BatchAuthenticatorRoleGrantedIterator struct {
	Event *BatchAuthenticatorRoleGranted // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorRoleGranted)
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
		it.Event = new(BatchAuthenticatorRoleGranted)
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
func (it *BatchAuthenticatorRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorRoleGranted represents a RoleGranted event raised by the BatchAuthenticator contract.
type BatchAuthenticatorRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BatchAuthenticatorRoleGrantedIterator, error) {

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

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorRoleGrantedIterator{contract: _BatchAuthenticator.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorRoleGranted)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseRoleGranted(log types.Log) (*BatchAuthenticatorRoleGranted, error) {
	event := new(BatchAuthenticatorRoleGranted)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the BatchAuthenticator contract.
type BatchAuthenticatorRoleRevokedIterator struct {
	Event *BatchAuthenticatorRoleRevoked // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorRoleRevoked)
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
		it.Event = new(BatchAuthenticatorRoleRevoked)
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
func (it *BatchAuthenticatorRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorRoleRevoked represents a RoleRevoked event raised by the BatchAuthenticator contract.
type BatchAuthenticatorRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BatchAuthenticatorRoleRevokedIterator, error) {

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

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorRoleRevokedIterator{contract: _BatchAuthenticator.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorRoleRevoked)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseRoleRevoked(log types.Log) (*BatchAuthenticatorRoleRevoked, error) {
	event := new(BatchAuthenticatorRoleRevoked)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorSignerRegistrationInitiatedIterator is returned from FilterSignerRegistrationInitiated and is used to iterate over the raw logs and unpacked data for SignerRegistrationInitiated events raised by the BatchAuthenticator contract.
type BatchAuthenticatorSignerRegistrationInitiatedIterator struct {
	Event *BatchAuthenticatorSignerRegistrationInitiated // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorSignerRegistrationInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorSignerRegistrationInitiated)
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
		it.Event = new(BatchAuthenticatorSignerRegistrationInitiated)
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
func (it *BatchAuthenticatorSignerRegistrationInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorSignerRegistrationInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorSignerRegistrationInitiated represents a SignerRegistrationInitiated event raised by the BatchAuthenticator contract.
type BatchAuthenticatorSignerRegistrationInitiated struct {
	Caller common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSignerRegistrationInitiated is a free log retrieval operation binding the contract event 0x665b016a0ac50d1280744eaaff1cf21254d0fd30e4c3987d291913c32163416c.
//
// Solidity: event SignerRegistrationInitiated(address indexed caller)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterSignerRegistrationInitiated(opts *bind.FilterOpts, caller []common.Address) (*BatchAuthenticatorSignerRegistrationInitiatedIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "SignerRegistrationInitiated", callerRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorSignerRegistrationInitiatedIterator{contract: _BatchAuthenticator.contract, event: "SignerRegistrationInitiated", logs: logs, sub: sub}, nil
}

// WatchSignerRegistrationInitiated is a free log subscription operation binding the contract event 0x665b016a0ac50d1280744eaaff1cf21254d0fd30e4c3987d291913c32163416c.
//
// Solidity: event SignerRegistrationInitiated(address indexed caller)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchSignerRegistrationInitiated(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorSignerRegistrationInitiated, caller []common.Address) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "SignerRegistrationInitiated", callerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorSignerRegistrationInitiated)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "SignerRegistrationInitiated", log); err != nil {
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

// ParseSignerRegistrationInitiated is a log parse operation binding the contract event 0x665b016a0ac50d1280744eaaff1cf21254d0fd30e4c3987d291913c32163416c.
//
// Solidity: event SignerRegistrationInitiated(address indexed caller)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseSignerRegistrationInitiated(log types.Log) (*BatchAuthenticatorSignerRegistrationInitiated, error) {
	event := new(BatchAuthenticatorSignerRegistrationInitiated)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "SignerRegistrationInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BatchAuthenticatorTeeBatcherUpdatedIterator is returned from FilterTeeBatcherUpdated and is used to iterate over the raw logs and unpacked data for TeeBatcherUpdated events raised by the BatchAuthenticator contract.
type BatchAuthenticatorTeeBatcherUpdatedIterator struct {
	Event *BatchAuthenticatorTeeBatcherUpdated // Event containing the contract specifics and raw log

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
func (it *BatchAuthenticatorTeeBatcherUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BatchAuthenticatorTeeBatcherUpdated)
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
		it.Event = new(BatchAuthenticatorTeeBatcherUpdated)
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
func (it *BatchAuthenticatorTeeBatcherUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BatchAuthenticatorTeeBatcherUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BatchAuthenticatorTeeBatcherUpdated represents a TeeBatcherUpdated event raised by the BatchAuthenticator contract.
type BatchAuthenticatorTeeBatcherUpdated struct {
	OldTeeBatcher common.Address
	NewTeeBatcher common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterTeeBatcherUpdated is a free log retrieval operation binding the contract event 0x5186a10c46a3a9c7ec5470c24b80c6414eba1320cf76bf72ef5135773c7b3327.
//
// Solidity: event TeeBatcherUpdated(address indexed oldTeeBatcher, address indexed newTeeBatcher)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) FilterTeeBatcherUpdated(opts *bind.FilterOpts, oldTeeBatcher []common.Address, newTeeBatcher []common.Address) (*BatchAuthenticatorTeeBatcherUpdatedIterator, error) {

	var oldTeeBatcherRule []interface{}
	for _, oldTeeBatcherItem := range oldTeeBatcher {
		oldTeeBatcherRule = append(oldTeeBatcherRule, oldTeeBatcherItem)
	}
	var newTeeBatcherRule []interface{}
	for _, newTeeBatcherItem := range newTeeBatcher {
		newTeeBatcherRule = append(newTeeBatcherRule, newTeeBatcherItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.FilterLogs(opts, "TeeBatcherUpdated", oldTeeBatcherRule, newTeeBatcherRule)
	if err != nil {
		return nil, err
	}
	return &BatchAuthenticatorTeeBatcherUpdatedIterator{contract: _BatchAuthenticator.contract, event: "TeeBatcherUpdated", logs: logs, sub: sub}, nil
}

// WatchTeeBatcherUpdated is a free log subscription operation binding the contract event 0x5186a10c46a3a9c7ec5470c24b80c6414eba1320cf76bf72ef5135773c7b3327.
//
// Solidity: event TeeBatcherUpdated(address indexed oldTeeBatcher, address indexed newTeeBatcher)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) WatchTeeBatcherUpdated(opts *bind.WatchOpts, sink chan<- *BatchAuthenticatorTeeBatcherUpdated, oldTeeBatcher []common.Address, newTeeBatcher []common.Address) (event.Subscription, error) {

	var oldTeeBatcherRule []interface{}
	for _, oldTeeBatcherItem := range oldTeeBatcher {
		oldTeeBatcherRule = append(oldTeeBatcherRule, oldTeeBatcherItem)
	}
	var newTeeBatcherRule []interface{}
	for _, newTeeBatcherItem := range newTeeBatcher {
		newTeeBatcherRule = append(newTeeBatcherRule, newTeeBatcherItem)
	}

	logs, sub, err := _BatchAuthenticator.contract.WatchLogs(opts, "TeeBatcherUpdated", oldTeeBatcherRule, newTeeBatcherRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BatchAuthenticatorTeeBatcherUpdated)
				if err := _BatchAuthenticator.contract.UnpackLog(event, "TeeBatcherUpdated", log); err != nil {
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

// ParseTeeBatcherUpdated is a log parse operation binding the contract event 0x5186a10c46a3a9c7ec5470c24b80c6414eba1320cf76bf72ef5135773c7b3327.
//
// Solidity: event TeeBatcherUpdated(address indexed oldTeeBatcher, address indexed newTeeBatcher)
func (_BatchAuthenticator *BatchAuthenticatorFilterer) ParseTeeBatcherUpdated(log types.Log) (*BatchAuthenticatorTeeBatcherUpdated, error) {
	event := new(BatchAuthenticatorTeeBatcherUpdated)
	if err := _BatchAuthenticator.contract.UnpackLog(event, "TeeBatcherUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
