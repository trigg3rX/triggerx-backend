package types

type NodeConfig struct {
	// Keeper settings
	KeeperName               string `yaml:"keeper_name"`
	KeeperAddress           string `yaml:"keeper_address"`
	EcdsaPrivateKeyStorePath string `yaml:"keeper_ecdsa_keystore_path"`
	EcdsaPassphrase         string `yaml:"keeper_ecdsa_passphrase"`
	BlsPrivateKeyStorePath  string `yaml:"keeper_bls_keystore_path"`
	BlsPassphrase          string `yaml:"keeper_bls_passphrase"`

	// Network settings
	EthRpcUrl string `yaml:"ethrpcurl"`
	EthWsUrl  string `yaml:"ethwsurl"`

	// Port Settings
	MetricsPort int `yaml:"metrics_port"`
	P2pPort     int `yaml:"p2p_port"`

	// Connection settings
	P2pPeerId         string `yaml:"keeper_p2p_peer_id"`
	ConnectionAddress string `yaml:"keeper_connection_address"`

	// Core settings
	AvsName string `yaml:"avs_name"`
	Version string `yaml:"version"`

	// Contract addresses
	AvsDirectoryAddress           string `yaml:"avs_directory_address"`
	DelegationManagerAddress      string `yaml:"delegation_manager_address"`
	StrategyManagerAddress        string `yaml:"strategy_manager_address"`
	RegistryCoordinatorAddress    string `yaml:"registry_coordinator_address"`
	ServiceManagerAddress         string `yaml:"service_manager_address"`
	OperatorStateRetrieverAddress string `yaml:"operator_state_retriever"`

	// Metrics settings
	EnableMetrics bool `yaml:"enable_metrics"`
}