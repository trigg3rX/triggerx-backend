package execution

import (
	// "bytes"
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	// "os"
	// "os/exec"
	// "path/filepath"
	// "strings"
)

// func (e *JobExecutor) evaluateConditionScript(scriptUrl string) (bool, error) {
// 	// Fetch script content from IPFS
// 	scriptContent, err := e.fetchFromIPFS(scriptUrl)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to fetch condition script: %v", err)
// 	}

// 	// Create a temporary file for the script
// 	tempFile, err := ioutil.TempFile("", "condition-*.go")
// 	if err != nil {
// 		return false, fmt.Errorf("failed to create temporary file: %v", err)
// 	}
// 	defer os.Remove(tempFile.Name())

// 	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
// 		return false, fmt.Errorf("failed to write script to file: %v", err)
// 	}
// 	if err := tempFile.Close(); err != nil {
// 		return false, fmt.Errorf("failed to close temporary file: %v", err)
// 	}

// 	// Create a temp directory for the script's build output
// 	tempDir, err := ioutil.TempDir("", "condition-build")
// 	if err != nil {
// 		return false, fmt.Errorf("failed to create temporary build directory: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	// Compile the script
// 	outputBinary := filepath.Join(tempDir, "condition")
// 	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
// 	var stderr bytes.Buffer
// 	cmd.Stderr = &stderr
// 	if err := cmd.Run(); err != nil {
// 		return false, fmt.Errorf("failed to compile condition script: %v, stderr: %s", err, stderr.String())
// 	}

// 	// Run the compiled script
// 	result := exec.Command(outputBinary)
// 	stdout, err := result.Output()
// 	if err != nil {
// 		return false, fmt.Errorf("failed to run condition script: %v", err)
// 	}

// 	// Parse the output to determine if condition is satisfied
// 	// Look for a line containing "Condition satisfied: true" or "Condition satisfied: false"
// 	lines := strings.Split(string(stdout), "\n")
// 	for _, line := range lines {
// 		if strings.Contains(line, "Condition satisfied: true") {
// 			return true, nil
// 		} else if strings.Contains(line, "Condition satisfied: false") {
// 			return false, nil
// 		}
// 	}

// 	// If no explicit condition found, try parsing as JSON
// 	var conditionResult struct {
// 		Satisfied bool `json:"satisfied"`
// 	}
// 	if err := json.Unmarshal(stdout, &conditionResult); err != nil {
// 		return false, fmt.Errorf("could not determine condition result from output: %s", string(stdout))
// 	}

// 	return conditionResult.Satisfied, nil
// }

// func (e *JobExecutor) fetchArgumentsFromIPFS(scriptIPFSUrl string) ([]string, error) {
// 	// Fetch script content from IPFS
// 	scriptContent, err := e.fetchFromIPFS(scriptIPFSUrl)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch arguments script: %v", err)
// 	}

// 	// Create a temporary file for the script
// 	tempFile, err := ioutil.TempFile("", "args-*.go")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create temporary file: %v", err)
// 	}
// 	defer os.Remove(tempFile.Name())

// 	if _, err := tempFile.Write([]byte(scriptContent)); err != nil {
// 		return nil, fmt.Errorf("failed to write script to file: %v", err)
// 	}
// 	if err := tempFile.Close(); err != nil {
// 		return nil, fmt.Errorf("failed to close temporary file: %v", err)
// 	}

// 	// Create a temp directory for the script's build output
// 	tempDir, err := ioutil.TempDir("", "args-build")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create temporary build directory: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	// Compile the script
// 	outputBinary := filepath.Join(tempDir, "args")
// 	cmd := exec.Command("go", "build", "-o", outputBinary, tempFile.Name())
// 	var stderr bytes.Buffer
// 	cmd.Stderr = &stderr
// 	if err := cmd.Run(); err != nil {
// 		return nil, fmt.Errorf("failed to compile args script: %v, stderr: %s", err, stderr.String())
// 	}

// 	// Run the compiled script
// 	result := exec.Command(outputBinary)
// 	stdout, err := result.Output()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to run args script: %v", err)
// 	}

// 	// Parse the output to get the arguments
// 	// First try parsing as JSON array
// 	var jsonOutput []string
// 	if err := json.Unmarshal(stdout, &jsonOutput); err == nil {
// 		return jsonOutput, nil
// 	}

// 	// If JSON parsing fails, look for checker function payload format
// 	lines := strings.Split(string(stdout), "\n")
// 	for _, line := range lines {
// 		if strings.Contains(line, "Payload received:") {
// 			payload := strings.TrimSpace(strings.TrimPrefix(line, "Payload received:"))
// 			return []string{payload}, nil
// 		}
// 	}

// 	// If no structured format is found, use the entire output as a single argument
// 	return []string{string(stdout)}, nil
// }

// func (e *JobExecutor) parseScriptOutput(output string) (interface{}, error) {
// 	// Try to parse as JSON first
// 	var jsonData interface{}
// 	if err := json.Unmarshal([]byte(output), &jsonData); err == nil {
// 		return jsonData, nil
// 	}

// 	// Fallback to line-based parsing
// 	lines := strings.Split(output, "\n")
// 	for _, line := range lines {
// 		if strings.Contains(line, "Payload received:") {
// 			payload := strings.TrimSpace(strings.TrimPrefix(line, "Payload received:"))
// 			return payload, nil
// 		}
// 	}

// 	// Return raw output as last resort
// 	return output, nil
// }
