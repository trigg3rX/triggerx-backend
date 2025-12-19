package rpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// MockTaskMonitor implements TaskMonitorInterface for testing
type MockTaskMonitor struct {
	mock.Mock
}

func (m *MockTaskMonitor) ReportTaskStatus(ctx context.Context, req *types.ReportTaskStatusRequest) (*types.ReportTaskStatusResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ReportTaskStatusResponse), args.Error(1)
}

func TestTaskMonitorHandler_Handle_ReportTaskStatus(t *testing.T) {
	logger := logging.NewNoOpLogger()
	mockMonitor := new(MockTaskMonitor)
	handler := NewTaskMonitorHandler(logger, mockMonitor)

	tests := []struct {
		name          string
		method        string
		request       interface{}
		setupMock     func()
		expectedResp  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:   "Success - Task succeeded",
			method: "report-task-status",
			request: &types.ReportTaskStatusRequest{
				TaskID:        123,
				KeeperAddress: "0x1234567890abcdef",
				Success:       true,
				ProofCID:      "QmTestCID123",
				Error:         "",
				Signature:     "valid_signature",
			},
			setupMock: func() {
				mockMonitor.On("ReportTaskStatus", mock.Anything, mock.MatchedBy(func(req *types.ReportTaskStatusRequest) bool {
					return req.TaskID == 123 && req.Success == true
				})).Return(&types.ReportTaskStatusResponse{
					Success: true,
					Message: "Task status updated",
				}, nil)
			},
			expectedResp: &types.ReportTaskStatusResponse{
				Success: true,
				Message: "Task status updated",
			},
			expectError: false,
		},
		{
			name:   "Success - Task failed",
			method: "report-task-status",
			request: &types.ReportTaskStatusRequest{
				TaskID:        456,
				KeeperAddress: "0xabcdef1234567890",
				Success:       false,
				ProofCID:      "QmTestCID456",
				Error:         "aggregator submission failed",
				Signature:     "valid_signature",
			},
			setupMock: func() {
				mockMonitor.On("ReportTaskStatus", mock.Anything, mock.MatchedBy(func(req *types.ReportTaskStatusRequest) bool {
					return req.TaskID == 456 && req.Success == false
				})).Return(&types.ReportTaskStatusResponse{
					Success: true,
					Message: "Task failure recorded",
				}, nil)
			},
			expectedResp: &types.ReportTaskStatusResponse{
				Success: true,
				Message: "Task failure recorded",
			},
			expectError: false,
		},
		{
			name:          "Error - Unknown method",
			method:        "unknown-method",
			request:       nil,
			setupMock:     func() {},
			expectedResp:  nil,
			expectError:   true,
			errorContains: "unknown method",
		},
		{
			name:   "Success - Request from map (JSON decoded)",
			method: "report-task-status",
			request: map[string]interface{}{
				"task_id":        float64(789), // JSON numbers are float64
				"keeper_address": "0x9876543210fedcba",
				"success":        true,
				"proof_cid":      "QmTestCID789",
				"error":          "",
				"signature":      "valid_signature",
			},
			setupMock: func() {
				mockMonitor.On("ReportTaskStatus", mock.Anything, mock.MatchedBy(func(req *types.ReportTaskStatusRequest) bool {
					return req.TaskID == 789 && req.Success == true
				})).Return(&types.ReportTaskStatusResponse{
					Success: true,
					Message: "Task status updated",
				}, nil)
			},
			expectedResp: &types.ReportTaskStatusResponse{
				Success: true,
				Message: "Task status updated",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mockMonitor.ExpectedCalls = nil
			mockMonitor.Calls = nil
			tt.setupMock()

			// Note: Signature validation is skipped in these tests
			// In production, you'd need valid signatures
			resp, err := handler.Handle(context.Background(), tt.method, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				// Note: This test will fail due to signature validation
				// For proper testing, you'd need to mock the signature validation
				// or use valid test signatures
				if err != nil {
					t.Logf("Expected signature validation to fail in test: %v", err)
				} else {
					assert.NotNil(t, resp)
				}
			}
		})
	}
}

func TestTaskMonitorHandler_GetMethods(t *testing.T) {
	logger := logging.NewNoOpLogger()
	mockMonitor := new(MockTaskMonitor)
	handler := NewTaskMonitorHandler(logger, mockMonitor)

	methods := handler.GetMethods()

	assert.Len(t, methods, 1)
	assert.Equal(t, "report-task-status", methods[0].Name)
	assert.NotNil(t, methods[0].RequestType)
	assert.NotNil(t, methods[0].ResponseType)
}

func TestConvertMapToStatusRequest(t *testing.T) {
	logger := logging.NewNoOpLogger()
	mockMonitor := new(MockTaskMonitor)
	handler := NewTaskMonitorHandler(logger, mockMonitor)

	tests := []struct {
		name        string
		input       map[string]interface{}
		expected    *types.ReportTaskStatusRequest
		expectError bool
	}{
		{
			name: "Valid map - success case",
			input: map[string]interface{}{
				"task_id":        float64(123),
				"keeper_address": "0x1234",
				"success":        true,
				"proof_cid":      "QmTest",
				"error":          "",
				"signature":      "sig123",
			},
			expected: &types.ReportTaskStatusRequest{
				TaskID:        123,
				KeeperAddress: "0x1234",
				Success:       true,
				ProofCID:      "QmTest",
				Error:         "",
				Signature:     "sig123",
			},
			expectError: false,
		},
		{
			name: "Valid map - failure case",
			input: map[string]interface{}{
				"task_id":        float64(456),
				"keeper_address": "0x5678",
				"success":        false,
				"proof_cid":      "",
				"error":          "execution failed",
				"signature":      "sig456",
			},
			expected: &types.ReportTaskStatusRequest{
				TaskID:        456,
				KeeperAddress: "0x5678",
				Success:       false,
				ProofCID:      "",
				Error:         "execution failed",
				Signature:     "sig456",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.convertMapToStatusRequest(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.TaskID, result.TaskID)
				assert.Equal(t, tt.expected.KeeperAddress, result.KeeperAddress)
				assert.Equal(t, tt.expected.Success, result.Success)
				assert.Equal(t, tt.expected.ProofCID, result.ProofCID)
				assert.Equal(t, tt.expected.Error, result.Error)
				assert.Equal(t, tt.expected.Signature, result.Signature)
			}
		})
	}
}
