import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    // We will control VUs and duration from command line
    thresholds: {
        http_req_failed: ['rate<0.10'],  // Allow 10% for extreme local load
        http_req_duration: ['p(95)<5000'], // 5s is ok for high load local test
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://cloud-api';

export default function () {
    const id = uuidv4().substring(0, 8);
    const email = `u-${id}@test.io`;
    const headers = { 'Content-Type': 'application/json' };

    // 1. Register (fast path)
    const regRes = http.post(`${BASE_URL}/auth/register`,
        JSON.stringify({ email, password: 'Pass123!', name: id }),
        { headers, timeout: '10s' });

    if (regRes.status !== 201 && regRes.status !== 200) return;

    // 2. Login (get token)
    const loginRes = http.post(`${BASE_URL}/auth/login`,
        JSON.stringify({ email, password: 'Pass123!' }),
        { headers, timeout: '10s' });

    if (loginRes.status !== 200) return;

    const apiKey = loginRes.json('data.api_key');
    if (!apiKey) return;

    const authHeaders = { ...headers, 'X-API-Key': apiKey };

    // 3. Create VPC (quick operation)
    const vpcRes = http.post(`${BASE_URL}/vpcs`,
        JSON.stringify({ name: `v-${id}`, cidr_block: '10.0.0.0/16' }),
        { headers: authHeaders, timeout: '10s' });

    check(vpcRes, { 'vpc ok': r => r.status === 201 });

    if (vpcRes.status !== 201) return;
    const vpcId = vpcRes.json('data.id');

    // 4. Launch instance (async, don't wait)
    const instRes = http.post(`${BASE_URL}/instances`,
        JSON.stringify({ name: `i-${id}`, image: 'alpine:latest', vpc_id: vpcId }),
        { headers: authHeaders, timeout: '10s' });

    check(instRes, { 'inst accepted': r => r.status === 202 });

    // 5. Cleanup immediately (no polling)
    if (instRes.status === 202) {
        const instId = instRes.json('data.id');
        http.del(`${BASE_URL}/instances/${instId}`, null, { headers: authHeaders });
    }

    http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
}
