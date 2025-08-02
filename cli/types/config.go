package types

type NodeConfig struct {
	// used to set the logger level (true = info, false = debug)
	Production                       bool   `yaml:"production"`
	AVSOwnerAddress                  string `yaml:"avs_owner_address"`
	OperatorAddress                  string `yaml:"operator_address"`
	AVSAddress                       string `yaml:"avs_address"`
	EthRpcUrl                        string `yaml:"eth_rpc_url"`
	EthWsUrl                         string `yaml:"eth_ws_url"`
	BlsPrivateKeyStorePath           string `yaml:"bls_private_key_store_path"`
	OperatorEcdsaPrivateKeyStorePath string `yaml:"operator_ecdsa_private_key_store_path"`
	AVSEcdsaPrivateKeyStorePath      string `yaml:"avs_ecdsa_private_key_store_path"`
	RegisterOperatorOnStartup        bool   `yaml:"register_operator_on_startup"`
	NodeApiIpPortAddress             string `yaml:"node_api_ip_port_address"`
	EnableNodeApi                    bool   `yaml:"enable_node_api"`

	// register avs parameters
	// AvsName            string   `yaml:"avs_name"`
	// MinStakeAmount     uint64   `yaml:"min_stake_amount"`
	// AvsOwnerAddresses  []string `yaml:"avs_owner_addresses"`
	// WhitelistAddresses []string `yaml:"whitelist_addresses"`
	// AssetIDs           []string `yaml:"asset_ids"`
	// AvsUnbondingPeriod uint64   `yaml:"avs_unbonding_period"`
	// MinSelfDelegation  uint64   `yaml:"min_self_delegation"`
	// EpochIdentifier    string   `yaml:"epoch_identifier"`
	// TaskAddress        string   `yaml:"task_address"`
	// AVSRewardAddress   string   `yaml:"avs_reward_address"`
	// AVSSlashAddress    string   `yaml:"avs_slash_address"`

	// create new task parameters
	// CreateTaskInterval    int64  `yaml:"create_task_interval"`
	// TaskResponsePeriod    uint64 `yaml:"task_response_period"`
	// TaskChallengePeriod   uint64 `yaml:"task_challenge_period"`
	// ThresholdPercentage   uint8  `yaml:"threshold_percentage"`
	// TaskStatisticalPeriod uint64 `yaml:"task_statistical_period"`
	// MiniOptInOperators    uint64 `yaml:"mini_opt_in_operators"`  // the minimum number of opt-in operators
	// MinTotalStakeAmount   uint64 `yaml:"min_total_stake_amount"` // the minimum total amount of stake by all operators
	// AvsRewardProportion   uint64 `yaml:"avs_reward_proportion"`  // the proportion of reward for AVS
	// AvsSlashProportion    uint64 `yaml:"avs_slash_proportion"`   // the proportion of slash for AVS

	// deposit and delegation
	DepositAmount  int64  `yaml:"deposit_amount"`
	DelegateAmount int64  `yaml:"delegate_amount"`
	Staker         string `yaml:"staker"`
}