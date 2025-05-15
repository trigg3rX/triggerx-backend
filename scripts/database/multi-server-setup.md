# ScyllaDB Multi-Server Deployment Guide

This guide outlines how to set up ScyllaDB on two separate servers with data synchronization for production.

## Prerequisites

- Two servers (referred to as Server1 and Server2)
- Docker installed on both servers
- Network connectivity between the servers (ports 7000, 7001, 9042 open)

## Step 1: Server Configuration

### On Server1 (Primary Node)

Create a `docker-compose.yaml` file with the following content:

```yaml
services:
  scylla:
    image: scylladb/scylla
    container_name: triggerx-scylla
    ports:
      - "9042:9042"
      - "7000:7000"
      - "7001:7001"
    volumes:
      - scylla_data:/var/lib/scylla
    command: --smp 2 --memory 4G --overprovisioned 1 --seeds=SERVER1_IP --cluster-name triggerx_cluster
    restart: always
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"

volumes:
  scylla_data:
```

Replace `SERVER1_IP` with the actual IP address of Server1 (public or private IP that Server2 can reach).

### On Server2 (Secondary Node)

Create a `docker-compose.yaml` file with the following content:

```yaml
services:
  scylla:
    image: scylladb/scylla
    container_name: triggerx-scylla
    ports:
      - "9043:9042"
      - "7000:7000"
      - "7001:7001"
    volumes:
      - scylla_data:/var/lib/scylla
    command: --smp 2 --memory 4G --overprovisioned 1 --seeds=SERVER1_IP --cluster-name triggerx_cluster
    restart: always
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"

volumes:
  scylla_data:
```

Replace `SERVER1_IP` with the actual IP address of Server1.

## Step 2: Start the Cluster

### On Server1

```bash
docker-compose up -d
```

Wait for the node to start (approximately 30 seconds).

### On Server2

```bash
docker-compose up -d
```

## Step 3: Verify Cluster Formation

### On Server1

```bash
docker exec -it triggerx-scylla nodetool status
```

You should see both nodes listed in the output. If they don't appear immediately, wait a few minutes for the cluster to form.

## Step 4: Create Keyspace with SimpleStrategy

Connect to Server1 and create the keyspace:

```bash
docker exec -it triggerx-scylla cqlsh
```

In the CQL shell:

```cql
CREATE KEYSPACE IF NOT EXISTS triggerx
WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 2
};
```

## Step 5: Initialize the Schema

Create an `init-db.cql` file on Server1 with your schema definitions (tables, etc.) and run:

```bash
docker exec -i triggerx-scylla cqlsh < init-db.cql
```

## Step 6: Update Application Configuration

Update your application's database configuration to include both server addresses:

```go
func NewConfig() *Config {
    return &Config{
        Hosts:       []string{
            "SERVER1_IP:9042",
            "SERVER2_IP:9043"
        },
        Keyspace:    "triggerx",
        Timeout:     time.Second * 30,
        Retries:     5,
        ConnectWait: time.Second * 10,
    }
}
```

Replace `SERVER1_IP` and `SERVER2_IP` with the actual IP addresses of your servers.

## Step 7: Setup Monitoring (Optional but Recommended)

For production, it's recommended to set up monitoring for your ScyllaDB cluster.

```bash
# Install Prometheus and Grafana for monitoring
docker run -d --name prometheus -p 9090:9090 -v /path/to/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus
docker run -d --name grafana -p 3000:3000 grafana/grafana
```

Configure prometheus.yml to scrape ScyllaDB metrics:

```yaml
scrape_configs:
  - job_name: 'scylla'
    static_configs:
      - targets: ['SERVER1_IP:9180', 'SERVER2_IP:9180']
```

## Maintenance Operations

### Regular Tasks

Add these to your server crontab for maintenance:

```
# Run repair weekly
0 2 * * 0 docker exec triggerx-scylla nodetool repair -pr

# Take snapshot daily
0 3 * * * docker exec triggerx-scylla nodetool snapshot -t backup-$(date +\%Y\%m\%d) triggerx
```

### Check Cluster Status

```bash
docker exec -it triggerx-scylla nodetool status
```

### Backup Data

```bash
docker exec -it triggerx-scylla nodetool snapshot -t triggerx_backup triggerx
```

These snapshots are stored in `/var/lib/scylla/data/triggerx/*/snapshots/triggerx_backup`.

### Modify Replication Factor

If you need to change the replication factor:

```cql
ALTER KEYSPACE triggerx WITH replication = {
    'class': 'SimpleStrategy',
    'replication_factor': 2
};
```

Then run:

```bash
docker exec -it triggerx-scylla nodetool repair
```

## Testing Failover

To test that your application can handle node failures:

1. Stop one of the ScyllaDB nodes:
   ```bash
   docker stop triggerx-scylla
   ```

2. Verify your application can still function using the other node
3. Restart the stopped node:
   ```bash
   docker start triggerx-scylla
   ```

## Troubleshooting

### Node Not Joining Cluster

1. Check firewall settings (ports 7000, 7001, 9042 should be open between nodes)
2. Verify the `--seeds` parameter contains the correct IP address
3. Check logs: `docker logs triggerx-scylla`

### Data Inconsistency

Run repair operation:

```bash
docker exec -it triggerx-scylla nodetool repair -pr
```

## Production Security Considerations

For production deployments, consider these additional security measures:

1. Configure authentication:
   ```cql
   CREATE ROLE admin WITH PASSWORD = 'secure_password' AND SUPERUSER = true;
   ```

2. Enable client-to-node encryption:
   ```yaml
   command: >
     --smp 2 --memory 4G --overprovisioned 1 
     --seeds=SERVER1_IP 
     --cluster-name triggerx_cluster
     --authenticator PasswordAuthenticator
   ```

3. Use a dedicated network for inter-node communication:
   ```yaml
   networks:
     - scylla-internal
     - scylla-external
   ``` 