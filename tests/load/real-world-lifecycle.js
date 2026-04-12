import http from 'k6/http';
import { check, sleep, fail } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE_URL, LIFECYCLE_THRESHOLDS } from './common/config.js';
import { registerAndLogin } from './common/auth.js';
import { lifecycleProfile } from './common/profiles.js';

export const options = {
    ...lifecycleProfile,
    thresholds: LIFECYCLE_THRESHOLDS,
};

function waitForInstance(authHeaders, instanceId) {
    for (let i = 0; i < 60; i++) {
        const getRes = http.get(`${BASE_URL}/instances/${instanceId}`, { headers: authHeaders });
        if (getRes.status === 200 && getRes.json('data.status') === 'running') {
            return true;
        }
        sleep(1);
    }
    return false;
}

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const { authHeaders } = registerAndLogin(uniqueId);

    // 1. CREATE VPC
    const vpcPayload = JSON.stringify({ name: `vpc-${uniqueId}`, cidr_block: '10.0.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'vpc created': (r) => r.status === 201 });
    if (vpcRes.status !== 201) {
        console.error(`VPC Creation Failed: ${vpcRes.status} ${vpcRes.body}`);
        return;
    }
    const vpcId = vpcRes.json('data.id');

    // 2. CREATE SUBNET
    const subnetPayload = JSON.stringify({
        name: `subnet-${uniqueId}`,
        vpc_id: vpcId,
        cidr_block: '10.0.1.0/24'
    });
    const subnetRes = http.post(`${BASE_URL}/vpcs/${vpcId}/subnets`, subnetPayload, { headers: authHeaders });
    check(subnetRes, { 'subnet created': (r) => r.status === 201 });
    if (subnetRes.status !== 201) return;
    const subnetId = subnetRes.json('data.id');

    // 3. LAUNCH INSTANCE
    const instPayload = JSON.stringify({
        name: `inst-${uniqueId}`,
        image: 'alpine:latest',
        vpc_id: vpcId,
        subnet_id: subnetId,
        ports: '80:80'
    });
    const instRes = http.post(`${BASE_URL}/instances`, instPayload, { headers: authHeaders });
    check(instRes, { 'instance launch accepted': (r) => r.status === 202 });
    if (instRes.status !== 202) return;
    const instId = instRes.json('data.id');

    // 4. WAIT FOR RUNNING
    const isRunning = waitForInstance(authHeaders, instId);
    check(isRunning, { 'instance is running': (val) => val === true });

    if (!isRunning) {
        console.error(`Instance ${instId} never reached running state`);
        fail('Instance never reached running state');
    }

    // 5. GET STATS
    if (isRunning) {
        const statsRes = http.get(`${BASE_URL}/instances/${instId}/stats`, { headers: authHeaders });
        check(statsRes, { 'stats retrieved': (r) => r.status === 200 });
    }

    // 6. CLEANUP
    const delInstRes = http.del(`${BASE_URL}/instances/${instId}`, null, { headers: authHeaders });
    check(delInstRes, { 'inst deleted': (r) => r.status === 204 || r.status === 200 });

    const delVpcRes = http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
    check(delVpcRes, { 'vpc deleted': (r) => r.status === 204 || r.status === 200 });

    sleep(1);
}
