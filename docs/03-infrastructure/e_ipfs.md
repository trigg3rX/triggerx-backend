# IPFS Infrastructure

TriggerX uses [Pinata](https://www.pinata.cloud/) as the IPFS provider for distributed storage of scripts and execution proofs.

## Use Cases

- **Agent Scripts**: Stored on IPFS when jobs are created, fetched during task execution
- **Execution Proofs**: Complete execution data uploaded for validation and attestation

## IPFS Client

The client (`pkg/ipfs/client.go`) uses Pinata's v3 API:

```go
config := ipfs.NewConfig(pinataHost, pinataJWT)
client, err := ipfs.NewClient(config)

// Upload
cid, err := client.Upload(ctx, "script.go", data)

// Fetch with gateway fallback
data, err := client.Fetch(ctx, cid)

// Delete pin
err := client.Delete(ctx, cid)
```

## Gateway Fallback

Fetch operations try multiple gateways:

1. **Primary**: Configured Pinata gateway
2. **Fallback**: `ipfs.io` public gateway

## Configuration

```bash
# Environment variables
IPFS_HOST=your-gateway.mypinata.cloud
PINATA_JWT=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Content Addressing

- **CID**: Content Identifier uniquely identifies data
- **Immutable**: Same content always produces same CID
- **Public**: All uploads set to `network=public` for distributed access

## Resources

- [Pinata Documentation](https://docs.pinata.cloud/)
- [IPFS Documentation](https://docs.ipfs.tech/)
- [Content Addressing](https://docs.ipfs.tech/concepts/content-addressing/)
