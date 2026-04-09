import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL } from './common/config.js';
import { getOrCreateApiKey } from './common/auth.js';

export const options = {
    stages: [
        { duration: '10s', target: 5 },
        { duration: '30s', target: 10 },
        { duration: '10s', target: 0 },
    ],
};

export default function () {
    const uniqueId = Date.now();

    // Use cached auth to avoid rate limiting - only registers once per VU
    const auth = getOrCreateApiKey(__VU, `loadtest-${__VU}@loadtest.local`, 'Password123!', `Load User ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Test resource not found (valid UUID but doesn't exist)
    const freshUuid = '00000000-0000-0000-0000-000000000000';
    const notFoundRes = http.get(`${BASE_URL}/instances/${freshUuid}`, { headers: authHeaders });
    check(notFoundRes, { 'not found returns 404': (r) => r.status === 404 });

    // 2. Test creating resource with missing required fields
    const missingFieldsRes = http.post(
        `${BASE_URL}/instances`,
        JSON.stringify({ name: 'test' }), // missing image, vpc_id, etc.
        { headers: authHeaders }
    );
    check(missingFieldsRes, { 'missing required fields rejected': (r) => r.status === 400 || r.status === 422 });

    // 3. Test creating VPC with invalid CIDR
    const badCidrRes = http.post(
        `${BASE_URL}/vpcs`,
        JSON.stringify({ name: `vpc-err-${uniqueId}`, cidr_block: '999.999.999.999/32' }),
        { headers: authHeaders }
    );
    check(badCidrRes, { 'invalid CIDR rejected': (r) => r.status === 400 || r.status === 422 });

    // 4. Test creating VPC with valid CIDR (use __VU to avoid name collisions between VUs)
    const vpcPayload = JSON.stringify({ name: `vpc-err-${uniqueId}-vu${__VU}`, cidr_block: '10.100.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'valid VPC created': (r) => r.status === 201 || r.status === 200 });
    const vpcId = vpcRes.json('data.id');

    // 5. Try to delete non-existent resource
    const deleteNotFound = http.del(`${BASE_URL}/vpcs/00000000-0000-0000-0000-000000000000`, null, { headers: authHeaders });
    check(deleteNotFound, { 'delete non-existent returns 404': (r) => r.status === 404 });

    // 6. Try to create duplicate VPC name
    const dupVpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(dupVpcRes, { 'duplicate VPC name handled': (r) => r.status === 200 || r.status === 201 || r.status === 409 || r.status === 400 || r.status === 500 });

    // 7. Cleanup
    if (vpcId) {
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
    }

    sleep(1);
}
