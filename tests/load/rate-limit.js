import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '10s', target: 50 },
        { duration: '30s', target: 50 },
        { duration: '10s', target: 0 },
    ],
    thresholds: {
        http_req_failed: ['rate<0.10'], // Allow failures as we're testing rate limits
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

function getApiKey() {
    const email = `ratelimit-test-${Date.now()}@loadtest.local`;
    const password = 'Password123!';
    const headers = { 'Content-Type': 'application/json' };

    const regPayload = JSON.stringify({ email, password, name: 'RateLimitTest' });
    const regRes = http.post(`${BASE_URL}/auth/register`, regPayload, { headers });

    const loginRes = http.post(`${BASE_URL}/auth/login`, regPayload, { headers });
    if (loginRes.status === 200) {
        return { apiKey: loginRes.json('data.api_key'), headers };
    }
    return { apiKey: __ENV.API_KEY || '', headers };
}

export default function () {
    const { apiKey, headers } = getApiKey();
    const authHeaders = { ...headers, 'X-API-Key': apiKey };

    // Burst requests to the same endpoint to trigger rate limiting
    // Using instances list as it's a simple read endpoint
    const batchRequests = [];
    for (let i = 0; i < 100; i++) {
        batchRequests.push(['GET', `${BASE_URL}/instances`, null, { headers: authHeaders }]);
    }

    const batchRes = http.batch(batchRequests);

    let successCount = 0;
    let rateLimitedCount = 0;
    let otherCount = 0;

    for (const res of batchRes) {
        if (res.status === 200) {
            successCount++;
        } else if (res.status === 429) {
            rateLimitedCount++;
        } else {
            otherCount++;
        }
    }

    check({}, { 'some requests succeeded': () => successCount > 0 });
    check({}, { 'rate limit detected': () => rateLimitedCount > 0 || successCount < 100 });

    // Also test health endpoint which may have different rate limits
    const healthBatch = [];
    for (let i = 0; i < 200; i++) {
        healthBatch.push(['GET', `${BASE_URL}/health`, null, null]);
    }
    const healthBatchRes = http.batch(healthBatch);

    let healthSuccess = 0;
    let healthLimited = 0;
    for (const res of healthBatchRes) {
        if (res.status === 200) healthSuccess++;
        else if (res.status === 429) healthLimited++;
    }

    check({}, { 'health endpoint tested': () => healthSuccess > 0 || healthLimited > 0 });

    sleep(1);
}
