package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "math/big"
    "net/http"
    "strings"
	"os"

    "github.com/ethereum/go-ethereum"
    // "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// Event signature for OperatorRegistered
const OperatorRegisteredSig = "OperatorRegistered(address,uint256[4])"

type OperatorRegisteredEvent struct {
    Operator    common.Address
    BlsKey      [4]*big.Int
}

type KeeperData struct {
    KeeperID          int64     `json:"keeper_id"`
    KeeperAddress     string    `json:"keeper_address"`
    KeeperType        int       `json:"keeper_type"`
    RegisteredTx      string    `json:"registered_tx"`
    RewardsAddress    string    `json:"rewards_address"`
    Stakes            []float64 `json:"stakes"`
    Strategies        []string  `json:"strategies"`
    Verified          bool      `json:"verified"`
    Status            bool      `json:"status"`
    ConsensusKeys     []string  `json:"consensus_keys"`
    ConnectionAddress string    `json:"connection_address"`
}

func main() {
    // Connect to Ethereum node
	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    contractAddress := os.Getenv("AVS_GOVERNANCE_ADDRESS")


    client, err := ethclient.Dial(os.Getenv("L1_RPC"))
    if err != nil {
        log.Fatalf("Failed to connect to the Ethereum client: %v", err)
    }

    // Contract address

    // Create a filter query for the OperatorRegistered event
    query := ethereum.FilterQuery{
        Addresses: []common.Address{common.HexToAddress(contractAddress)},
        Topics: [][]common.Hash{{
            common.HexToHash("0x" + generateEventSignatureHash(OperatorRegisteredSig)),
        }},
    }

    // Create a channel to receive the logs
    logs := make(chan types.Log)

    // Subscribe to the events
    sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        log.Fatalf("Failed to subscribe to logs: %v", err)
    }

    fmt.Println("Listening for OperatorRegistered events...")

    // Listen for events
    for {
        select {
        case err := <-sub.Err():
            log.Fatal(err)
        case vLog := <-logs:
            // Parse the event
            event, err := parseOperatorRegisteredEvent(vLog)
            if err != nil {
                log.Printf("Error parsing event: %v", err)
                continue
            }

            // Convert BLS key to strings
            consensusKeys := make([]string, 4)
            for i, key := range event.BlsKey {
                consensusKeys[i] = key.String()
            }

            // Create keeper data
            keeperData := KeeperData{
                KeeperAddress: event.Operator.Hex(),
                KeeperType:   1, // Assuming 1 for Keeper
                RegisteredTx: vLog.TxHash.Hex(),
                ConsensusKeys: consensusKeys,
                Status:       true,
                Verified:     true,
            }

            // Send to API
            if err := sendToAPI(keeperData); err != nil {
                log.Printf("Error sending to API: %v", err)
                continue
            }

            fmt.Printf("Processed operator registration: %s\n", event.Operator.Hex())
        }
    }
}

func generateEventSignatureHash(sig string) string {
    return strings.TrimPrefix(common.HexToHash(sig).Hex(), "0x")
}

func parseOperatorRegisteredEvent(vLog types.Log) (*OperatorRegisteredEvent, error) {
    event := new(OperatorRegisteredEvent)
    
    // Parse operator address from first topic
    event.Operator = common.HexToAddress(vLog.Topics[1].Hex())
    
    // Parse BLS key from data
    blsKeyData := vLog.Data
    for i := 0; i < 4; i++ {
        start := i * 32
        end := start + 32
        if end > len(blsKeyData) {
            return nil, fmt.Errorf("invalid data length")
        }
        event.BlsKey[i] = new(big.Int).SetBytes(blsKeyData[start:end])
    }
    
    return event, nil
}

func sendToAPI(data KeeperData) error {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("error marshaling data: %v", err)
    }

    resp, err := http.Post("http://data.triggerx.network/api/keepers", 
        "application/json", 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("error sending request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    return nil
}