package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	datastoreMocks "github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	pkgErrors "github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func setupKeeperTestHandler() (*Handler, *datastoreMocks.MockGenericRepository[types.KeeperDataEntity], *logging.MockLogger) {
	mockKeeperRepo := new(datastoreMocks.MockGenericRepository[types.KeeperDataEntity])
	mockLogger := new(logging.MockLogger)
	mockLogger.SetupDefaultExpectations()

	handler := &Handler{
		logger:           mockLogger,
		keeperRepository: mockKeeperRepo,
	}

	return handler, mockKeeperRepo, mockLogger
}

func TestCreateKeeperDataFromGoogleForm(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*datastoreMocks.MockGenericRepository[types.KeeperDataEntity])
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Create New Keeper",
			requestBody: types.CreateKeeperDataFromGoogleFormRequest{
				KeeperAddress:  "0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
				RewardsAddress: "0x1234567890ABCDEF1234567890ABCDEF12345678",
				KeeperName:     "Test Keeper",
				EmailID:        "keeper@example.com",
				OnImua:         true,
			},
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.KeeperDataEntity]) {
				// Check if keeper exists - return nil (doesn't exist)
				mockRepo.On("GetByID", mock.Anything, "0xabcdef1234567890abcdef1234567890abcdef12").Return(nil, nil).Once()

				// Create keeper
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(keeper *types.KeeperDataEntity) bool {
					return keeper.KeeperAddress == "0xabcdef1234567890abcdef1234567890abcdef12" &&
						keeper.RewardsAddress == "0x1234567890abcdef1234567890abcdef12345678" &&
						keeper.KeeperName == "Test Keeper" &&
						keeper.EmailID == "keeper@example.com" &&
						keeper.RewardsBooster == "1" &&
						keeper.KeeperPoints == "0"
				})).Return(nil).Once()
			},
			expectedCode: http.StatusCreated,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				// Address should be lowercase in response
				assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", response["keeper_address"])
			},
		},
		{
			name: "Success - Create Keeper with Lowercase Address",
			requestBody: types.CreateKeeperDataFromGoogleFormRequest{
				KeeperAddress:  "0xabcdef1234567890abcdef1234567890abcdef12",
				RewardsAddress: "0x1234567890abcdef1234567890abcdef12345678",
				KeeperName:     "Lowercase Keeper",
				EmailID:        "lowercase@example.com",
				OnImua:         false,
			},
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.KeeperDataEntity]) {
				mockRepo.On("GetByID", mock.Anything, "0xabcdef1234567890abcdef1234567890abcdef12").Return(nil, nil).Once()
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusCreated,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", response["keeper_address"])
			},
		},
		{
			name: "Success - Create Keeper with Mixed Case Address",
			requestBody: types.CreateKeeperDataFromGoogleFormRequest{
				KeeperAddress:  "0xAbCdEf1234567890aBcDeF1234567890AbCdEf12",
				RewardsAddress: "0x1234567890AbCdEf1234567890aBcDeF12345678",
				KeeperName:     "Mixed Case Keeper",
				EmailID:        "mixed@example.com",
				OnImua:         false,
			},
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.KeeperDataEntity]) {
				// Should search with lowercase
				mockRepo.On("GetByID", mock.Anything, "0xabcdef1234567890abcdef1234567890abcdef12").Return(nil, nil).Once()
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(keeper *types.KeeperDataEntity) bool {
					// Verify addresses are stored in lowercase
					return keeper.KeeperAddress == "0xabcdef1234567890abcdef1234567890abcdef12" &&
						keeper.RewardsAddress == "0x1234567890abcdef1234567890abcdef12345678"
				})).Return(nil).Once()
			},
			expectedCode: http.StatusCreated,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				// Should return lowercase address
				assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", response["keeper_address"])
			},
		},
		{
			name: "Success - Create Keeper with Minimal Data",
			requestBody: types.CreateKeeperDataFromGoogleFormRequest{
				KeeperAddress:  "0xMinimal1234567890Minimal1234567890Minimal12",
				RewardsAddress: "",
				KeeperName:     "Min Keeper",
				EmailID:        "min@example.com",
				OnImua:         false,
			},
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.KeeperDataEntity]) {
				mockRepo.On("GetByID", mock.Anything, "0xminimal1234567890minimal1234567890minimal12").Return(nil, nil).Once()
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(keeper *types.KeeperDataEntity) bool {
					// Verify default values
					return keeper.RewardsBooster == "1" && keeper.KeeperPoints == "0"
				})).Return(nil).Once()
			},
			expectedCode: http.StatusCreated,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["keeper_address"])
			},
		},
		{
			name:        "Error - Invalid Request Body (malformed JSON)",
			requestBody: "invalid json {malformed",
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.KeeperDataEntity]) {
				// No mock setup needed - validation fails before repository call
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: pkgErrors.ErrInvalidRequestBody,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrInvalidRequestBody, response["error"])
			},
		},
		{
			name: "Error - Database Error on GetByID",
			requestBody: types.CreateKeeperDataFromGoogleFormRequest{
				KeeperAddress:  "0xDBError1234567890DBError1234567890DBError12",
				RewardsAddress: "0x1234567890123456789012345678901234567890",
				KeeperName:     "DB Error Keeper",
				EmailID:        "dberror@example.com",
				OnImua:         false,
			},
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.KeeperDataEntity]) {
				// GetByID returns error (not nil, not found)
				mockRepo.On("GetByID", mock.Anything, "0xdberror1234567890dberror1234567890dberror12").
					Return(nil, errors.New("database connection error")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: pkgErrors.ErrDBOperationFailed,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBOperationFailed, response["error"])
			},
		},
		{
			name: "Error - Keeper Already Exists (Duplicate)",
			requestBody: types.CreateKeeperDataFromGoogleFormRequest{
				KeeperAddress:  "0xDuplicate1234567890Duplicate1234567890Dupl12",
				RewardsAddress: "0x1234567890123456789012345678901234567890",
				KeeperName:     "Duplicate Keeper",
				EmailID:        "duplicate@example.com",
				OnImua:         false,
			},
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.KeeperDataEntity]) {
				// Keeper already exists
				existingKeeper := &types.KeeperDataEntity{
					KeeperAddress:  "0xduplicate1234567890duplicate1234567890dupl12",
					RewardsAddress: "0x1234567890123456789012345678901234567890",
					KeeperName:     "Existing Keeper",
					EmailID:        "existing@example.com",
				}
				mockRepo.On("GetByID", mock.Anything, "0xduplicate1234567890duplicate1234567890dupl12").
					Return(existingKeeper, nil).Once()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: pkgErrors.ErrDBDuplicateRecord,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBDuplicateRecord, response["error"])
			},
		},
		{
			name: "Error - Database Error on Create",
			requestBody: types.CreateKeeperDataFromGoogleFormRequest{
				KeeperAddress:  "0xCreateErr1234567890CreateErr1234567890Creat",
				RewardsAddress: "0x1234567890123456789012345678901234567890",
				KeeperName:     "Create Error Keeper",
				EmailID:        "createerror@example.com",
				OnImua:         false,
			},
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.KeeperDataEntity]) {
				// GetByID succeeds (no existing keeper)
				mockRepo.On("GetByID", mock.Anything, "0xcreateerr1234567890createerr1234567890creat").Return(nil, nil).Once()

				// Create fails with database error
				mockRepo.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("database write error")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: pkgErrors.ErrDBOperationFailed,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBOperationFailed, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockKeeperRepo, mockLogger := setupKeeperTestHandler()
			defer mockLogger.AssertExpectations(t)
			defer mockKeeperRepo.AssertExpectations(t)

			// Reset mocks
			mockKeeperRepo.ExpectedCalls = nil
			mockKeeperRepo.Calls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request body
			var reqBody []byte
			var err error
			if strBody, ok := tt.requestBody.(string); ok {
				reqBody = []byte(strBody)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			c.Request = httptest.NewRequest("POST", "/keepers/google-form", bytes.NewBuffer(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// Setup mocks
			tt.mockSetup(mockKeeperRepo)

			// Execute
			handler.CreateKeeperDataFromGoogleForm(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}

			mockKeeperRepo.AssertExpectations(t)
		})
	}
}

func TestCreateKeeperDataFromGoogleForm_AddressNormalization(t *testing.T) {
	// This test specifically verifies that addresses are normalized to lowercase
	handler, mockKeeperRepo, mockLogger := setupKeeperTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockKeeperRepo.AssertExpectations(t)

	// Setup
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := types.CreateKeeperDataFromGoogleFormRequest{
		KeeperAddress:  "0xABCDEF1234567890ABCDEF1234567890ABCDEF12", // Uppercase
		RewardsAddress: "0xFEDCBA9876543210FEDCBA9876543210FEDCBA98", // Uppercase
		KeeperName:     "Test Keeper",
		EmailID:        "test@example.com",
		OnImua:         false,
	}

	// Expect lowercase addresses in all calls
	expectedKeeperAddr := "0xabcdef1234567890abcdef1234567890abcdef12"
	expectedRewardsAddr := "0xfedcba9876543210fedcba9876543210fedcba98"

	mockKeeperRepo.On("GetByID", mock.Anything, expectedKeeperAddr).Return(nil, nil).Once()
	mockKeeperRepo.On("Create", mock.Anything, mock.MatchedBy(func(keeper *types.KeeperDataEntity) bool {
		return keeper.KeeperAddress == expectedKeeperAddr &&
			keeper.RewardsAddress == expectedRewardsAddr
	})).Return(nil).Once()

	reqBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	c.Request = httptest.NewRequest("POST", "/keepers/google-form", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateKeeperDataFromGoogleForm(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedKeeperAddr, response["keeper_address"])

	mockKeeperRepo.AssertExpectations(t)
}

func TestCreateKeeperDataFromGoogleForm_DefaultValues(t *testing.T) {
	// This test verifies that default values are correctly set
	handler, mockKeeperRepo, mockLogger := setupKeeperTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockKeeperRepo.AssertExpectations(t)

	// Setup
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	requestBody := types.CreateKeeperDataFromGoogleFormRequest{
		KeeperAddress:  "0x1234567890123456789012345678901234567890",
		RewardsAddress: "0x0987654321098765432109876543210987654321",
		KeeperName:     "Test Keeper",
		EmailID:        "test@example.com",
		OnImua:         true,
	}

	mockKeeperRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
	mockKeeperRepo.On("Create", mock.Anything, mock.MatchedBy(func(keeper *types.KeeperDataEntity) bool {
		// Verify default values
		if keeper.RewardsBooster != "1" {
			t.Errorf("Expected RewardsBooster to be '1', got '%s'", keeper.RewardsBooster)
			return false
		}
		if keeper.KeeperPoints != "0" {
			t.Errorf("Expected KeeperPoints to be '0', got '%s'", keeper.KeeperPoints)
			return false
		}
		// Verify other fields are correctly set
		if keeper.KeeperName != "Test Keeper" {
			return false
		}
		if keeper.EmailID != "test@example.com" {
			return false
		}
		return true
	})).Return(nil).Once()

	reqBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	c.Request = httptest.NewRequest("POST", "/keepers/google-form", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateKeeperDataFromGoogleForm(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	mockKeeperRepo.AssertExpectations(t)
}
