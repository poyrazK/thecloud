import http from 'k6/http';
import { check, sleep, fail } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE_URL, LIFECYCLE_THRESHOLDS } from './common/config.js';
import { getOrCreateApiKey } from './common/auth.js';

export const options = {
    stages: [
        { duration: '30s', target: 2 },
        { duration: '2m', target: 2 },
        { duration: '1m', target: 0 },
    ],
    thresholds: LIFECYCLE_THRESHOLDS,
};

function waitForSnapshot(authHeaders, snapshotId, targetStatus = 'available') {
    for (let i = 0; i < 60; i++) {
        const getRes = http.get(`${BASE_URL}/snapshots/${snapshotId}`, { headers: authHeaders });
        if (getRes.status === 200 && getRes.json('data.status') === targetStatus) {
            return true;
        }
        sleep(1);
    }
    return false;
}

export default function () {
    const uniqueId = uuidv4().substring(0, 8);

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `snapshotest-${__VU}@loadtest.local`, 'Password123!', `SnapshotUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create VPC (required for volume operations)
    const vpcPayload = JSON.stringify({ name: `vpc-snap-${uniqueId}`, cidr_block: '10.9.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'vpc created': (r) => r.status === 201 || r.status === 200 });
    if (vpcRes.status !== 201 && vpcRes.status !== 200) {
        console.error(`VPC creation failed: ${vpcRes.status} ${vpcRes.body}`);
        return;
    }
    const vpcId = vpcRes.json('data.id');

    // 2. Create volume
    const volPayload = JSON.stringify({
        name: `vol-snap-${uniqueId}`,
        size_gb: 10,
    });
    const volRes = http.post(`${BASE_URL}/volumes`, volPayload, { headers: authHeaders });
    check(volRes, { 'volume created': (r) => r.status === 201 || r.status === 200 });
    if (volRes.status !== 201 && volRes.status !== 200) {
        console.error(`Volume creation failed: ${volRes.status} ${volRes.body}`);
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const volId = volRes.json('data.id');

    // 3. Poll for volume to be available
    let isVolumeAvailable = false;
    for (let i = 0; i < 60; i++) {
        const getVolRes = http.get(`${BASE_URL}/volumes/${volId}`, { headers: authHeaders });
        if (getVolRes.status === 200 && getVolRes.json('data.status') === 'available') {
            isVolumeAvailable = true;
            break;
        }
        sleep(1);
    }
    check(isVolumeAvailable, { 'volume is available': (val) => val === true });

    if (!isVolumeAvailable) {
        console.error(`Volume ${volId} never reached available state`);
        http.del(`${BASE_URL}/volumes/${volId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        fail('Volume never reached available state');
        return;
    }

    // 4. Create snapshot from volume
    const snapPayload = JSON.stringify({
        volume_id: volId,
        description: `Test snapshot ${uniqueId}`,
    });
    const snapRes = http.post(`${BASE_URL}/snapshots`, snapPayload, { headers: authHeaders });
    check(snapRes, { 'snapshot created': (r) => r.status === 201 || r.status === 200 });

    if (snapRes.status !== 201 && snapRes.status !== 200) {
        console.error(`Snapshot creation failed: ${snapRes.status} ${snapRes.body}`);
        http.del(`${BASE_URL}/volumes/${volId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const snapId = snapRes.json('data.id');

    // 5. Poll for snapshot to be available
    const isAvailable = waitForSnapshot(authHeaders, snapId, 'available');
    check(isAvailable, { 'snapshot is available': (val) => val === true });

    if (!isAvailable) {
        console.error(`Snapshot ${snapId} never reached available state`);
        http.del(`${BASE_URL}/snapshots/${snapId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/volumes/${volId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        fail('Snapshot never reached available state');
        return;
    }

    // 6. Get snapshot details
    const getSnapRes = http.get(`${BASE_URL}/snapshots/${snapId}`, { headers: authHeaders });
    check(getSnapRes, { 'snapshot retrieved': (r) => r.status === 200 });

    // 7. List snapshots
    const listSnapRes = http.get(`${BASE_URL}/snapshots`, { headers: authHeaders });
    check(listSnapRes, { 'snapshots listed': (r) => r.status === 200 });

    // 8. Restore snapshot to new volume
    const restorePayload = JSON.stringify({
        new_volume_name: `restored-vol-${uniqueId}`,
    });
    const restoreRes = http.post(`${BASE_URL}/snapshots/${snapId}/restore`, restorePayload, { headers: authHeaders });
    check(restoreRes, { 'snapshot restored': (r) => r.status === 201 || r.status === 200 || r.status === 202 });

    // 9. Delete snapshot
    const delSnapRes = http.del(`${BASE_URL}/snapshots/${snapId}`, null, { headers: authHeaders });
    check(delSnapRes, { 'snapshot deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });

    // 10. Cleanup volumes and VPC
    sleep(2);
    http.del(`${BASE_URL}/volumes/${volId}`, null, { headers: authHeaders });
    http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });

    sleep(1);
}
