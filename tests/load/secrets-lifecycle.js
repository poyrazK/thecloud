import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE_URL, LIFECYCLE_THRESHOLDS } from './common/config.js';
import { getOrCreateApiKey } from './common/auth.js';

export const options = {
    stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 20 },
        { duration: '2m', target: 20 },
        { duration: '30s', target: 0 },
    ],
    thresholds: LIFECYCLE_THRESHOLDS,
};

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const secretName = `secret-${uniqueId}`;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `secrettest-${__VU}@loadtest.local`, 'Password123!', `SecretUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create secret
    const secretPayload = JSON.stringify({
        name: secretName,
        value: `super-secret-value-${uniqueId}`,
        description: `Test secret ${uniqueId}`,
    });
    const secretRes = http.post(`${BASE_URL}/secrets`, secretPayload, { headers: authHeaders });
    check(secretRes, { 'secret created': (r) => r.status === 201 || r.status === 200 });

    if (secretRes.status !== 201 && secretRes.status !== 200) {
        console.error(`Secret creation failed: ${secretRes.status} ${secretRes.body}`);
        return;
    }
    const secretId = secretRes.json('data.id');

    // 2. Get secret (by ID)
    const getRes = http.get(`${BASE_URL}/secrets/${secretId}`, { headers: authHeaders });
    check(getRes, { 'secret retrieved by id': (r) => r.status === 200 });

    // 3. Get secret by name
    const getByNameRes = http.get(`${BASE_URL}/secrets/${secretName}`, { headers: authHeaders });
    check(getByNameRes, { 'secret retrieved by name': (r) => r.status === 200 });

    // 4. List secrets
    const listRes = http.get(`${BASE_URL}/secrets`, { headers: authHeaders });
    check(listRes, { 'secrets listed': (r) => r.status === 200 });

    // 5. Delete secret
    const delRes = http.del(`${BASE_URL}/secrets/${secretId}`, null, { headers: authHeaders });
    check(delRes, { 'secret deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });

    sleep(1);
}
