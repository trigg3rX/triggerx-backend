package attestation

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// AttestationCenterABI contains the ABI for AttestationCenter contract events
const AttestationCenterABI = `[
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "operator",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint32",
				"name": "taskNumber",
				"type": "uint32"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "proofOfTask",
				"type": "string"
			},
			{
				"indexed": false,
				"internalType": "bytes",
				"name": "data",
				"type": "bytes"
			},
			{
				"indexed": true,
				"internalType": "uint16",
				"name": "taskDefinitionId",
				"type": "uint16"
			},
			{
				"indexed": false,
				"internalType": "uint256[]",
				"name": "attestersIds",
				"type": "uint256[]"
			}
		],
		"name": "TaskSubmitted",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "operator",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint32",
				"name": "taskNumber",
				"type": "uint32"
			},
			{
				"indexed": false,
				"internalType": "string",
				"name": "proofOfTask",
				"type": "string"
			},
			{
				"indexed": false,
				"internalType": "bytes",
				"name": "data",
				"type": "bytes"
			},
			{
				"indexed": true,
				"internalType": "uint16",
				"name": "taskDefinitionId",
				"type": "uint16"
			},
			{
				"indexed": false,
				"internalType": "uint256[]",
				"name": "attestersIds",
				"type": "uint256[]"
			}
		],
		"name": "TaskRejected",
		"type": "event"
	}
]`

var (
	// AttestationCenterParsedABI is the parsed ABI for AttestationCenter contract
	AttestationCenterParsedABI = func() abi.ABI {
		parsed, err := abi.JSON(strings.NewReader(AttestationCenterABI))
		if err != nil {
			panic("failed to parse AttestationCenter ABI: " + err.Error())
		}
		return parsed
	}()
	// TaskSubmittedEvent is the TaskSubmitted event from the ABI
	TaskSubmittedEvent = AttestationCenterParsedABI.Events["TaskSubmitted"]
	// TaskRejectedEvent is the TaskRejected event from the ABI
	TaskRejectedEvent = AttestationCenterParsedABI.Events["TaskRejected"]
)
