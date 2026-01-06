package config

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/trigg3rX/triggerx-backend/pkg/yaml"
)

type Config struct {
	devMode bool

	// RPC URLs for Ethereum and Base
	ethRPCUrl  string
	baseRPCUrl string

	// API Keys for Alchemy and Etherscan
	alchemyAPIKey   string
	etherscanAPIKey string

	// Controller Key and Keeper Address
	privateKeyController string
	keeperAddress        string

	// Consensus Key and Address
	privateKeyConsensus string
	consensusAddress    string

	// Public IP Address and Peer ID
	publicIPV4Address string
	peerID            string

	// Ports for Keeper API server, P2P connections, metrics and Grafana
	keeperRPCPort     string
	keeperP2PPort     string
	keeperMetricsPort string
	grafanaPort       string

	// IPFS configuration
	ipfsHost  string
	pinataJWT string

	// TLS Proof configuration
	tlsProofHost string
	tlsProofPort string

	// Manager Signing Address
	managerSigningAddress string

	// Backend Service URLs
	aggregatorRPCUrl  string
	healthRPCUrl      string
	taskMonitorRPCUrl string

	// AVS Contract Address
	avsGovernanceAddress     string
	attestationCenterAddress string
	taskExecutionAddress     string

	// Othentic Bootstrap ID
	othenticBootstrapID string

	// Observability configuration
	otelExporterEndpoint   string
	enablePrometheusExport bool

	// YAML-loaded settings
	api      APIConfig
	health   HealthConfig
	shutdown ShutdownConfig
	version  yaml.VersionConfig
}

type APIConfig struct {
	ReadTimeout  yaml.Duration `yaml:"read_timeout"`
	WriteTimeout yaml.Duration `yaml:"write_timeout"`
}

type HealthConfig struct {
	CheckInterval  yaml.Duration `yaml:"check_interval"`
	RequestTimeout yaml.Duration `yaml:"request_timeout"`
}

type ShutdownConfig struct {
	Timeout yaml.Duration `yaml:"timeout"`
}

type YAMLConfig struct {
	API      APIConfig          `yaml:"api"`
	Health   HealthConfig       `yaml:"health"`
	Shutdown ShutdownConfig     `yaml:"shutdown"`
	Version  yaml.VersionConfig `yaml:"version"`
}

var cfg Config

func Init(configPath string) error {
	// Load secrets from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Load YAML config
	var yamlConfig YAMLConfig
	if err := yaml.LoadYAML(configPath, &yamlConfig); err != nil {
		return fmt.Errorf("error loading configuration file: %w", err)
	}

	cfg = Config{
		devMode:              env.GetEnvBool("DEV_MODE", false),
		ethRPCUrl:            env.GetEnvString("L1_RPC", ""),
		baseRPCUrl:           env.GetEnvString("L2_RPC", ""),
		privateKeyConsensus:  env.GetEnvString("PRIVATE_KEY", ""),
		privateKeyController: env.GetEnvString("OPERATOR_PRIVATE_KEY", ""),
		keeperAddress:        env.GetEnvString("OPERATOR_ADDRESS", ""),
		consensusAddress:     crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(common.FromHex(env.GetEnvString("PRIVATE_KEY", ""))).PublicKey).Hex(),
		publicIPV4Address:    env.GetEnvString("PUBLIC_IPV4_ADDRESS", ""),
		peerID:               env.GetEnvString("PEER_ID", ""),
		keeperRPCPort:        env.GetEnvString("OPERATOR_RPC_PORT", "9011"),
		keeperP2PPort:        env.GetEnvString("OPERATOR_P2P_PORT", "9012"),
		keeperMetricsPort:    env.GetEnvString("OPERATOR_METRICS_PORT", "9013"),
		grafanaPort:          env.GetEnvString("GRAFANA_PORT", "3000"),
		aggregatorRPCUrl:     env.GetEnvString("OTHENTIC_CLIENT_RPC_ADDRESS", "https://aggregator.triggerx.network"),
		healthRPCUrl:         env.GetEnvString("HEALTH_IP_ADDRESS", "https://health.triggerx.network"),
		taskMonitorRPCUrl:    env.GetEnvString("TASK_MONITOR_RPC_URL", "https://task.triggerx.network"),
		tlsProofHost:         "www.google.com",
		tlsProofPort:         "443",
		// Test Attestation Center Address for Base Sepolia
		attestationCenterAddress: env.GetEnvString("ATTESTATION_CENTER_ADDRESS", "0xB3c01C8BaEF65436B0d01F891d00B25CA9d7D383"),
		// Base Mainnet Attestation Center Address
		// attestationCenterAddress: env.GetEnvString("ATTESTATION_CENTER_ADDRESS", "0x6DFee10D13d5B43AaF97bDA908C1D76d4313aF5f"),
		othenticBootstrapID:    env.GetEnvString("OTHENTIC_BOOTSTRAP_ID", "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB"),
		otelExporterEndpoint:   env.GetOTELExporterEndpoint(),
		enablePrometheusExport: env.GetEnvBool("ENABLE_PROMETHEUS_EXPORT", true),
		api:                    yamlConfig.API,
		health:                 yamlConfig.Health,
		shutdown:               yamlConfig.Shutdown,
		version:                yamlConfig.Version,
	}
	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if err := yaml.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if !cfg.devMode {
		gin.SetMode(gin.ReleaseMode)
	}
	if err := checkKeeperRegistration(); err != nil {
		fmt.Println("Keeper address is not yet registered on L2. Please register the address before continuing. If registered, please wait for the registration to be confirmed.")
		return fmt.Errorf("keeper address is not registered on L2")
	}
	return nil
}

func validateConfig(cfg Config) error {
	if env.IsEmpty(cfg.ethRPCUrl) {
		return fmt.Errorf("invalid eth rpc url: %s", cfg.ethRPCUrl)
	}
	if env.IsEmpty(cfg.baseRPCUrl) {
		return fmt.Errorf("invalid base rpc url: %s", cfg.baseRPCUrl)
	}
	if !env.IsValidIPAddress(cfg.publicIPV4Address) {
		return fmt.Errorf("invalid public ipv4 address: %s", cfg.publicIPV4Address)
	}
	if !env.IsValidPrivateKey(cfg.privateKeyConsensus) {
		return fmt.Errorf("invalid private key consensus: %s", cfg.privateKeyConsensus)
	}
	if !env.IsValidEthAddress(cfg.keeperAddress) {
		return fmt.Errorf("invalid keeper address: %s", cfg.keeperAddress)
	}
	// if !env.IsValidPeerID(cfg.peerID) {
	// 	return fmt.Errorf("invalid peer id: %s", cfg.peerID)
	// }
	if !env.IsValidHostPort(cfg.otelExporterEndpoint) {
		return fmt.Errorf("invalid OTEL exporter endpoint: %s (must be a valid host:port, e.g., localhost:4318)", cfg.otelExporterEndpoint)
	}
	return nil
}

func GetEthRPCUrl() string {
	return cfg.ethRPCUrl
}

func GetBaseRPCUrl() string {
	return cfg.baseRPCUrl
}

// Only sets it if there was no key in env file
func SetAlchemyAPIKey(key string) {
	if !env.IsEmpty(cfg.alchemyAPIKey) {
		return
	}
	cfg.alchemyAPIKey = key
}

func GetAlchemyAPIKey() string {
	return cfg.alchemyAPIKey
}

func SetEtherscanAPIKey(key string) {
	cfg.etherscanAPIKey = key
}

func GetEtherscanAPIKey() string {
	return cfg.etherscanAPIKey
}

func GetPrivateKeyConsensus() string {
	return cfg.privateKeyConsensus
}

func GetPrivateKeyController() string {
	return cfg.privateKeyController
}

func GetKeeperAddress() string {
	return cfg.keeperAddress
}

func GetConsensusAddress() string {
	return cfg.consensusAddress
}

func GetPublicIPV4Address() string {
	return cfg.publicIPV4Address
}

func GetPeerID() string {
	return cfg.peerID
}

func GetOperatorRPCPort() string {
	return cfg.keeperRPCPort
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetKeeperRPCPort() string {
	return cfg.keeperRPCPort
}

func GetAggregatorRPCUrl() string {
	url := cfg.aggregatorRPCUrl
	// Auto-prepend http:// if scheme is missing (for backward compatibility)
	if url != "" && !hasScheme(url) {
		return "http://" + url
	}
	return url
}

// hasScheme checks if a URL string has a scheme (http://, https://, etc.)
func hasScheme(url string) bool {
	for i := 0; i < len(url); i++ {
		if url[i] == ':' {
			// Check if it's followed by // (scheme separator)
			if i+2 < len(url) && url[i+1] == '/' && url[i+2] == '/' {
				return true
			}
			// If we hit a colon before //, it's likely a port, not a scheme
			return false
		}
		if url[i] == '/' {
			// If we hit / before :, no scheme
			return false
		}
	}
	return false
}

func GetHealthRPCUrl() string {
	return cfg.healthRPCUrl
}

func GetTaskMonitorRPCUrl() string {
	return cfg.taskMonitorRPCUrl
}

func GetAvsGovernanceAddress() string {
	return cfg.avsGovernanceAddress
}

func GetAttestationCenterAddress() string {
	return cfg.attestationCenterAddress
}

func GetVersion() string {
	return cfg.version.Version
}

func IsImua() bool {
	// Check environment variable, default to false
	return env.GetEnvBool("IS_IMUA", false)
}

// IPFS configuration
func SetIpfsHost(host string) {
	cfg.ipfsHost = host
}

func GetIpfsHost() string {
	return cfg.ipfsHost
}

func SetPinataJWT(jwt string) {
	cfg.pinataJWT = jwt
}

func GetPinataJWT() string {
	return cfg.pinataJWT
}

// TLS Proof configuration
func SetTLSProofHost(host string) {
	cfg.tlsProofHost = host
}

func SetTLSProofPort(port string) {
	cfg.tlsProofPort = port
}

func GetTLSProofHost() string {
	return cfg.tlsProofHost
}

func GetTLSProofPort() string {
	return cfg.tlsProofPort
}

// Manager Signing Address
func SetManagerSigningAddress(addr string) {
	cfg.managerSigningAddress = addr
}

func GetManagerSigningAddress() string {
	return cfg.managerSigningAddress
}

func SetTaskExecutionAddress(addr string) {
	cfg.taskExecutionAddress = addr
}

func GetTaskExecutionAddress() string {
	return cfg.taskExecutionAddress
}

// SetKeeperAddress sets the keeper address in the config (for testing)
func SetKeeperAddress(addr string) {
	cfg.keeperAddress = addr
}

// Observability configuration getters
func GetOTELExporterEndpoint() string {
	return cfg.otelExporterEndpoint
}

func GetEnablePrometheusExport() bool {
	return cfg.enablePrometheusExport
}

func GetAPIReadTimeout() time.Duration {
	return cfg.api.ReadTimeout.ToDuration()
}

func GetAPIWriteTimeout() time.Duration {
	return cfg.api.WriteTimeout.ToDuration()
}

func GetHealthCheckInterval() time.Duration {
	return cfg.health.CheckInterval.ToDuration()
}

func GetHealthRequestTimeout() time.Duration {
	return cfg.health.RequestTimeout.ToDuration()
}

func GetShutdownTimeout() time.Duration {
	return cfg.shutdown.Timeout.ToDuration()
}
