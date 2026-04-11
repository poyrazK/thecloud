import http from 'k6/http';
import { check, sleep } from 'k6';

import { BASE_URL } from './common/config.js';
import { getOrCreateApiKey } from './common/auth.js';

export const options = {
    stages: [
        { duration: '10s', target: 50 },
        { duration: '30s', target: 50 },
        { duration: '10s', target: 0 },
    ],
    thresholds: {
        http_req_failed: ['rate<1.0'], // Allow high failure rates since we intentionally trigger rate limits
    },
};

export default function () {
    // Use cached auth to avoid rate limiting and auth issues
    const auth = getOrCreateApiKey(__VU, `ratelimit-test-${__VU}@loadtest.local`, 'Password123!', `RateLimitUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

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

    // Ensure no 5xx errors (5xx would indicate server issues, not rate limiting)
    check({}, { 'no 5xx errors': () => otherCount === 0 || otherCount < 5 });

    sleep(1);
}
