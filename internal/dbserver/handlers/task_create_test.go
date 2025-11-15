package handlers

import (
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	dbtypes "github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	pkgtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type fakeTaskRepoCreate struct{}

func (f *fakeTaskRepoCreate) CreateTaskDataInDB(task *dbtypes.CreateTaskDataRequest) (int64, error) {
	return 0, assertErr{}
}
func (f *fakeTaskRepoCreate) AddTaskPerformerID(taskID int64, performerID int64) error { return nil }
func (f *fakeTaskRepoCreate) UpdateTaskExecutionDataInDB(task *dbtypes.UpdateTaskExecutionDataRequest) error {
	return nil
}
func (f *fakeTaskRepoCreate) UpdateTaskAttestationDataInDB(task *dbtypes.UpdateTaskAttestationDataRequest) error {
	return nil
}
func (f *fakeTaskRepoCreate) UpdateTaskNumberAndStatus(taskID int64, taskNumber int64, status string, txHash string) error {
	return nil
}
func (f *fakeTaskRepoCreate) GetTaskDataByID(taskID int64) (pkgtypes.TaskData, error) {
	return pkgtypes.TaskData{}, nil
}
func (f *fakeTaskRepoCreate) GetTasksByJobID(jobID *big.Int) ([]dbtypes.GetTasksByJobID, error) {
	return nil, nil
}
func (f *fakeTaskRepoCreate) AddTaskIDToJob(jobID *big.Int, taskID int64) error       { return nil }
func (f *fakeTaskRepoCreate) UpdateTaskFee(taskID int64, fee float64) error           { return nil }
func (f *fakeTaskRepoCreate) GetTaskFee(taskID int64) (float64, error)                { return 0, nil }
func (f *fakeTaskRepoCreate) GetCreatedChainIDByJobID(jobID *big.Int) (string, error) { return "", nil }

func TestCreateTaskData_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{taskRepository: &fakeTaskRepoCreate{}, logger: &MockLogger{}}
	r := gin.New()
	r.POST("/tasks", h.CreateTaskData)

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
