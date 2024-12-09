package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "context"
    "path/filepath"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/joho/godotenv"
)

type ContractAddresses struct {
    Proxy          string `json:"proxy"`
    Implementation string `json:"implementation"`
}

type DeploymentConfig struct {
    ProxyAdmin              string            `json:"proxyAdmin"`
    PauserRegistry         string            `json:"pauserRegistry"`
    IndexRegistry          ContractAddresses `json:"indexRegistry"`
    StakeRegistry         ContractAddresses `json:"stakeRegistry"`
    ApkRegistry           ContractAddresses `json:"apkRegistry"`
    SocketRegistry        ContractAddresses `json:"socketRegistry"`
    RegistryCoordinator   ContractAddresses `json:"registryCoordinator"`
    OperatorStateRetriever string            `json:"operatorStateRetriever"`
    TriggerXServiceManager ContractAddresses `json:"triggerXServiceManager"`
    TriggerXTaskManager    ContractAddresses `json:"triggerXTaskManager"`
}

type StakeDeploymentConfig struct {
    TriggerXStakeRegistry ContractAddresses `json:"triggerXStakeRegistry"`
}

const (
    OUTPUT_DIR = "pkg/avsinterface/abis"
)

func main() {
    if err := godotenv.Load(); err != nil {
        panic("Error loading .env file")
    }

    deploymentPath := "./pkg/avsinterface/deployments.holesky.json"
    var deploymentData []byte
    var err error
    if _, err := os.Stat(deploymentPath); err == nil {
        deploymentData, err = ioutil.ReadFile(deploymentPath)
        if err != nil {
            panic(err)
        }
    } else {
        resp, err := http.Get("https://raw.githubusercontent.com/trigg3rX/triggerx-contracts/main/contracts/script/output/deployment.holesky.json")
        if err != nil {
            panic(err)
        }
        defer resp.Body.Close()

        deploymentData, err = ioutil.ReadAll(resp.Body)
        if err != nil {
            panic(err)
        }
    }

    stakeRegistry := "./pkg/avsinterface/stake.opsepolia.json"
    var stakeRegistryData []byte
    if _, err := os.Stat(stakeRegistry); err == nil {
        stakeRegistryData, err = ioutil.ReadFile(stakeRegistry)
        if err != nil {
            panic(err)
        }
    } else {
        resp, err := http.Get("https://raw.githubusercontent.com/trigg3rX/triggerx-contracts/main/contracts/script/output/stake.opsepolia.json")
        if err != nil {
            panic(err)
        }
        defer resp.Body.Close()

        stakeRegistryData, err = ioutil.ReadAll(resp.Body)
        if err != nil {
            panic(err)
        }
    }

    err = ioutil.WriteFile("./pkg/avsinterface/deployments.holesky.json", deploymentData, 0644)
    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile("./pkg/avsinterface/stake.opsepolia.json", stakeRegistryData, 0644)
    if err != nil {
        panic(err)
    }

    holeskyData, err := ioutil.ReadFile("./pkg/avsinterface/deployments.holesky.json")
    if err != nil {
        panic(err)
    }

    var deployments DeploymentConfig
    if err := json.Unmarshal(holeskyData, &deployments); err != nil {
        panic(err)
    }

    var stakeDeployments StakeDeploymentConfig
    if err := json.Unmarshal(stakeRegistryData, &stakeDeployments); err != nil {
        panic(err)
    }

    alchemyKey := os.Getenv("ALCHEMY_API_KEY")
    if alchemyKey == "" {
        panic("ALCHEMY_API_KEY not set in environment")
    }

    holeskyClient, err := ethclient.Dial("https://eth-holesky.g.alchemy.com/v2/" + alchemyKey)
    if err != nil {
        panic(err)
    }

    opSepoliaClient, err := ethclient.Dial("https://opt-sepolia.g.alchemy.com/v2/" + alchemyKey)
    if err != nil {
        panic(err)
    }

    os.MkdirAll(OUTPUT_DIR, 0755)

    fetchBytecode := func(name string, address string, client *ethclient.Client, network string) {
        binPath := filepath.Join(OUTPUT_DIR, name+"."+network+".bin")
        if _, err := os.Stat(binPath); err == nil {
            return
        }

        addr := common.HexToAddress(address)
        bytecode, err := client.CodeAt(context.Background(), addr, nil)
        if err != nil {
            panic(err)
        }

        err = ioutil.WriteFile(
            binPath,
            []byte(fmt.Sprintf("%x", bytecode)),
            0644,
        )
        if err != nil {
            panic(err)
        }
        fmt.Printf("Fetched bytecode for %s on %s\n", name, network)
    }

    holeskyContracts := map[string]string{
        "ProxyAdmin":              deployments.ProxyAdmin,
        "PauserRegistry":          deployments.PauserRegistry,
        "IndexRegistry":           deployments.IndexRegistry.Implementation,
        "ApkRegistry":            deployments.ApkRegistry.Implementation,
        "SocketRegistry":         deployments.SocketRegistry.Implementation,
        "RegistryCoordinator":    deployments.RegistryCoordinator.Implementation,
        "OperatorStateRetriever": deployments.OperatorStateRetriever,
        "TriggerXServiceManager": deployments.TriggerXServiceManager.Implementation,
        "TriggerXTaskManager":    deployments.TriggerXTaskManager.Implementation,
    }

    opSepoliaContracts := map[string]string{
        "TriggerXStakeRegistry": stakeDeployments.TriggerXStakeRegistry.Implementation,
    }

    for name, addr := range holeskyContracts {
        fetchBytecode(name, addr, holeskyClient, "holesky")
        abi, err := fetchABIFromBlockscout(addr, "holesky")
        if err != nil {
            panic(err)
        }

        abiPath := filepath.Join(OUTPUT_DIR, name+".holesky.abi")
        if _, err := os.Stat(abiPath); err == nil {
            continue
        }

        err = ioutil.WriteFile(
            abiPath,
            []byte(abi),
            0644,
        )
        if err != nil {
            panic(err)
        }
        fmt.Printf("Fetched ABI for %s on holesky\n", name)
    }

    for name, addr := range opSepoliaContracts {
        fetchBytecode(name, addr, opSepoliaClient, "opsepolia")
        abi, err := fetchABIFromBlockscout(addr, "opsepolia")
        if err != nil {
            panic(err)
        }

        abiPath := filepath.Join(OUTPUT_DIR, name+".opsepolia.abi")
        if _, err := os.Stat(abiPath); err == nil {
            continue
        }

        err = ioutil.WriteFile(
            abiPath,
            []byte(abi),
            0644,
        )
        if err != nil {
            panic(err)
        }
        fmt.Printf("Fetched ABI for %s on opsepolia\n", name)
    }
}

func fetchABIFromBlockscout(address string, network string) (string, error) {
    var baseURL string
    switch network {
    case "holesky":
        baseURL = "https://eth-holesky.blockscout.com/api"
    case "opsepolia":
        baseURL = "https://optimism-sepolia.blockscout.com/api"
    default:
        return "", fmt.Errorf("unsupported network: %s", network)
    }

    url := fmt.Sprintf("%s?module=contract&action=getabi&address=%s", baseURL, address)
    
    resp, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("failed to fetch ABI: %v", err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response body: %v", err)
    }

    var result struct {
        Status  string `json:"status"`
        Message string `json:"message"`
        Result  string `json:"result"`
    }

    if err := json.Unmarshal(body, &result); err != nil {
        return "", fmt.Errorf("failed to parse JSON response: %v", err)
    }

    if result.Status != "1" {
        return "", fmt.Errorf("blockscout API error: %s ||| %s" , result.Message, address)
    }

    return result.Result, nil
}