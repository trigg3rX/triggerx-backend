package types

type NodeConfig struct {
	// used to set the logger level (true = info, false = debug)
	Passphrase                    string `yaml:"passphrase"`
	Production                    bool   `yaml:"production"`
	KeeperAddress                 string `yaml:"keeper_address"`
	OperatorStateRetrieverAddress string `yaml:"operator_state_retriever_address"`
	ServiceManagerAddress         string `yaml:"service_manager_address"`
	TokenStrategyAddr             string `yaml:"token_strategy_addr"`
	EthRpcUrl                     string `yaml:"eth_rpc_url"`
	EthWsUrl                      string `yaml:"eth_ws_url"`
	BlsPrivateKeyStorePath        string `yaml:"bls_private_key_store_path"`
	EcdsaPrivateKeyStorePath      string `yaml:"ecdsa_private_key_store_path"`
	AggregatorServerIpPortAddress string `yaml:"aggregator_server_ip_port_address"`
	EigenMetricsIpPortAddress     string `yaml:"prometheus_port_address"`
	EnableMetrics                 bool   `yaml:"enable_metrics"`
	NodeApiIpPortAddress          string `yaml:"node_api_ip_port_address"`
	EnableNodeApi                 bool   `yaml:"enable_node_api"`
}

type KeeperStatus struct {
	EcdsaAddress      string
	PubkeysRegistered bool
	G1Pubkey          string
	G2Pubkey          string
	RegisteredWithAvs bool
	KeeperId          string
}
