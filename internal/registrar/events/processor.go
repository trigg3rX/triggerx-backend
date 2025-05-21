package events

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// EventProcessor handles blockchain event processing
type EventProcessor struct {
	logger logging.Logger
}

// NewEventProcessor creates a new event processor instance
func NewEventProcessor(logger logging.Logger) *EventProcessor {
	if logger == nil {
		panic("logger cannot be nil")
	}

	// Add component tag to logger
	logger = logger.With("component", "event_processor")

	return &EventProcessor{
		logger: logger,
	}
}

func init() {
	avsGovernanceABIJSON, err := os.ReadFile("pkg/bindings/abi/AvsGovernance.json")
	if err != nil {
		panic("failed to read AvsGovernance ABI: " + err.Error())
	}
	AvsGovernanceABI, err = abi.JSON(strings.NewReader(string(avsGovernanceABIJSON)))
	if err != nil {
		panic("failed to parse AvsGovernance ABI: " + err.Error())
	}

	attestationCenterABIJSON, err := os.ReadFile("pkg/bindings/abi/AttestationCenter.json")
	if err != nil {
		panic("failed to read AttestationCenter ABI: " + err.Error())
	}
	AttestationCenterABI, err = abi.JSON(strings.NewReader(string(attestationCenterABIJSON)))
	if err != nil {
		panic("failed to parse AttestationCenter ABI: " + err.Error())
	}
}
