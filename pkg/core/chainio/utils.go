package chainio

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	registrycoordinator "github.com/trigg3rX/triggerx-contracts/bindings/contracts/RegistryCoordinator"
	txservicemanager "github.com/trigg3rX/triggerx-contracts/bindings/contracts/TriggerXServiceManager"
	txtaskmanager "github.com/trigg3rX/triggerx-contracts/bindings/contracts/TriggerXTaskManager"
	"golang.org/x/crypto/sha3"
)

// this hardcodes abi.encode() for txtaskmanager.ITriggerXTaskManagerTaskResponse
func AbiEncodeTaskResponse(h *txtaskmanager.ITriggerXTaskManagerTaskResponse) ([]byte, error) {
	taskResponseType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{
			Name: "referenceTaskIndex",
			Type: "uint32",
		},
		{
			Name: "numberSquared",
			Type: "uint256",
		},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{
		{
			Type: taskResponseType,
		},
	}

	bytes, err := arguments.Pack(h)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// GetTaskResponseDigest returns the hash of the TaskResponse, which is what operators sign over
func GetTaskResponseDigest(h *txtaskmanager.ITriggerXTaskManagerTaskResponse) ([32]byte, error) {

	encodeTaskResponseByte, err := AbiEncodeTaskResponse(h)
	if err != nil {
		return [32]byte{}, err
	}

	var taskResponseDigest [32]byte
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(encodeTaskResponseByte)
	copy(taskResponseDigest[:], hasher.Sum(nil)[:32])

	return taskResponseDigest, nil
}

// ABI encoding for BLS registry types
func AbiEncodePubkeyRegistrationParams(params *registrycoordinator.IBLSApkRegistryPubkeyRegistrationParams) ([]byte, error) {
	pubkeyRegistrationParamsType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "pubkeyG1", Type: "tuple", Components: []abi.ArgumentMarshaling{
			{Name: "X", Type: "uint256"},
			{Name: "Y", Type: "uint256"},
		}},
		{Name: "pubkeyG2", Type: "tuple", Components: []abi.ArgumentMarshaling{
			{Name: "X", Type: "uint256[2]"},
			{Name: "Y", Type: "uint256[2]"},
		}},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: pubkeyRegistrationParamsType}}
	return arguments.Pack(params)
}

func AbiEncodeOperatorKickParams(params *registrycoordinator.IRegistryCoordinatorOperatorKickParam) ([]byte, error) {
	operatorKickParamsType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "operator", Type: "address"},
		{Name: "quorumNumbers", Type: "bytes"},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: operatorKickParamsType}}
	return arguments.Pack(params)
}

// ABI encoding for signature types
func AbiEncodeSignatureWithSaltAndExpiry(sig *registrycoordinator.ISignatureUtilsSignatureWithSaltAndExpiry) ([]byte, error) {
	signatureType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "signature", Type: "bytes"},
		{Name: "salt", Type: "bytes32"},
		{Name: "expiry", Type: "uint256"},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: signatureType}}
	return arguments.Pack(sig)
}

// ABI encoding for rewards types
func AbiEncodeRewardsSubmission(submission *txservicemanager.IRewardsCoordinatorTypesRewardsSubmission) ([]byte, error) {
	rewardsSubmissionType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "rewardToken", Type: "address"},
		{Name: "amount", Type: "uint256"},
		{Name: "recipient", Type: "address"},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: rewardsSubmissionType}}
	return arguments.Pack(submission)
}

func AbiEncodeOperatorDirectedRewardsSubmission(submission *txservicemanager.IRewardsCoordinatorTypesRewardsSubmission) ([]byte, error) {
	operatorDirectedRewardsSubmissionType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "rewardToken", Type: "address"},
		{Name: "amount", Type: "uint256"},
		{Name: "operator", Type: "address"},
		{Name: "recipient", Type: "address"},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: operatorDirectedRewardsSubmissionType}}
	return arguments.Pack(submission)
}

// ABI encoding for BLS signature checking types
func AbiEncodeNonSignerStakesAndSignature(data *txtaskmanager.IBLSSignatureCheckerNonSignerStakesAndSignature) ([]byte, error) {
	nonSignerStakesAndSignatureType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "nonSignerQuorumBitmapIndices", Type: "uint32[]"},
		{Name: "nonSignerPubkeys", Type: "tuple[]", Components: []abi.ArgumentMarshaling{
			{Name: "X", Type: "uint256"},
			{Name: "Y", Type: "uint256"},
		}},
		{Name: "quorumApks", Type: "tuple[]", Components: []abi.ArgumentMarshaling{
			{Name: "X", Type: "uint256"},
			{Name: "Y", Type: "uint256"},
		}},
		{Name: "apkG2", Type: "tuple", Components: []abi.ArgumentMarshaling{
			{Name: "X", Type: "uint256[2]"},
			{Name: "Y", Type: "uint256[2]"},
		}},
		{Name: "sigma", Type: "tuple", Components: []abi.ArgumentMarshaling{
			{Name: "X", Type: "uint256"},
			{Name: "Y", Type: "uint256"},
		}},
		{Name: "quorumApkIndices", Type: "uint32[]"},
		{Name: "totalStakeIndices", Type: "uint32[]"},
		{Name: "nonSignerStakeIndices", Type: "uint32[][]"},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: nonSignerStakesAndSignatureType}}
	return arguments.Pack(data)
}

func AbiEncodeQuorumStakeTotals(totals *txtaskmanager.IBLSSignatureCheckerQuorumStakeTotals) ([]byte, error) {
	quorumStakeTotalsType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "signedStakeForQuorum", Type: "uint96[]"},
		{Name: "totalStakeForQuorum", Type: "uint96[]"},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: quorumStakeTotalsType}}
	return arguments.Pack(totals)
}

// ABI encoding for BN254 point types
func AbiEncodeBN254G1Point(point *txtaskmanager.BN254G1Point) ([]byte, error) {
	g1PointType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "X", Type: "uint256"},
		{Name: "Y", Type: "uint256"},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: g1PointType}}
	return arguments.Pack(point)
}

func AbiEncodeBN254G2Point(point *txtaskmanager.BN254G2Point) ([]byte, error) {
	g2PointType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "X", Type: "uint256[2]"},
		{Name: "Y", Type: "uint256[2]"},
	})
	if err != nil {
		return nil, err
	}
	arguments := abi.Arguments{{Type: g2PointType}}
	return arguments.Pack(point)
}
