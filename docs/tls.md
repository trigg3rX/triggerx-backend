# TLS Proof with Custom Certificate

## Introduction

Using a custom certificate on your own server is actually a much better approach for your use case. This gives you full control over the certificate lifecycle and ensures consistency across all keepers. Let me show you how to implement this.

## Custom Certificate Implementation for TLS Proof

### 1. Enhanced TLS Configuration

```go
// Enhanced TLSProofConfig with custom certificate support
type TLSProofConfig struct {
    TargetHost         string        // Host to establish TLS connection with
    TargetPort         string        // Port to connect to (default: "443")
    Timeout            time.Duration // Connection timeout (default: 5s)
    VerifyPeer         bool          // Whether to verify peer certificates (default: false)
    ServerName         string        // Server name for SNI (optional, defaults to TargetHost)
    CustomCAPath       string        // Path to custom CA certificate file
    CustomCACert       []byte        // Custom CA certificate data (alternative to CustomCAPath)
    PinnedCertHash     string        // Expected certificate hash for pinning
    InsecureSkipVerify bool          // Skip certificate verification (for testing only)
}
```

### 2. Configuration Setup

Add these to your keeper configuration:

```go
// In internal/keeper/config/config.go
type Config struct {
    // ... existing fields ...
    // TLS Proof configuration
    tlsProofHost       string
    tlsProofPort       string
    tlsProofCAPath     string  // Path to custom CA certificate
    tlsProofCertHash   string  // Expected certificate hash for pinning
}
```

### 3. Enhanced Proof Generation Functions

Here are the additional functions you'll need to add to `pkg/proof/proof.go`:

```go
// Update the existing functions to support custom certificates

// LoadCustomCACertificate loads a custom CA certificate from file
func LoadCustomCACertificate(certPath string) ([]byte, error) {
    certData, err := ioutil.ReadFile(certPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read certificate file: %w", err)
    }
    
    // Validate that it's a valid PEM certificate
    block, _ := pem.Decode(certData)
    if block == nil {
        return nil, errors.New("failed to parse certificate PEM")
    }
    
    if block.Type != "CERTIFICATE" {
        return nil, fmt.Errorf("invalid certificate type: %s", block.Type)
    }
    
    // Parse the certificate to ensure it's valid
    _, err = x509.ParseCertificate(block.Bytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse certificate: %w", err)
    }
    
    return certData, nil
}

// GetCertificateHash returns the SHA256 hash of a certificate
func GetCertificateHash(cert *x509.Certificate) string {
    hash := sha256.Sum256(cert.Raw)
    return hex.EncodeToString(hash[:])
}
```

### 4. Enhanced EstablishTLSConnection Function

Update the `EstablishTLSConnection` function to support custom certificates:

```go
// EstablishTLSConnection creates a real TLS connection and returns the connection state
func EstablishTLSConnection(config *TLSProofConfig) (*tls.ConnectionState, error) {
    // ... existing code ...
    
    // Setup custom CA if provided
    if len(config.CustomCACert) > 0 || config.CustomCAPath != "" {
        caCertPool := x509.NewCertPool()
        
        var caCert []byte
        var err error
        
        if len(config.CustomCACert) > 0 {
            caCert = config.CustomCACert
        } else if config.CustomCAPath != "" {
            caCert, err = ioutil.ReadFile(config.CustomCAPath)
            if err != nil {
                return nil, fmt.Errorf("failed to read custom CA certificate: %w", err)
            }
        }
        
        if !caCertPool.AppendCertsFromPEM(caCert) {
            return nil, errors.New("failed to parse custom CA certificate")
        }
        
        tlsConfig.RootCAs = caCertPool
    }
    
    // Establish connection with timeout
    dialer := &net.Dialer{
        Timeout: config.Timeout,
    }

    address := net.JoinHostPort(config.TargetHost, config.TargetPort)
    conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to establish TLS connection to %s: %w", address, err)
    }
    defer conn.Close()

    // Get connection state
    connState := conn.ConnectionState()
    
    // Verify we have certificates
    if len(connState.PeerCertificates) == 0 {
        return nil, errors.New("no peer certificates found in TLS connection")
    }

    // Perform certificate pinning if configured
    if config.PinnedCertHash != "" {
        actualCertHash := sha256.Sum256(connState.PeerCertificates[0].Raw)
        actualCertHashStr := hex.EncodeToString(actualCertHash[:])
        
        if actualCertHashStr != config.PinnedCertHash {
            return nil, fmt.Errorf("certificate pinning failed: expected %s, got %s", 
                config.PinnedCertHash, actualCertHashStr)
        }
    }
    
    return &connState, nil
}
```

### 5. Updated Executor Implementation

Update your executor to use the custom certificate:

```go
// In internal/keeper/core/execution/executor.go
func (e *TaskExecutor) ExecuteTask(ctx context.Context, task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
    // ... existing code ...
    
    // Generate TLS proof using custom certificate
    var tlsConfig *proof.TLSProofConfig
    var err error
    
    // Load custom CA certificate if configured
    caPath := config.GetTLSProofCAPath()
    if caPath != "" {
        customCA, err := proof.LoadCustomCACertificate(caPath)
        if err != nil {
            e.logger.Error("Failed to load custom CA certificate", "trace_id", traceID, "error", err)
            return false, err
        }

        // Use custom certificate with pinning if hash is configured
        certHash := config.GetTLSProofCertHash()
        if certHash != "" {
            tlsConfig = proof.CustomTLSProofConfigWithPinning(
                config.GetTLSProofHost(),
                config.GetTLSProofPort(),
                certHash,
                customCA,
            )
        } else {
            tlsConfig = proof.CustomTLSProofConfig(
                config.GetTLSProofHost(),
                config.GetTLSProofPort(),
                customCA,
            )
        }
    } else {
        // Fallback to default configuration
        tlsConfig = proof.DefaultTLSProofConfig(config.GetTLSProofHost())
        tlsConfig.TargetPort = config.GetTLSProofPort()
    }
    
    proofData, err := proof.GenerateProofWithTLSConnection(ipfsData, tlsConfig)
    if err != nil {
        e.logger.Error("Failed to generate TLS proof", "trace_id", traceID, "error", err)
        return false, err
    }

    // ... rest of the function ...
}
```

### 6. Environment Configuration

Set up your environment variables:

```bash
# .env file
TLS_PROOF_HOST=proof.yourdomain.com
TLS_PROOF_PORT=443
TLS_PROOF_CA_PATH=/path/to/your/ca-certificate.pem
TLS_PROOF_CERT_HASH=abc123def456...  # Optional: for certificate pinning
```

### 7. Setting Up Your Custom Certificate Server

Here's a simple example of how to set up a TLS server with a custom certificate:

```go
// Example TLS proof server
func main() {
    // Load your custom certificate and key
    cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
    if err != nil {
        log.Fatal(err)
    }

    // Configure TLS
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
    }

    // Simple handler
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "TLS Proof Server - Certificate Hash: %s", getCertHash())
    })

    // Start HTTPS server
    server := &http.Server{
        Addr:      ":443",
        TLSConfig: tlsConfig,
    }
}
```

### 8. Benefits of Custom Certificate Approach

1. **Full Control**: You control certificate lifecycle and rotation
2. **Consistency**: Same certificate across all keepers
3. **Security**: Certificate pinning prevents MITM attacks
4. **Reliability**: Your server, your uptime guarantee
5. **Privacy**: No dependency on external services
6. **Auditability**: Complete control over the proof infrastructure

### 9. Certificate Generation Example

```bash
# Generate a self-signed certificate for testing
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=proof.yourdomain.com"

# Generate CA certificate for production
openssl genrsa -out ca.key 4096
openssl req -new -x509 -key ca.key -sha256 -subj "/C=US/ST=State/L=City/O=Organization/CN=TriggerX CA" -days 3650 -out ca.crt
```

This approach gives you complete control over your TLS proof infrastructure while maintaining the cryptographic integrity of the proof system. All keepers will use the same certificate, ensuring consistent and verifiable proofs across your network.
