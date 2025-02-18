package types

type IPFSData struct {
	JobData 	Job 		`json:"job_data"`
	
	TriggerData TriggerData `json:"trigger_data"`
	
	ActionData 	ActionData 	`json:"action_data"`

	ProofData 	ProofData 	`json:"proof_data"`
}