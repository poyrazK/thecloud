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

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const sgName = `sg-${uniqueId}`;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `sgtest-${__VU}@loadtest.local`, 'Password123!', `SGUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create VPC
    const vpcPayload = JSON.stringify({ name: `vpc-sg-${uniqueId}`, cidr_block: '10.6.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'vpc created': (r) => r.status === 201 || r.status === 200 });
    if (vpcRes.status !== 201 && vpcRes.status !== 200) {
        console.error(`VPC creation failed: ${vpcRes.status} ${vpcRes.body}`);
        return;
    }
    const vpcId = vpcRes.json('data.id');

    // 2. Create security group
    const sgPayload = JSON.stringify({
        vpc_id: vpcId,
        name: sgName,
        description: `Test security group ${uniqueId}`,
    });
    const sgRes = http.post(`${BASE_URL}/security-groups`, sgPayload, { headers: authHeaders });
    check(sgRes, { 'security group created': (r) => r.status === 201 });

    if (sgRes.status !== 201) {
        console.error(`Security group creation failed: ${sgRes.status} ${sgRes.body}`);
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const sgId = sgRes.json('data.id');

    // 3. Create subnet and instance for attachment
    const subnetPayload = JSON.stringify({
        name: `subnet-sg-${uniqueId}`,
        vpc_id: vpcId,
        cidr_block: '10.6.1.0/24',
    });
    const subnetRes = http.post(`${BASE_URL}/vpcs/${vpcId}/subnets`, subnetPayload, { headers: authHeaders });
    check(subnetRes, { 'subnet created': (r) => r.status === 201 });
    if (subnetRes.status !== 201) {
        console.error(`Subnet creation failed: ${subnetRes.status} ${subnetRes.body}`);
        http.del(`${BASE_URL}/security-groups/${sgId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const subnetId = subnetRes.json('data.id');

    const instPayload = JSON.stringify({
        name: `inst-sg-${uniqueId}`,
        image: 'alpine:latest',
        vpc_id: vpcId,
        subnet_id: subnetId,
        ports: '22:22',
    });
    const instRes = http.post(`${BASE_URL}/instances`, instPayload, { headers: authHeaders });
    check(instRes, { 'instance launch accepted': (r) => r.status === 202 });
    if (instRes.status !== 202) {
        console.error(`Instance launch failed: ${instRes.status} ${instRes.body}`);
        http.del(`${BASE_URL}/security-groups/${sgId}`, null, { headers: authHeaders });
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

    // 4. Add rules to security group
    const rulePayload = JSON.stringify({
        direction: 'ingress',
        protocol: 'tcp',
        port: 22,
        cidr: '0.0.0.0/0',
    });
    const ruleRes = http.post(`${BASE_URL}/security-groups/${sgId}/rules`, rulePayload, { headers: authHeaders });
    check(ruleRes, { 'rule added': (r) => r.status === 201 });
    let ruleId = null;
    if (ruleRes.status === 201) {
        ruleId = ruleRes.json('data.id');
    }

    // 5. List security groups for VPC
    const listRes = http.get(`${BASE_URL}/security-groups?vpc_id=${vpcId}`, { headers: authHeaders });
    check(listRes, { 'security groups listed': (r) => r.status === 200 });

    // 6. Get security group details
    const getRes = http.get(`${BASE_URL}/security-groups/${sgId}`, { headers: authHeaders });
    check(getRes, { 'security group retrieved': (r) => r.status === 200 });

    // 7. Attach security group to instance
    if (isInstanceRunning) {
        const attachPayload = JSON.stringify({
            instance_id: instId,
            group_id: sgId,
        });
        const attachRes = http.post(`${BASE_URL}/security-groups/attach`, attachPayload, { headers: authHeaders });
        check(attachRes, { 'security group attached': (r) => r.status === 200 });
    }

    // 8. Remove rule (if created)
    if (ruleId) {
        const removeRuleRes = http.del(`${BASE_URL}/security-groups/rules/${ruleId}`, null, { headers: authHeaders });
        check(removeRuleRes, { 'rule removed': (r) => r.status === 204 });
    }

    // 9. Cleanup - delete instance and VPC
    http.del(`${BASE_URL}/instances/${instId}`, null, { headers: authHeaders });
    sleep(2);
    http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });

    sleep(1);
}
