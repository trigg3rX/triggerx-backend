package handlers

import (
	"context"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	dexconfig "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/config"
	dexexec "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/execution"
	dextypes "github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor/types"
)

// helper to create execution result with given cost and success
func newExecResult(cost int64, success bool, execErr error) *dextypes.ExecutionResult {
	return &dextypes.ExecutionResult{
		Stats:   dextypes.DockerResourceStats{TotalCost: big.NewInt(cost)},
		Success: success,
		Error:   execErr,
	}
}

// FakeDockerExecutor is a lightweight test double for DockerExecutorAPI
type FakeDockerExecutor struct {
	responses map[string]*dextypes.ExecutionResult
	errors    map[string]error
}

func NewFakeDockerExecutor() *FakeDockerExecutor {
	return &FakeDockerExecutor{
		responses: make(map[string]*dextypes.ExecutionResult),
		errors:    make(map[string]error),
	}
}

func (f *FakeDockerExecutor) Initialize(ctx context.Context) error { return nil }
func (f *FakeDockerExecutor) Execute(ctx context.Context, fileURL string, fileLanguage string, noOfAttesters int) (*dextypes.ExecutionResult, error) {
	if err, ok := f.errors[fileURL]; ok {
		return nil, err
	}
	if res, ok := f.responses[fileURL]; ok {
		return res, nil
	}
	return newExecResult(0, true, nil), nil
}
func (f *FakeDockerExecutor) ExecuteSource(ctx context.Context, code string, language string) (*dextypes.ExecutionResult, error) {
	if err, ok := f.errors[code]; ok {
		return nil, err
	}
	if res, ok := f.responses[code]; ok {
		return res, nil
	}
	return newExecResult(0, true, nil), nil
}
func (f *FakeDockerExecutor) GetHealthStatus() *dexexec.HealthStatus { return &dexexec.HealthStatus{} }
func (f *FakeDockerExecutor) GetExecutionFeeConfig() dexconfig.ExecutionFeeConfig {
	return dexconfig.ExecutionFeeConfig{}
}
func (f *FakeDockerExecutor) GetStats() *dextypes.PerformanceMetrics {
	return &dextypes.PerformanceMetrics{}
}
func (f *FakeDockerExecutor) GetAllPoolStats() map[dextypes.Language]*dextypes.PoolStats {
	return map[dextypes.Language]*dextypes.PoolStats{}
}
func (f *FakeDockerExecutor) GetPoolStats(language dextypes.Language) *dextypes.PoolStats {
	return &dextypes.PoolStats{}
}
func (f *FakeDockerExecutor) GetLanguageStats(language dextypes.Language) (*dextypes.PoolStats, bool) {
	return &dextypes.PoolStats{}, true
}
func (f *FakeDockerExecutor) GetSupportedLanguages() []dextypes.Language {
	return []dextypes.Language{}
}
func (f *FakeDockerExecutor) IsLanguageSupported(language dextypes.Language) bool { return true }
func (f *FakeDockerExecutor) GetActiveExecutions() []*dextypes.ExecutionContext {
	return []*dextypes.ExecutionContext{}
}
func (f *FakeDockerExecutor) CancelExecution(executionID string) error { return nil }
func (f *FakeDockerExecutor) GetAlerts(severity string, limit int) []dexexec.Alert {
	return []dexexec.Alert{}
}
func (f *FakeDockerExecutor) ClearAlerts()                    {}
func (f *FakeDockerExecutor) Close(ctx context.Context) error { return nil }

func TestCalculateTaskFees_EmptyInput(t *testing.T) {
	fake := NewFakeDockerExecutor()

	h := &Handler{
		dockerExecutor: fake,
		logger:         &MockLogger{},
	}

	total, err := h.CalculateTaskFees("")
	if err == nil {
		t.Fatalf("expected error for empty input, got nil")
	}
	if total.Cmp(big.NewInt(0)) != 0 {
		t.Fatalf("expected zero total, got %s", total.String())
	}
}

func TestCalculateTaskFees_SingleURL_Success(t *testing.T) {
	fake := NewFakeDockerExecutor()
	fake.responses["ipfs://file1"] = newExecResult(123, true, nil)

	h := &Handler{
		dockerExecutor: fake,
		logger:         &MockLogger{},
	}

	total, err := h.CalculateTaskFees("ipfs://file1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total.Cmp(big.NewInt(123)) != 0 {
		t.Fatalf("expected 123, got %s", total.String())
	}
}

func TestCalculateTaskFees_MultipleURLs_PartialFailures(t *testing.T) {
	fake := NewFakeDockerExecutor()
	fake.responses["ipfs://file1"] = newExecResult(100, true, nil)
	fake.responses["ipfs://file2"] = newExecResult(999, false, nil)
	fake.errors["ipfs://file3"] = context.DeadlineExceeded

	h := &Handler{
		dockerExecutor: fake,
		logger:         &MockLogger{},
	}

	total, err := h.CalculateTaskFees("ipfs://file1, ipfs://file2, ipfs://file3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("expected 100, got %s", total.String())
	}
}

func TestGetTaskFees_HTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fake := NewFakeDockerExecutor()
	fake.responses["ipfs://a"] = newExecResult(50, true, nil)
	fake.responses["ipfs://b"] = newExecResult(70, true, nil)

	h := &Handler{dockerExecutor: fake, logger: &MockLogger{}}

	r := gin.New()
	r.GET("/fees", h.GetTaskFees)

	req := httptest.NewRequest(http.MethodGet, "/fees?ipfs_url="+strings.ReplaceAll("ipfs://a,ipfs://b", " ", "%20"), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// The handler returns a big.Int; encoding can vary. Ensure body contains 120.
	if !strings.Contains(w.Body.String(), "120") {
		t.Fatalf("expected response to contain 120, got %s", w.Body.String())
	}
}
