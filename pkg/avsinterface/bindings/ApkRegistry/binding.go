// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractApkRegistry

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

// IBLSApkRegistryApkUpdate is an auto generated low-level Go binding around an user-defined struct.
type IBLSApkRegistryApkUpdate struct {
	ApkHash               [24]byte
	UpdateBlockNumber     uint32
	NextUpdateBlockNumber uint32
}

// IBLSApkRegistryPubkeyRegistrationParams is an auto generated low-level Go binding around an user-defined struct.
type IBLSApkRegistryPubkeyRegistrationParams struct {
	PubkeyRegistrationSignature BN254G1Point
	PubkeyG1                    BN254G1Point
	PubkeyG2                    BN254G2Point
}

// ContractApkRegistryMetaData contains all meta data concerning the ContractApkRegistry contract.
var ContractApkRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIRegistryCoordinator\",\"name\":\"_registryCoordinator\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structBN254.G1Point\",\"name\":\"pubkeyG1\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"indexed\":false,\"internalType\":\"structBN254.G2Point\",\"name\":\"pubkeyG2\",\"type\":\"tuple\"}],\"name\":\"NewPubkeyRegistration\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"operatorId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"quorumNumbers\",\"type\":\"bytes\"}],\"name\":\"OperatorAddedToQuorums\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"operatorId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"quorumNumbers\",\"type\":\"bytes\"}],\"name\":\"OperatorRemovedFromQuorums\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"apkHistory\",\"outputs\":[{\"internalType\":\"bytes24\",\"name\":\"apkHash\",\"type\":\"bytes24\"},{\"internalType\":\"uint32\",\"name\":\"updateBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nextUpdateBlockNumber\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"name\":\"currentApk\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"quorumNumbers\",\"type\":\"bytes\"}],\"name\":\"deregisterOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"quorumNumber\",\"type\":\"uint8\"}],\"name\":\"getApk\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"quorumNumber\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getApkHashAtBlockNumberAndIndex\",\"outputs\":[{\"internalType\":\"bytes24\",\"name\":\"\",\"type\":\"bytes24\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"quorumNumber\",\"type\":\"uint8\"}],\"name\":\"getApkHistoryLength\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"quorumNumbers\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"getApkIndicesAtBlockNumber\",\"outputs\":[{\"internalType\":\"uint32[]\",\"name\":\"\",\"type\":\"uint32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"quorumNumber\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getApkUpdateAtIndex\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes24\",\"name\":\"apkHash\",\"type\":\"bytes24\"},{\"internalType\":\"uint32\",\"name\":\"updateBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nextUpdateBlockNumber\",\"type\":\"uint32\"}],\"internalType\":\"structIBLSApkRegistry.ApkUpdate\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"pubkeyHash\",\"type\":\"bytes32\"}],\"name\":\"getOperatorFromPubkeyHash\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"getOperatorId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"getRegisteredPubkey\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point\",\"name\":\"\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"quorumNumber\",\"type\":\"uint8\"}],\"name\":\"initializeQuorum\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"operatorToPubkey\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"operatorToPubkeyHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"pubkeyHashToOperator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"components\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point\",\"name\":\"pubkeyRegistrationSignature\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point\",\"name\":\"pubkeyG1\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structBN254.G2Point\",\"name\":\"pubkeyG2\",\"type\":\"tuple\"}],\"internalType\":\"structIBLSApkRegistry.PubkeyRegistrationParams\",\"name\":\"params\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structBN254.G1Point\",\"name\":\"pubkeyRegistrationMessageHash\",\"type\":\"tuple\"}],\"name\":\"registerBLSPublicKey\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"operatorId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"quorumNumbers\",\"type\":\"bytes\"}],\"name\":\"registerOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registryCoordinator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6080604052600436101561001257600080fd5b60003560e01c8062a1f4cb1461012c57806313542a4e146100eb57806326d941f214610127578063377ed99d146101225780633fb279521461011d57806347b314e8146100e65780635f61a88414610118578063605747d51461011357806368bccaac1461010e5780636d14a987146101095780637916cea6146101045780637ff81a87146100ff578063a3db80e2146100fa578063bf79ce58146100f5578063d5254a8c146100f0578063de29fac0146100eb578063e8bb9ae6146100e65763f4e24fe5146100e157600080fd5b610bba565b6104ba565b6101b7565b610b4d565b6108a9565b61085c565b61081b565b6107a6565b610719565b6105dc565b61054c565b6104e8565b610431565b6102de565b610201565b610160565b600435906001600160a01b038216820361014757565b600080fd5b35906001600160a01b038216820361014757565b34610147576020366003190112610147576001600160a01b03610181610131565b166000526003602052604060002060018154910154906101b36040519283928360209093929193604081019481520152565b0390f35b34610147576020366003190112610147576001600160a01b036101d8610131565b1660005260016020526020604060002054604051908152f35b6004359060ff8216820361014757565b346101475760203660031901126101475761021a6101f1565b610222611419565b60ff8116600052600460205260406000205461027a576102526102789160ff166000526004602052604060002090565b61025a61038d565b600081524363ffffffff166020820152905b60006040830152610c0d565b005b60405162461bcd60e51b815260206004820152603660248201527f424c5341706b52656769737472792e696e697469616c697a6551756f72756d3a6044820152752071756f72756d20616c72656164792065786973747360501b6064820152608490fd5b346101475760203660031901126101475760ff6102f96101f1565b166000526004602052602063ffffffff60406000205416604051908152f35b634e487b7160e01b600052604160045260246000fd5b6040810190811067ffffffffffffffff82111761034a57604052565b610318565b6060810190811067ffffffffffffffff82111761034a57604052565b90601f8019910116810190811067ffffffffffffffff82111761034a57604052565b6040519061039c60608361036b565b565b9061039c604051928361036b565b906040600319830112610147576103c3600461014c565b9160243567ffffffffffffffff811161014757816023820112156101475780600401359067ffffffffffffffff821161034a576040519261040e601f8401601f19166020018561036b565b828452602483830101116101475781600092602460209301838601378301015290565b34610147577f73a2b7fb844724b971802ae9b15db094d4b7192df9d7350e14eb466b9b22eb4e610460366103ac565b90610469611419565b61047c61047582610dd0565b5083611550565b60018060a01b03811660005260016020526104a36040600020549260405193849384610c85565b0390a1005b60209060031901126101475760043590565b34610147576104c8366104a8565b6000526002602052602060018060a01b0360406000205416604051908152f35b346101475760203660031901126101475760ff6105036101f1565b61050b610ce2565b5016600052600560205260408060002060018251916105298361032e565b805483520154602082015261054a8251809260208091805184520151910152565bf35b34610147576040366003190112610147576105a461059e61056b6101f1565b60ff602435916000604080516105808161034f565b8281528260208201520152166000526004602052604060002061078e565b50610d19565b604051809163ffffffff6040606084019267ffffffffffffffff19815116855282602082015116602086015201511660408301520390f35b34610147576060366003190112610147576105f56101f1565b6024359063ffffffff82168092036101475761059e61062d9160ff61061960443590565b91166000526004602052604060002061078e565b9063ffffffff60208301511681106106ae57816106736106829261065b60406101b396015163ffffffff1690565b9063ffffffff82161591821561069e575b5050610d52565b5167ffffffffffffffff191690565b60405167ffffffffffffffff1990911681529081906020820190565b63ffffffff16119050388061066c565b60405162461bcd60e51b815260206004820152603e60248201527f424c5341706b52656769737472792e5f76616c696461746541706b486173684160448201527f74426c6f636b4e756d6265723a20696e64657820746f6f20726563656e7400006064820152608490fd5b34610147576000366003190112610147576040517f00000000000000000000000013a05d12b8061f8f12beca62a42b9815310214396001600160a01b03168152602090f35b634e487b7160e01b600052603260045260246000fd5b80541561078957600052602060002090600090565b61075e565b80548210156107895760005260206000200190600090565b34610147576040366003190112610147576107bf6101f1565b60ff602435911660005260046020526040600020908154811015610147576107e69161078e565b50546040805182821b67ffffffffffffffff1916815260c083901c63ffffffff16602082015260e09290921c90820152606090f35b3461014757602036600319011261014757606061083e610839610131565b610dd0565b610855604051809360208091805184520151910152565b6040820152f35b346101475760203660031901126101475760ff6108776101f1565b166000526005602052604060002060018154910154906101b36040519283928360209093929193604081019481520152565b3461014757610160366003190112610147576108c3610131565b61010036602319011261014757604036610123190112610147576101b3906108e9611419565b6109096108f536610e8f565b805160005260200151602052604060002090565b906109367fad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5831415610f0a565b6001600160a01b0381166000908152600160205260409020610959905415610f69565b60008281526002602052604090205461097b906001600160a01b031615610fd9565b604051610a4290610a3d906109e79060208101906109be816109b061014435610124356084356064356044356024358a611044565b03601f19810183528261036b565b5190207f30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001900690565b610a0b6109f336610eb7565b610a0583610a0036610e8f565b61171d565b90611765565b90610a2d610a176117ef565b91610a05610a2436610edf565b91610a006118e6565b90610a37366110c6565b926119cb565b6110ff565b6001600160a01b0381166000908152600360205260409020610a6d9060643581556001608435910155565b6001600160a01b0381166000908152600160205260409020829055610abf81610aa0846000526002602052604060002090565b80546001600160a01b0319166001600160a01b03909216919091179055565b6040516001600160a01b03909116907fe3fb6613af2e8930cf85d47fcf6db10192224a64c6cbe8023e0eee1ba38280419080610afa8161119a565b0390a26040519081529081906020820190565b602060408183019282815284518094520192019060005b818110610b315750505090565b825163ffffffff16845260209384019390920191600101610b24565b346101475760403660031901126101475760043567ffffffffffffffff8111610147573660238201121561014757806004013567ffffffffffffffff8111610147573660248284010111610147576101b391610bae9160248035920161125f565b60405191829182610b0d565b34610147577ff843ecd53a563675e62107be1494fdde4a3d49aeedaf8d88c616d85346e3500e610be9366103ac565b90610bf2611419565b61047c610c07610c0183610dd0565b50611b14565b83611550565b80546801000000000000000081101561034a57610c2f9160018201815561078e565b610c6f578151602083015160409384015163ffffffff60c01b60c09290921b919091169190931c1760e09290921b6001600160e01b031916919091179055565b634e487b7160e01b600052600060045260246000fd5b919093929360018060a01b03168252602082015260606040820152825180606083015260005b818110610ccc575060809293506000838284010152601f8019910116010190565b8060208092870101516080828601015201610cab565b60405190610cef8261032e565b60006020838281520152565b90604051610d088161032e565b602060018294805484520154910152565b90604051610d268161034f565b604081935467ffffffffffffffff1981831b16835263ffffffff8160c01c16602084015260e01c910152565b15610d5957565b60405162461bcd60e51b815260206004820152604360248201527f424c5341706b52656769737472792e5f76616c696461746541706b486173684160448201527f74426c6f636b4e756d6265723a206e6f74206c61746573742061706b2075706460648201526261746560e81b608482015260a490fd5b610dd8610ce2565b5060018060a01b031690816000526003602052604060002091600160405193610e008561032e565b80548552015460208401526000526001602052604060002054918215610e24579190565b60405162461bcd60e51b815260206004820152603e60248201527f424c5341706b52656769737472792e676574526567697374657265645075626b60448201527f65793a206f70657261746f72206973206e6f74207265676973746572656400006064820152608490fd5b60409060631901126101475760405190610ea88261032e565b60643582526084356020830152565b60409060231901126101475760405190610ed08261032e565b60243582526044356020830152565b6040906101231901126101475760405190610ef98261032e565b610124358252610144356020830152565b15610f1157565b608460405162461bcd60e51b81526020600482015260406024820152600080516020611bd783398151915260448201527f4b65793a2063616e6e6f74207265676973746572207a65726f207075626b65796064820152fd5b15610f7057565b60405162461bcd60e51b81526020600482015260476024820152600080516020611bd783398151915260448201527f4b65793a206f70657261746f7220616c72656164792072656769737465726564606482015266207075626b657960c81b608482015260a490fd5b15610fe057565b60405162461bcd60e51b81526020600482015260426024820152600080516020611bd783398151915260448201527f4b65793a207075626c6963206b657920616c7265616479207265676973746572606482015261195960f21b608482015260a490fd5b949290916101409694928652602086015260408501526060840152604060a46080850137604060e460c08501376101008301526101208201520190565b9080601f8301121561014757604080519261109c828561036b565b8391810192831161014757905b8282106110b65750505090565b81358152602091820191016110a9565b90608060a319830112610147576040516110df8161032e565b60206110fa82946110f18160a4611081565b845260e4611081565b910152565b1561110657565b60405162461bcd60e51b815260206004820152606c6024820152600080516020611bd783398151915260448201527f4b65793a2065697468657220746865204731207369676e61747572652069732060648201527f77726f6e672c206f7220473120616e642047322070726976617465206b65792060848201526b0c8de40dcdee840dac2e8c6d60a31b60a482015260c490fd5b90604060e4608060c0850194606435815260843560208201528360a4818301370137565b67ffffffffffffffff811161034a5760051b60200190565b906111e0826111be565b6111ed604051918261036b565b82815280926111fe601f19916111be565b0190602036910137565b90821015610789570190565b634e487b7160e01b600052601160045260246000fd5b8015611237576000190190565b611214565b60001981019190821161123757565b80518210156107895760209160051b010190565b91909161126b836111d6565b9260005b81811061127d575050505090565b6112a261129c61128e838587611208565b356001600160f81b03191690565b60f81c90565b6112b98160ff166000526004602052604060002090565b54801580156113ec575b61136757805b6112d8575b505060010161126f565b8563ffffffff61131861130a6112fb8660ff166000526004602052604060002090565b6113048661123c565b9061078e565b505460c01c63ffffffff1690565b16111561132e576113289061122a565b806112c9565b600192915061134b6113426113609261123c565b63ffffffff1690565b611355838961124b565b9063ffffffff169052565b90386112ce565b60405162461bcd60e51b815260206004820152605160248201527f424c5341706b52656769737472792e67657441706b496e64696365734174426c60448201527f6f636b4e756d6265723a20626c6f636b4e756d626572206973206265666f7265606482015270207468652066697273742075706461746560781b608482015260a490fd5b5061141261134261130a61140d8560ff166000526004602052604060002090565b610774565b86106112c3565b7f00000000000000000000000013a05d12b8061f8f12beca62a42b9815310214396001600160a01b0316330361144b57565b60405162461bcd60e51b815260206004820152604e60248201527f424c5341706b52656769737472792e6f6e6c795265676973747279436f6f726460448201527f696e61746f723a2063616c6c6572206973206e6f74207468652072656769737460648201526d393c9031b7b7b93234b730ba37b960911b608482015260a490fd5b908151811015610789570160200190565b156114e557565b60405162461bcd60e51b815260206004820152603d60248201527f424c5341706b52656769737472792e5f70726f6365737351756f72756d41706b60448201527f5570646174653a2071756f72756d20646f6573206e6f742065786973740000006064820152608490fd5b919061155a610ce2565b504363ffffffff169060005b84518110156116f057808361159161129c6115836001958a6114cd565b516001600160f81b03191690565b6115a88160ff166000526004602052604060002090565b54906115b58215156114de565b6116416116206116126115e6896115e16115dc8760ff166000526005602052604060002090565b610cfb565b611765565b6108f5816116018760ff166000526005602052604060002090565b906020600191805184550151910155565b67ffffffffffffffff191690565b9261130461163b8460ff166000526004602052604060002090565b9161123c565b509083611659611342845463ffffffff9060c01c1690565b03611682575061167c92509060401c67ffffffffffffffff60c01b825416179055565b01611566565b81546001600160e01b031660e09490941b6001600160e01b0319169390931790556116eb916116be9060ff166000526004602052604060002090565b6116da6116c961038d565b67ffffffffffffffff199093168352565b63ffffffff8716602083015261026c565b61167c565b5050509050565b60405190610180611708818461036b565b368337565b604051906020611708818461036b565b60409092919261172b610ce2565b9384916060916020855192611740858561036b565b8436853780518452015160208301528482015260076107cf195a01fa1561176357565bfe5b604090929192611773610ce2565b938491602060809281865193611789868661036b565b85368637805185520151828401528051868401520151606082015260066107cf195a01fa801561176357156117ba57565b60405162461bcd60e51b815260206004820152600d60248201526c1958cb5859190b59985a5b1959609a1b6044820152606490fd5b6040516117fb8161032e565b604090815161180a838261036b565b823682378152602082519161181f848461036b565b8336843701528051611831828261036b565b7f198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c281527f1800deef121f1e76426a00665e5c4479674322d4f75edadd46debd5cd992f6ed6020820152815190611887838361036b565b7f275dc4a288d1afb3cbb1ac09187524c7db36395df7be3b99e673b13a075a65ec82527f1d9befcd05a5323e6da4d435f3b617cdb3af83285c2df711ef39c01571827f9d60208301526118dc8351938461036b565b8252602082015290565b6118ee610ce2565b506040516118fb8161032e565b600181526002602082015290565b9060068202918083046006149015171561123757565b9060028110156107895760051b0190565b906001820180921161123757565b906002820180921161123757565b906003820180921161123757565b906004820180921161123757565b906005820180921161123757565b90600c8110156107895760051b0190565b1561198e57565b60405162461bcd60e51b81526020600482015260156024820152741c185a5c9a5b99cb5bdc18dbd9194b59985a5b1959605a1b6044820152606490fd5b9290916119d8604061039e565b93845260208401526119ea604061039e565b91825260208201526119fa6116f7565b9160005b60028110611a3857505050602061018091611a1761170d565b92839160086107cf195a01fa801561176357611a3290611987565b51151590565b80611a44600192611909565b611a4e828561191f565b5151600090611a5d8389611976565b526020611a6a848761191f565b510151611a7f611a7984611930565b89611976565b52611a8a838761191f565b515151611a99611a798461193e565b52611aaf611aa7848861191f565b515160200190565b51611abc611a798461194c565b526020611ac9848861191f565b510151905051611ae1611adb8361195a565b88611976565b52611b0d611b07611b006020611af7868a61191f565b51015160200190565b5192611968565b87611976565b52016119fe565b611b1c610ce2565b50805190811580611bca575b15611b4b575050604051611b3d60408261036b565b600081526000602082015290565b60207f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47910151067f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47037f30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47811161123757604051916118dc60408461036b565b50602081015115611b2856fe424c5341706b52656769737472792e7265676973746572424c535075626c6963a2646970667358221220e82a286020eb484c3c3b39d94a806193e88535ff7ede1441a7e91aab51c0379464736f6c634300081a0033",
}

// ContractApkRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractApkRegistryMetaData.ABI instead.
var ContractApkRegistryABI = ContractApkRegistryMetaData.ABI

// ContractApkRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractApkRegistryMetaData.Bin instead.
var ContractApkRegistryBin = ContractApkRegistryMetaData.Bin

// DeployContractApkRegistry deploys a new Ethereum contract, binding an instance of ContractApkRegistry to it.
func DeployContractApkRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, _registryCoordinator common.Address) (common.Address, *types.Transaction, *ContractApkRegistry, error) {
	parsed, err := ContractApkRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractApkRegistryBin), backend, _registryCoordinator)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractApkRegistry{ContractApkRegistryCaller: ContractApkRegistryCaller{contract: contract}, ContractApkRegistryTransactor: ContractApkRegistryTransactor{contract: contract}, ContractApkRegistryFilterer: ContractApkRegistryFilterer{contract: contract}}, nil
}

// ContractApkRegistryMethods is an auto generated interface around an Ethereum contract.
type ContractApkRegistryMethods interface {
	ContractApkRegistryCalls
	ContractApkRegistryTransacts
	ContractApkRegistryFilters
}

// ContractApkRegistryCalls is an auto generated interface that defines the call methods available for an Ethereum contract.
type ContractApkRegistryCalls interface {
	ApkHistory(opts *bind.CallOpts, arg0 uint8, arg1 *big.Int) (struct {
		ApkHash               [24]byte
		UpdateBlockNumber     uint32
		NextUpdateBlockNumber uint32
	}, error)

	CurrentApk(opts *bind.CallOpts, arg0 uint8) (struct {
		X *big.Int
		Y *big.Int
	}, error)

	GetApk(opts *bind.CallOpts, quorumNumber uint8) (BN254G1Point, error)

	GetApkHashAtBlockNumberAndIndex(opts *bind.CallOpts, quorumNumber uint8, blockNumber uint32, index *big.Int) ([24]byte, error)

	GetApkHistoryLength(opts *bind.CallOpts, quorumNumber uint8) (uint32, error)

	GetApkIndicesAtBlockNumber(opts *bind.CallOpts, quorumNumbers []byte, blockNumber *big.Int) ([]uint32, error)

	GetApkUpdateAtIndex(opts *bind.CallOpts, quorumNumber uint8, index *big.Int) (IBLSApkRegistryApkUpdate, error)

	GetOperatorFromPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error)

	GetOperatorId(opts *bind.CallOpts, operator common.Address) ([32]byte, error)

	GetRegisteredPubkey(opts *bind.CallOpts, operator common.Address) (BN254G1Point, [32]byte, error)

	OperatorToPubkey(opts *bind.CallOpts, arg0 common.Address) (struct {
		X *big.Int
		Y *big.Int
	}, error)

	OperatorToPubkeyHash(opts *bind.CallOpts, arg0 common.Address) ([32]byte, error)

	PubkeyHashToOperator(opts *bind.CallOpts, arg0 [32]byte) (common.Address, error)

	RegistryCoordinator(opts *bind.CallOpts) (common.Address, error)
}

// ContractApkRegistryTransacts is an auto generated interface that defines the transact methods available for an Ethereum contract.
type ContractApkRegistryTransacts interface {
	DeregisterOperator(opts *bind.TransactOpts, operator common.Address, quorumNumbers []byte) (*types.Transaction, error)

	InitializeQuorum(opts *bind.TransactOpts, quorumNumber uint8) (*types.Transaction, error)

	RegisterBLSPublicKey(opts *bind.TransactOpts, operator common.Address, params IBLSApkRegistryPubkeyRegistrationParams, pubkeyRegistrationMessageHash BN254G1Point) (*types.Transaction, error)

	RegisterOperator(opts *bind.TransactOpts, operator common.Address, quorumNumbers []byte) (*types.Transaction, error)
}

// ContractApkRegistryFilterer is an auto generated interface that defines the log filtering methods available for an Ethereum contract.
type ContractApkRegistryFilters interface {
	FilterInitialized(opts *bind.FilterOpts) (*ContractApkRegistryInitializedIterator, error)
	WatchInitialized(opts *bind.WatchOpts, sink chan<- *ContractApkRegistryInitialized) (event.Subscription, error)
	ParseInitialized(log types.Log) (*ContractApkRegistryInitialized, error)

	FilterNewPubkeyRegistration(opts *bind.FilterOpts, operator []common.Address) (*ContractApkRegistryNewPubkeyRegistrationIterator, error)
	WatchNewPubkeyRegistration(opts *bind.WatchOpts, sink chan<- *ContractApkRegistryNewPubkeyRegistration, operator []common.Address) (event.Subscription, error)
	ParseNewPubkeyRegistration(log types.Log) (*ContractApkRegistryNewPubkeyRegistration, error)

	FilterOperatorAddedToQuorums(opts *bind.FilterOpts) (*ContractApkRegistryOperatorAddedToQuorumsIterator, error)
	WatchOperatorAddedToQuorums(opts *bind.WatchOpts, sink chan<- *ContractApkRegistryOperatorAddedToQuorums) (event.Subscription, error)
	ParseOperatorAddedToQuorums(log types.Log) (*ContractApkRegistryOperatorAddedToQuorums, error)

	FilterOperatorRemovedFromQuorums(opts *bind.FilterOpts) (*ContractApkRegistryOperatorRemovedFromQuorumsIterator, error)
	WatchOperatorRemovedFromQuorums(opts *bind.WatchOpts, sink chan<- *ContractApkRegistryOperatorRemovedFromQuorums) (event.Subscription, error)
	ParseOperatorRemovedFromQuorums(log types.Log) (*ContractApkRegistryOperatorRemovedFromQuorums, error)
}

// ContractApkRegistry is an auto generated Go binding around an Ethereum contract.
type ContractApkRegistry struct {
	ContractApkRegistryCaller     // Read-only binding to the contract
	ContractApkRegistryTransactor // Write-only binding to the contract
	ContractApkRegistryFilterer   // Log filterer for contract events
}

// ContractApkRegistry implements the ContractApkRegistryMethods interface.
var _ ContractApkRegistryMethods = (*ContractApkRegistry)(nil)

// ContractApkRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractApkRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractApkRegistryCaller implements the ContractApkRegistryCalls interface.
var _ ContractApkRegistryCalls = (*ContractApkRegistryCaller)(nil)

// ContractApkRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractApkRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractApkRegistryTransactor implements the ContractApkRegistryTransacts interface.
var _ ContractApkRegistryTransacts = (*ContractApkRegistryTransactor)(nil)

// ContractApkRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractApkRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractApkRegistryFilterer implements the ContractApkRegistryFilters interface.
var _ ContractApkRegistryFilters = (*ContractApkRegistryFilterer)(nil)

// ContractApkRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractApkRegistrySession struct {
	Contract     *ContractApkRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ContractApkRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractApkRegistryCallerSession struct {
	Contract *ContractApkRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// ContractApkRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractApkRegistryTransactorSession struct {
	Contract     *ContractApkRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// ContractApkRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractApkRegistryRaw struct {
	Contract *ContractApkRegistry // Generic contract binding to access the raw methods on
}

// ContractApkRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractApkRegistryCallerRaw struct {
	Contract *ContractApkRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ContractApkRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractApkRegistryTransactorRaw struct {
	Contract *ContractApkRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractApkRegistry creates a new instance of ContractApkRegistry, bound to a specific deployed contract.
func NewContractApkRegistry(address common.Address, backend bind.ContractBackend) (*ContractApkRegistry, error) {
	contract, err := bindContractApkRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractApkRegistry{ContractApkRegistryCaller: ContractApkRegistryCaller{contract: contract}, ContractApkRegistryTransactor: ContractApkRegistryTransactor{contract: contract}, ContractApkRegistryFilterer: ContractApkRegistryFilterer{contract: contract}}, nil
}

// NewContractApkRegistryCaller creates a new read-only instance of ContractApkRegistry, bound to a specific deployed contract.
func NewContractApkRegistryCaller(address common.Address, caller bind.ContractCaller) (*ContractApkRegistryCaller, error) {
	contract, err := bindContractApkRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractApkRegistryCaller{contract: contract}, nil
}

// NewContractApkRegistryTransactor creates a new write-only instance of ContractApkRegistry, bound to a specific deployed contract.
func NewContractApkRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractApkRegistryTransactor, error) {
	contract, err := bindContractApkRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractApkRegistryTransactor{contract: contract}, nil
}

// NewContractApkRegistryFilterer creates a new log filterer instance of ContractApkRegistry, bound to a specific deployed contract.
func NewContractApkRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractApkRegistryFilterer, error) {
	contract, err := bindContractApkRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractApkRegistryFilterer{contract: contract}, nil
}

// bindContractApkRegistry binds a generic wrapper to an already deployed contract.
func bindContractApkRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractApkRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractApkRegistry *ContractApkRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractApkRegistry.Contract.ContractApkRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractApkRegistry *ContractApkRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.ContractApkRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractApkRegistry *ContractApkRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.ContractApkRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractApkRegistry *ContractApkRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractApkRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractApkRegistry *ContractApkRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractApkRegistry *ContractApkRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.contract.Transact(opts, method, params...)
}

// ApkHistory is a free data retrieval call binding the contract method 0x7916cea6.
//
// Solidity: function apkHistory(uint8 , uint256 ) view returns(bytes24 apkHash, uint32 updateBlockNumber, uint32 nextUpdateBlockNumber)
func (_ContractApkRegistry *ContractApkRegistryCaller) ApkHistory(opts *bind.CallOpts, arg0 uint8, arg1 *big.Int) (struct {
	ApkHash               [24]byte
	UpdateBlockNumber     uint32
	NextUpdateBlockNumber uint32
}, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "apkHistory", arg0, arg1)

	outstruct := new(struct {
		ApkHash               [24]byte
		UpdateBlockNumber     uint32
		NextUpdateBlockNumber uint32
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ApkHash = *abi.ConvertType(out[0], new([24]byte)).(*[24]byte)
	outstruct.UpdateBlockNumber = *abi.ConvertType(out[1], new(uint32)).(*uint32)
	outstruct.NextUpdateBlockNumber = *abi.ConvertType(out[2], new(uint32)).(*uint32)

	return *outstruct, err

}

// ApkHistory is a free data retrieval call binding the contract method 0x7916cea6.
//
// Solidity: function apkHistory(uint8 , uint256 ) view returns(bytes24 apkHash, uint32 updateBlockNumber, uint32 nextUpdateBlockNumber)
func (_ContractApkRegistry *ContractApkRegistrySession) ApkHistory(arg0 uint8, arg1 *big.Int) (struct {
	ApkHash               [24]byte
	UpdateBlockNumber     uint32
	NextUpdateBlockNumber uint32
}, error) {
	return _ContractApkRegistry.Contract.ApkHistory(&_ContractApkRegistry.CallOpts, arg0, arg1)
}

// ApkHistory is a free data retrieval call binding the contract method 0x7916cea6.
//
// Solidity: function apkHistory(uint8 , uint256 ) view returns(bytes24 apkHash, uint32 updateBlockNumber, uint32 nextUpdateBlockNumber)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) ApkHistory(arg0 uint8, arg1 *big.Int) (struct {
	ApkHash               [24]byte
	UpdateBlockNumber     uint32
	NextUpdateBlockNumber uint32
}, error) {
	return _ContractApkRegistry.Contract.ApkHistory(&_ContractApkRegistry.CallOpts, arg0, arg1)
}

// CurrentApk is a free data retrieval call binding the contract method 0xa3db80e2.
//
// Solidity: function currentApk(uint8 ) view returns(uint256 X, uint256 Y)
func (_ContractApkRegistry *ContractApkRegistryCaller) CurrentApk(opts *bind.CallOpts, arg0 uint8) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "currentApk", arg0)

	outstruct := new(struct {
		X *big.Int
		Y *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.X = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Y = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// CurrentApk is a free data retrieval call binding the contract method 0xa3db80e2.
//
// Solidity: function currentApk(uint8 ) view returns(uint256 X, uint256 Y)
func (_ContractApkRegistry *ContractApkRegistrySession) CurrentApk(arg0 uint8) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _ContractApkRegistry.Contract.CurrentApk(&_ContractApkRegistry.CallOpts, arg0)
}

// CurrentApk is a free data retrieval call binding the contract method 0xa3db80e2.
//
// Solidity: function currentApk(uint8 ) view returns(uint256 X, uint256 Y)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) CurrentApk(arg0 uint8) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _ContractApkRegistry.Contract.CurrentApk(&_ContractApkRegistry.CallOpts, arg0)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 quorumNumber) view returns((uint256,uint256))
func (_ContractApkRegistry *ContractApkRegistryCaller) GetApk(opts *bind.CallOpts, quorumNumber uint8) (BN254G1Point, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "getApk", quorumNumber)

	if err != nil {
		return *new(BN254G1Point), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)

	return out0, err

}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 quorumNumber) view returns((uint256,uint256))
func (_ContractApkRegistry *ContractApkRegistrySession) GetApk(quorumNumber uint8) (BN254G1Point, error) {
	return _ContractApkRegistry.Contract.GetApk(&_ContractApkRegistry.CallOpts, quorumNumber)
}

// GetApk is a free data retrieval call binding the contract method 0x5f61a884.
//
// Solidity: function getApk(uint8 quorumNumber) view returns((uint256,uint256))
func (_ContractApkRegistry *ContractApkRegistryCallerSession) GetApk(quorumNumber uint8) (BN254G1Point, error) {
	return _ContractApkRegistry.Contract.GetApk(&_ContractApkRegistry.CallOpts, quorumNumber)
}

// GetApkHashAtBlockNumberAndIndex is a free data retrieval call binding the contract method 0x68bccaac.
//
// Solidity: function getApkHashAtBlockNumberAndIndex(uint8 quorumNumber, uint32 blockNumber, uint256 index) view returns(bytes24)
func (_ContractApkRegistry *ContractApkRegistryCaller) GetApkHashAtBlockNumberAndIndex(opts *bind.CallOpts, quorumNumber uint8, blockNumber uint32, index *big.Int) ([24]byte, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "getApkHashAtBlockNumberAndIndex", quorumNumber, blockNumber, index)

	if err != nil {
		return *new([24]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([24]byte)).(*[24]byte)

	return out0, err

}

// GetApkHashAtBlockNumberAndIndex is a free data retrieval call binding the contract method 0x68bccaac.
//
// Solidity: function getApkHashAtBlockNumberAndIndex(uint8 quorumNumber, uint32 blockNumber, uint256 index) view returns(bytes24)
func (_ContractApkRegistry *ContractApkRegistrySession) GetApkHashAtBlockNumberAndIndex(quorumNumber uint8, blockNumber uint32, index *big.Int) ([24]byte, error) {
	return _ContractApkRegistry.Contract.GetApkHashAtBlockNumberAndIndex(&_ContractApkRegistry.CallOpts, quorumNumber, blockNumber, index)
}

// GetApkHashAtBlockNumberAndIndex is a free data retrieval call binding the contract method 0x68bccaac.
//
// Solidity: function getApkHashAtBlockNumberAndIndex(uint8 quorumNumber, uint32 blockNumber, uint256 index) view returns(bytes24)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) GetApkHashAtBlockNumberAndIndex(quorumNumber uint8, blockNumber uint32, index *big.Int) ([24]byte, error) {
	return _ContractApkRegistry.Contract.GetApkHashAtBlockNumberAndIndex(&_ContractApkRegistry.CallOpts, quorumNumber, blockNumber, index)
}

// GetApkHistoryLength is a free data retrieval call binding the contract method 0x377ed99d.
//
// Solidity: function getApkHistoryLength(uint8 quorumNumber) view returns(uint32)
func (_ContractApkRegistry *ContractApkRegistryCaller) GetApkHistoryLength(opts *bind.CallOpts, quorumNumber uint8) (uint32, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "getApkHistoryLength", quorumNumber)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetApkHistoryLength is a free data retrieval call binding the contract method 0x377ed99d.
//
// Solidity: function getApkHistoryLength(uint8 quorumNumber) view returns(uint32)
func (_ContractApkRegistry *ContractApkRegistrySession) GetApkHistoryLength(quorumNumber uint8) (uint32, error) {
	return _ContractApkRegistry.Contract.GetApkHistoryLength(&_ContractApkRegistry.CallOpts, quorumNumber)
}

// GetApkHistoryLength is a free data retrieval call binding the contract method 0x377ed99d.
//
// Solidity: function getApkHistoryLength(uint8 quorumNumber) view returns(uint32)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) GetApkHistoryLength(quorumNumber uint8) (uint32, error) {
	return _ContractApkRegistry.Contract.GetApkHistoryLength(&_ContractApkRegistry.CallOpts, quorumNumber)
}

// GetApkIndicesAtBlockNumber is a free data retrieval call binding the contract method 0xd5254a8c.
//
// Solidity: function getApkIndicesAtBlockNumber(bytes quorumNumbers, uint256 blockNumber) view returns(uint32[])
func (_ContractApkRegistry *ContractApkRegistryCaller) GetApkIndicesAtBlockNumber(opts *bind.CallOpts, quorumNumbers []byte, blockNumber *big.Int) ([]uint32, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "getApkIndicesAtBlockNumber", quorumNumbers, blockNumber)

	if err != nil {
		return *new([]uint32), err
	}

	out0 := *abi.ConvertType(out[0], new([]uint32)).(*[]uint32)

	return out0, err

}

// GetApkIndicesAtBlockNumber is a free data retrieval call binding the contract method 0xd5254a8c.
//
// Solidity: function getApkIndicesAtBlockNumber(bytes quorumNumbers, uint256 blockNumber) view returns(uint32[])
func (_ContractApkRegistry *ContractApkRegistrySession) GetApkIndicesAtBlockNumber(quorumNumbers []byte, blockNumber *big.Int) ([]uint32, error) {
	return _ContractApkRegistry.Contract.GetApkIndicesAtBlockNumber(&_ContractApkRegistry.CallOpts, quorumNumbers, blockNumber)
}

// GetApkIndicesAtBlockNumber is a free data retrieval call binding the contract method 0xd5254a8c.
//
// Solidity: function getApkIndicesAtBlockNumber(bytes quorumNumbers, uint256 blockNumber) view returns(uint32[])
func (_ContractApkRegistry *ContractApkRegistryCallerSession) GetApkIndicesAtBlockNumber(quorumNumbers []byte, blockNumber *big.Int) ([]uint32, error) {
	return _ContractApkRegistry.Contract.GetApkIndicesAtBlockNumber(&_ContractApkRegistry.CallOpts, quorumNumbers, blockNumber)
}

// GetApkUpdateAtIndex is a free data retrieval call binding the contract method 0x605747d5.
//
// Solidity: function getApkUpdateAtIndex(uint8 quorumNumber, uint256 index) view returns((bytes24,uint32,uint32))
func (_ContractApkRegistry *ContractApkRegistryCaller) GetApkUpdateAtIndex(opts *bind.CallOpts, quorumNumber uint8, index *big.Int) (IBLSApkRegistryApkUpdate, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "getApkUpdateAtIndex", quorumNumber, index)

	if err != nil {
		return *new(IBLSApkRegistryApkUpdate), err
	}

	out0 := *abi.ConvertType(out[0], new(IBLSApkRegistryApkUpdate)).(*IBLSApkRegistryApkUpdate)

	return out0, err

}

// GetApkUpdateAtIndex is a free data retrieval call binding the contract method 0x605747d5.
//
// Solidity: function getApkUpdateAtIndex(uint8 quorumNumber, uint256 index) view returns((bytes24,uint32,uint32))
func (_ContractApkRegistry *ContractApkRegistrySession) GetApkUpdateAtIndex(quorumNumber uint8, index *big.Int) (IBLSApkRegistryApkUpdate, error) {
	return _ContractApkRegistry.Contract.GetApkUpdateAtIndex(&_ContractApkRegistry.CallOpts, quorumNumber, index)
}

// GetApkUpdateAtIndex is a free data retrieval call binding the contract method 0x605747d5.
//
// Solidity: function getApkUpdateAtIndex(uint8 quorumNumber, uint256 index) view returns((bytes24,uint32,uint32))
func (_ContractApkRegistry *ContractApkRegistryCallerSession) GetApkUpdateAtIndex(quorumNumber uint8, index *big.Int) (IBLSApkRegistryApkUpdate, error) {
	return _ContractApkRegistry.Contract.GetApkUpdateAtIndex(&_ContractApkRegistry.CallOpts, quorumNumber, index)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_ContractApkRegistry *ContractApkRegistryCaller) GetOperatorFromPubkeyHash(opts *bind.CallOpts, pubkeyHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "getOperatorFromPubkeyHash", pubkeyHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_ContractApkRegistry *ContractApkRegistrySession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _ContractApkRegistry.Contract.GetOperatorFromPubkeyHash(&_ContractApkRegistry.CallOpts, pubkeyHash)
}

// GetOperatorFromPubkeyHash is a free data retrieval call binding the contract method 0x47b314e8.
//
// Solidity: function getOperatorFromPubkeyHash(bytes32 pubkeyHash) view returns(address)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) GetOperatorFromPubkeyHash(pubkeyHash [32]byte) (common.Address, error) {
	return _ContractApkRegistry.Contract.GetOperatorFromPubkeyHash(&_ContractApkRegistry.CallOpts, pubkeyHash)
}

// GetOperatorId is a free data retrieval call binding the contract method 0x13542a4e.
//
// Solidity: function getOperatorId(address operator) view returns(bytes32)
func (_ContractApkRegistry *ContractApkRegistryCaller) GetOperatorId(opts *bind.CallOpts, operator common.Address) ([32]byte, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "getOperatorId", operator)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetOperatorId is a free data retrieval call binding the contract method 0x13542a4e.
//
// Solidity: function getOperatorId(address operator) view returns(bytes32)
func (_ContractApkRegistry *ContractApkRegistrySession) GetOperatorId(operator common.Address) ([32]byte, error) {
	return _ContractApkRegistry.Contract.GetOperatorId(&_ContractApkRegistry.CallOpts, operator)
}

// GetOperatorId is a free data retrieval call binding the contract method 0x13542a4e.
//
// Solidity: function getOperatorId(address operator) view returns(bytes32)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) GetOperatorId(operator common.Address) ([32]byte, error) {
	return _ContractApkRegistry.Contract.GetOperatorId(&_ContractApkRegistry.CallOpts, operator)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_ContractApkRegistry *ContractApkRegistryCaller) GetRegisteredPubkey(opts *bind.CallOpts, operator common.Address) (BN254G1Point, [32]byte, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "getRegisteredPubkey", operator)

	if err != nil {
		return *new(BN254G1Point), *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(BN254G1Point)).(*BN254G1Point)
	out1 := *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)

	return out0, out1, err

}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_ContractApkRegistry *ContractApkRegistrySession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _ContractApkRegistry.Contract.GetRegisteredPubkey(&_ContractApkRegistry.CallOpts, operator)
}

// GetRegisteredPubkey is a free data retrieval call binding the contract method 0x7ff81a87.
//
// Solidity: function getRegisteredPubkey(address operator) view returns((uint256,uint256), bytes32)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) GetRegisteredPubkey(operator common.Address) (BN254G1Point, [32]byte, error) {
	return _ContractApkRegistry.Contract.GetRegisteredPubkey(&_ContractApkRegistry.CallOpts, operator)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address ) view returns(uint256 X, uint256 Y)
func (_ContractApkRegistry *ContractApkRegistryCaller) OperatorToPubkey(opts *bind.CallOpts, arg0 common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "operatorToPubkey", arg0)

	outstruct := new(struct {
		X *big.Int
		Y *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.X = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Y = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address ) view returns(uint256 X, uint256 Y)
func (_ContractApkRegistry *ContractApkRegistrySession) OperatorToPubkey(arg0 common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _ContractApkRegistry.Contract.OperatorToPubkey(&_ContractApkRegistry.CallOpts, arg0)
}

// OperatorToPubkey is a free data retrieval call binding the contract method 0x00a1f4cb.
//
// Solidity: function operatorToPubkey(address ) view returns(uint256 X, uint256 Y)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) OperatorToPubkey(arg0 common.Address) (struct {
	X *big.Int
	Y *big.Int
}, error) {
	return _ContractApkRegistry.Contract.OperatorToPubkey(&_ContractApkRegistry.CallOpts, arg0)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address ) view returns(bytes32)
func (_ContractApkRegistry *ContractApkRegistryCaller) OperatorToPubkeyHash(opts *bind.CallOpts, arg0 common.Address) ([32]byte, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "operatorToPubkeyHash", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address ) view returns(bytes32)
func (_ContractApkRegistry *ContractApkRegistrySession) OperatorToPubkeyHash(arg0 common.Address) ([32]byte, error) {
	return _ContractApkRegistry.Contract.OperatorToPubkeyHash(&_ContractApkRegistry.CallOpts, arg0)
}

// OperatorToPubkeyHash is a free data retrieval call binding the contract method 0xde29fac0.
//
// Solidity: function operatorToPubkeyHash(address ) view returns(bytes32)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) OperatorToPubkeyHash(arg0 common.Address) ([32]byte, error) {
	return _ContractApkRegistry.Contract.OperatorToPubkeyHash(&_ContractApkRegistry.CallOpts, arg0)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 ) view returns(address)
func (_ContractApkRegistry *ContractApkRegistryCaller) PubkeyHashToOperator(opts *bind.CallOpts, arg0 [32]byte) (common.Address, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "pubkeyHashToOperator", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 ) view returns(address)
func (_ContractApkRegistry *ContractApkRegistrySession) PubkeyHashToOperator(arg0 [32]byte) (common.Address, error) {
	return _ContractApkRegistry.Contract.PubkeyHashToOperator(&_ContractApkRegistry.CallOpts, arg0)
}

// PubkeyHashToOperator is a free data retrieval call binding the contract method 0xe8bb9ae6.
//
// Solidity: function pubkeyHashToOperator(bytes32 ) view returns(address)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) PubkeyHashToOperator(arg0 [32]byte) (common.Address, error) {
	return _ContractApkRegistry.Contract.PubkeyHashToOperator(&_ContractApkRegistry.CallOpts, arg0)
}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractApkRegistry *ContractApkRegistryCaller) RegistryCoordinator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractApkRegistry.contract.Call(opts, &out, "registryCoordinator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractApkRegistry *ContractApkRegistrySession) RegistryCoordinator() (common.Address, error) {
	return _ContractApkRegistry.Contract.RegistryCoordinator(&_ContractApkRegistry.CallOpts)
}

// RegistryCoordinator is a free data retrieval call binding the contract method 0x6d14a987.
//
// Solidity: function registryCoordinator() view returns(address)
func (_ContractApkRegistry *ContractApkRegistryCallerSession) RegistryCoordinator() (common.Address, error) {
	return _ContractApkRegistry.Contract.RegistryCoordinator(&_ContractApkRegistry.CallOpts)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0xf4e24fe5.
//
// Solidity: function deregisterOperator(address operator, bytes quorumNumbers) returns()
func (_ContractApkRegistry *ContractApkRegistryTransactor) DeregisterOperator(opts *bind.TransactOpts, operator common.Address, quorumNumbers []byte) (*types.Transaction, error) {
	return _ContractApkRegistry.contract.Transact(opts, "deregisterOperator", operator, quorumNumbers)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0xf4e24fe5.
//
// Solidity: function deregisterOperator(address operator, bytes quorumNumbers) returns()
func (_ContractApkRegistry *ContractApkRegistrySession) DeregisterOperator(operator common.Address, quorumNumbers []byte) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.DeregisterOperator(&_ContractApkRegistry.TransactOpts, operator, quorumNumbers)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0xf4e24fe5.
//
// Solidity: function deregisterOperator(address operator, bytes quorumNumbers) returns()
func (_ContractApkRegistry *ContractApkRegistryTransactorSession) DeregisterOperator(operator common.Address, quorumNumbers []byte) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.DeregisterOperator(&_ContractApkRegistry.TransactOpts, operator, quorumNumbers)
}

// InitializeQuorum is a paid mutator transaction binding the contract method 0x26d941f2.
//
// Solidity: function initializeQuorum(uint8 quorumNumber) returns()
func (_ContractApkRegistry *ContractApkRegistryTransactor) InitializeQuorum(opts *bind.TransactOpts, quorumNumber uint8) (*types.Transaction, error) {
	return _ContractApkRegistry.contract.Transact(opts, "initializeQuorum", quorumNumber)
}

// InitializeQuorum is a paid mutator transaction binding the contract method 0x26d941f2.
//
// Solidity: function initializeQuorum(uint8 quorumNumber) returns()
func (_ContractApkRegistry *ContractApkRegistrySession) InitializeQuorum(quorumNumber uint8) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.InitializeQuorum(&_ContractApkRegistry.TransactOpts, quorumNumber)
}

// InitializeQuorum is a paid mutator transaction binding the contract method 0x26d941f2.
//
// Solidity: function initializeQuorum(uint8 quorumNumber) returns()
func (_ContractApkRegistry *ContractApkRegistryTransactorSession) InitializeQuorum(quorumNumber uint8) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.InitializeQuorum(&_ContractApkRegistry.TransactOpts, quorumNumber)
}

// RegisterBLSPublicKey is a paid mutator transaction binding the contract method 0xbf79ce58.
//
// Solidity: function registerBLSPublicKey(address operator, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) params, (uint256,uint256) pubkeyRegistrationMessageHash) returns(bytes32 operatorId)
func (_ContractApkRegistry *ContractApkRegistryTransactor) RegisterBLSPublicKey(opts *bind.TransactOpts, operator common.Address, params IBLSApkRegistryPubkeyRegistrationParams, pubkeyRegistrationMessageHash BN254G1Point) (*types.Transaction, error) {
	return _ContractApkRegistry.contract.Transact(opts, "registerBLSPublicKey", operator, params, pubkeyRegistrationMessageHash)
}

// RegisterBLSPublicKey is a paid mutator transaction binding the contract method 0xbf79ce58.
//
// Solidity: function registerBLSPublicKey(address operator, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) params, (uint256,uint256) pubkeyRegistrationMessageHash) returns(bytes32 operatorId)
func (_ContractApkRegistry *ContractApkRegistrySession) RegisterBLSPublicKey(operator common.Address, params IBLSApkRegistryPubkeyRegistrationParams, pubkeyRegistrationMessageHash BN254G1Point) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.RegisterBLSPublicKey(&_ContractApkRegistry.TransactOpts, operator, params, pubkeyRegistrationMessageHash)
}

// RegisterBLSPublicKey is a paid mutator transaction binding the contract method 0xbf79ce58.
//
// Solidity: function registerBLSPublicKey(address operator, ((uint256,uint256),(uint256,uint256),(uint256[2],uint256[2])) params, (uint256,uint256) pubkeyRegistrationMessageHash) returns(bytes32 operatorId)
func (_ContractApkRegistry *ContractApkRegistryTransactorSession) RegisterBLSPublicKey(operator common.Address, params IBLSApkRegistryPubkeyRegistrationParams, pubkeyRegistrationMessageHash BN254G1Point) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.RegisterBLSPublicKey(&_ContractApkRegistry.TransactOpts, operator, params, pubkeyRegistrationMessageHash)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x3fb27952.
//
// Solidity: function registerOperator(address operator, bytes quorumNumbers) returns()
func (_ContractApkRegistry *ContractApkRegistryTransactor) RegisterOperator(opts *bind.TransactOpts, operator common.Address, quorumNumbers []byte) (*types.Transaction, error) {
	return _ContractApkRegistry.contract.Transact(opts, "registerOperator", operator, quorumNumbers)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x3fb27952.
//
// Solidity: function registerOperator(address operator, bytes quorumNumbers) returns()
func (_ContractApkRegistry *ContractApkRegistrySession) RegisterOperator(operator common.Address, quorumNumbers []byte) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.RegisterOperator(&_ContractApkRegistry.TransactOpts, operator, quorumNumbers)
}

// RegisterOperator is a paid mutator transaction binding the contract method 0x3fb27952.
//
// Solidity: function registerOperator(address operator, bytes quorumNumbers) returns()
func (_ContractApkRegistry *ContractApkRegistryTransactorSession) RegisterOperator(operator common.Address, quorumNumbers []byte) (*types.Transaction, error) {
	return _ContractApkRegistry.Contract.RegisterOperator(&_ContractApkRegistry.TransactOpts, operator, quorumNumbers)
}

// ContractApkRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ContractApkRegistry contract.
type ContractApkRegistryInitializedIterator struct {
	Event *ContractApkRegistryInitialized // Event containing the contract specifics and raw log

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
func (it *ContractApkRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApkRegistryInitialized)
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
		it.Event = new(ContractApkRegistryInitialized)
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
func (it *ContractApkRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApkRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApkRegistryInitialized represents a Initialized event raised by the ContractApkRegistry contract.
type ContractApkRegistryInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ContractApkRegistry *ContractApkRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*ContractApkRegistryInitializedIterator, error) {

	logs, sub, err := _ContractApkRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ContractApkRegistryInitializedIterator{contract: _ContractApkRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ContractApkRegistry *ContractApkRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ContractApkRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _ContractApkRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApkRegistryInitialized)
				if err := _ContractApkRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_ContractApkRegistry *ContractApkRegistryFilterer) ParseInitialized(log types.Log) (*ContractApkRegistryInitialized, error) {
	event := new(ContractApkRegistryInitialized)
	if err := _ContractApkRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApkRegistryNewPubkeyRegistrationIterator is returned from FilterNewPubkeyRegistration and is used to iterate over the raw logs and unpacked data for NewPubkeyRegistration events raised by the ContractApkRegistry contract.
type ContractApkRegistryNewPubkeyRegistrationIterator struct {
	Event *ContractApkRegistryNewPubkeyRegistration // Event containing the contract specifics and raw log

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
func (it *ContractApkRegistryNewPubkeyRegistrationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApkRegistryNewPubkeyRegistration)
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
		it.Event = new(ContractApkRegistryNewPubkeyRegistration)
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
func (it *ContractApkRegistryNewPubkeyRegistrationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApkRegistryNewPubkeyRegistrationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApkRegistryNewPubkeyRegistration represents a NewPubkeyRegistration event raised by the ContractApkRegistry contract.
type ContractApkRegistryNewPubkeyRegistration struct {
	Operator common.Address
	PubkeyG1 BN254G1Point
	PubkeyG2 BN254G2Point
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNewPubkeyRegistration is a free log retrieval operation binding the contract event 0xe3fb6613af2e8930cf85d47fcf6db10192224a64c6cbe8023e0eee1ba3828041.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_ContractApkRegistry *ContractApkRegistryFilterer) FilterNewPubkeyRegistration(opts *bind.FilterOpts, operator []common.Address) (*ContractApkRegistryNewPubkeyRegistrationIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractApkRegistry.contract.FilterLogs(opts, "NewPubkeyRegistration", operatorRule)
	if err != nil {
		return nil, err
	}
	return &ContractApkRegistryNewPubkeyRegistrationIterator{contract: _ContractApkRegistry.contract, event: "NewPubkeyRegistration", logs: logs, sub: sub}, nil
}

// WatchNewPubkeyRegistration is a free log subscription operation binding the contract event 0xe3fb6613af2e8930cf85d47fcf6db10192224a64c6cbe8023e0eee1ba3828041.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_ContractApkRegistry *ContractApkRegistryFilterer) WatchNewPubkeyRegistration(opts *bind.WatchOpts, sink chan<- *ContractApkRegistryNewPubkeyRegistration, operator []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ContractApkRegistry.contract.WatchLogs(opts, "NewPubkeyRegistration", operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApkRegistryNewPubkeyRegistration)
				if err := _ContractApkRegistry.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
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

// ParseNewPubkeyRegistration is a log parse operation binding the contract event 0xe3fb6613af2e8930cf85d47fcf6db10192224a64c6cbe8023e0eee1ba3828041.
//
// Solidity: event NewPubkeyRegistration(address indexed operator, (uint256,uint256) pubkeyG1, (uint256[2],uint256[2]) pubkeyG2)
func (_ContractApkRegistry *ContractApkRegistryFilterer) ParseNewPubkeyRegistration(log types.Log) (*ContractApkRegistryNewPubkeyRegistration, error) {
	event := new(ContractApkRegistryNewPubkeyRegistration)
	if err := _ContractApkRegistry.contract.UnpackLog(event, "NewPubkeyRegistration", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApkRegistryOperatorAddedToQuorumsIterator is returned from FilterOperatorAddedToQuorums and is used to iterate over the raw logs and unpacked data for OperatorAddedToQuorums events raised by the ContractApkRegistry contract.
type ContractApkRegistryOperatorAddedToQuorumsIterator struct {
	Event *ContractApkRegistryOperatorAddedToQuorums // Event containing the contract specifics and raw log

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
func (it *ContractApkRegistryOperatorAddedToQuorumsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApkRegistryOperatorAddedToQuorums)
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
		it.Event = new(ContractApkRegistryOperatorAddedToQuorums)
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
func (it *ContractApkRegistryOperatorAddedToQuorumsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApkRegistryOperatorAddedToQuorumsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApkRegistryOperatorAddedToQuorums represents a OperatorAddedToQuorums event raised by the ContractApkRegistry contract.
type ContractApkRegistryOperatorAddedToQuorums struct {
	Operator      common.Address
	OperatorId    [32]byte
	QuorumNumbers []byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorAddedToQuorums is a free log retrieval operation binding the contract event 0x73a2b7fb844724b971802ae9b15db094d4b7192df9d7350e14eb466b9b22eb4e.
//
// Solidity: event OperatorAddedToQuorums(address operator, bytes32 operatorId, bytes quorumNumbers)
func (_ContractApkRegistry *ContractApkRegistryFilterer) FilterOperatorAddedToQuorums(opts *bind.FilterOpts) (*ContractApkRegistryOperatorAddedToQuorumsIterator, error) {

	logs, sub, err := _ContractApkRegistry.contract.FilterLogs(opts, "OperatorAddedToQuorums")
	if err != nil {
		return nil, err
	}
	return &ContractApkRegistryOperatorAddedToQuorumsIterator{contract: _ContractApkRegistry.contract, event: "OperatorAddedToQuorums", logs: logs, sub: sub}, nil
}

// WatchOperatorAddedToQuorums is a free log subscription operation binding the contract event 0x73a2b7fb844724b971802ae9b15db094d4b7192df9d7350e14eb466b9b22eb4e.
//
// Solidity: event OperatorAddedToQuorums(address operator, bytes32 operatorId, bytes quorumNumbers)
func (_ContractApkRegistry *ContractApkRegistryFilterer) WatchOperatorAddedToQuorums(opts *bind.WatchOpts, sink chan<- *ContractApkRegistryOperatorAddedToQuorums) (event.Subscription, error) {

	logs, sub, err := _ContractApkRegistry.contract.WatchLogs(opts, "OperatorAddedToQuorums")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApkRegistryOperatorAddedToQuorums)
				if err := _ContractApkRegistry.contract.UnpackLog(event, "OperatorAddedToQuorums", log); err != nil {
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

// ParseOperatorAddedToQuorums is a log parse operation binding the contract event 0x73a2b7fb844724b971802ae9b15db094d4b7192df9d7350e14eb466b9b22eb4e.
//
// Solidity: event OperatorAddedToQuorums(address operator, bytes32 operatorId, bytes quorumNumbers)
func (_ContractApkRegistry *ContractApkRegistryFilterer) ParseOperatorAddedToQuorums(log types.Log) (*ContractApkRegistryOperatorAddedToQuorums, error) {
	event := new(ContractApkRegistryOperatorAddedToQuorums)
	if err := _ContractApkRegistry.contract.UnpackLog(event, "OperatorAddedToQuorums", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractApkRegistryOperatorRemovedFromQuorumsIterator is returned from FilterOperatorRemovedFromQuorums and is used to iterate over the raw logs and unpacked data for OperatorRemovedFromQuorums events raised by the ContractApkRegistry contract.
type ContractApkRegistryOperatorRemovedFromQuorumsIterator struct {
	Event *ContractApkRegistryOperatorRemovedFromQuorums // Event containing the contract specifics and raw log

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
func (it *ContractApkRegistryOperatorRemovedFromQuorumsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractApkRegistryOperatorRemovedFromQuorums)
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
		it.Event = new(ContractApkRegistryOperatorRemovedFromQuorums)
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
func (it *ContractApkRegistryOperatorRemovedFromQuorumsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractApkRegistryOperatorRemovedFromQuorumsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractApkRegistryOperatorRemovedFromQuorums represents a OperatorRemovedFromQuorums event raised by the ContractApkRegistry contract.
type ContractApkRegistryOperatorRemovedFromQuorums struct {
	Operator      common.Address
	OperatorId    [32]byte
	QuorumNumbers []byte
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOperatorRemovedFromQuorums is a free log retrieval operation binding the contract event 0xf843ecd53a563675e62107be1494fdde4a3d49aeedaf8d88c616d85346e3500e.
//
// Solidity: event OperatorRemovedFromQuorums(address operator, bytes32 operatorId, bytes quorumNumbers)
func (_ContractApkRegistry *ContractApkRegistryFilterer) FilterOperatorRemovedFromQuorums(opts *bind.FilterOpts) (*ContractApkRegistryOperatorRemovedFromQuorumsIterator, error) {

	logs, sub, err := _ContractApkRegistry.contract.FilterLogs(opts, "OperatorRemovedFromQuorums")
	if err != nil {
		return nil, err
	}
	return &ContractApkRegistryOperatorRemovedFromQuorumsIterator{contract: _ContractApkRegistry.contract, event: "OperatorRemovedFromQuorums", logs: logs, sub: sub}, nil
}

// WatchOperatorRemovedFromQuorums is a free log subscription operation binding the contract event 0xf843ecd53a563675e62107be1494fdde4a3d49aeedaf8d88c616d85346e3500e.
//
// Solidity: event OperatorRemovedFromQuorums(address operator, bytes32 operatorId, bytes quorumNumbers)
func (_ContractApkRegistry *ContractApkRegistryFilterer) WatchOperatorRemovedFromQuorums(opts *bind.WatchOpts, sink chan<- *ContractApkRegistryOperatorRemovedFromQuorums) (event.Subscription, error) {

	logs, sub, err := _ContractApkRegistry.contract.WatchLogs(opts, "OperatorRemovedFromQuorums")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractApkRegistryOperatorRemovedFromQuorums)
				if err := _ContractApkRegistry.contract.UnpackLog(event, "OperatorRemovedFromQuorums", log); err != nil {
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

// ParseOperatorRemovedFromQuorums is a log parse operation binding the contract event 0xf843ecd53a563675e62107be1494fdde4a3d49aeedaf8d88c616d85346e3500e.
//
// Solidity: event OperatorRemovedFromQuorums(address operator, bytes32 operatorId, bytes quorumNumbers)
func (_ContractApkRegistry *ContractApkRegistryFilterer) ParseOperatorRemovedFromQuorums(log types.Log) (*ContractApkRegistryOperatorRemovedFromQuorums, error) {
	event := new(ContractApkRegistryOperatorRemovedFromQuorums)
	if err := _ContractApkRegistry.contract.UnpackLog(event, "OperatorRemovedFromQuorums", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
