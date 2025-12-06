package handlers

import (
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	dbtypes "github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	pkgtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type fakeTaskRepoUpdate struct {
	updateExecErr error
	updateAttErr  error
	updateFeeErr  error
}

func (f *fakeTaskRepoUpdate) CreateTaskDataInDB(task *dbtypes.CreateTaskDataRequest) (int64, error) {
	return 0, nil
}
func (f *fakeTaskRepoUpdate) AddTaskPerformerID(taskID int64, performerID int64) error { return nil }
func (f *fakeTaskRepoUpdate) UpdateTaskExecutionDataInDB(task *dbtypes.UpdateTaskExecutionDataRequest) error {
	return f.updateExecErr
}
func (f *fakeTaskRepoUpdate) UpdateTaskAttestationDataInDB(task *dbtypes.UpdateTaskAttestationDataRequest) error {
	return f.updateAttErr
}
func (f *fakeTaskRepoUpdate) UpdateTaskNumberAndStatus(taskID int64, taskNumber int64, status string, txHash string) error {
	return nil
}

// Methods to satisfy repository.TaskRepository
func (f *fakeTaskRepoUpdate) GetTaskDataByID(taskID int64) (pkgtypes.TaskData, error) {
	return pkgtypes.TaskData{}, nil
}
func (f *fakeTaskRepoUpdate) GetTasksByJobID(jobID *big.Int) ([]dbtypes.GetTasksByJobID, error) {
	return nil, nil
}
func (f *fakeTaskRepoUpdate) AddTaskIDToJob(jobID *big.Int, taskID int64) error       { return nil }
func (f *fakeTaskRepoUpdate) UpdateTaskFee(taskID int64, fee float64) error           { return f.updateFeeErr }
func (f *fakeTaskRepoUpdate) GetTaskFee(taskID int64) (float64, error)                { return 0, nil }
func (f *fakeTaskRepoUpdate) GetCreatedChainIDByJobID(jobID *big.Int) (string, error) { return "", nil }
func (f *fakeTaskRepoUpdate) GetRecentTasks(limit int) ([]dbtypes.RecentTaskResponse, error) {
	return []dbtypes.RecentTaskResponse{}, nil
}

func TestUpdateTaskExecutionData_ValidationAndRepoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// invalid body
	{
		h := &Handler{taskRepository: &fakeTaskRepoUpdate{}, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/execution/:id", h.UpdateTaskExecutionData)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/execution/1", strings.NewReader("{"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	// missing required fields
	{
		h := &Handler{taskRepository: &fakeTaskRepoUpdate{}, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/execution/:id", h.UpdateTaskExecutionData)

		body := `{"task_id":0}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/execution/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	// repo error
	{
		fake := &fakeTaskRepoUpdate{updateExecErr: assertErr{}}
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/execution/:id", h.UpdateTaskExecutionData)

		ts := time.Now().UTC()
		body := `{"task_id":1,"execution_timestamp":"` + ts.Format(time.RFC3339Nano) + `","execution_tx_hash":"0xabc"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/execution/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 on repo error, got %d", w.Code)
		}
	}
}

func TestUpdateTaskAttestationData_ValidationAndRepoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// invalid body
	{
		h := &Handler{taskRepository: &fakeTaskRepoUpdate{}, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/:id/attestation", h.UpdateTaskAttestationData)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/1/attestation", strings.NewReader("{"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	// missing required fields
	{
		h := &Handler{taskRepository: &fakeTaskRepoUpdate{}, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/:id/attestation", h.UpdateTaskAttestationData)

		body := `{"task_id":0}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/1/attestation", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	// repo error
	{
		fake := &fakeTaskRepoUpdate{updateAttErr: assertErr{}}
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/:id/attestation", h.UpdateTaskAttestationData)

		body := `{"task_id":1,"task_number":1,"task_attester_ids":[1],"tp_signature":"a","ta_signature":"b","task_submission_tx_hash":"0xabc"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/1/attestation", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 on repo error, got %d", w.Code)
		}
	}
}

func TestUpdateTaskFee_ValidationAndRepoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// invalid body
	{
		h := &Handler{taskRepository: &fakeTaskRepoUpdate{}, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/:id/fee", h.UpdateTaskFee)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/1/fee", strings.NewReader("{"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	// invalid id
	{
		h := &Handler{taskRepository: &fakeTaskRepoUpdate{}, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/:id/fee", h.UpdateTaskFee)

		body := `{"fee":1.2}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/abc/fee", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	// repo error
	{
		fake := &fakeTaskRepoUpdate{updateFeeErr: assertErr{}}
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.PUT("/tasks/:id/fee", h.UpdateTaskFee)

		body := `{"fee":1.2}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/1/fee", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 on repo error, got %d", w.Code)
		}
	}
}
