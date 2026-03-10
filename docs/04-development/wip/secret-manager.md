# Secret Manager Implementation Guide

## Overview

This guide explains how to implement a secret manager system that allows users to securely store and use secrets (API keys, private keys, etc.) in keeper-executed tasks. Secrets are encrypted at rest, only accessible to the secret owner, and injected as environment variables into Docker containers during task execution.

## Architecture Overview

The secret manager consists of several components:

1. **Database Schema**: Tables for storing encrypted secrets and job-secret associations
2. **Encryption Service**: AES-256-GCM encryption for secrets at rest
3. **Secret Manager Service**: Business logic for CRUD operations on secrets
4. **Docker Executor Integration**: Environment variable injection during container execution
5. **API Endpoints**: REST API for secret management

## 1. Database Schema

### User Secrets Table

Add to `scripts/database/init-db.cql`:

```cql
-- User secrets table (encrypted at rest)
CREATE TABLE IF NOT EXISTS triggerx.user_secrets (
    secret_id uuid,
    user_id bigint,
    secret_name text,
    secret_type text,           -- 'api_key', 'private_key', 'password', 'token', 'other'
    encrypted_value blob,       -- AES-256 encrypted secret value
    key_version int,            -- For key rotation support
    is_active boolean,
    description text,
    created_at timestamp,
    updated_at timestamp,
    last_used_at timestamp,
    PRIMARY KEY (user_id, secret_id)
);

-- Index for quick lookup by secret_name
CREATE INDEX IF NOT EXISTS user_secrets_name_idx ON triggerx.user_secrets (secret_name);
```

### Job-Secret Associations Table

```cql
-- Job-secret associations (many-to-many)
CREATE TABLE IF NOT EXISTS triggerx.job_secrets (
    job_id varint,
    secret_id uuid,
    env_var_name text,          -- Name of environment variable (e.g., "OPENAI_API_KEY")
    created_at timestamp,
    PRIMARY KEY (job_id, secret_id)
);
```

## 2. Encryption Service

Create a new package for encryption/decryption:

### Package Structure

```bash
pkg/secretmanager/
├── encryption.go    # AES-256-GCM encryption service
├── manager.go       # Secret manager business logic
└── types.go         # Type definitions
```

### Encryption Service Implementation

```go
// pkg/secretmanager/encryption.go

package secretmanager

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "io"
)

type EncryptionService struct {
    masterKey []byte // Load from environment or key management service
}

func NewEncryptionService(masterKey string) (*EncryptionService, error) {
    hash := sha256.Sum256([]byte(masterKey))
    return &EncryptionService{masterKey: hash[:]}, nil
}

func (e *EncryptionService) Encrypt(plaintext string) ([]byte, error) {
    block, err := aes.NewCipher(e.masterKey)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return ciphertext, nil
}

func (e *EncryptionService) Decrypt(ciphertext []byte) (string, error) {
    block, err := aes.NewCipher(e.masterKey)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return "", errors.New("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
```

## 3. Secret Manager Service

### Type Definitions

```go
// pkg/secretmanager/types.go

package secretmanager

import (
    "time"
    "github.com/google/uuid"
)

type Secret struct {
    SecretID      uuid.UUID
    UserID        int64
    SecretName    string
    SecretType    string // 'api_key', 'private_key', 'password', 'token', 'other'
    EncryptedValue []byte
    KeyVersion    int
    IsActive      bool
    Description   string
    CreatedAt     time.Time
    UpdatedAt     time.Time
    LastUsedAt    time.Time
}

type SecretManager struct {
    db         *database.Connection
    encryption *EncryptionService
    logger     observability.Logger
}

func NewSecretManager(
    db *database.Connection, 
    encryption *EncryptionService, 
    logger observability.Logger,
) *SecretManager {
    return &SecretManager{
        db:         db,
        encryption: encryption,
        logger:     logger,
    }
}
```

### Core Operations

```go
// pkg/secretmanager/manager.go

package secretmanager

import (
    "context"
    "math/big"
    "time"
    "github.com/google/uuid"
    "github.com/trigg3rX/triggerx-backend/pkg/database"
    "github.com/trigg3rX/triggerx-backend/pkg/observability"
)

// CreateSecret encrypts and stores a secret
func (sm *SecretManager) CreateSecret(
    ctx context.Context,
    userID int64,
    name, secretType, value, description string,
) (*Secret, error) {
    encryptedValue, err := sm.encryption.Encrypt(value)
    if err != nil {
        return nil, err
    }

    secretID := uuid.New()
    now := time.Now()
    secret := &Secret{
        SecretID:      secretID,
        UserID:        userID,
        SecretName:    name,
        SecretType:    secretType,
        EncryptedValue: encryptedValue,
        KeyVersion:    1,
        IsActive:      true,
        Description:   description,
        CreatedAt:     now,
        UpdatedAt:     now,
    }

    query := `INSERT INTO triggerx.user_secrets 
              (secret_id, user_id, secret_name, secret_type, encrypted_value, key_version, is_active, description, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    
    err = sm.db.Session().Query(query,
        secret.SecretID,
        secret.UserID,
        secret.SecretName,
        secret.SecretType,
        secret.EncryptedValue,
        secret.KeyVersion,
        secret.IsActive,
        secret.Description,
        secret.CreatedAt,
        secret.UpdatedAt,
    ).Exec()

    if err != nil {
        sm.logger.Error(ctx, "Failed to create secret", 
            observability.String("user_id", string(rune(userID))),
            observability.String("secret_name", name),
            observability.Error(err))
        return nil, err
    }

    sm.logger.Info(ctx, "Created secret", 
        observability.String("secret_id", secretID.String()),
        observability.String("user_id", string(rune(userID))))

    return secret, nil
}

// GetSecretsByUserID retrieves all secrets for a user (without decrypting values)
func (sm *SecretManager) GetSecretsByUserID(ctx context.Context, userID int64) ([]*Secret, error) {
    query := `SELECT secret_id, user_id, secret_name, secret_type, encrypted_value, key_version, 
              is_active, description, created_at, updated_at, last_used_at
              FROM triggerx.user_secrets 
              WHERE user_id = ?`
    
    iter := sm.db.Session().Query(query, userID).Iter()
    var secrets []*Secret
    
    var secret Secret
    for iter.Scan(
        &secret.SecretID,
        &secret.UserID,
        &secret.SecretName,
        &secret.SecretType,
        &secret.EncryptedValue,
        &secret.KeyVersion,
        &secret.IsActive,
        &secret.Description,
        &secret.CreatedAt,
        &secret.UpdatedAt,
        &secret.LastUsedAt,
    ) {
        secretCopy := secret
        secrets = append(secrets, &secretCopy)
    }
    
    return secrets, iter.Close()
}

// GetSecretsByJobID retrieves and decrypts secrets for a job
func (sm *SecretManager) GetSecretsByJobID(ctx context.Context, jobID *big.Int) (map[string]string, error) {
    // Get secret associations
    query := `SELECT secret_id, env_var_name FROM triggerx.job_secrets WHERE job_id = ?`
    iter := sm.db.Session().Query(query, jobID).Iter()
    
    secrets := make(map[string]string)
    var secretID uuid.UUID
    var envVarName string
    
    for iter.Scan(&secretID, &envVarName) {
        // Get encrypted secret
        secret, err := sm.getSecretByID(ctx, secretID)
        if err != nil || !secret.IsActive {
            sm.logger.Warn(ctx, "Secret not found or inactive", 
                observability.String("secret_id", secretID.String()))
            continue
        }
        
        // Decrypt
        plaintext, err := sm.encryption.Decrypt(secret.EncryptedValue)
        if err != nil {
            sm.logger.Error(ctx, "Failed to decrypt secret", 
                observability.String("secret_id", secretID.String()),
                observability.Error(err))
            continue
        }
        
        secrets[envVarName] = plaintext
        
        // Update last_used_at
        sm.updateLastUsed(ctx, secretID)
    }
    
    return secrets, iter.Close()
}

// getSecretByID retrieves a secret by ID
func (sm *SecretManager) getSecretByID(ctx context.Context, secretID uuid.UUID) (*Secret, error) {
    query := `SELECT secret_id, user_id, secret_name, secret_type, encrypted_value, key_version,
              is_active, description, created_at, updated_at, last_used_at
              FROM triggerx.user_secrets 
              WHERE secret_id = ? ALLOW FILTERING`
    
    var secret Secret
    err := sm.db.Session().Query(query, secretID).Scan(
        &secret.SecretID,
        &secret.UserID,
        &secret.SecretName,
        &secret.SecretType,
        &secret.EncryptedValue,
        &secret.KeyVersion,
        &secret.IsActive,
        &secret.Description,
        &secret.CreatedAt,
        &secret.UpdatedAt,
        &secret.LastUsedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &secret, nil
}

// AssociateSecretWithJob links a secret to a job with an environment variable name
func (sm *SecretManager) AssociateSecretWithJob(
    ctx context.Context,
    jobID *big.Int,
    secretID uuid.UUID,
    envVarName string,
) error {
    // Verify secret belongs to the job owner
    secret, err := sm.getSecretByID(ctx, secretID)
    if err != nil {
        return err
    }
    
    // Get job owner
    var jobUserID int64
    jobQuery := `SELECT user_id FROM triggerx.job_data WHERE job_id = ?`
    err = sm.db.Session().Query(jobQuery, jobID).Scan(&jobUserID)
    if err != nil {
        return err
    }
    
    // Verify ownership
    if secret.UserID != jobUserID {
        return errors.New("secret does not belong to job owner")
    }
    
    query := `INSERT INTO triggerx.job_secrets (job_id, secret_id, env_var_name, created_at)
              VALUES (?, ?, ?, ?)`
    return sm.db.Session().Query(query, jobID, secretID, envVarName, time.Now()).Exec()
}

// UpdateLastUsed updates the last_used_at timestamp
func (sm *SecretManager) updateLastUsed(ctx context.Context, secretID uuid.UUID) {
    query := `UPDATE triggerx.user_secrets SET last_used_at = ? WHERE secret_id = ?`
    go func() {
        if err := sm.db.Session().Query(query, time.Now(), secretID).Exec(); err != nil {
            sm.logger.Warn(context.Background(), "Failed to update last_used_at", 
                observability.String("secret_id", secretID.String()),
                observability.Error(err))
        }
    }()
}

// DeleteSecret soft-deletes a secret
func (sm *SecretManager) DeleteSecret(ctx context.Context, userID int64, secretID uuid.UUID) error {
    query := `UPDATE triggerx.user_secrets SET is_active = false, updated_at = ? 
              WHERE user_id = ? AND secret_id = ?`
    return sm.db.Session().Query(query, time.Now(), userID, secretID).Exec()
}
```

## 4. Integration with Docker Executor

### Modify Keeper Execution Flow

Update `internal/keeper/core/execution/action.go`:

```go
// In ExecutePerformerAction method, after getting job_id:

// Get user secrets for this job
userSecrets, err := e.getUserSecrets(ctx, targetData.JobID)
if err != nil {
    e.logger.Warn(ctx, "Failed to retrieve secrets for job", 
        observability.String("job_id", targetData.JobID.String()), 
        observability.Error(err))
    userSecrets = make(map[string]string)
}

result, execErr = e.validator.GetDockerExecutor().Execute(
    context.Background(), 
    targetData.DynamicArgumentsScriptUrl, 
    "go", 
    1, 
    config.GetAlchemyAPIKey(), 
    metadata,
    userSecrets, // Add this parameter
)

// Add helper function to TaskExecutor:
func (e *TaskExecutor) getUserSecrets(ctx context.Context, jobID *big.Int) (map[string]string, error) {
    return e.secretManager.GetSecretsByJobID(ctx, jobID)
}
```

### Update Docker Executor Interface

Modify `pkg/dockerexecutor/dockerexecutor.go`:

```go
func (de *DockerExecutor) Execute(
    ctx context.Context, 
    fileURL string, 
    fileLanguage string, 
    noOfAttesters int, 
    alchemyAPIKey string, 
    metadata map[string]string,
    userSecrets map[string]string, // Add this parameter
) (*types.ExecutionResult, error) {
    // ... existing validation code ...
    
    // Pass userSecrets to executor
    result, err := de.executor.Execute(ctx, fileURL, fileLanguage, noOfAttesters, alchemyAPIKey, metadataMap, userSecrets)
    // ... rest of code ...
}
```

### Update Container Pool

Modify `pkg/dockerexecutor/container/pool.go` in the `createContainer` method:

```go
func (p *containerPool) createContainer(ctx context.Context, codePath string, userSecrets map[string]string) (string, error) {
    // ... existing code ...
    
    // Merge environment variables from DockerConfig and LanguageConfig
    envVars := make([]string, 0, len(p.config.DockerConfig.Environment)+len(p.config.LanguageConfig.Environment))
    envVars = append(envVars, p.config.DockerConfig.Environment...)
    envVars = append(envVars, p.config.LanguageConfig.Environment...)
    
    // Add user secrets as environment variables
    for envName, envValue := range userSecrets {
        envVars = append(envVars, fmt.Sprintf("%s=%s", envName, envValue))
    }
    
    config := &container.Config{
        Image:      p.config.LanguageConfig.ImageName,
        Cmd:        []string{"sh", "-c", keepAliveCommand},
        Tty:        true,
        WorkingDir: "/code",
        Env:        envVars, // Now includes user secrets
    }
    
    // ... rest of container creation code ...
}
```

### Update Execution Pipeline

Modify `pkg/dockerexecutor/execution/pipeline.go` to pass userSecrets through:

```go
func (ep *executionPipeline) execute(
    ctx context.Context,
    fileURL string,
    fileLanguage string,
    noOfAttesters int,
    alchemyAPIKey string,
    metadata map[string]string,
    userSecrets map[string]string, // Add parameter
) (*types.ExecutionResult, error) {
    // Pass userSecrets to container manager
    // ... implementation ...
}
```

## 5. API Endpoints

### Handler Types

Create `internal/dbserver/types/secrets_types.go`:

```go
package types

type CreateSecretRequest struct {
    Name        string `json:"name" binding:"required"`
    SecretType  string `json:"secret_type" binding:"required"` // 'api_key', 'private_key', etc.
    Value       string `json:"value" binding:"required"`
    Description string `json:"description"`
}

type AssociateSecretRequest struct {
    SecretID    string `json:"secret_id" binding:"required"`
    EnvVarName  string `json:"env_var_name" binding:"required"`
}

type SecretResponse struct {
    SecretID    string    `json:"secret_id"`
    Name        string    `json:"name"`
    SecretType  string    `json:"secret_type"`
    IsActive    bool      `json:"is_active"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    LastUsedAt  *time.Time `json:"last_used_at"`
}
```

### Handler Implementation

Create `internal/dbserver/handlers/secrets.go`:

```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/trigg3rX/triggerx-backend/pkg/observability"
    commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// CreateSecret creates a new secret for the authenticated user
func (h *Handler) CreateSecret(c *gin.Context) {
    var req types.CreateSecretRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error(c.Request.Context(), "[CreateSecret] Error decoding request body", observability.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    // Get user_id from API key context
    apiKey := c.MustGet("apiKey").(*commonTypes.ApiKey)
    userID, err := h.userRepository.GetUserIDByAddress(apiKey.Owner)
    if err != nil {
        h.logger.Error(c.Request.Context(), "[CreateSecret] Failed to get user ID", observability.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
        return
    }

    secret, err := h.secretManager.CreateSecret(
        c.Request.Context(),
        userID,
        req.Name,
        req.SecretType,
        req.Value,
        req.Description,
    )
    if err != nil {
        h.logger.Error(c.Request.Context(), "[CreateSecret] Failed to create secret", observability.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create secret"})
        return
    }

    response := types.SecretResponse{
        SecretID:    secret.SecretID.String(),
        Name:        secret.SecretName,
        SecretType:  secret.SecretType,
        IsActive:    secret.IsActive,
        Description: secret.Description,
        CreatedAt:   secret.CreatedAt,
        UpdatedAt:   secret.UpdatedAt,
        LastUsedAt:  &secret.LastUsedAt,
    }

    c.JSON(http.StatusCreated, response)
    h.logger.Info(c.Request.Context(), "[CreateSecret] Created secret", 
        observability.String("secret_id", secret.SecretID.String()),
        observability.String("user_id", string(rune(userID))))
}

// ListSecrets lists all secrets for the authenticated user
func (h *Handler) ListSecrets(c *gin.Context) {
    apiKey := c.MustGet("apiKey").(*commonTypes.ApiKey)
    userID, err := h.userRepository.GetUserIDByAddress(apiKey.Owner)
    if err != nil {
        h.logger.Error(c.Request.Context(), "[ListSecrets] Failed to get user ID", observability.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
        return
    }

    secrets, err := h.secretManager.GetSecretsByUserID(c.Request.Context(), userID)
    if err != nil {
        h.logger.Error(c.Request.Context(), "[ListSecrets] Failed to list secrets", observability.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list secrets"})
        return
    }

    responses := make([]types.SecretResponse, len(secrets))
    for i, secret := range secrets {
        responses[i] = types.SecretResponse{
            SecretID:    secret.SecretID.String(),
            Name:        secret.SecretName,
            SecretType:  secret.SecretType,
            IsActive:    secret.IsActive,
            Description: secret.Description,
            CreatedAt:   secret.CreatedAt,
            UpdatedAt:   secret.UpdatedAt,
            LastUsedAt:  &secret.LastUsedAt,
        }
    }

    c.JSON(http.StatusOK, responses)
}

// AssociateSecretWithJob associates a secret with a job
func (h *Handler) AssociateSecretWithJob(c *gin.Context) {
    jobIDParam := c.Param("job_id")
    jobID := new(big.Int)
    _, ok := jobID.SetString(jobIDParam, 10)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job_id"})
        return
    }

    var req types.AssociateSecretRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error(c.Request.Context(), "[AssociateSecretWithJob] Error decoding request body", observability.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    secretID, err := uuid.Parse(req.SecretID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret_id"})
        return
    }

    err = h.secretManager.AssociateSecretWithJob(
        c.Request.Context(),
        jobID,
        secretID,
        req.EnvVarName,
    )
    if err != nil {
        h.logger.Error(c.Request.Context(), "[AssociateSecretWithJob] Failed to associate secret", observability.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate secret"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Secret associated with job"})
}

// DeleteSecret soft-deletes a secret
func (h *Handler) DeleteSecret(c *gin.Context) {
    secretIDParam := c.Param("secret_id")
    secretID, err := uuid.Parse(secretIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret_id"})
        return
    }

    apiKey := c.MustGet("apiKey").(*commonTypes.ApiKey)
    userID, err := h.userRepository.GetUserIDByAddress(apiKey.Owner)
    if err != nil {
        h.logger.Error(c.Request.Context(), "[DeleteSecret] Failed to get user ID", observability.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID"})
        return
    }

    err = h.secretManager.DeleteSecret(c.Request.Context(), userID, secretID)
    if err != nil {
        h.logger.Error(c.Request.Context(), "[DeleteSecret] Failed to delete secret", observability.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete secret"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Secret deleted"})
}
```

### Add Routes

Update `internal/dbserver/router/router.go`:

```go
// Secret management routes
secrets := api.Group("/secrets")
secrets.Use(middleware.ApiKeyAuth.GinMiddleware())
{
    secrets.POST("", handler.CreateSecret)
    secrets.GET("", handler.ListSecrets)
    secrets.DELETE("/:secret_id", handler.DeleteSecret)
}

// Job-secret associations
jobs := api.Group("/jobs")
jobs.Use(middleware.ApiKeyAuth.GinMiddleware())
{
    jobs.POST("/:job_id/secrets", handler.AssociateSecretWithJob)
}
```

## 6. Initialization

### Initialize Secret Manager in DBServer

Update `cmd/dbserver/main.go` or wherever services are initialized:

```go
import (
    "github.com/trigg3rX/triggerx-backend/pkg/secretmanager"
)

// In initialization:
encryptionService, err := secretmanager.NewEncryptionService(os.Getenv("SECRET_MANAGER_MASTER_KEY"))
if err != nil {
    log.Fatal("Failed to initialize encryption service", err)
}

secretManager := secretmanager.NewSecretManager(db, encryptionService, logger)
handler.SetSecretManager(secretManager) // Add setter to Handler
```

### Initialize Secret Manager in Keeper

Update `cmd/keeper/main.go`:

```go
// Initialize secret manager for task execution
encryptionService, err := secretmanager.NewEncryptionService(os.Getenv("SECRET_MANAGER_MASTER_KEY"))
if err != nil {
    log.Fatal("Failed to initialize encryption service", err)
}

secretManager := secretmanager.NewSecretManager(dbClient, encryptionService, logger)
taskExecutor.SetSecretManager(secretManager) // Add setter to TaskExecutor
```

## 7. Environment Configuration

Add to environment configuration:

```bash
# Secret Manager Master Key (256-bit key, base64 encoded)
SECRET_MANAGER_MASTER_KEY=<base64-encoded-256-bit-key>
```

Generate a master key:

```bash
# Generate a secure 256-bit key
openssl rand -base64 32
```

## 8. Security Considerations

### 1. **Encryption at Rest**

- All secrets are encrypted using AES-256-GCM before storage
- Master key should be stored securely (environment variable or key management service)

### 2. **Access Control**

- Users can only access their own secrets
- Job-secret associations verify ownership before linking

### 3. **Audit Logging**

- Log all secret access and usage
- Track `last_used_at` timestamp for monitoring

### 4. **Key Rotation**

- Support `key_version` field for future key rotation
- Plan for re-encryption of existing secrets during rotation

### 5. **Secret Expiration**

- Optionally implement expiration policies for unused secrets
- Consider implementing automatic cleanup of inactive secrets

### 6. **Container Isolation**

- Secrets are only injected into containers during execution
- Secrets are not persisted in container images or logs
- Ensure containers are properly cleaned up after execution

### 7. **Master Key Management**

- Consider using a key management service (AWS KMS, HashiCorp Vault) for production
- Never commit master keys to version control
- Use separate keys for different environments (dev/staging/prod)

## 9. Usage Examples

### Create a Secret

```bash
POST /api/v1/secrets
Content-Type: application/json
X-Api-Key: TGRX-...

{
  "name": "openai_api_key",
  "secret_type": "api_key",
  "value": "sk-proj-...",
  "description": "OpenAI API key for my bot"
}

Response:
{
  "secret_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "openai_api_key",
  "secret_type": "api_key",
  "is_active": true,
  "description": "OpenAI API key for my bot",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z",
  "last_used_at": null
}
```

### Associate Secret with Job

```bash
POST /api/v1/jobs/12345/secrets
Content-Type: application/json
X-Api-Key: TGRX-...

{
  "secret_id": "550e8400-e29b-41d4-a716-446655440000",
  "env_var_name": "OPENAI_API_KEY"
}

Response:
{
  "message": "Secret associated with job"
}
```

### Use in Keeper Script

```typescript
// In your keeper script (TypeScript example)
const apiKey = process.env.OPENAI_API_KEY;

if (!apiKey) {
    console.error("OPENAI_API_KEY not found");
    process.exit(1);
}

// Use the API key securely
const response = await fetch('https://api.openai.com/v1/chat/completions', {
    headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json'
    },
    // ... rest of request
});
```

```go
// In your keeper script (Go example)
package main

import (
    "fmt"
    "os"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        fmt.Fprintf(os.Stderr, "OPENAI_API_KEY not found\n")
        os.Exit(1)
    }
    
    // Use the API key securely
    // ... rest of your code
}
```

## 10. Testing

### Unit Tests

Test encryption/decryption:

```go
func TestEncryption(t *testing.T) {
    masterKey := "test-master-key-32-bytes-long!!"
    service, err := NewEncryptionService(masterKey)
    require.NoError(t, err)
    
    plaintext := "my-secret-api-key"
    encrypted, err := service.Encrypt(plaintext)
    require.NoError(t, err)
    
    decrypted, err := service.Decrypt(encrypted)
    require.NoError(t, err)
    require.Equal(t, plaintext, decrypted)
}
```

### Integration Tests

Test secret manager operations:

```go
func TestSecretManager(t *testing.T) {
    // Setup test database and encryption service
    // ...
    
    // Create secret
    secret, err := manager.CreateSecret(ctx, userID, "test_key", "api_key", "secret-value", "")
    require.NoError(t, err)
    
    // Retrieve secrets for job
    secrets, err := manager.GetSecretsByJobID(ctx, jobID)
    require.NoError(t, err)
    require.Contains(t, secrets, "TEST_API_KEY")
    require.Equal(t, "secret-value", secrets["TEST_API_KEY"])
}
```

## 11. Migration Steps

1. **Add database tables**: Run migration script to add `user_secrets` and `job_secrets` tables
2. **Deploy encryption service**: Create `pkg/secretmanager` package
3. **Deploy secret manager**: Create secret manager service
4. **Update keeper**: Integrate secret manager into keeper execution flow
5. **Update API**: Add secret management endpoints
6. **Configure master key**: Set `SECRET_MANAGER_MASTER_KEY` environment variable
7. **Test**: Verify secrets are encrypted, stored, and injected correctly

## 12. Future Enhancements

1. **Key Rotation**: Implement automatic key rotation and re-encryption
2. **Secret Versioning**: Support multiple versions of secrets
3. **Access Policies**: Fine-grained access control (time-based, IP-based)
4. **Secret Sharing**: Allow sharing secrets between users/jobs (with permissions)
5. **Audit Dashboard**: UI for viewing secret access logs
6. **Automatic Cleanup**: Clean up unused or expired secrets
7. **Secret Templates**: Pre-defined secret types with validation rules
8. **Integration with External Key Managers**: Support for AWS KMS, Vault, etc.
