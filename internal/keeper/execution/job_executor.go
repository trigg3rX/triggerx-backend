package execution

import (
	// "bytes"
	// "context"
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	// "log"
	// "math/big"
	// "net/http"
	// "os"
	// "os/exec"
	// "path/filepath"
	// "reflect"
	// "strconv"
	// "strings"
	// "time"

	// "github.com/ethereum/go-ethereum/accounts/abi"
	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	// ethcommon "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/ethclient"
	// "github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	// "github.com/trigg3rX/triggerx-backend/internal/keeper/validation"
	// "github.com/trigg3rX/triggerx-backend/pkg/common"
	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"

	// dockertypes "github.com/docker/docker/api/types"
	// "github.com/docker/docker/client"
	// "github.com/trigg3rX/triggerx-backend/pkg/resources"
)

// ValidatorInterface defines the contract for job validation
type ValidatorInterface interface {
	ValidateTimeBasedJob(job *jobtypes.HandleCreateJobData) (bool, error)
	ValidateEventBasedJob(job *jobtypes.HandleCreateJobData, ipfsData *jobtypes.IPFSData) (bool, error)
	ValidateAndPrepareJob(job *jobtypes.HandleCreateJobData, triggerData *jobtypes.TriggerData) (bool, error)
}









// type JobExecutor struct {
// 	ethClient       common.EthClientInterface
// 	etherscanAPIKey string
// 	argConverter    *ArgumentConverter
// 	validator       common.ValidatorInterface
// 	logger          common.Logger
// }

// func NewJobExecutor(ethClient *ethclient.Client, etherscanAPIKey string) *JobExecutor {
// 	return &JobExecutor{
// 		ethClient:       ethClient,
// 		etherscanAPIKey: etherscanAPIKey,
// 		argConverter:    &ArgumentConverter{},
// 		validator:       validation.NewJobValidator(logger, ethClient),
// 		logger:          logger,
// 	}
// }

// const (
// 	executionContractAddress = "0x68605feB94a8FeBe5e1fBEF0A9D3fE6e80cEC126"
// )

// // Execute routes jobs to appropriate handlers based on the target function
// Currently supports 'transfer' for token transfers and 'execute' for generic contract calls






// func (e *JobExecutor) executeGoScript(scriptContent string) (string, error) {
// 	// Create a temporary file for the script
// 	tempFile, err := ioutil.TempFile("", "script-*.go")
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create temporary file: %v", err)
// 	}
// 	defer os.Remove(tempFile.Name())

// 	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
// 		return "", fmt.Errorf("failed to write script to file: %v", err)
// 	}
// 	if err := tempFile.Close(); err != nil {
// 		return "", fmt.Errorf("failed to close temporary file: %v", err)
// 	}

// 	// Create a temp directory for the script's build output
// 	tempDir, err := ioutil.TempDir("", "script-build")
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create temporary build directory: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	// Compile the script
// 	outputBinary := filepath.Join(tempDir, "script")
// 	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
// 	var stderr bytes.Buffer
// 	cmd.Stderr = &stderr
// 	if err := cmd.Run(); err != nil {
// 		return "", fmt.Errorf("failed to compile script: %v, stderr: %s", err, stderr.String())
// 	}

// 	// Run the compiled script
// 	result := exec.Command(outputBinary)
// 	stdout, err := result.Output()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to run script: %v", err)
// 	}

// 	return string(stdout), nil
// }










