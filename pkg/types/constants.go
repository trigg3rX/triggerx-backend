package types

type TaskDefinitionID int

const (
	TaskDefTimeBasedStatic TaskDefinitionID = 1
	TaskDefTimeBasedDynamic TaskDefinitionID = 2
	TaskDefEventBasedStatic TaskDefinitionID = 3
	TaskDefEventBasedDynamic TaskDefinitionID = 4
	TaskDefConditionBasedStatic TaskDefinitionID = 5
	TaskDefConditionBasedDynamic TaskDefinitionID = 6
)

type JobStatus string

const (
	JobStatusCreated JobStatus = "created"      // Created but not yet scheduled
	JobStatusRunning JobStatus = "running"      // Scheduled and running
	JobStatusCompleted JobStatus = "completed"  // Ran with atleast one task completed
	JobStatusFailed JobStatus = "failed"        // Ran with no successful task execution
	JobStatusExpired JobStatus = "expired"      // Ran with no triggers in time frame
	JobStatusDeleted JobStatus = "deleted"      // Deleted by user (was active when deleted)
)

type JobChainStatus int

const (
	JobChainStatusNone JobChainStatus = 0
	JobChainStatusHead JobChainStatus = 1
	JobChainStatusBlock JobChainStatus = 2
)

type JobType string

const (
	JobTypeSDK JobType = "sdk"
	JobTypeFrontend JobType = "frontend"
	JobTypeContract JobType = "contract"
	JobTypeTemplate JobType = "template"
)

type JobArgType int

const (
	JobArgTypeNone JobArgType = 0
	JobArgTypeStatic JobArgType = 1
	JobArgTypeDynamic JobArgType = 2
)

type TimeJobScheduleType string

const (
	TimeJobScheduleTypeInterval TimeJobScheduleType = "interval"
	TimeJobScheduleTypeCron TimeJobScheduleType = "cron"
	TimeJobScheduleTypeSpecific TimeJobScheduleType = "specific"
)

type ConditionJobConditionType string

const (
	ConditionJobConditionTypeEquals ConditionJobConditionType = "equals"
	ConditionJobConditionTypeNotEquals ConditionJobConditionType = "not_equals"
	ConditionJobConditionTypeGreaterThan ConditionJobConditionType = "greater_than"
	ConditionJobConditionTypeGreaterEqual ConditionJobConditionType = "greater_equal"
	ConditionJobConditionTypeLessThan ConditionJobConditionType = "less_than"
	ConditionJobConditionTypeLessEqual ConditionJobConditionType = "less_equal"
	ConditionJobConditionTypeBetween ConditionJobConditionType = "between"
)

type ConditionJobValueSourceType string

const (
	ConditionJobValueSourceTypeAPI ConditionJobValueSourceType = "api"
	ConditionJobValueSourceTypeOracle ConditionJobValueSourceType = "oracle"
	ConditionJobValueSourceTypeStatic ConditionJobValueSourceType = "static"
)

// Ethereum, Base, Arbitrum, Optimism are supported testnet for Jobs
type SupportedTestnetChainID string

const (
	SupportedTestnetChainIDEthereumSepolia SupportedTestnetChainID = "11155111"
	SupportedTestnetChainIDBaseSepolia SupportedTestnetChainID = "84532"
	SupportedTestnetChainIDArbitrumSepolia SupportedTestnetChainID = "421614"
	SupportedTestnetChainIDOptimismSepolia SupportedTestnetChainID = "11155420"
)

// Ethereum and Base have the architecture deployed and Arbitrum is supported mainnet for Jobs
type SupportedMainnetChainID string

const (
	SupportedMainnetChainIDEthereum SupportedMainnetChainID = "1"
	SupportedMainnetChainIDBase SupportedMainnetChainID = "8453"
	SupportedMainnetChainIDArbitrum SupportedMainnetChainID = "42161"
	// SupportedMainnetChainIDOptimism SupportedMainnetChainID = "10"
)

// Store the values as strings as the caluculations are done in Wei
const TG_PER_ETH = "1000"
const WEI_PER_ETH = "1000000000000000000" // 1e18
const WEI_PER_TG = "1000000000000000" // 1e15
