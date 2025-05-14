package config

import "os"

var (
	KeeperRPCPort        string
	KeeperP2PPort        string
	KeeperMetricsPort    string
	GrafanaPort          string
	
	PinataApiKey             string
	PinataSecretApiKey       string
	IpfsHost                 string
	
	AggregatorRPCAddress      string
	HealthRPCAddress          string
	
	L1Chain                  string
	L2Chain                  string
	
	AVSGovernanceAddress     string
	AttestationCenterAddress string
	
	OthenticBootstrapID      string
)

func checkDefaultValues() {
	KeeperRPCPort = os.Getenv("KEEPER_RPC_PORT")
	if KeeperRPCPort == "" {
		KeeperRPCPort = "9005"
	}
	KeeperP2PPort = os.Getenv("KEEPER_P2P_PORT")
	if KeeperP2PPort == "" {
		KeeperP2PPort = "9006"
	}
	KeeperMetricsPort = os.Getenv("KEEPER_METRICS_PORT")
	if KeeperMetricsPort == "" {
		KeeperMetricsPort = "9009"
	}
	GrafanaPort = os.Getenv("GRAFANA_PORT")
	if GrafanaPort == "" {
		GrafanaPort = "3000"
	}
	PinataApiKey = os.Getenv("PINATA_API_KEY")
	if PinataApiKey == "" {
		PinataApiKey = "3e1b278b99bd95877625"
	}
	PinataSecretApiKey = os.Getenv("PINATA_SECRET_API_KEY")
	if PinataSecretApiKey == "" {
		PinataSecretApiKey = "8e41503276cd848b4f95fcde1f30e325652e224e7233dcc1910e5a226675ace4"
	}
	IpfsHost = os.Getenv("IPFS_HOST")
	if IpfsHost == "" {
		IpfsHost = "apricot-voluntary-fowl-585.mypinata.cloud"
	}
	AggregatorRPCAddress = os.Getenv("OTHENTIC_CLIENT_RPC_ADDRESS")
	if AggregatorRPCAddress == "" {
		AggregatorRPCAddress = "http://127.0.0.1:9001"
	}
	HealthRPCAddress = os.Getenv("HEALTH_RPC_ADDRESS")
	if HealthRPCAddress == "" {
		HealthRPCAddress = "http://127.0.0.1:9004"
	}
	L1Chain = os.Getenv("L1_CHAIN")
	if L1Chain == "" {
		L1Chain = "17000"
	}
	L2Chain = os.Getenv("L2_CHAIN")
	if L2Chain == "" {
		L2Chain = "84532"
	}
	AVSGovernanceAddress = os.Getenv("AVS_GOVERNANCE_ADDRESS")
	if AVSGovernanceAddress == "" {
		AVSGovernanceAddress = "0xE52De62Bf743493d3c4E1ac8db40f342FEb11fEa"
	}
	AttestationCenterAddress = os.Getenv("ATTESTATION_CENTER_ADDRESS")
	if AttestationCenterAddress == "" {
		AttestationCenterAddress = "0x8256F235Ed6445fb9f8177a847183A8C8CD97cF1"
	}
	OthenticBootstrapID = os.Getenv("OTHE_BOOTSTRAP_ID")
	if OthenticBootstrapID == "" {
		OthenticBootstrapID = "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB"
	}	
}
