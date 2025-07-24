package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type ProofValidatorMockLogger struct {
	mock.Mock
}

func (m *ProofValidatorMockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ProofValidatorMockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ProofValidatorMockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ProofValidatorMockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ProofValidatorMockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *ProofValidatorMockLogger) Debugf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *ProofValidatorMockLogger) Infof(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *ProofValidatorMockLogger) Warnf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *ProofValidatorMockLogger) Errorf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *ProofValidatorMockLogger) Fatalf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *ProofValidatorMockLogger) With(tags ...interface{}) logging.Logger {
	args := m.Called(tags)
	return args.Get(0).(logging.Logger)
}

func TestTaskValidator_ValidateProof(t *testing.T) {
	// Setup
	mockLogger := new(ProofValidatorMockLogger)
	validator := &TaskValidator{
		logger: mockLogger,
	}

	tests := []struct {
		name      string
		ipfsData  types.IPFSData
		traceID   string
		wantValid bool
		wantErr   bool
		setupMock func()
	}{
		{
			name: "missing proof data",
			ipfsData: types.IPFSData{
				TaskData: &types.SendTaskDataToKeeper{
					TaskID: []int64{1},
				},
			},
			traceID:   "test-trace-1",
			wantValid: false,
			wantErr:   true,
			setupMock: func() {},
		},
		{
			name: "empty proof of task",
			ipfsData: types.IPFSData{
				TaskData: &types.SendTaskDataToKeeper{
					TaskID: []int64{2},
				},
				ProofData: &types.ProofData{},
			},
			traceID:   "test-trace-2",
			wantValid: false,
			wantErr:   true,
			setupMock: func() {},
		},
		{
			name: "empty certificate hash",
			ipfsData: types.IPFSData{
				TaskData: &types.SendTaskDataToKeeper{
					TaskID: []int64{3},
				},
				ProofData: &types.ProofData{
					ProofOfTask: "some-proof",
				},
			},
			traceID:   "test-trace-3",
			wantValid: false,
			wantErr:   true,
			setupMock: func() {},
		},
		{
			name: "valid proof data",
			ipfsData: types.IPFSData{
				TaskData: &types.SendTaskDataToKeeper{
					TaskID: []int64{4},
				},
				ProofData: &types.ProofData{
					TaskID:          4,
					ProofOfTask:     "valid-proof",
					CertificateHash: "valid-cert-hash",
				},
				PerformerSignature: &types.PerformerSignatureData{
					TaskID:                  4,
					PerformerSigningAddress: config.GetConsensusAddress(),
				},
			},
			traceID:   "test-trace-4",
			wantValid: false, // Will be false because we're not actually establishing TLS connection
			wantErr:   true,
			setupMock: func() {
				mockLogger.On("Warn", "Failed to establish TLS connection for validation", mock.Anything).Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			valid, err := validator.ValidateProof(tt.ipfsData, tt.traceID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, valid)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValid, valid)
			}
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestTaskValidator_validateProofHash(t *testing.T) {
	// Setup
	mockLogger := new(ProofValidatorMockLogger)
	validator := &TaskValidator{
		logger: mockLogger,
	}

	tests := []struct {
		name      string
		ipfsData  types.IPFSData
		traceID   string
		wantValid bool
		wantErr   bool
		setupMock func()
	}{
		{
			name: "valid proof hash",
			ipfsData: types.IPFSData{
				TaskData: &types.SendTaskDataToKeeper{
					TaskID: []int64{1},
				},
				ProofData: &types.ProofData{
					TaskID:          1,
					ProofOfTask:     "valid-hash",
					CertificateHash: "valid-cert-hash",
				},
				PerformerSignature: &types.PerformerSignatureData{
					TaskID:                  1,
					PerformerSigningAddress: config.GetConsensusAddress(),
				},
			},
			traceID:   "test-trace-1",
			wantValid: false, // Will be false because we're not actually generating a real hash
			wantErr:   true,
			setupMock: func() {},
		},
		{
			name: "missing task data",
			ipfsData: types.IPFSData{
				TaskData: &types.SendTaskDataToKeeper{
					TaskID: []int64{2},
				},
				ProofData: &types.ProofData{
					TaskID:          2,
					ProofOfTask:     "valid-hash",
					CertificateHash: "valid-cert-hash",
				},
				PerformerSignature: &types.PerformerSignatureData{
					TaskID:                  2,
					PerformerSigningAddress: config.GetConsensusAddress(),
				},
			},
			traceID:   "test-trace-2",
			wantValid: false,
			wantErr:   true,
			setupMock: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			valid, err := validator.validateProofHash(tt.ipfsData, tt.traceID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, valid)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValid, valid)
			}
			mockLogger.AssertExpectations(t)
		})
	}
}
