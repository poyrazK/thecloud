import http from 'k6/http';
import { check, sleep } from 'k6';
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

function waitForInstance(apiKey, headers, instanceId) {
    for (let i = 0; i < 60; i++) {
        const getRes = http.get(`${BASE_URL}/instances/${instanceId}`, { headers });
        if (getRes.status === 200 && getRes.json('data.status') === 'running') {
            return true;
        }
        sleep(1);
    }
    return false;
}

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const lbName = `lb-${uniqueId}`;
    const vpcName = `vpc-lb-${uniqueId}`;
    const subnetName = `subnet-lb-${uniqueId}`;
    const instanceName = `inst-lb-${uniqueId}`;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `loadtest-${__VU}@loadtest.local`, 'Password123!', `Load User ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { apiKey, authHeaders } = auth;

    // 1. Create VPC
    const vpcPayload = JSON.stringify({ name: vpcName, cidr_block: '10.2.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'vpc created': (r) => r.status === 201 });
    if (vpcRes.status !== 201) {
        console.error(`VPC creation failed: ${vpcRes.status} ${vpcRes.body}`);
        return;
    }
    const vpcId = vpcRes.json('data.id');

    // 2. Create subnet
    const subnetPayload = JSON.stringify({
        name: subnetName,
        vpc_id: vpcId,
        cidr_block: '10.2.1.0/24'
    });
    const subnetRes = http.post(`${BASE_URL}/vpcs/${vpcId}/subnets`, subnetPayload, { headers: authHeaders });
    check(subnetRes, { 'subnet created': (r) => r.status === 201 });
    if (subnetRes.status !== 201) {
        console.error(`Subnet creation failed: ${subnetRes.status} ${subnetRes.body}`);
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const subnetId = subnetRes.json('data.id');

    // 3. Launch instance (needed as LB target)
    const instPayload = JSON.stringify({
        name: instanceName,
        image: 'alpine:latest',
        vpc_id: vpcId,
        subnet_id: subnetId,
        ports: '8080:80'
    });
    const instRes = http.post(`${BASE_URL}/instances`, instPayload, { headers: authHeaders });
    check(instRes, { 'instance launch accepted': (r) => r.status === 202 });
    if (instRes.status !== 202) {
        console.error(`Instance launch failed: ${instRes.status} ${instRes.body}`);
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const instId = instRes.json('data.id');

    // 4. Wait for instance to be running
    const isRunning = waitForInstance(apiKey, authHeaders, instId);
    check(isRunning, { 'instance is running for LB': (val) => val === true });
    if (!isRunning) {
        console.error(`Instance ${instId} never reached running state`);
        http.del(`${BASE_URL}/instances/${instId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }

    // 5. Create load balancer
    const lbPayload = JSON.stringify({
        name: lbName,
        vpc_id: vpcId,
        port: 80,
        algorithm: 'round-robin'
    });
    const lbRes = http.post(`${BASE_URL}/lb`, lbPayload, { headers: authHeaders });
    check(lbRes, { 'lb creation accepted': (r) => r.status === 202 });

    if (lbRes.status !== 202) {
        console.error(`LB creation failed: ${lbRes.status} ${lbRes.body}`);
        http.del(`${BASE_URL}/instances/${instId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const lbId = lbRes.json('data.id');

    // 6. Poll for LB to be ACTIVE
    let isActive = false;
    for (let i = 0; i < 30; i++) {
        const getLbRes = http.get(`${BASE_URL}/lb/${lbId}`, { headers: authHeaders });
        if (getLbRes.status === 200) {
            const status = getLbRes.json('data.status');
            if (status === 'active') {
                isActive = true;
                break;
            }
            if (status === 'failed') {
                console.error(`LB failed: ${getLbRes.body}`);
                break;
            }
        }
        sleep(1);
    }
    check(isActive, { 'lb is active': (val) => val === true });

    if (!isActive) {
        console.error(`LB ${lbId} never reached active state`);
        http.del(`${BASE_URL}/lb/${lbId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/instances/${instId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }

    // 7. Add target to LB
    const targetPayload = JSON.stringify({
        instance_id: instId,
        port: 8080,
        weight: 1
    });
    const addTargetRes = http.post(`${BASE_URL}/lb/${lbId}/targets`, targetPayload, { headers: authHeaders });
    check(addTargetRes, { 'target added': (r) => r.status === 201 || r.status === 200 });

    // 8. List LB targets
    const listTargetsRes = http.get(`${BASE_URL}/lb/${lbId}/targets`, { headers: authHeaders });
    check(listTargetsRes, { 'targets listed': (r) => r.status === 200 });

    // 9. List all LBs
    const listLbRes = http.get(`${BASE_URL}/lb`, { headers: authHeaders });
    check(listLbRes, { 'lbs listed': (r) => r.status === 200 });

    // 10. Remove target
    const removeTargetRes = http.del(`${BASE_URL}/lb/${lbId}/targets/${instId}`, null, { headers: authHeaders });
    check(removeTargetRes, { 'target removed': (r) => r.status === 200 || r.status === 204 });

    // 11. Delete LB
    const deleteLbRes = http.del(`${BASE_URL}/lb/${lbId}`, null, { headers: authHeaders });
    check(deleteLbRes, { 'lb deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });

    // 12. Cleanup instance and VPC
    sleep(2); // Brief wait for LB deletion to complete
    http.del(`${BASE_URL}/instances/${instId}`, null, { headers: authHeaders });
    http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });

    sleep(1);
}
