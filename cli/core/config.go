package config

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	sdklogging "github.com/imua-xyz/imua-avs-sdk/logging"
	"github.com/imua-xyz/imua-avs-sdk/signer"

	//"github.com/trigg3rX/triggerx-backend/core/chainio/eth"

	blscommon "github.com/prysmaticlabs/prysm/v5/crypto/bls/common"
)

type Config struct {
	Production      bool `yaml:"production"`
	EcdsaPrivateKey *ecdsa.PrivateKey
	BlsPrivateKey   *blscommon.SecretKey
	Logger          sdklogging.Logger
	// we need the url for the sdk currently... eventually standardize api to
	// only take an ethclient or an rpcUrl (and build the ethclient at each constructor site)
	EthHttpRpcUrl              string
	EthWsRpcUrl                string
	EthHttpClient              eth.EthClient
	EthWsClient                eth.EthClient
	OperatorStateRetrieverAddr common.Address
	AvsRegistryCoordinatorAddr common.Address
	AggregatorServerIpPortAddr string
	RegisterOperatorOnStartup  bool
	// json:"-" skips this field when marshaling (only used for logging to stdout), since SignerFn doesnt implement marshalJson
	SignerFn                 signer.SignerFn `json:"-"`
	TxMgr                    txmgr.TxManager
	AggregatorAddress        common.Address
	EcdsaPrivateKeyStorePath string `yaml:"ecdsa_private_key_store_path"`
}

var (
	FileFlag = cli.StringFlag{
		Name:     "config",
		Required: true,
		Usage:    "Load configuration from `FILE`",
	}

	EcdsaPrivateKeyFlag = cli.StringFlag{
		Name:     "ecdsa-private-key",
		Usage:    "Ethereum private key",
		Required: true,
		EnvVar:   "ECDSA_PRIVATE_KEY",
	}
	/* Optional Flags */
	TaskIDFlag = &cli.Uint64Flag{
		Name:     "task-ID",
		Usage:    "task ID",
		Required: false,
	}
	NumberToBeSquaredFlag = &cli.Uint64Flag{
		Name:     "NumberToBeSquared",
		Usage:    "number to be squared",
		Required: false,
	}
	ExecTypeFlag = &cli.IntFlag{
		Name:     "ExecType",
		Usage:    "Execution type: If the input is 1, then automatic execution, if it is 2, then manual execution",
		Required: true,
	}
)

var requiredFlags = []cli.Flag{
	FileFlag,
	EcdsaPrivateKeyFlag,
}

var optionalFlags = []cli.Flag{}

func init() {
	Flags = append(requiredFlags, optionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag
