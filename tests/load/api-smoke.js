import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, SMOKE_THRESHOLDS } from './common/config.js';
import { registerAndLogin } from './common/auth.js';
import { smokeProfile } from './common/profiles.js';

export const options = {
    ...smokeProfile,
    thresholds: SMOKE_THRESHOLDS,
};

export default function () {
    // Health check (no auth required)
    const healthRes = http.get(`${BASE_URL}/health`);
    check(healthRes, { 'health is 200 or 503': (r) => r.status === 200 || r.status === 503 });

    // Login with admin credentials (may fail, that's ok for smoke)
    const loginPayload = JSON.stringify({
        email: 'admin@thecloud.local',
        password: 'Password123!',
    });
    const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    let apiKey = '';
    if (loginRes.status === 200) {
        apiKey = loginRes.json('api_key') || loginRes.json('data.api_key') || '';
    }
    apiKey = apiKey || __ENV.API_KEY || '';

    const authHeaders = { 'X-API-Key': apiKey };

    // List instances
    const instancesRes = http.get(`${BASE_URL}/instances`, { headers: authHeaders });
    check(instancesRes, {
        'instances status is 200 or 401': (r) => [200, 401].includes(r.status),
    });

    // Dashboard summary
    const dashRes = http.get(`${BASE_URL}/api/dashboard/summary`, { headers: authHeaders });
    check(dashRes, {
        'dashboard status is 200 or 401': (r) => [200, 401].includes(r.status),
    });

    // Create and immediately delete an instance (every 5th iteration)
    if (__ITER % 5 === 0) {
        const uniqueId = Date.now();
        const { authHeaders: regAuthHeaders } = registerAndLogin(uniqueId);

        const instPayload = JSON.stringify({
            name: `smoke-${uniqueId}`,
            image: 'alpine:latest',
        });
        const instRes = http.post(`${BASE_URL}/instances`, instPayload, { headers: regAuthHeaders });
        check(instRes, { 'smoke instance created': (r) => r.status === 202 || r.status === 200 });
        if (instRes.status === 202 || instRes.status === 200) {
            const instId = instRes.json('data.id');
            http.del(`${BASE_URL}/instances/${instId}`, null, { headers: regAuthHeaders });
        }
    }

    sleep(1);
}
