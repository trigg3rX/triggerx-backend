package types

import (
	"encoding/json"
)

// ConvertJobDataEntityToDTO converts JobDataEntity to JobDataDTO
func ConvertJobDataEntityToDTO(entity *JobDataEntity) *JobDataDTO {
	if entity == nil {
		return nil
	}
	return &JobDataDTO{
		JobID:             entity.JobID,
		JobTitle:          entity.JobTitle,
		TaskDefinitionID:  entity.TaskDefinitionID,
		CreatedChainID:    entity.CreatedChainID,
		UserAddress:       entity.UserAddress,
		LinkJobID:         entity.LinkJobID,
		ChainStatus:       entity.ChainStatus,
		Timezone:          entity.Timezone,
		IsImua:            entity.IsImua,
		JobType:           entity.JobType,
		TimeFrame:         entity.TimeFrame,
		Recurring:         entity.Recurring,
		Status:            entity.Status,
		JobCostPrediction: entity.JobCostPrediction,
		JobCostActual:     entity.JobCostActual,
		TaskIDs:           entity.TaskIDs,
		CreatedAt:         entity.CreatedAt,
		UpdatedAt:         entity.UpdatedAt,
		LastExecutedAt:    entity.LastExecutedAt,
	}
}

// ConvertTimeJobDataEntityToDTO converts TimeJobDataEntity to TimeJobDataDTO
func ConvertTimeJobDataEntityToDTO(entity *TimeJobDataEntity) *TimeJobDataDTO {
	if entity == nil {
		return nil
	}
	return &TimeJobDataDTO{
		JobID:                     entity.JobID,
		TaskDefinitionID:          entity.TaskDefinitionID,
		ScheduleType:              entity.ScheduleType,
		TimeInterval:              entity.TimeInterval,
		CronExpression:            entity.CronExpression,
		SpecificSchedule:          entity.SpecificSchedule,
		NextExecutionTimestamp:    entity.NextExecutionTimestamp,
		TargetChainID:             entity.TargetChainID,
		TargetContractAddress:     entity.TargetContractAddress,
		TargetFunction:            entity.TargetFunction,
		ABI:                       entity.ABI,
		ArgType:                   entity.ArgType,
		Arguments:                 entity.Arguments,
		DynamicArgumentsScriptURL: entity.DynamicArgumentsScriptURL,
		IsCompleted:               entity.IsCompleted,
		LastExecutedAt:            entity.LastExecutedAt,
		ExpirationTime:            entity.ExpirationTime,
	}
}

// ConvertEventJobDataEntityToDTO converts EventJobDataEntity to EventJobDataDTO
func ConvertEventJobDataEntityToDTO(entity *EventJobDataEntity) *EventJobDataDTO {
	if entity == nil {
		return nil
	}
	return &EventJobDataDTO{
		JobID:                      entity.JobID,
		TaskDefinitionID:           entity.TaskDefinitionID,
		Recurring:                  entity.Recurring,
		TriggerChainID:             entity.TriggerChainID,
		TriggerContractAddress:     entity.TriggerContractAddress,
		TriggerEvent:               entity.TriggerEvent,
		TriggerEventFilterParaName: entity.TriggerEventFilterParaName,
		TriggerEventFilterValue:    entity.TriggerEventFilterValue,
		TargetChainID:              entity.TargetChainID,
		TargetContractAddress:      entity.TargetContractAddress,
		TargetFunction:             entity.TargetFunction,
		ABI:                        entity.ABI,
		ArgType:                    entity.ArgType,
		Arguments:                  entity.Arguments,
		DynamicArgumentsScriptURL:  entity.DynamicArgumentsScriptURL,
		IsCompleted:                entity.IsCompleted,
		LastExecutedAt:             entity.LastExecutedAt,
		ExpirationTime:             entity.ExpirationTime,
	}
}

// ConvertConditionJobDataEntityToDTO converts ConditionJobDataEntity to ConditionJobDataDTO
func ConvertConditionJobDataEntityToDTO(entity *ConditionJobDataEntity) *ConditionJobDataDTO {
	if entity == nil {
		return nil
	}
	return &ConditionJobDataDTO{
		JobID:                     entity.JobID,
		TaskDefinitionID:          entity.TaskDefinitionID,
		Recurring:                 entity.Recurring,
		ConditionType:             entity.ConditionType,
		UpperLimit:                entity.UpperLimit,
		LowerLimit:                entity.LowerLimit,
		ValueSourceType:           entity.ValueSourceType,
		ValueSourceURL:            entity.ValueSourceURL,
		SelectedKeyRoute:          entity.SelectedKeyRoute,
		TargetChainID:             entity.TargetChainID,
		TargetContractAddress:     entity.TargetContractAddress,
		TargetFunction:            entity.TargetFunction,
		ABI:                       entity.ABI,
		ArgType:                   entity.ArgType,
		Arguments:                 entity.Arguments,
		DynamicArgumentsScriptURL: entity.DynamicArgumentsScriptURL,
		IsCompleted:               entity.IsCompleted,
		LastExecutedAt:            entity.LastExecutedAt,
		ExpirationTime:            entity.ExpirationTime,
	}
}

// ConvertUserDataEntityToDTO converts UserDataEntity to UserDataDTO
func ConvertUserDataEntityToDTO(entity *UserDataEntity) *UserDataDTO {
	if entity == nil {
		return nil
	}

	return &UserDataDTO{
		UserAddress:   entity.UserAddress,
		EmailID:       entity.EmailID,
		JobIDs:        entity.JobIDs,
		UserPoints:    entity.UserPoints,
		TotalJobs:     entity.TotalJobs,
		TotalTasks:    entity.TotalTasks,
		CreatedAt:     entity.CreatedAt,
		LastUpdatedAt: entity.LastUpdatedAt,
	}
}

// ConvertTaskDataEntityToDTO converts TaskDataEntity to TaskDataDTO
func ConvertTaskDataEntityToDTO(entity *TaskDataEntity) *TaskDataDTO {
	if entity == nil {
		return nil
	}

	// Convert ConvertedArguments from string to []interface{}
	var convertedArgs []interface{}
	if entity.ConvertedArguments != "" {
		_ = json.Unmarshal([]byte(entity.ConvertedArguments), &convertedArgs)
	}

	return &TaskDataDTO{
		TaskID:               entity.TaskID,
		TaskNumber:           entity.TaskNumber,
		JobID:                entity.JobID,
		TaskDefinitionID:     entity.TaskDefinitionID,
		CreatedAt:            entity.CreatedAt,
		TaskOpxPredictedCost: entity.TaskOpxPredictedCost,
		TaskOpxActualCost:    entity.TaskOpxActualCost,
		ExecutionTimestamp:   entity.ExecutionTimestamp,
		ExecutionTxHash:      entity.ExecutionTxHash,
		TaskPerformerID:      entity.TaskPerformerID,
		TaskAttesterIDs:      entity.TaskAttesterIDs,
		ConvertedArguments:   convertedArgs,
		ProofOfTask:          entity.ProofOfTask,
		SubmissionTxHash:     entity.SubmissionTxHash,
		IsSuccessful:         entity.IsSuccessful,
		IsAccepted:           entity.IsAccepted,
		IsImua:               entity.IsImua,
	}
}

// ConvertKeeperDataEntityToDTO converts KeeperDataEntity to KeeperDataDTO
func ConvertKeeperDataEntityToDTO(entity *KeeperDataEntity) *KeeperDataDTO {
	if entity == nil {
		return nil
	}

	return &KeeperDataDTO{
		KeeperName:       entity.KeeperName,
		KeeperAddress:    entity.KeeperAddress,
		RewardsAddress:   entity.RewardsAddress,
		ConsensusAddress: entity.ConsensusAddress,
		RegisteredTx:     entity.RegisteredTx,
		OperatorID:       entity.OperatorID,
		VotingPower:      entity.VotingPower,
		Whitelisted:      entity.Whitelisted,
		Registered:       entity.Registered,
		Online:           entity.Online,
		Version:          entity.Version,
		OnImua:           entity.OnImua,
		PublicIP:         entity.PublicIP,
		ChatID:           entity.ChatID,
		EmailID:          entity.EmailID,
		RewardsBooster:   entity.RewardsBooster,
		NoExecutedTasks:  entity.NoExecutedTasks,
		NoAttestedTasks:  entity.NoAttestedTasks,
		Uptime:           entity.Uptime,
		KeeperPoints:     entity.KeeperPoints,
		LastCheckedIn:    entity.LastCheckedIn,
	}
}

// ConvertApiKeyDataEntityToDTO converts ApiKeyDataEntity to ApiKeyDataDTO
func ConvertApiKeyDataEntityToDTO(entity *ApiKeyDataEntity) *ApiKeyDataDTO {
	if entity == nil {
		return nil
	}
	return &ApiKeyDataDTO{
		Key:          entity.Key,
		Owner:        entity.Owner,
		IsActive:     entity.IsActive,
		RateLimit:    entity.RateLimit,
		SuccessCount: entity.SuccessCount,
		FailedCount:  entity.FailedCount,
		LastUsed:     entity.LastUsed,
		CreatedAt:    entity.CreatedAt,
	}
}
