// Simple Node.js WebSocket test client for TriggerX backend
// Usage: node test_client.js <API_KEY> <JOB_ID>

const WebSocket = require('ws');

const API_KEY = process.argv[2] || 'YOUR_API_KEY';
const JOB_ID = process.argv[3] || '123';

const wsUrl = `ws://localhost:9002/api/ws/tasks?api_key=${API_KEY}`;

console.log('Connecting to:', wsUrl);
const ws = new WebSocket(wsUrl);

ws.on('open', () => {
  console.log('WebSocket connected');
  // Subscribe to a job room
  ws.send(JSON.stringify({
    type: 'SUBSCRIBE',
    data: { room: `job:${JOB_ID}`, job_id: JOB_ID }
  }));
});

ws.on('message', (data) => {
  try {
    const message = JSON.parse(data);
    console.log('Received:', message);
  } catch (e) {
    console.log('Received non-JSON message:', data);
  }
});

ws.on('error', (err) => {
  console.error('WebSocket error:', err);
});

ws.on('close', () => {
  console.log('WebSocket closed');
});
