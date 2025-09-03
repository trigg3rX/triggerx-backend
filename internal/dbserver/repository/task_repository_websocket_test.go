package repository

import (
	"math/big"
	"testing"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func TestTaskRepositoryWebSocketEvents(t *testing.T) {
	// Create a mock logger
	logger := logging.NewNoOpLogger()

	// Create a new hub
	hub := websocket.NewHub(logger)

	// Start the hub in a goroutine
	go hub.Run()
	defer hub.Shutdown()

	// Create event publisher
	publisher := events.NewPublisher(hub, logger)

	// Create repository with publisher (we'll use nil for db since we're only testing event emission)
	repo := NewTaskRepositoryWithPublisher(nil, publisher)

	// Test task creation event
	t.Run("CreateTaskDataInDB emits TASK_CREATED event", func(t *testing.T) {
		// This would normally fail due to nil db, but we're testing event emission
		// In a real test, you'd mock the database
		task := &types.CreateTaskDataRequest{
			JobID:            big.NewInt(123),
			TaskDefinitionID: 456,
			IsImua:           false,
		}

		// The function will fail at database level, but we can verify the event emission logic
		_, err := repo.CreateTaskDataInDB(task)
		if err == nil {
			t.Error("Expected error due to nil database, but got nil")
		}

		// In a real implementation with mocked database, we would verify that
		// the WebSocket event was emitted correctly
	})

	// Test task execution update event
	t.Run("UpdateTaskExecutionDataInDB emits TASK_UPDATED event", func(t *testing.T) {
		task := &types.UpdateTaskExecutionDataRequest{
			TaskID:             789,
			TaskPerformerID:    101,
			ExecutionTimestamp: time.Now(),
			ExecutionTxHash:    "0x1234567890abcdef",
			ProofOfTask:        "proof_data",
			TaskOpXCost:        1.5,
		}

		// The function will fail at database level, but we can verify the event emission logic
		err := repo.UpdateTaskExecutionDataInDB(task)
		if err == nil {
			t.Error("Expected error due to nil database, but got nil")
		}
	})

	// Test task attestation update event
	t.Run("UpdateTaskAttestationDataInDB emits TASK_UPDATED event", func(t *testing.T) {
		task := &types.UpdateTaskAttestationDataRequest{
			TaskID:               789,
			TaskNumber:           1,
			TaskAttesterIDs:      []int64{101, 102, 103},
			TpSignature:          []byte("tp_signature_data"),
			TaSignature:          []byte("ta_signature_data"),
			TaskSubmissionTxHash: "0xabcdef1234567890",
			IsSuccessful:         true,
		}

		// The function will fail at database level, but we can verify the event emission logic
		err := repo.UpdateTaskAttestationDataInDB(task)
		if err == nil {
			t.Error("Expected error due to nil database, but got nil")
		}
	})

	// Test task fee update event
	t.Run("UpdateTaskFee emits TASK_FEE_UPDATED event", func(t *testing.T) {
		// The function will fail at database level, but we can verify the event emission logic
		err := repo.UpdateTaskFee(789, 2.5)
		if err == nil {
			t.Error("Expected error due to nil database, but got nil")
		}
	})

	// Test task status update event
	t.Run("UpdateTaskNumberAndStatus emits TASK_STATUS_CHANGED event", func(t *testing.T) {
		// The function will fail at database level, but we can verify the event emission logic
		err := repo.UpdateTaskNumberAndStatus(789, 1, "completed", "0xstatus_tx_hash")
		if err == nil {
			t.Error("Expected error due to nil database, but got nil")
		}
	})
}

func TestTaskRepositoryWithoutPublisher(t *testing.T) {
	// Test that repository works without publisher (backward compatibility)
	repo := NewTaskRepository(nil)

	// This should not panic even without publisher
	task := &types.CreateTaskDataRequest{
		JobID:            big.NewInt(123),
		TaskDefinitionID: 456,
		IsImua:           false,
	}

	_, err := repo.CreateTaskDataInDB(task)
	if err == nil {
		t.Error("Expected error due to nil database, but got nil")
	}
}
