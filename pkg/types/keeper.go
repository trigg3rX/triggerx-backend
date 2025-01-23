package types

type NodeConfig struct {
	// Core settings
	Production bool   `yaml:"production"`
	AvsName    string `yaml:"avs_name"`
	SemVer     string `yaml:"sem_ver"`
	ServerIpAddress string `yaml:"server_ip_address"`

	// Keeper settings
	KeeperName 				 string `yaml:"keeper_name"`
	KeeperAddress            string `yaml:"address"`
	EcdsaPrivateKeyStorePath string `yaml:"ecdsa_keystore_path"`
	BlsPrivateKeyStorePath   string `yaml:"bls_keystore_path"`
	P2pPeerId                string `yaml:"p2p_peer_id"`
	ConnectionAddress        string `yaml:"connection_address"`
	P2pPort                  string `yaml:"p2p_port"`
	
	// Network settings
	EthRpcUrl string `yaml:"ethrpcurl"`
	EthWsUrl  string `yaml:"ethwsurl"`

	// Contract addresses
	AvsDirectoryAddress         string `yaml:"avs_directory_address"`
	StrategyManagerAddress      string `yaml:"strategy_manager_address"`
	RegistryCoordinatorAddress  string `yaml:"registry_coordinator_address"`
	ServiceManagerAddress       string `yaml:"service_manager_address"`
	OperatorStateRetrieverAddress string `yaml:"operator_state_retriever"`

	// Metrics and API settings
	EnableMetrics        bool   `yaml:"enable_metrics"`
	MetricsIpPortAddress string `yaml:"port_address"`
}