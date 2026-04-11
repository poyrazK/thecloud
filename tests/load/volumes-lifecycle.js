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

function waitForVolume(authHeaders, volumeId, targetStatus = 'available') {
    for (let i = 0; i < 30; i++) {
        const getRes = http.get(`${BASE_URL}/volumes/${volumeId}`, { headers: authHeaders });
        if (getRes.status === 200 && getRes.json('data.status') === targetStatus) {
            return true;
        }
        sleep(1);
    }
    return false;
}

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const volumeName = `vol-${uniqueId}`;
    const sizeGb = 10;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `volumetest-${__VU}@loadtest.local`, 'Password123!', `VolumeUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create VPC (required for volume operations)
    const vpcPayload = JSON.stringify({ name: `vpc-vol-${uniqueId}`, cidr_block: '10.5.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'vpc created': (r) => r.status === 201 || r.status === 200 });
    if (vpcRes.status !== 201 && vpcRes.status !== 200) {
        console.error(`VPC creation failed: ${vpcRes.status} ${vpcRes.body}`);
        return;
    }
    const vpcId = vpcRes.json('data.id');

    // 2. Create volume
    const volPayload = JSON.stringify({
        name: volumeName,
        size_gb: sizeGb,
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
    const isAvailable = waitForVolume(authHeaders, volId, 'available');
    check(isAvailable, { 'volume is available': (val) => val === true });

    if (!isAvailable) {
        console.error(`Volume ${volId} never reached available state`);
        http.del(`${BASE_URL}/volumes/${volId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        fail('Volume never reached available state');
        return;
    }

    // 4. Create subnet and instance for attachment
    const subnetPayload = JSON.stringify({
        name: `subnet-vol-${uniqueId}`,
        vpc_id: vpcId,
        cidr_block: '10.5.1.0/24',
    });
    const subnetRes = http.post(`${BASE_URL}/vpcs/${vpcId}/subnets`, subnetPayload, { headers: authHeaders });
    check(subnetRes, { 'subnet created': (r) => r.status === 201 });
    if (subnetRes.status !== 201) {
        console.error(`Subnet creation failed: ${subnetRes.status} ${subnetRes.body}`);
        http.del(`${BASE_URL}/volumes/${volId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const subnetId = subnetRes.json('data.id');

    const instPayload = JSON.stringify({
        name: `inst-vol-${uniqueId}`,
        image: 'alpine:latest',
        vpc_id: vpcId,
        subnet_id: subnetId,
        ports: '80:80',
    });
    const instRes = http.post(`${BASE_URL}/instances`, instPayload, { headers: authHeaders });
    check(instRes, { 'instance launch accepted': (r) => r.status === 202 });
    if (instRes.status !== 202) {
        console.error(`Instance launch failed: ${instRes.status} ${instRes.body}`);
        http.del(`${BASE_URL}/volumes/${volId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const instId = instRes.json('data.id');

    // Wait for instance to be running
    let isInstanceRunning = false;
    for (let i = 0; i < 60; i++) {
        const getInstRes = http.get(`${BASE_URL}/instances/${instId}`, { headers: authHeaders });
        if (getInstRes.status === 200 && getInstRes.json('data.status') === 'running') {
            isInstanceRunning = true;
            break;
        }
        sleep(1);
    }
    check(isInstanceRunning, { 'instance is running': (val) => val === true });

    // 5. Attach volume to instance
    let attached = false;
    if (isInstanceRunning) {
        const attachPayload = JSON.stringify({
            instance_id: instId,
            mount_path: '/mnt/volumes',
        });
        const attachRes = http.post(`${BASE_URL}/volumes/${volId}/attach`, attachPayload, { headers: authHeaders });
        check(attachRes, { 'volume attached': (r) => r.status === 200 || r.status === 201 });
        attached = attachRes.status === 200 || attachRes.status === 201;
    }

    // 6. List volumes
    const listRes = http.get(`${BASE_URL}/volumes`, { headers: authHeaders });
    check(listRes, { 'volumes listed': (r) => r.status === 200 });

    // 7. Detach volume (if attached)
    if (attached) {
        const detachRes = http.post(`${BASE_URL}/volumes/${volId}/detach`, null, { headers: authHeaders });
        check(detachRes, { 'volume detached': (r) => r.status === 200 });
    }

    // 8. Delete volume
    const delVolRes = http.del(`${BASE_URL}/volumes/${volId}`, null, { headers: authHeaders });
    check(delVolRes, { 'volume deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });

    // 9. Cleanup instance and VPC
    http.del(`${BASE_URL}/instances/${instId}`, null, { headers: authHeaders });
    sleep(2); // Brief wait for instance deletion
    http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });

    sleep(1);
}
