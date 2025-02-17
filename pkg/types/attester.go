package types

type IPFSResponse struct {
	JobID            string `json:"job_id"`
	JobType          string `json:"job_type"`
	TaskID           string `json:"task_id"`
	TaskDefinitionID string `json:"task_definition_id"`
	Trigger          struct {
		Timestamp               string `json:"timestamp"`
		Value                   string `json:"value"`
		TxHash                  string `json:"txHash"`
		EventName               string `json:"eventName"`
		ConditionEndpoint       string `json:"conditionEndpoint"`
		ConditionValue          string `json:"conditionValue"`
		CustomTriggerDefinition struct {
			Type   string                 `json:"type"`
			Params map[string]interface{} `json:"params"`
		} `json:"customTriggerDefinition"`
	} `json:"trigger"`
	Action struct {
		Timestamp string `json:"timestamp"`
		TxHash    string `json:"txHash"`
		GasUsed   string `json:"gasUsed"`
		Status    string `json:"status"`
	} `json:"action"`
	Proof struct {
		CertificateHash string `json:"certificateHash"`
		ResponseHash    string `json:"responseHash"`
		Timestamp       string `json:"timestamp"`
	} `json:"proof"`
}