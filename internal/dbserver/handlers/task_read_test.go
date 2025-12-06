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

// fakeTaskRepo implements the subset of TaskRepository used by task_read handlers
type fakeTaskRepo struct {
	// controls
	getTaskErr        error
	getTasksErr       error
	createdChainID    string
	createdChainIDErr error

	taskData           pkgtypes.TaskData
	tasksByJob         []dbtypes.GetTasksByJobID
	getRecentTasksResp []dbtypes.RecentTaskResponse
	getRecentTasksErr  error
}

func (f *fakeTaskRepo) CreateTaskDataInDB(task *dbtypes.CreateTaskDataRequest) (int64, error) {
	return 0, nil
}
func (f *fakeTaskRepo) AddTaskPerformerID(taskID int64, performerID int64) error { return nil }
func (f *fakeTaskRepo) UpdateTaskExecutionDataInDB(task *dbtypes.UpdateTaskExecutionDataRequest) error {
	return nil
}
func (f *fakeTaskRepo) UpdateTaskAttestationDataInDB(task *dbtypes.UpdateTaskAttestationDataRequest) error {
	return nil
}
func (f *fakeTaskRepo) UpdateTaskNumberAndStatus(taskID int64, taskNumber int64, status string, txHash string) error {
	return nil
}
func (f *fakeTaskRepo) GetTaskDataByID(taskID int64) (pkgtypes.TaskData, error) {
	return f.taskData, f.getTaskErr
}
func (f *fakeTaskRepo) GetTasksByJobID(jobID *big.Int) ([]dbtypes.GetTasksByJobID, error) {
	return f.tasksByJob, f.getTasksErr
}
func (f *fakeTaskRepo) AddTaskIDToJob(jobID *big.Int, taskID int64) error { return nil }
func (f *fakeTaskRepo) UpdateTaskFee(taskID int64, fee float64) error     { return nil }
func (f *fakeTaskRepo) GetTaskFee(taskID int64) (float64, error)          { return 0, nil }
func (f *fakeTaskRepo) GetCreatedChainIDByJobID(jobID *big.Int) (string, error) {
	return f.createdChainID, f.createdChainIDErr
}
func (f *fakeTaskRepo) GetRecentTasks(limit int) ([]dbtypes.RecentTaskResponse, error) {
	return f.getRecentTasksResp, f.getRecentTasksErr
}

func TestGetTaskDataByID_ErrorsAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Missing ID
	{
		h := &Handler{taskRepository: &fakeTaskRepo{}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/:id", h.GetTaskDataByID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound { // gin treats missing param route as 404
			t.Fatalf("expected 404 for missing id route, got %d", w.Code)
		}
	}

	// Invalid ID
	{
		h := &Handler{taskRepository: &fakeTaskRepo{}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/:id", h.GetTaskDataByID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid id, got %d", w.Code)
		}
	}

	// Repo not found error
	{
		fake := &fakeTaskRepo{getTaskErr: assertErr{}}
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/:id", h.GetTaskDataByID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 when repo returns error, got %d", w.Code)
		}
	}

	// Success
	{
		td := pkgtypes.TaskData{TaskID: 1, TaskNumber: 2}
		fake := &fakeTaskRepo{taskData: td}
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/:id", h.GetTaskDataByID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

// assertErr provides a non-nil error for testing
type assertErr struct{}

func (assertErr) Error() string { return "err" }

func TestGetTasksByJobID_ErrorsAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Missing job_id param path: gin will 404; also test invalid format explicitly
	{
		h := &Handler{taskRepository: &fakeTaskRepo{}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/job/:job_id", h.GetTasksByJobID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/job/", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for missing param route, got %d", w.Code)
		}
	}

	// Invalid job_id format
	{
		h := &Handler{taskRepository: &fakeTaskRepo{}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/job/:job_id", h.GetTasksByJobID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/job/abc", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid job_id, got %d", w.Code)
		}
	}

	// Repo error on GetTasksByJobID
	{
		fake := &fakeTaskRepo{getTasksErr: assertErr{}}
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/job/:job_id", h.GetTasksByJobID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/job/123", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 when repo returns error, got %d", w.Code)
		}
	}

	// Success with tx url build
	{
		tasks := []dbtypes.GetTasksByJobID{{
			TaskID:          1,
			TaskNumber:      1,
			TaskOpXCost:     1.5,
			ExecutionTxHash: "0xabc",
		}}
		fake := &fakeTaskRepo{tasksByJob: tasks, createdChainID: "11155111"}
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/job/:job_id", h.GetTasksByJobID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/job/123", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		// ensure explorer base added
		if !strings.Contains(w.Body.String(), "blockscout.com/tx/") {
			t.Fatalf("expected explorer URL, got %s", w.Body.String())
		}
	}
}
func TestGetRecentTasks_SuccessAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Default limit (200)
	{
		tasks := []dbtypes.RecentTaskResponse{
			{TaskID: 1, TaskNumber: 1, JobID: "123", TaskDefinitionID: 1},
			{TaskID: 2, TaskNumber: 2, JobID: "456", TaskDefinitionID: 2},
		}
		fake := &fakeTaskRepo{}
		fake.getRecentTasksResp = tasks
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/recent", h.GetRecentTasks)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/recent", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "\"count\":2") {
			t.Fatalf("expected count:2 in response, got %s", w.Body.String())
		}
	}

	// Custom limit
	{
		tasks := []dbtypes.RecentTaskResponse{
			{TaskID: 1, TaskNumber: 1, JobID: "123"},
		}
		fake := &fakeTaskRepo{}
		fake.getRecentTasksResp = tasks
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/recent", h.GetRecentTasks)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/recent?limit=50", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "\"limit\":50") {
			t.Fatalf("expected limit:50 in response, got %s", w.Body.String())
		}
	}

	// Limit exceeding max (should cap at 200)
	{
		tasks := []dbtypes.RecentTaskResponse{}
		fake := &fakeTaskRepo{}
		fake.getRecentTasksResp = tasks
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/recent", h.GetRecentTasks)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/recent?limit=300", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "\"limit\":200") {
			t.Fatalf("expected limit capped at 200, got %s", w.Body.String())
		}
	}

	// Invalid limit
	{
		h := &Handler{taskRepository: &fakeTaskRepo{}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/recent", h.GetRecentTasks)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/recent?limit=invalid", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid limit, got %d", w.Code)
		}
	}

	// Repository error
	{
		fake := &fakeTaskRepo{getRecentTasksErr: assertErr{}}
		h := &Handler{taskRepository: fake, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/tasks/recent", h.GetRecentTasks)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/recent", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 on repo error, got %d", w.Code)
		}
	}
}
