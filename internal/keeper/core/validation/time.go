package validation

import (
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// ValidateTimeBasedJob checks if a time-based job (task definitions 1 and 2) should be executed
// based on its time interval, timeframe, and last execution time
func (v *TaskValidator) ValidateTimeBasedTask(ipfsData types.IPFSData) (bool, error) {
	
	return false, nil
}
