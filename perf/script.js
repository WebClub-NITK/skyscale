import http from 'k6/http';
import { sleep, check, group } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Configuration
const API_ENDPOINT = 'http://localhost:8080/api'; // Adjust this to match your API endpoint
const API_KEY = 'your-api-key-here'; // Replace with your API key

// Test options
export const options = {
  stages: [
    { duration: '1m', target: 10 }, // Ramp up to 10 users over 1 minute
    { duration: '3m', target: 10 }, // Stay at 10 users for 3 minutes
    { duration: '1m', target: 0 },  // Ramp down to 0 users over 1 minute
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],   // Less than 1% of requests should fail
  },
};

// Helper function to generate test function code
function generateTestFunction() {
  return `
def handler(event, context):
    return {"message": "Hello from test function!"}
  `.trim();
}

// Main test function
export default function() {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${API_KEY}`,
  };

  group('Health Check', function() {
    const healthCheck = http.get(`${API_ENDPOINT}/health`);
    check(healthCheck, {
      'health check returns 200': (r) => r.status === 200,
      'health check body is OK': (r) => r.body.includes('OK'),
    });
  });

  group('Function Management', function() {
    // Register a new function
    const functionName = `test-function-${randomString(8)}`;
    const registerPayload = JSON.stringify({
      name: functionName,
      runtime: 'python3.9',
      memory: 128,
      timeout: 30,
      code: generateTestFunction(),
      requirements: '',
      config: '{}',
    });

    const registerRes = http.post(
      `${API_ENDPOINT}/functions`,
      registerPayload,
      { headers }
    );

    check(registerRes, {
      'function registration successful': (r) => r.status === 200,
      'function has valid ID': (r) => JSON.parse(r.body).id !== undefined,
    });

    if (registerRes.status === 200) {
      const functionId = JSON.parse(registerRes.body).id;

      // Get function details
      const getRes = http.get(
        `${API_ENDPOINT}/functions/${functionId}`,
        { headers }
      );

      check(getRes, {
        'get function successful': (r) => r.status === 200,
        'function details match': (r) => JSON.parse(r.body).name === functionName,
      });

      // Invoke function
      const invokePayload = JSON.stringify({
        input: { test: "data" },
        sync: true,
      });

      const invokeRes = http.post(
        `${API_ENDPOINT}/functions/${functionId}/invoke`,
        invokePayload,
        { headers }
      );

      check(invokeRes, {
        'function invocation successful': (r) => r.status === 200,
        'invocation has request ID': (r) => JSON.parse(r.body).request_id !== undefined,
      });

      // Check execution status if async invocation
      if (!JSON.parse(invokePayload).sync) {
        sleep(1); // Wait a bit for async execution
        const requestId = JSON.parse(invokeRes.body).request_id;
        const execRes = http.get(
          `${API_ENDPOINT}/executions/${requestId}`,
          { headers }
        );

        check(execRes, {
          'execution status check successful': (r) => r.status === 200,
        });
      }
    }
  });

  group('VM Management', function() {
    // List VMs
    const listVMsRes = http.get(
      `${API_ENDPOINT}/vms`,
      { headers }
    );

    check(listVMsRes, {
      'list VMs successful': (r) => r.status === 200,
      'VMs response is array': (r) => Array.isArray(JSON.parse(r.body)),
    });
  });

  // Sleep between iterations
  sleep(1);
}
