# Logging Package Mocks

This document explains how to use the mock system provided by the logging package for testing components that depend on logging.

## Overview

The logging package provides a comprehensive mock system that allows other packages to test their components without depending on actual logging infrastructure. This includes:

- `MockLogger` - A mock implementation of the Logger interface
- `NoOpLogger` - A logger that does nothing (useful for tests that don't care about logging)
- `MockLoggerBuilder` - A fluent builder for creating mock loggers with expectations
- `MockSequentialRotator` - A mock implementation of the log rotator
- `MockLoggerFactory` - A factory for creating mock loggers

## Basic Usage

### Simple Mock Logger

For basic testing where you just need a logger that doesn't do anything:

```go
import "github.com/trigg3rX/triggerx-backend/pkg/logging"

func TestMyComponent(t *testing.T) {
    // Create a no-op logger
    logger := logging.NewNoOpLogger()

    // Use it in your component
    component := NewMyComponent(logger)

    // Test your component logic
    result := component.DoSomething()
    assert.Equal(t, expected, result)
}
```

### Mock Logger with Expectations

For tests where you want to verify that specific log messages are written:

```go
import "github.com/trigg3rX/triggerx-backend/pkg/logging"

func TestMyComponent_LogsCorrectly(t *testing.T) {
    // Create a mock logger with expectations
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectInfo("User logged in", "user_id", "123").
        ExpectDebug("Processing request", "method", "GET").
        ExpectError("Database error", "error", "connection failed").
        Build()

    // Use it in your component
    component := NewMyComponent(mockLogger)

    // Trigger the behavior that should log
    component.HandleUserLogin("123")

    // Verify all expected log calls were made
    mockLogger.AssertExpectations(t)
}
```

### Mock Logger with Call Count Verification

For tests where you want to verify the number of times a method is called:

```go
func TestMyComponent_LogsMultipleTimes(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectInfo("Processing item", "item_id", mock.Anything).
        Build()

    component := NewMyComponent(mockLogger)

    // Process multiple items
    component.ProcessItems([]string{"item1", "item2", "item3"})

    // Verify Info was called exactly 3 times
    assert.True(t, mockLogger.AssertNumberOfCalls(t, "Info", 3))
}
```

## Advanced Usage

### Testing Different Log Levels

```go
func TestMyComponent_LogLevels(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectDebug("Debug info", "detail", "verbose").
        ExpectInfo("Normal operation", "status", "ok").
        ExpectWarn("Warning condition", "issue", "minor").
        ExpectError("Error occurred", "code", 500).
        Build()

    component := NewMyComponent(mockLogger)
    component.RunWithDifferentLogLevels()

    mockLogger.AssertExpectations(t)
}
```

### Testing Formatted Logging

```go
func TestMyComponent_FormattedLogging(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectInfof("User %s logged in from %s", "john", "192.168.1.1").
        ExpectErrorf("Failed to process %d items", 5).
        Build()

    component := NewMyComponent(mockLogger)
    component.LogFormattedMessages()

    mockLogger.AssertExpectations(t)
}
```

### Testing With Tags

```go
func TestMyComponent_WithTags(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectWith("request_id", "req-123", "user_id", "user-456").
        Build()

    component := NewMyComponent(mockLogger)

    // The component should create a tagged logger
    taggedLogger := component.CreateTaggedLogger("req-123", "user-456")
    taggedLogger.Info("Request processed")

    mockLogger.AssertExpectations(t)
}
```

### Testing Error Conditions

```go
func TestMyComponent_ErrorHandling(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectError("Failed to connect to database", "error", "timeout").
        ExpectFatal("Critical error, shutting down", "reason", "database_unavailable").
        Build()

    component := NewMyComponent(mockLogger)

    // Simulate a critical error
    component.HandleCriticalError("database_unavailable")

    mockLogger.AssertExpectations(t)
}
```

## Integration Examples

### Testing an HTTP Handler

```go
func TestUserHandler_Login(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectInfo("Login attempt", "username", "john", "ip", "192.168.1.1").
        ExpectInfo("Login successful", "username", "john", "session_id", mock.Anything).
        Build()

    handler := NewUserHandler(mockLogger)

    req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"john","password":"secret"}`))
    req.Header.Set("X-Forwarded-For", "192.168.1.1")
    w := httptest.NewRecorder()

    handler.Login(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    mockLogger.AssertExpectations(t)
}
```

### Testing a Database Service

```go
func TestUserService_CreateUser(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectDebug("Creating new user", "username", "john").
        ExpectInfo("User created successfully", "user_id", "123", "username", "john").
        Build()

    service := NewUserService(mockLogger, mockDB)

    user, err := service.CreateUser("john", "john@example.com")

    assert.NoError(t, err)
    assert.NotNil(t, user)
    mockLogger.AssertExpectations(t)
}
```

### Testing a Background Worker

```go
func TestBackgroundWorker_ProcessJob(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectInfo("Starting job processing", "job_id", "job-123").
        ExpectDebug("Processing step 1", "step", 1, "progress", "10%").
        ExpectDebug("Processing step 2", "step", 2, "progress", "50%").
        ExpectInfo("Job completed successfully", "job_id", "job-123", "duration", mock.Anything).
        Build()

    worker := NewBackgroundWorker(mockLogger)

    err := worker.ProcessJob("job-123")

    assert.NoError(t, err)
    mockLogger.AssertExpectations(t)
}
```

## Best Practices

### 1. Use NoOpLogger for Tests That Don't Care About Logging

```go
func TestPureLogic_NoLoggingNeeded(t *testing.T) {
    logger := logging.NewNoOpLogger()
    calculator := NewCalculator(logger)

    result := calculator.Add(2, 3)
    assert.Equal(t, 5, result)
}
```

### 2. Use MockLoggerBuilder for Tests That Verify Logging Behavior

```go
func TestBusinessLogic_LoggingIsImportant(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectInfo("Transaction started", "amount", 100).
        ExpectInfo("Transaction completed", "transaction_id", mock.Anything).
        Build()

    service := NewTransactionService(mockLogger)
    err := service.ProcessTransaction(100)

    assert.NoError(t, err)
    mockLogger.AssertExpectations(t)
}
```

### 3. Use Mock.Anything for Dynamic Values

```go
func TestDynamicValues(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectInfo("Request processed", "request_id", mock.Anything, "duration", mock.Anything).
        Build()

    // The actual request_id and duration will be generated at runtime
    service := NewRequestService(mockLogger)
    service.ProcessRequest()

    mockLogger.AssertExpectations(t)
}
```

### 4. Test Error Scenarios

```go
func TestErrorScenarios(t *testing.T) {
    mockLogger := logging.NewMockLoggerBuilder().
        ExpectError("Failed to process request", "error", "database timeout").
        Build()

    service := NewRequestService(mockLogger)

    // Simulate database error
    err := service.ProcessRequestWithError()

    assert.Error(t, err)
    mockLogger.AssertExpectations(t)
}
```

### 5. Use Table-Driven Tests with Mocks

```go
func TestMultipleScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        logs     func(*logging.MockLoggerBuilder) *logging.MockLoggerBuilder
    }{
        {
            name:     "success case",
            input:    "valid input",
            expected: "success",
            logs: func(b *logging.MockLoggerBuilder) *logging.MockLoggerBuilder {
                return b.ExpectInfo("Processing successful", "input", "valid input")
            },
        },
        {
            name:     "error case",
            input:    "invalid input",
            expected: "error",
            logs: func(b *logging.MockLoggerBuilder) *logging.MockLoggerBuilder {
                return b.ExpectError("Processing failed", "input", "invalid input", "error", "validation failed")
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockLogger := tt.logs(logging.NewMockLoggerBuilder()).Build()

            service := NewProcessingService(mockLogger)
            result := service.Process(tt.input)

            assert.Equal(t, tt.expected, result)
            mockLogger.AssertExpectations(t)
        })
    }
}
```

## Mock Sequential Rotator

For testing components that depend on log rotation:

```go
func TestLogRotation(t *testing.T) {
    mockRotator := &logging.MockSequentialRotator{}

    mockRotator.On("Write", []byte("log entry 1\n")).Return(12, nil)
    mockRotator.On("Write", []byte("log entry 2\n")).Return(12, nil)
    mockRotator.On("Close").Return(nil)

    logger := NewCustomLogger(mockRotator)

    logger.Info("log entry 1")
    logger.Info("log entry 2")
    logger.Close()

    mockRotator.AssertExpectations(t)
}
```

## Conclusion

The mock system provided by the logging package makes it easy to test components that depend on logging without the complexity of setting up actual logging infrastructure. Use `NoOpLogger` when logging is not important for the test, and use `MockLoggerBuilder` when you need to verify specific logging behavior.

Remember to always call `AssertExpectations(t)` at the end of your tests to ensure all expected log calls were made.
