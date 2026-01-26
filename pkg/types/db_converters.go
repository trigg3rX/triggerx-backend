package types

// ApiKeyDataEntityToDTO converts ApiKeyDataEntity to ApiKeyDataDTO.
func ApiKeyDataEntityToDTO(entity *ApiKeyDataEntity) *ApiKeyDataDTO {
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

// UserDataEntityToDTO converts UserDataEntity to UserDataDTO.
func UserDataEntityToDTO(entity *UserDataEntity) *UserDataDTO {
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

// SafeAddressDataEntityToDTO converts SafeAddressDataEntity to SafeAddressDataDTO. (unused)
func SafeAddressDataEntityToDTO(entity *SafeAddressDataEntity) *SafeAddressDataDTO {
	if entity == nil {
		return nil
	}
	return &SafeAddressDataDTO{
		UserAddress: entity.UserAddress,
		SafeAddress: entity.SafeAddress,
		SafeName:    entity.SafeName,
		CreatedAt:   entity.CreatedAt,
	}
}

// JobDataEntityToDTO converts JobDataEntity to JobDataDTO.
func JobDataEntityToDTO(entity *JobDataEntity) *JobDataDTO {
	if entity == nil {
		return nil
	}
	return &JobDataDTO{
		JobID:             entity.JobID,
		JobTitle:          entity.JobTitle,
		TaskDefinitionID:  TaskDefinitionID(entity.TaskDefinitionID),
		CreatedChainID:    entity.CreatedChainID,
		UserAddress:       entity.UserAddress,
		LinkJobID:         entity.LinkJobID,
		ChainStatus:       ChainStatus(entity.ChainStatus),
		SafeAddress:       entity.SafeAddress,
		Timezone:          entity.Timezone,
		JobType:           JobType(entity.JobType),
		TimeFrame:         entity.TimeFrame,
		Recurring:         entity.Recurring,
		Status:            JobStatus(entity.Status),
		JobCostPrediction: entity.JobCostPrediction,
		JobCostActual:     entity.JobCostActual,
		TaskIDs:           entity.TaskIDs,
		CreatedAt:         entity.CreatedAt,
		UpdatedAt:         entity.UpdatedAt,
		LastExecutedAt:    entity.LastExecutedAt,
	}
}

// TimeJobDataEntityToDTO converts TimeJobDataEntity to TimeJobDataDTO.
func TimeJobDataEntityToDTO(entity *TimeJobDataEntity) *TimeJobDataDTO {
	if entity == nil {
		return nil
	}
	return &TimeJobDataDTO{
		JobID:                     entity.JobID,
		TaskDefinitionID:          TaskDefinitionID(entity.TaskDefinitionID),
		Network:                   Network(entity.Network),
		ScheduleType:              ScheduleType(entity.ScheduleType),
		TimeInterval:              entity.TimeInterval,
		CronExpression:            entity.CronExpression,
		SpecificSchedule:          entity.SpecificSchedule,
		Timezone:                  entity.Timezone,
		NextExecutionTimestamp:    entity.NextExecutionTimestamp,
		TargetChainID:             entity.TargetChainID,
		TargetContractAddress:     entity.TargetContractAddress,
		TargetFunction:            entity.TargetFunction,
		ABI:                       entity.ABI,
		ArgType:                   ArgType(entity.ArgType),
		Arguments:                 entity.Arguments,
		ExecutionScriptURL:        entity.ExecutionScriptURL,
		ExecutionScriptLanguage:   ScriptLanguage(entity.ExecutionScriptLanguage),
		ExecutionScriptHash:       entity.ExecutionScriptHash,
		MaxExecutionTime:          entity.MaxExecutionTime,
		ChallengePeriod:           entity.ChallengePeriod,
		IsActive:                  entity.IsActive,
		LastExecutedAt:            entity.LastExecutedAt,
		ExpirationTime:            entity.ExpirationTime,
	}
}

// EventJobDataEntityToDTO converts EventJobDataEntity to EventJobDataDTO.
func EventJobDataEntityToDTO(entity *EventJobDataEntity) *EventJobDataDTO {
	if entity == nil {
		return nil
	}
	return &EventJobDataDTO{
		JobID:                     entity.JobID,
		TaskDefinitionID:          TaskDefinitionID(entity.TaskDefinitionID),
		Network:                   Network(entity.Network),
		Recurring:                 entity.Recurring,
		TriggerChainID:            entity.TriggerChainID,
		TriggerContractAddress:    entity.TriggerContractAddress,
		TriggerEvent:              entity.TriggerEvent,
		EventFilterParaName:       entity.EventFilterParaName,
		EventFilterValue:          entity.EventFilterValue,
		TargetChainID:             entity.TargetChainID,
		TargetContractAddress:     entity.TargetContractAddress,
		TargetFunction:            entity.TargetFunction,
		ABI:                       entity.ABI,
		ArgType:                   ArgType(entity.ArgType),
		Arguments:                 entity.Arguments,
		ExecutionScriptURL:        entity.ExecutionScriptURL,
		ExecutionScriptLanguage:   ScriptLanguage(entity.ExecutionScriptLanguage),
		ExecutionScriptHash:       entity.ExecutionScriptHash,
		MaxExecutionTime:          entity.MaxExecutionTime,
		ChallengePeriod:           entity.ChallengePeriod,
		IsActive:                  entity.IsActive,
		LastExecutedAt:            entity.LastExecutedAt,
		ExpirationTime:            entity.ExpirationTime,
	}
}

// ConditionJobDataEntityToDTO converts ConditionJobDataEntity to ConditionJobDataDTO.
func ConditionJobDataEntityToDTO(entity *ConditionJobDataEntity) *ConditionJobDataDTO {
	if entity == nil {
		return nil
	}
	return &ConditionJobDataDTO{
		JobID:                     entity.JobID,
		TaskDefinitionID:          TaskDefinitionID(entity.TaskDefinitionID),
		Network:                   Network(entity.Network),
		Recurring:                 entity.Recurring,
		ConditionType:             entity.ConditionType,
		UpperLimit:                entity.UpperLimit,
		LowerLimit:                entity.LowerLimit,
		ValueSourceType:           ValueSourceType(entity.ValueSourceType),
		ValueSourceURL:            entity.ValueSourceURL,
		SelectedKeyRoute:          entity.SelectedKeyRoute,
		TargetChainID:             entity.TargetChainID,
		TargetContractAddress:     entity.TargetContractAddress,
		TargetFunction:            entity.TargetFunction,
		ABI:                       entity.ABI,
		ArgType:                   ArgType(entity.ArgType),
		Arguments:                 entity.Arguments,
		ExecutionScriptURL:        entity.ExecutionScriptURL,
		ExecutionScriptLanguage:   ScriptLanguage(entity.ExecutionScriptLanguage),
		ExecutionScriptHash:       entity.ExecutionScriptHash,
		MaxExecutionTime:          entity.MaxExecutionTime,
		ChallengePeriod:           entity.ChallengePeriod,
		IsActive:                  entity.IsActive,
		LastExecutedAt:            entity.LastExecutedAt,
		ExpirationTime:            entity.ExpirationTime,
	}
}

// TaskDataEntityToDTO converts TaskDataEntity to TaskDataDTO.
func TaskDataEntityToDTO(entity *TaskDataEntity) *TaskDataDTO {
	if entity == nil {
		return nil
	}
	return &TaskDataDTO{
		TaskID:               entity.TaskID,
		TaskNumber:           entity.TaskNumber,
		TaskStatus:           TaskStatus(entity.TaskStatus),
		TaskError:            entity.TaskError,
		JobID:                entity.JobID,
		TaskDefinitionID:     TaskDefinitionID(entity.TaskDefinitionID),
		CreatedAt:            entity.CreatedAt,
		ExecutedAt:           entity.ExecutedAt,
		SubmittedAt:          entity.SubmittedAt,
		TaskOpxPredictedCost: entity.TaskOpxPredictedCost,
		TaskOpxActualCost:    entity.TaskOpxActualCost,
		ExecutionTxHash:      entity.ExecutionTxHash,
		SubmissionTxHash:     entity.SubmissionTxHash,
		ConvertedArguments:   entity.ConvertedArguments,
		TaskPerformerAddress: entity.TaskPerformerAddress,
		TaskAttesterAddress:  entity.TaskAttesterAddress,
		ProofOfTask:          entity.ProofOfTask,
		IsSuccessful:         entity.IsSuccessful,
		IsAccepted:           entity.IsAccepted,
		Network:              Network(entity.Network),
	}
}

// KeeperDataEntityToDTO converts KeeperDataEntity to KeeperDataDTO.
func KeeperDataEntityToDTO(entity *KeeperDataEntity) *KeeperDataDTO {
	if entity == nil {
		return nil
	}
	return &KeeperDataDTO{
		KeeperAddress:     entity.KeeperAddress,
		KeeperName:        entity.KeeperName,
		RewardsAddress:    entity.RewardsAddress,
		ConsensusAddress:  entity.ConsensusAddress,
		OperatorID:        entity.OperatorID,
		VotingPower:       entity.VotingPower,
		Registered:        entity.Registered,
		RegisteredAt:      entity.RegisteredAt,
		Online:            entity.Online,
		Version:           entity.Version,
		Network:           Network(entity.Network),
		ConnectionAddress: entity.ConnectionAddress,
		PeerID:            entity.PeerID,
		Uptime:            entity.Uptime,
		LastCheckedIn:     entity.LastCheckedIn,
		RewardsBooster:    entity.RewardsBooster,
		NoExecutedTasks:   entity.NoExecutedTasks,
		NoAttestedTasks:   entity.NoAttestedTasks,
		KeeperPoints:      entity.KeeperPoints,
		ChatID:            entity.ChatID,
		EmailID:           entity.EmailID,
	}
}

// AgentScriptExecutionsEntityToDTO converts AgentScriptExecutionsEntity to AgentScriptExecutionsDTO.
func AgentScriptExecutionsEntityToDTO(entity *AgentScriptExecutionsEntity) *AgentScriptExecutionsDTO {
	if entity == nil {
		return nil
	}
	return &AgentScriptExecutionsDTO{
		ExecutionID:        entity.ExecutionID,
		JobID:              entity.JobID,
		TaskID:             entity.TaskID,
		TaskDefinitionID:   entity.TaskDefinitionID,
		ScheduledTime:      entity.ScheduledTime,
		ActualTime:         entity.ActualTime,
		PerformerAddress:   entity.PerformerAddress,
		InputTimestamp:     entity.InputTimestamp,
		InputStorage:       entity.InputStorage,
		InputHash:          entity.InputHash,
		TriggerData:        entity.TriggerData,
		ShouldExecute:      entity.ShouldExecute,
		TargetContract:     entity.TargetContract,
		Calldata:           entity.Calldata,
		OutputHash:         entity.OutputHash,
		ExecutionMetadata:  entity.ExecutionMetadata,
		ScriptHash:         entity.ScriptHash,
		Signature:          entity.Signature,
		TxHash:             entity.TxHash,
		ExecutionStatus:    entity.ExecutionStatus,
		ExecutionError:     entity.ExecutionError,
		VerificationStatus: VerificationStatus(entity.VerificationStatus),
		ChallengeDeadline:  entity.ChallengeDeadline,
		IsChallenged:       entity.IsChallenged,
		ChallengeCount:     entity.ChallengeCount,
		CreatedAt:          entity.CreatedAt,
	}
}

// ScriptStorageEntityToDTO converts ScriptStorageEntity to ScriptStorageDTO.
func ScriptStorageEntityToDTO(entity *ScriptStorageEntity) *ScriptStorageDTO {
	if entity == nil {
		return nil
	}
	return &ScriptStorageDTO{
		JobID:        entity.JobID,
		StorageKey:   entity.StorageKey,
		StorageValue: entity.StorageValue,
		UpdatedAt:    entity.UpdatedAt,
	}
}

// ExecutionChallengesEntityToDTO converts ExecutionChallengesEntity to ExecutionChallengesDTO.
func ExecutionChallengesEntityToDTO(entity *ExecutionChallengesEntity) *ExecutionChallengesDTO {
	if entity == nil {
		return nil
	}
	return &ExecutionChallengesDTO{
		ChallengeID:              entity.ChallengeID,
		ExecutionID:              entity.ExecutionID,
		ChallengerAddress:        entity.ChallengerAddress,
		ChallengeReason:          ChallengeReason(entity.ChallengeReason),
		ChallengerOutputHash:     entity.ChallengerOutputHash,
		ChallengerShouldExecute:  entity.ChallengerShouldExecute,
		ChallengerTargetContract: entity.ChallengerTargetContract,
		ChallengerCalldata:       entity.ChallengerCalldata,
		ChallengerSignature:      entity.ChallengerSignature,
		BondAmount:               entity.BondAmount,
		ResolutionStatus:         ResolutionStatus(entity.ResolutionStatus),
		ResolutionTime:           entity.ResolutionTime,
		ValidatorCount:           entity.ValidatorCount,
		ApproveCount:             entity.ApproveCount,
		RejectCount:              entity.RejectCount,
		CreatedAt:                entity.CreatedAt,
	}
}
