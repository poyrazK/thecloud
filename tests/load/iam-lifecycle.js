import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE_URL, LIFECYCLE_THRESHOLDS } from './common/config.js';
import { getOrCreateApiKey } from './common/auth.js';

export const options = {
    stages: [
        { duration: '30s', target: 5 },
        { duration: '1m', target: 10 },
        { duration: '2m', target: 10 },
        { duration: '30s', target: 0 },
    ],
    thresholds: LIFECYCLE_THRESHOLDS,
};

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const policyName = `policy-${uniqueId}`;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `iam-${__VU}@loadtest.local`, 'Password123!', `IAMUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create policy
    const policyPayload = JSON.stringify({
        name: policyName,
        description: `Test policy ${uniqueId}`,
        statements: [
            {
                effect: 'Allow',
                action: ['instance:read', 'instance:list', 'vpc:read'],
                resource: ['*'],
            },
        ],
    });
    const policyRes = http.post(`${BASE_URL}/iam/policies`, policyPayload, { headers: authHeaders });
    check(policyRes, { 'policy created': (r) => r.status === 201 || r.status === 200 });

    if (policyRes.status !== 201 && policyRes.status !== 200) {
        console.error(`Policy creation failed: ${policyRes.status} ${policyRes.body}`);
        sleep(1);
        return;
    }
    const policyId = policyRes.json('data.id');

    // 2. List policies
    const listPoliciesRes = http.get(`${BASE_URL}/iam/policies`, { headers: authHeaders });
    check(listPoliciesRes, { 'policies listed': (r) => r.status === 200 });

    // 3. Get policy details
    const getPolicyRes = http.get(`${BASE_URL}/iam/policies/${policyId}`, { headers: authHeaders });
    check(getPolicyRes, { 'policy retrieved': (r) => r.status === 200 });

    // 4. Update policy
    const updatePayload = JSON.stringify({
        description: `Updated policy ${uniqueId}`,
        statements: [
            {
                effect: 'Allow',
                action: ['instance:read', 'instance:list', 'instance:create', 'vpc:read'],
                resource: ['*'],
            },
        ],
    });
    const updateRes = http.put(`${BASE_URL}/iam/policies/${policyId}`, updatePayload, { headers: authHeaders });
    check(updateRes, { 'policy updated': (r) => r.status === 200 });

    // 5. Get user ID from the API key response to attach policy
    // The auth token contains user info; we use the current user's ID
    const userIdFromToken = auth.userId || auth.apiKey.split('-')[0];

    // 6. Attach policy to user (using the authenticated user)
    // Note: We need a valid user ID. Since we're using API key auth,
    // we'll get the current user's ID from the auth response if available
    const attachRes = http.post(`${BASE_URL}/iam/users/${auth.userId}/policies/${policyId}`, null, { headers: authHeaders });
    check(attachRes, { 'policy attached': (r) => r.status === 200 || r.status === 201 || r.status === 204 });

    // 7. List user policies
    const listUserPoliciesRes = http.get(`${BASE_URL}/iam/users/${auth.userId}/policies`, { headers: authHeaders });
    check(listUserPoliciesRes, { 'user policies listed': (r) => r.status === 200 });

    // 8. Detach policy from user
    const detachRes = http.del(`${BASE_URL}/iam/users/${auth.userId}/policies/${policyId}`, null, { headers: authHeaders });
    check(detachRes, { 'policy detached': (r) => r.status === 200 || r.status === 204 || r.status === 202 });

    // 9. Delete policy
    const deletePolicyRes = http.del(`${BASE_URL}/iam/policies/${policyId}`, null, { headers: authHeaders });
    check(deletePolicyRes, { 'policy deleted': (r) => r.status === 200 || r.status === 204 || r.status === 202 });

    sleep(1);
}