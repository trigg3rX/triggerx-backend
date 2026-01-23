# Docker Infrastructure

Docker provides sandboxed execution environments for user-provided scripts in TriggerX.

## Supported Languages

| Language   | Image                  | Use Case                                   |
| ---------- | ---------------------- | ------------------------------------------ |
| Go         | `golang:1.25.5-alpine` | Dynamic argument generation, agent scripts |
| TypeScript | `node:25.3.0-alpine`   | Dynamic argument generation, agent scripts |

## Container Pool Architecture

The Docker executor (`pkg/dockerexecutor/`) uses pre-warmed container pools:

- **Max Containers**: 5 per language
- **Min Containers**: 2 kept warm
- **Health Check Interval**: 10 minutes
- **Max Wait Time**: 60 seconds for container acquisition

## Resource Limits

Per-container limits (configurable in `config/services/docker-executor.yaml`):

- **Memory**: 1024MB
- **CPU**: 1.0 core
- **Timeout**: 300 seconds
- **Security**: `no-new-privileges` enforced

## Execution Flow

1. **Fetch Script**: Download from IPFS
2. **Acquire Container**: Get from pool or create new
3. **Mount Code**: Script mounted at `/code`
4. **Execute**: Run with language-specific command
5. **Parse Output**: Extract results from stdout
6. **Release**: Return container to pool

## Configuration

```yaml
# config/services/docker-executor.yaml
languages:
  go:
    base_config:
      max_containers: 5
      min_containers: 2
    docker_config:
      image: "golang:1.25.5-alpine"
      memory_limit: "1024m"
      cpu_limit: 1.0
    language_config:
      setup_script: "apk add --no-cache git && go mod init test && go mod tidy"
      run_command: "go run code.go"
```

## Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Security](https://docs.docker.com/engine/security/)
