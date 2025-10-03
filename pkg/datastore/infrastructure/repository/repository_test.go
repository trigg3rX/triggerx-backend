package repository

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TestGenericRepository tests the generic repository functionality
func TestGenericRepository(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.MockLogger{}
	tableName := "test_table"
	primaryKey := "id"

	// Set up expectations for repository creation
	mockGocqlxSessioner := mocks.NewMockGocqlxSessioner(ctrl)
	mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSessioner).AnyTimes()

	// Create repository
	repo := NewGenericRepository[types.UserDataEntity](
		mockConnection,
		mockLogger,
		tableName,
		primaryKey,
	)

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		user := &types.UserDataEntity{
			UserID:      1,
			UserAddress: "0x123",
			EmailID:     "test@example.com",
			CreatedAt:   time.Now(),
		}

		// Test successful creation
		mockGocqlxQueryer := mocks.NewMockGocqlxQueryer(ctrl)
		mockGocqlxSessioner.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().WithContext(gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().BindStruct(gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().ExecRelease().Return(nil)

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		user := &types.UserDataEntity{
			UserID:      1,
			UserAddress: "0x456",
			EmailID:     "updated@example.com",
			CreatedAt:   time.Now(),
		}

		// Test successful update
		mockGocqlxQueryer := mocks.NewMockGocqlxQueryer(ctrl)
		mockGocqlxSessioner.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().WithContext(gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().BindStruct(gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().ExecRelease().Return(nil)

		err := repo.Update(ctx, user)
		assert.NoError(t, err)
	})

	// Verify all expectations were met
}

// TestRepositoryWithDifferentTypes tests repository with different entity types
func TestRepositoryWithDifferentTypes(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnection := mocks.NewMockConnection(ctrl)
	mockLogger := &logging.MockLogger{}

	// Set up expectations for repository creation
	mockGocqlxSessioner := mocks.NewMockGocqlxSessioner(ctrl)
	mockConnection.EXPECT().GetGocqlxSession().Return(mockGocqlxSessioner).AnyTimes()

	t.Run("JobDataEntity", func(t *testing.T) {
		repo := NewGenericRepository[types.JobDataEntity](
			mockConnection,
			mockLogger,
			"job_data",
			"job_id",
		)

		ctx := context.Background()
		job := &types.JobDataEntity{
			JobID:            *big.NewInt(1),
			JobTitle:         "Test Job",
			TaskDefinitionID: 1,
			CreatedChainID:   "chain1",
			UserID:           1,
			CreatedAt:        time.Now(),
		}

		// Test successful creation
		mockGocqlxQueryer := mocks.NewMockGocqlxQueryer(ctrl)
		mockGocqlxSessioner.EXPECT().Query(gomock.Any(), gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().WithContext(gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().BindStruct(gomock.Any()).Return(mockGocqlxQueryer)
		mockGocqlxQueryer.EXPECT().ExecRelease().Return(nil)

		err := repo.Create(ctx, job)
		assert.NoError(t, err)
	})

	// Verify all expectations were met
}
