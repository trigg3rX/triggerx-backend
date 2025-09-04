# WebSocket Real-time Task Data System

This WebSocket implementation provides real-time updates for task data changes in the TriggerX backend system. It allows frontend clients to receive instant notifications when task data is created, updated, or modified.

## Features

- **Real-time Updates**: Instant notifications for task data changes
- **Room-based Subscriptions**: Subscribe to specific jobs, tasks, or users
- **Authentication**: API key-based authentication for WebSocket connections
- **Rate Limiting**: Connection limits per IP address
- **Graceful Shutdown**: Proper cleanup of connections and resources
- **Event Types**: Support for task creation, updates, status changes, and fee updates

## Architecture

### Core Components

1. **Hub**: Central WebSocket connection manager
2. **Client**: Individual WebSocket client connection
3. **Message**: WebSocket message types and structures
4. **Publisher**: Event publisher for task data changes
5. **Middleware**: Authentication and connection management

### Event Types

- `TASK_CREATED`: New task created
- `TASK_UPDATED`: Task execution/attestation data updated
- `TASK_STATUS_CHANGED`: Task status changes
- `TASK_FEE_UPDATED`: Task fee changes

## API Endpoints

### WebSocket Connection
```
GET /api/ws/tasks?api_key=YOUR_API_KEY
```

### WebSocket Statistics
```
GET /api/ws/stats
```

### WebSocket Health Check
```
GET /api/ws/health
```

## Usage

### 1. Connect to WebSocket

```javascript
const ws = new WebSocket('ws://localhost:9002/api/ws/tasks?api_key=YOUR_API_KEY');

ws.onopen = function() {
    console.log('WebSocket connected');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('Received message:', message);
};
```

### 2. Subscribe to Rooms

```javascript
// Subscribe to a specific job
ws.send(JSON.stringify({
    type: 'SUBSCRIBE',
    data: {
        room: 'job:123',
        job_id: '123'
    }
}));

// Subscribe to a specific task
ws.send(JSON.stringify({
    type: 'SUBSCRIBE',
    data: {
        room: 'task:456',
        task_id: '456'
    }
}));

// Subscribe to user's tasks
ws.send(JSON.stringify({
    type: 'SUBSCRIBE',
    data: {
        room: 'user:user_123',
        user_id: 'user_123'
    }
}));
```

### 3. Unsubscribe from Rooms

```javascript
ws.send(JSON.stringify({
    type: 'UNSUBSCRIBE',
    data: {
        room: 'job:123'
    }
}));
```

### 4. Handle Real-time Updates

```javascript
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    switch(message.type) {
        case 'TASK_CREATED':
            console.log('New task created:', message.data);
            break;
        case 'TASK_UPDATED':
            console.log('Task updated:', message.data);
            break;
        case 'TASK_STATUS_CHANGED':
            console.log('Task status changed:', message.data);
            break;
        case 'TASK_FEE_UPDATED':
            console.log('Task fee updated:', message.data);
            break;
    }
};
```

## Message Format

### Incoming Messages (Client to Server)

```json
{
    "type": "SUBSCRIBE",
    "data": {
        "room": "job:123",
        "job_id": "123"
    }
}
```

### Outgoing Messages (Server to Client)

```json
{
    "type": "TASK_CREATED",
    "data": {
        "task_id": 123,
        "job_id": "456",
        "user_id": "user_789",
        "changes": {
            "task_id": 123,
            "job_id": "456",
            "task_definition_id": 789,
            "is_imua": false,
            "created_at": "2024-01-01T00:00:00Z"
        },
        "timestamp": "2024-01-01T00:00:00Z"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

## Room Types

### Job Rooms
- **Format**: `job:{job_id}`
- **Purpose**: Receive updates for all tasks in a specific job
- **Example**: `job:123`

### Task Rooms
- **Format**: `task:{task_id}`
- **Purpose**: Receive updates for a specific task
- **Example**: `task:456`

### User Rooms
- **Format**: `user:{user_id}`
- **Purpose**: Receive updates for all tasks belonging to a user
- **Example**: `user:user_789`

## Authentication

WebSocket connections use the **same authentication system** as the REST API, ensuring consistency across your application. The API key can be provided in two ways:

1. **Query Parameter**: `?api_key=YOUR_API_KEY` (WebSocket specific)
2. **Header**: `X-Api-Key: YOUR_API_KEY` (Standard API key location)

### Authentication Features

- **Database Validation**: API keys are validated against the database
- **Active Status Check**: Only active API keys are accepted
- **User Identification**: User ID is extracted from the API key owner
- **Last Used Tracking**: API key usage is tracked for monitoring
- **Consistent Logic**: Uses the same `ApiKeyAuth` middleware as REST endpoints

## Rate Limiting

WebSocket connections use the **same rate limiting system** as the REST API:

- **API Key Based**: Rate limits are applied per API key (not just IP)
- **Redis Backend**: Uses Redis for distributed rate limiting
- **Configurable Limits**: Each API key has its own rate limit configuration
- **Connection Limits**: Maximum 10 WebSocket connections per IP address
- **Automatic Cleanup**: Connections are automatically cleaned up on disconnect
- **Consistent Logic**: Uses the same `RateLimiter` middleware as REST endpoints

## Error Handling

### Connection Errors
```json
{
    "type": "ERROR",
    "data": {
        "code": "AUTHENTICATION_FAILED",
        "message": "Invalid API key"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

### Common Error Codes
- `MISSING_API_KEY`: API key not provided
- `INVALID_API_KEY`: Invalid or expired API key
- `RATE_LIMIT_EXCEEDED`: Too many connections from IP
- `INVALID_ROOM`: Invalid room format
- `ACCESS_DENIED`: No permission to access room

## Integration with Repository Layer

The WebSocket system integrates at the **repository layer**, ensuring that **ALL** task data changes emit WebSocket events, regardless of how they are triggered:

### Repository-Level Event Emission

WebSocket events are emitted directly from the task repository functions:

- `CreateTaskDataInDB()` - Emits `TASK_CREATED` event
- `UpdateTaskExecutionDataInDB()` - Emits `TASK_UPDATED` event  
- `UpdateTaskAttestationDataInDB()` - Emits `TASK_UPDATED` event
- `UpdateTaskFee()` - Emits `TASK_FEE_UPDATED` event
- `UpdateTaskNumberAndStatus()` - Emits `TASK_STATUS_CHANGED` event

### Benefits of Repository-Level Integration

1. **Complete Coverage**: Events are emitted for ALL task data changes, whether triggered by:
   - REST API calls
   - Internal backend services
   - Scheduled jobs
   - Direct repository calls

2. **No Missed Events**: Unlike API-level integration, repository-level ensures no task changes are missed

3. **Backward Compatibility**: Existing code continues to work without modification

4. **Centralized Logic**: All WebSocket event logic is centralized in the repository layer

### Repository Configuration

The task repository can be created with or without WebSocket publisher:

```go
// Without WebSocket (backward compatible)
repo := repository.NewTaskRepository(db)

// With WebSocket events
publisher := events.NewPublisher(hub, logger)
repo := repository.NewTaskRepositoryWithPublisher(db, publisher)
```

### API Endpoints That Trigger Events

The following REST API endpoints will trigger WebSocket events (via repository calls):

- `POST /api/tasks` - Triggers `TASK_CREATED` event
- `PUT /api/tasks/execution/:id` - Triggers `TASK_UPDATED` event
- `PUT /api/tasks/:id/fee` - Triggers `TASK_FEE_UPDATED` event

## Monitoring and Statistics

### WebSocket Statistics Endpoint
```bash
curl http://localhost:9002/api/ws/stats
```

Response:
```json
{
    "status": "success",
    "data": {
        "total_clients": 5,
        "total_rooms": 12,
        "rooms": {
            "job:123": 2,
            "task:456": 1,
            "user:user_789": 2
        }
    }
}
```

### Health Check Endpoint
```bash
curl http://localhost:9002/api/ws/health
```

Response:
```json
{
    "status": "healthy",
    "timestamp": 1704067200,
    "websocket": {
        "active_connections": 5,
        "active_rooms": 12
    }
}
```

## Development and Testing

### Running Tests
```bash
go test ./internal/dbserver/websocket/...
```

### Testing WebSocket Connection
```bash
# Using wscat (install with: npm install -g wscat)
wscat -c "ws://localhost:9002/api/ws/tasks?api_key=YOUR_API_KEY"
```

## Production Considerations

1. **Load Balancing**: Use sticky sessions for WebSocket connections
2. **Scaling**: Consider Redis for multi-instance WebSocket scaling
3. **Monitoring**: Monitor connection counts and message rates
4. **Security**: Implement proper origin checking and rate limiting
5. **SSL/TLS**: Use WSS (WebSocket Secure) in production

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check if WebSocket endpoint is accessible
2. **Authentication Failed**: Verify API key is valid and not expired
3. **Rate Limited**: Reduce number of concurrent connections
4. **No Messages**: Ensure you're subscribed to the correct rooms

### Debug Mode

Enable debug logging to see detailed WebSocket activity:
```go
logger := logging.NewLogger("websocket", "debug")
```
