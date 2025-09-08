package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockTaskDispatcherInterface is a mock implementation of TaskDispatcherInterface
type MockTaskDispatcherInterface struct {
	mock.Mock
}

func (m *MockTaskDispatcherInterface) SubmitTaskFromScheduler(ctx context.Context, req *types.SchedulerTaskRequest) (*types.TaskManagerAPIResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TaskManagerAPIResponse), args.Error(1)
}

func TestTaskDispatcherHandler_Handle_SubmitTask_Success(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockDispatcher := &MockTaskDispatcherInterface{}
	handler := NewTaskDispatcherHandler(logger, mockDispatcher)

	// Test data
	req := &types.SchedulerTaskRequest{
		SendTaskDataToKeeper: types.SendTaskDataToKeeper{
			TaskID: []int64{123},
			TargetData: []types.TaskTargetData{
				{
					TaskID: 123,
					JobID:  nil, // Will be set by the handler
				},
			},
			TriggerData: []types.TaskTriggerData{
				{
					TaskID: 123,
				},
			},
			SchedulerID: 1,
		},
		Source: "test_scheduler",
	}

	expectedResp := &types.TaskManagerAPIResponse{
		Success:   true,
		TaskID:    []int64{123},
		Message:   "Task submitted successfully",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Expectations
	mockDispatcher.On("SubmitTaskFromScheduler", mock.Anything, req).Return(expectedResp, nil)

	// Execute
	result, err := handler.Handle(context.Background(), "submit-task", req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(*types.TaskManagerAPIResponse)
	assert.True(t, ok)
	assert.Equal(t, expectedResp.Success, resp.Success)
	assert.Equal(t, expectedResp.TaskID, resp.TaskID)
	assert.Equal(t, expectedResp.Message, resp.Message)

	mockDispatcher.AssertExpectations(t)
}

func TestTaskDispatcherHandler_Handle_SubmitTask_InvalidRequestType(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockDispatcher := &MockTaskDispatcherInterface{}
	handler := NewTaskDispatcherHandler(logger, mockDispatcher)

	// Test data - wrong type
	req := "invalid request type"

	// Execute
	result, err := handler.Handle(context.Background(), "submit-task", req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid request type for submit-task")
}

func TestTaskDispatcherHandler_Handle_UnknownMethod(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockDispatcher := &MockTaskDispatcherInterface{}
	handler := NewTaskDispatcherHandler(logger, mockDispatcher)

	// Execute
	result, err := handler.Handle(context.Background(), "unknown-method", nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown method: unknown-method")
}

func TestTaskDispatcherHandler_GetMethods(t *testing.T) {
	// Setup
	logger := logging.NewNoOpLogger()
	mockDispatcher := &MockTaskDispatcherInterface{}
	handler := NewTaskDispatcherHandler(logger, mockDispatcher)

	// Execute
	methods := handler.GetMethods()

	// Assert
	assert.Len(t, methods, 1)
	assert.Equal(t, "submit-task", methods[0].Name)
	assert.Equal(t, "Submit a task from schedulers to the dispatcher", methods[0].Description)
	assert.Equal(t, 30*time.Second, methods[0].Timeout)
	assert.NotNil(t, methods[0].RequestType)
	assert.NotNil(t, methods[0].ResponseType)
}
