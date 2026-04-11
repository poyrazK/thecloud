import http from 'k6/http';
import { check, sleep, fail } from 'k6';
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

function waitForCache(authHeaders, cacheIdOrName, targetStatus = 'running') {
    for (let i = 0; i < 30; i++) {
        const getRes = http.get(`${BASE_URL}/caches/${cacheIdOrName}`, { headers: authHeaders });
        if (getRes.status === 200 && getRes.json('data.status') === targetStatus) {
            return true;
        }
        sleep(1);
    }
    return false;
}

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const cacheName = `cache-${uniqueId}`;
    const engine = 'redis';
    const version = '7';
    const memoryMb = 256;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `cachetest-${__VU}@loadtest.local`, 'Password123!', `CacheUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create VPC (caches need a VPC)
    const vpcPayload = JSON.stringify({ name: `vpc-cache-${uniqueId}`, cidr_block: '10.7.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'vpc created': (r) => r.status === 201 || r.status === 200 });
    if (vpcRes.status !== 201 && vpcRes.status !== 200) {
        console.error(`VPC creation failed: ${vpcRes.status} ${vpcRes.body}`);
        return;
    }
    const vpcId = vpcRes.json('data.id');

    // 2. Create cache
    const cachePayload = JSON.stringify({
        name: cacheName,
        version: version,
        memory_mb: memoryMb,
        vpc_id: vpcId,
    });
    const cacheRes = http.post(`${BASE_URL}/caches`, cachePayload, { headers: authHeaders });
    check(cacheRes, { 'cache creation accepted': (r) => r.status === 201 || r.status === 200 });

    if (cacheRes.status !== 201 && cacheRes.status !== 200) {
        console.error(`Cache creation failed: ${cacheRes.status} ${cacheRes.body}`);
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const cacheId = cacheRes.json('data.id');

    // 3. Poll for cache to be running
    const isRunning = waitForCache(authHeaders, cacheId, 'running');
    check(isRunning, { 'cache is running': (val) => val === true });

    if (!isRunning) {
        console.error(`Cache ${cacheId} never reached running state`);
        http.del(`${BASE_URL}/caches/${cacheId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        fail('Cache never reached running state');
        return;
    }

    // 4. Get connection string
    const connRes = http.get(`${BASE_URL}/caches/${cacheId}/connection`, { headers: authHeaders });
    check(connRes, { 'connection string retrieved': (r) => r.status === 200 });

    // 5. List caches
    const listRes = http.get(`${BASE_URL}/caches`, { headers: authHeaders });
    check(listRes, { 'caches listed': (r) => r.status === 200 });

    // 6. Flush cache
    const flushRes = http.post(`${BASE_URL}/caches/${cacheId}/flush`, null, { headers: authHeaders });
    check(flushRes, { 'cache flushed': (r) => r.status === 200 });

    // 7. Get cache stats
    const statsRes = http.get(`${BASE_URL}/caches/${cacheId}/stats`, { headers: authHeaders });
    check(statsRes, { 'cache stats retrieved': (r) => r.status === 200 });

    // 8. Delete cache
    const delRes = http.del(`${BASE_URL}/caches/${cacheId}`, null, { headers: authHeaders });
    check(delRes, { 'cache deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });

    // 9. Cleanup VPC
    sleep(2);
    http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });

    sleep(1);
}
