package types

type NodeConfig struct {
	// Core settings
	Production bool   `yaml:"production"`
	AvsName    string `yaml:"avs_name"`
	SemVer     string `yaml:"sem_ver"`

	// Keeper settings
	KeeperAddress            string `yaml:"address"`
	EcdsaPrivateKeyStorePath string `yaml:"ecdsa_keystore_path"`
	BlsPrivateKeyStorePath   string `yaml:"bls_keystore_path"`

	// Network settings
	EthRpcUrl string `yaml:"ethrpcurl"`
	EthWsUrl  string `yaml:"ethwsurl"`

	// Contract addresses
	ServiceManagerAddress         string `yaml:"service_manager_address"`
	OperatorStateRetrieverAddress string `yaml:"operator_state_retriever"`

	// Metrics and API settings
	EnableMetrics             bool   `yaml:"enable_metrics"`
	EigenMetricsIpPortAddress string `yaml:"port_address"`
	EnableNodeApi             bool   `yaml:"enable_node_api"`
	NodeApiIpPortAddress      string `yaml:"node_api_ip_port_address"`
}

type KeeperStatus struct {
	EcdsaAddress      string
	PubkeysRegistered bool
	G1Pubkey          string
	G2Pubkey          string
	RegisteredWithAvs bool
	KeeperId          string
}
