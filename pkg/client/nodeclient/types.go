package nodeclient

import (
	"encoding/json"
	"fmt"
)

// JSON-RPC 2.0 request structure
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// JSON-RPC 2.0 response structure
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// Block identifier types (hex string or tag)
type BlockNumber string

const (
	BlockEarliest  BlockNumber = "earliest"
	BlockFinalized BlockNumber = "finalized"
	BlockSafe      BlockNumber = "safe"
	BlockLatest    BlockNumber = "latest"
	BlockPending   BlockNumber = "pending"
)

// EthGetLogsParams represents parameters for eth_getLogs
type EthGetLogsParams struct {
	FromBlock *BlockNumber  `json:"fromBlock,omitempty"` // hex string or tag
	ToBlock   *BlockNumber  `json:"toBlock,omitempty"`   // hex string or tag
	Address   interface{}   `json:"address,omitempty"`   // string or []string for multiple addresses
	Topics    []interface{} `json:"topics,omitempty"`    // array of topic filters
}

// Log represents an event log
type Log struct {
	Removed          bool     `json:"removed"`
	LogIndex         string   `json:"logIndex"`
	TransactionIndex string   `json:"transactionIndex"`
	TransactionHash  string   `json:"transactionHash"`
	BlockHash        string   `json:"blockHash"`
	BlockNumber      string   `json:"blockNumber"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

// EthGetTransactionReceiptParams represents parameters for eth_getTransactionReceipt
type EthGetTransactionReceiptParams struct {
	TransactionHash string `json:"transactionHash"`
}

// TransactionReceipt represents a transaction receipt
type TransactionReceipt struct {
	TransactionHash   string `json:"transactionHash"`
	TransactionIndex  string `json:"transactionIndex"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	From              string `json:"from"`
	To                string `json:"to"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	EffectiveGasPrice string `json:"effectiveGasPrice,omitempty"`
	GasUsed           string `json:"gasUsed"`
	ContractAddress   string `json:"contractAddress,omitempty"`
	Logs              []Log  `json:"logs"`
	LogsBloom         string `json:"logsBloom"`
	Status            string `json:"status"` // "0x1" for success, "0x0" for failure
	Type              string `json:"type,omitempty"`
}

// EthGetBlockByNumberParams represents parameters for eth_getBlockByNumber
type EthGetBlockByNumberParams struct {
	BlockNumber BlockNumber `json:"blockNumber"` // hex string or tag
	FullTx      bool        `json:"fullTx"`      // true for full transaction objects, false for hashes only
}

// Block represents a block
type Block struct {
	Number                string        `json:"number"`
	Hash                  string        `json:"hash"`
	ParentHash            string        `json:"parentHash"`
	Nonce                 string        `json:"nonce,omitempty"`
	Sha3Uncles            string        `json:"sha3Uncles"`
	LogsBloom             string        `json:"logsBloom"`
	TransactionsRoot      string        `json:"transactionsRoot"`
	StateRoot             string        `json:"stateRoot"`
	ReceiptsRoot          string        `json:"receiptsRoot"`
	Miner                 string        `json:"miner"`
	Difficulty            string        `json:"difficulty"`
	TotalDifficulty       string        `json:"totalDifficulty,omitempty"`
	ExtraData             string        `json:"extraData"`
	Size                  string        `json:"size"`
	GasLimit              string        `json:"gasLimit"`
	GasUsed               string        `json:"gasUsed"`
	Timestamp             string        `json:"timestamp"`
	Transactions          []interface{} `json:"transactions"` // []string if FullTx=false, []Transaction if FullTx=true
	Uncles                []string      `json:"uncles"`
	BaseFeePerGas         string        `json:"baseFeePerGas,omitempty"`
	Withdrawals           []interface{} `json:"withdrawals,omitempty"`
	WithdrawalsRoot       string        `json:"withdrawalsRoot,omitempty"`
	BlobGasUsed           string        `json:"blobGasUsed,omitempty"`
	ExcessBlobGas         string        `json:"excessBlobGas,omitempty"`
	ParentBeaconBlockRoot string        `json:"parentBeaconBlockRoot,omitempty"`
}

// Transaction represents a transaction (when FullTx=true)
type Transaction struct {
	BlockHash            string        `json:"blockHash"`
	BlockNumber          string        `json:"blockNumber"`
	From                 string        `json:"from"`
	Gas                  string        `json:"gas"`
	GasPrice             string        `json:"gasPrice,omitempty"`
	MaxFeePerGas         string        `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string        `json:"maxPriorityFeePerGas,omitempty"`
	Hash                 string        `json:"hash"`
	Input                string        `json:"input"`
	Nonce                string        `json:"nonce"`
	To                   string        `json:"to"`
	TransactionIndex     string        `json:"transactionIndex"`
	Value                string        `json:"value"`
	Type                 string        `json:"type,omitempty"`
	V                    string        `json:"v"`
	R                    string        `json:"r"`
	S                    string        `json:"s"`
	ChainID              string        `json:"chainId,omitempty"`
	AccessList           []interface{} `json:"accessList,omitempty"`
}

// EthGetBlockReceiptsParams represents parameters for eth_getBlockReceipts
type EthGetBlockReceiptsParams struct {
	BlockNumber BlockNumber `json:"blockNumber"` // hex string or tag
}

// EthEstimateGasParams represents parameters for eth_estimateGas
type EthEstimateGasParams struct {
	From     string `json:"from,omitempty"`     // address
	To       string `json:"to,omitempty"`       // address
	Gas      string `json:"gas,omitempty"`      // hex string
	GasPrice string `json:"gasPrice,omitempty"` // hex string
	Value    string `json:"value,omitempty"`    // hex string
	Data     string `json:"data,omitempty"`     // hex string
}

// EthGetCodeParams represents parameters for eth_getCode
type EthGetCodeParams struct {
	Address     string      `json:"address"`     // address
	BlockNumber BlockNumber `json:"blockNumber"` // hex string or tag
}

// EthGetStorageAtParams represents parameters for eth_getStorageAt
type EthGetStorageAtParams struct {
	Address     string      `json:"address"`     // address
	Position    string      `json:"position"`    // hex string (storage slot)
	BlockNumber BlockNumber `json:"blockNumber"` // hex string or tag
}

// EthSendRawTransactionParams represents parameters for eth_sendRawTransaction
type EthSendRawTransactionParams struct {
	SignedTxData string `json:"signedTxData"` // hex string of signed transaction
}

// Network represents a blockchain network
type Network string

const (
	NetworkEthereum        Network = "eth-mainnet"
	NetworkEthereumSepolia Network = "eth-sepolia"
	NetworkArbitrum        Network = "arb-mainnet"
	NetworkArbitrumSepolia Network = "arb-sepolia"
	NetworkOptimism        Network = "opt-mainnet"
	NetworkOptimismSepolia Network = "opt-sepolia"
	NetworkBase            Network = "base-mainnet"
	NetworkBaseSepolia     Network = "base-sepolia"
)

// GetAlchemyURL returns the Alchemy API URL for a given network
func (n Network) GetAlchemyURL() string {
	return fmt.Sprintf("https://%s.g.alchemy.com/v2/", n)
}
