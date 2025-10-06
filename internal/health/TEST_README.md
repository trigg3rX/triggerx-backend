# Health Module Test Suite

This directory contains comprehensive unit tests for the TriggerX health monitoring module.

## Test Structure

### Core Test Files

- **`client/database_test.go`** - Tests for database operations including keeper health updates and verification
- **`keeper/persistence_test.go`** - Tests for keeper data persistence and state loading/dumping
- **`keeper/state_manager_test.go`** - Tests for in-memory keeper state management
- **`keeper/cleanup_test.go`** - Tests for inactive keeper cleanup routines
- **`keeper/update_test.go`** - Tests for keeper health status updates

### Mock Files

- **`mocks/database_mocks.go`** - Mock implementations for database operations
- **`mocks/telegram_mocks.go`** - Mock implementations for Telegram bot notifications

### Helper Files

- **`test_helpers.go`** - Utility functions for creating test data and common test patterns

## Running Tests

### Run All Health Module Tests
```bash
go test ./internal/health/...
```

### Run Specific Test Files
```bash
# Database tests
go test ./internal/health/client -v

# Keeper tests
go test ./internal/health/keeper -v
```

### Run with Coverage
```bash
go test -cover ./internal/health/...
```

### Run with Race Detection
```bash
go test -race ./internal/health/...
```

## Test Categories

### Unit Tests
- **Database Operations**: Testing keeper health updates, verification, and error handling
- **State Management**: Testing in-memory keeper state operations with thread safety
- **Persistence**: Testing data loading and dumping with proper error handling
- **Cleanup**: Testing inactive keeper detection and automatic cleanup
- **Updates**: Testing keeper health status updates with retry mechanisms

### Integration Tests
- **End-to-End Flows**: Testing complete keeper registration and health monitoring flows
- **Concurrency**: Testing thread-safe operations under concurrent access
- **Error Scenarios**: Testing error handling and recovery mechanisms

## Mock Usage

### Database Mocks
The `MockDatabaseManager` provides mocked database operations:
```go
mockDB := &mocks.MockDatabaseManager{}
mockDB.On("UpdateKeeperHealth", mock.Anything, true).Return(nil)
```

### Telegram Mocks
The `MockTelegramBot` provides mocked notification operations:
```go
mockBot := &mocks.MockTelegramBot{}
mockBot.On("SendMessage", chatID, message).Return(nil)
```

## Test Helpers

Use the helper functions in `test_helpers.go` to create test data:

```go
// Create a test keeper
keeper := CreateTestKeeperInfo("0x123", WithActiveKeeper(), WithImuaKeeper())

// Create a test health check-in
health := CreateTestKeeperHealthCheckIn("0x123", WithHealthVersion("2.0.0"))
```

## Best Practices

1. **Test Isolation** - Each test resets global state
2. **Mock Verification** - All expectations verified  
3. **Concurrency Testing** - Thread-safety verified
4. **Error Testing** - All error paths covered
5. **High Coverage** - Aim for >90% coverage
