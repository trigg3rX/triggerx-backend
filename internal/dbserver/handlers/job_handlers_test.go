package handlers

import (
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	dbtypes "github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	pkgtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Fakes for repositories used in job handlers
type fakeJobRepo struct {
	jobByID    *pkgtypes.JobData
	jobByIDErr error
	fees       []dbtypes.TaskFeeResponse
	feesErr    error
}

func (f *fakeJobRepo) CreateNewJob(job *pkgtypes.JobData) (*big.Int, error) { return nil, nil }
func (f *fakeJobRepo) UpdateJobFromUserInDB(jobID *big.Int, job *dbtypes.UpdateJobDataFromUserRequest) error {
	return nil
}
func (f *fakeJobRepo) UpdateJobLastExecutedAt(jobID *big.Int, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error {
	return nil
}
func (f *fakeJobRepo) UpdateJobStatus(jobID *big.Int, status string) error { return nil }
func (f *fakeJobRepo) GetJobByID(jobID *big.Int) (*pkgtypes.JobData, error) {
	return f.jobByID, f.jobByIDErr
}
func (f *fakeJobRepo) GetTaskDefinitionIDByJobID(jobID *big.Int) (int, error) { return 0, nil }
func (f *fakeJobRepo) GetTaskFeesByJobID(jobID *big.Int) ([]dbtypes.TaskFeeResponse, error) {
	return f.fees, f.feesErr
}
func (f *fakeJobRepo) GetJobsByUserIDAndChainID(userID int64, createdChainID string) ([]pkgtypes.JobData, error) {
	return nil, nil
}

type fakeUserRepo struct {
	userID int64
	jobIDs []*big.Int
	err    error

	userLeaderboard    []typesUserEntry
	userLeaderboardErr error
}

// Minimal interfaces for methods referenced in handlers
func (f *fakeUserRepo) GetUserJobIDsByAddress(addr string) (int64, []*big.Int, error) {
	return f.userID, f.jobIDs, f.err
}
func (f *fakeUserRepo) GetUserLeaderboard() ([]typesUserEntry, error) {
	return f.userLeaderboard, f.userLeaderboardErr
}
func (f *fakeUserRepo) GetUserLeaderboardByAddress(addr string) (typesUserEntry, error) {
	if len(f.userLeaderboard) > 0 {
		return f.userLeaderboard[0], nil
	}
	return typesUserEntry{}, errors.New("not found")
}

type fakeApiKeysRepo struct {
	owner string
	err   error
}

func (f *fakeApiKeysRepo) CreateApiKey(apiKey *pkgtypes.ApiKey) error { return nil }
func (f *fakeApiKeysRepo) GetApiKeyDataByOwner(owner string) ([]*pkgtypes.ApiKey, error) {
	return nil, nil
}
func (f *fakeApiKeysRepo) GetApiKeyDataByKey(apiKey string) (*pkgtypes.ApiKey, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &pkgtypes.ApiKey{Owner: f.owner, IsActive: true}, nil
}
func (f *fakeApiKeysRepo) GetApiKeyCounters(key string) (*dbtypes.ApiKeyCounters, error) {
	return nil, nil
}
func (f *fakeApiKeysRepo) GetApiKeyByOwner(owner string) (string, error)       { return "", nil }
func (f *fakeApiKeysRepo) GetApiOwnerByApiKey(key string) (string, error)      { return f.owner, nil }
func (f *fakeApiKeysRepo) UpdateApiKey(req *dbtypes.UpdateApiKeyRequest) error { return nil }
func (f *fakeApiKeysRepo) UpdateApiKeyStatus(req *dbtypes.UpdateApiKeyStatusRequest) error {
	return nil
}
func (f *fakeApiKeysRepo) UpdateApiKeyLastUsed(key string, isSuccess bool) error { return nil }
func (f *fakeApiKeysRepo) DeleteApiKey(key string) error                         { return nil }

// Structs mirroring leaderboard entries used by handlers
type typesUserEntry = dbtypes.UserLeaderboardEntry

func TestGetJobDataByJobID_ValidationAndRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// missing param -> 404 by router; we test invalid and repo error
	{
		h := &Handler{jobRepository: &fakeJobRepo{}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/:job_id", h.GetJobDataByJobID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/abc", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	{
		fr := &fakeJobRepo{jobByIDErr: errors.New("boom")}
		h := &Handler{jobRepository: fr, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/:job_id", h.GetJobDataByJobID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/123", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	}

	{
		jd := &pkgtypes.JobData{JobID: pkgtypes.FromBigInt(big.NewInt(123))}
		fr := &fakeJobRepo{jobByID: jd}
		h := &Handler{jobRepository: fr, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/:job_id", h.GetJobDataByJobID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/123", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func TestGetTaskFeesByJobID_ValidationAndRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// invalid id
	{
		h := &Handler{jobRepository: &fakeJobRepo{}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/:job_id/task-fees", h.GetTaskFeesByJobID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/abc/task-fees", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	// repo error
	{
		fr := &fakeJobRepo{feesErr: errors.New("fail")}
		h := &Handler{jobRepository: fr, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/:job_id/task-fees", h.GetTaskFeesByJobID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/123/task-fees", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	}

	// success
	{
		fr := &fakeJobRepo{fees: []dbtypes.TaskFeeResponse{{TaskID: 1, TaskOpxCost: 1.2}}}
		h := &Handler{jobRepository: fr, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/:job_id/task-fees", h.GetTaskFeesByJobID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/123/task-fees", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func TestGetJobsByApiKey_HeaderAndOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// missing header
	{
		h := &Handler{apiKeysRepository: &fakeApiKeysRepo{}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/by-apikey", h.GetJobsByApiKey)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/by-apikey", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}

	// invalid api key
	{
		h := &Handler{apiKeysRepository: &fakeApiKeysRepo{err: errors.New("bad key")}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/by-apikey", h.GetJobsByApiKey)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/by-apikey", nil)
		req.Header.Set("X-Api-Key", "x")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	}

	// no owner
	{
		h := &Handler{apiKeysRepository: &fakeApiKeysRepo{owner: ""}, logger: &MockLogger{}}
		r := gin.New()
		r.GET("/jobs/by-apikey", h.GetJobsByApiKey)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/jobs/by-apikey", nil)
		req.Header.Set("X-Api-Key", "x")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	}
}

