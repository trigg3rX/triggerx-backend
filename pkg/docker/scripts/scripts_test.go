package scripts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)

// TestGetInitializationScript tests that the correct initialization script is returned for each supported language.
func TestGetInitializationScript(t *testing.T) {
	testCases := []struct {
		name              string
		language          types.Language
		expectedSubstring string
		notExpected       string
	}{
		{
			name:              "Go Language",
			language:          types.LanguageGo,
			expectedSubstring: "go mod init code",
			notExpected:       "code.py",
		},
		{
			name:              "Python Language",
			language:          types.LanguagePy,
			expectedSubstring: "print(\"init\")",
			notExpected:       "code.go",
		},
		{
			name:              "JavaScript Language",
			language:          types.LanguageJS,
			expectedSubstring: "console.log(\"init\");",
			notExpected:       "code.go",
		},
		{
			name:              "Node Language",
			language:          types.LanguageNode,
			expectedSubstring: "console.log(\"init\");",
			notExpected:       "code.py",
		},
		{
			name:              "TypeScript Language",
			language:          types.LanguageTS,
			expectedSubstring: "npm install -g typescript",
			notExpected:       "code.js",
		},
		{
			name:              "Unknown Language",
			language:          "unknown",
			expectedSubstring: "Container initialized successfully",
			notExpected:       "go mod init",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			script := GetInitializationScript(tc.language)

			// All scripts should have this basic setup
			assert.Contains(t, script, "#!/bin/sh")
			assert.Contains(t, script, "set -e")
			assert.Contains(t, script, "mkdir -p /code")

			// Check for language-specific content
			assert.Contains(t, script, tc.expectedSubstring)
			if tc.notExpected != "" {
				assert.NotContains(t, script, tc.notExpected)
			}
		})
	}
}

// TestGetSetupScript tests that the correct setup script is returned for each supported language.
func TestGetSetupScript(t *testing.T) {
	testCases := []struct {
		name              string
		language          types.Language
		expectedSubstring string
		shouldReturnEmpty bool
	}{
		{
			name:              "Go Language",
			language:          types.LanguageGo,
			expectedSubstring: "go mod tidy",
		},
		{
			name:              "Python Language",
			language:          types.LanguagePy,
			expectedSubstring: "pip install -r requirements.txt",
		},
		{
			name:              "JavaScript Language",
			language:          types.LanguageJS,
			expectedSubstring: "npm install",
		},
		{
			name:              "Node Language",
			language:          types.LanguageNode,
			expectedSubstring: "npm install",
		},
		{
			name:              "TypeScript Language",
			language:          types.LanguageTS,
			expectedSubstring: "npm install",
		},
		{
			name:              "Unknown Language",
			language:          "unknown",
			shouldReturnEmpty: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			script := GetSetupScript(tc.language)

			if tc.shouldReturnEmpty {
				assert.Empty(t, script)
			} else {
				// All setup scripts should have this basic structure
				assert.Contains(t, script, "#!/bin/sh")
				assert.Contains(t, script, "set -e")
				assert.Contains(t, script, "cd /code")
				assert.Contains(t, script, "if [ ! -f /code/.warm ]")
				assert.Contains(t, script, "touch /code/.warm")

				// Check for language-specific content
				assert.Contains(t, script, tc.expectedSubstring)
			}
		})
	}
}

// TestGetExecutionScript tests that the correct execution script is returned for each supported language.
func TestGetExecutionScript(t *testing.T) {
	testCases := []struct {
		name              string
		language          types.Language
		expectedSubstring string
		shouldReturnError bool
	}{
		{
			name:              "Go Language",
			language:          types.LanguageGo,
			expectedSubstring: "go run code.go",
		},
		{
			name:              "Python Language",
			language:          types.LanguagePy,
			expectedSubstring: "python -u -B code.py",
		},
		{
			name:              "JavaScript Language",
			language:          types.LanguageJS,
			expectedSubstring: "node code.js",
		},
		{
			name:              "Node Language",
			language:          types.LanguageNode,
			expectedSubstring: "node code.js",
		},
		{
			name:              "TypeScript Language",
			language:          types.LanguageTS,
			expectedSubstring: "tsc code.ts",
		},
		{
			name:              "Unknown Language",
			language:          "unknown",
			shouldReturnError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			script := GetExecutionScript(tc.language)

			if tc.shouldReturnError {
				assert.Contains(t, script, "Error: Unsupported language.")
				assert.Contains(t, script, "exit 1")
			} else {
				// All execution scripts should have this basic structure
				assert.Contains(t, script, "#!/bin/sh")
				assert.Contains(t, script, "set -e")
				assert.Contains(t, script, "cd /code")

				// Check for language-specific content
				assert.Contains(t, script, tc.expectedSubstring)
			}
		})
	}
}

// TestGetCleanupScript tests that the correct cleanup script is returned for each supported language.
func TestGetCleanupScript(t *testing.T) {
	testCases := []struct {
		name              string
		language          types.Language
		expectedSubstring string
		notExpected       string
	}{
		{
			name:              "Go Language",
			language:          types.LanguageGo,
			expectedSubstring: "rm -f code.go",
			notExpected:       "rm -f code.py",
		},
		{
			name:              "Python Language",
			language:          types.LanguagePy,
			expectedSubstring: "rm -f code.py",
			notExpected:       "rm -f code.go",
		},
		{
			name:              "JavaScript Language",
			language:          types.LanguageJS,
			expectedSubstring: "rm -f code.js",
			notExpected:       "rm -f code.py",
		},
		{
			name:              "Node Language",
			language:          types.LanguageNode,
			expectedSubstring: "rm -f code.js",
			notExpected:       "rm -f code.py",
		},
		{
			name:              "TypeScript Language",
			language:          types.LanguageTS,
			expectedSubstring: "rm -f code.ts",
			notExpected:       "rm -f code.py",
		},
		{
			name:              "Unknown Language",
			language:          "unknown",
			expectedSubstring: "rm -f code.*",
			notExpected:       "go mod init",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			script := GetCleanupScript(tc.language)

			// All cleanup scripts should have this basic structure
			assert.Contains(t, script, "#!/bin/sh")
			assert.Contains(t, script, "cd /code")

			// Check for language-specific content
			assert.Contains(t, script, tc.expectedSubstring)
			if tc.notExpected != "" {
				assert.NotContains(t, script, tc.notExpected)
			}
		})
	}
}

// TestScriptConstants tests that the script constants are properly defined and contain expected content.
func TestScriptConstants(t *testing.T) {
	// Test initialization scripts
	assert.Contains(t, goInitializationScript, "go mod init code")
	assert.Contains(t, pythonInitializationScript, "print(\"init\")")
	assert.Contains(t, javascriptInitializationScript, "console.log(\"init\");")
	assert.Contains(t, typescriptInitializationScript, "npm install -g typescript")

	// Test setup scripts
	assert.Contains(t, goSetupScript, "go mod tidy")
	assert.Contains(t, pythonSetupScript, "pip install -r requirements.txt")
	assert.Contains(t, javascriptSetupScript, "npm install")
	assert.Contains(t, typescriptSetupScript, "npm install")

	// Test execution scripts
	assert.Contains(t, goExecutionScript, "go run code.go")
	assert.Contains(t, pythonExecutionScript, "python -u -B code.py")
	assert.Contains(t, javascriptExecutionScript, "node code.js")
	assert.Contains(t, typescriptExecutionScript, "tsc code.ts")

	// Test cleanup scripts
	assert.Contains(t, goCleanupScript, "rm -f code.go")
	assert.Contains(t, pythonCleanupScript, "rm -f code.py")
	assert.Contains(t, javascriptCleanupScript, "rm -f code.js")
	assert.Contains(t, typescriptCleanupScript, "rm -f code.ts")
}

// TestScriptConsistency tests that scripts for the same language are consistent across functions.
func TestScriptConsistency(t *testing.T) {
	// Test that JS and Node have the same scripts (as mentioned in comments)
	jsInit := GetInitializationScript(types.LanguageJS)
	nodeInit := GetInitializationScript(types.LanguageNode)
	assert.Equal(t, jsInit, nodeInit)

	jsSetup := GetSetupScript(types.LanguageJS)
	nodeSetup := GetSetupScript(types.LanguageNode)
	assert.Equal(t, jsSetup, nodeSetup)

	jsExec := GetExecutionScript(types.LanguageJS)
	nodeExec := GetExecutionScript(types.LanguageNode)
	assert.Equal(t, jsExec, nodeExec)

	jsCleanup := GetCleanupScript(types.LanguageJS)
	nodeCleanup := GetCleanupScript(types.LanguageNode)
	assert.Equal(t, jsCleanup, nodeCleanup)
}

// TestScriptSafety tests that scripts contain proper error handling and safety measures.
func TestScriptSafety(t *testing.T) {
	languages := []types.Language{types.LanguageGo, types.LanguagePy, types.LanguageJS, types.LanguageNode, types.LanguageTS}

	for _, lang := range languages {
		t.Run(string(lang), func(t *testing.T) {
			// Test initialization script safety
			initScript := GetInitializationScript(lang)
			assert.Contains(t, initScript, "set -e")
			assert.Contains(t, initScript, "mkdir -p /code")

			// Test setup script safety (if it exists)
			setupScript := GetSetupScript(lang)
			if setupScript != "" {
				assert.Contains(t, setupScript, "set -e")
				assert.Contains(t, setupScript, "cd /code")
			}

			// Test execution script safety
			execScript := GetExecutionScript(lang)
			assert.Contains(t, execScript, "set -e")
			assert.Contains(t, execScript, "cd /code")

			// Test cleanup script safety
			cleanupScript := GetCleanupScript(lang)
			assert.Contains(t, cleanupScript, "cd /code")
		})
	}
}

// TestMemoryLimits tests that memory limits are properly set for JavaScript and TypeScript.
func TestMemoryLimits(t *testing.T) {
	jsScript := GetExecutionScript(types.LanguageJS)
	tsScript := GetExecutionScript(types.LanguageTS)

	assert.Contains(t, jsScript, "V8_MEMORY_LIMIT=${V8_MEMORY_LIMIT:-256}")
	assert.Contains(t, jsScript, "--max-old-space-size=${V8_MEMORY_LIMIT}")

	assert.Contains(t, tsScript, "V8_MEMORY_LIMIT=${V8_MEMORY_LIMIT:-256}")
	assert.Contains(t, tsScript, "--max-old-space-size=${V8_MEMORY_LIMIT}")
}

// TestWarmupMechanism tests that warmup mechanisms are properly implemented in setup scripts.
func TestWarmupMechanism(t *testing.T) {
	languages := []types.Language{types.LanguageGo, types.LanguagePy, types.LanguageJS, types.LanguageNode, types.LanguageTS}

	for _, lang := range languages {
		t.Run(string(lang), func(t *testing.T) {
			setupScript := GetSetupScript(lang)
			if setupScript != "" {
				assert.Contains(t, setupScript, "if [ ! -f /code/.warm ]")
				assert.Contains(t, setupScript, "touch /code/.warm")
			}
		})
	}
}
