import http from 'k6/http';
import { check, sleep, fail } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE_URL } from './common/config.js';
import { getOrCreateApiKey } from './common/auth.js';

export const options = {
    stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 20 },
        { duration: '2m', target: 20 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<5000'],
    },
};

function createVpc(authHeaders, uniqueId) {
    const vpcPayload = JSON.stringify({ name: `vpc-db-${uniqueId}`, cidr_block: '10.1.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    if (vpcRes.status === 201 || vpcRes.status === 200) {
        return vpcRes.json('data.id');
    }
    return null;
}

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const dbName = `db-${uniqueId}`;
    const engine = 'postgres';
    const version = '15';

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `loadtest-${__VU}@loadtest.local`, 'Password123!', `Load User ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // Create VPC first (databases need a VPC)
    const vpcId = createVpc(authHeaders, uniqueId);

    // 1. Create database
    const createDbPayload = JSON.stringify({
        name: dbName,
        engine: engine,
        version: version,
        vpc_id: vpcId,
        allocated_storage: 10,
    });
    const createDbRes = http.post(`${BASE_URL}/databases`, createDbPayload, { headers: authHeaders });
    check(createDbRes, { 'database creation accepted': (r) => r.status === 201 || r.status === 200 });

    if (createDbRes.status !== 201 && createDbRes.status !== 200) {
        console.error(`Database creation failed: ${createDbRes.status} ${createDbRes.body}`);
        // Cleanup VPC if DB creation failed
        if (vpcId) {
            http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        }
        return;
    }

    const dbId = createDbRes.json('data.id');

    // 2. Poll for database to be RUNNING
    let isRunning = false;
    for (let i = 0; i < 60; i++) { // Wait up to 60s
        const getDbRes = http.get(`${BASE_URL}/databases/${dbId}`, { headers: authHeaders });
        if (getDbRes.status === 200) {
            const status = getDbRes.json('data.status');
            if (status === 'running') {
                isRunning = true;
                break;
            }
            // If FAILED, stop polling
            if (status === 'failed') {
                console.error(`Database failed to provision: ${getDbRes.body}`);
                break;
            }
        }
        sleep(1);
    }
    check(isRunning, { 'database is running': (val) => val === true });

    if (!isRunning) {
        console.error(`Database ${dbId} never reached running state`);
        // Try to cleanup anyway
        http.del(`${BASE_URL}/databases/${dbId}`, null, { headers: authHeaders });
        if (vpcId) {
            http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        }
        fail('Database never reached running state');
        return;
    }

    // 3. Get connection string
    const connStrRes = http.get(`${BASE_URL}/databases/${dbId}/connection-string`, { headers: authHeaders });
    check(connStrRes, { 'connection string retrieved': (r) => r.status === 200 });

    // 4. List databases
    const listDbRes = http.get(`${BASE_URL}/databases`, { headers: authHeaders });
    check(listDbRes, { 'databases listed': (r) => r.status === 200 });

    // 5. Rotate credentials
    const rotateRes = http.post(`${BASE_URL}/databases/${dbId}/rotate-credentials`, null, { headers: authHeaders });
    check(rotateRes, { 'credentials rotated': (r) => r.status === 200 || r.status === 202 });

    // 6. Get database details after rotation
    const getDbAfterRes = http.get(`${BASE_URL}/databases/${dbId}`, { headers: authHeaders });
    check(getDbAfterRes, { 'database details after rotation': (r) => r.status === 200 });

    // 7. Delete database
    const deleteDbRes = http.del(`${BASE_URL}/databases/${dbId}`, null, { headers: authHeaders });
    check(deleteDbRes, { 'database deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });

    // 8. Cleanup VPC
    if (vpcId) {
        // Wait for DB to be fully deleted first
        sleep(2);
        const deleteVpcRes = http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        check(deleteVpcRes, { 'vpc deleted': (r) => r.status === 204 || r.status === 200 });
    }

    sleep(1);
}
