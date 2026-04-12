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

function waitForCluster(authHeaders, clusterId, targetStatus = 'running') {
    for (let i = 0; i < 60; i++) {
        const getRes = http.get(`${BASE_URL}/clusters/${clusterId}`, { headers: authHeaders });
        if (getRes.status === 200 && getRes.json('data.status') === targetStatus) {
            return true;
        }
        sleep(1);
    }
    return false;
}

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const clusterName = `cluster-${uniqueId}`;
    const k8sVersion = '1.28';

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `clustertest-${__VU}@loadtest.local`, 'Password123!', `ClusterUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create VPC (clusters need a VPC)
    const vpcPayload = JSON.stringify({ name: `vpc-cluster-${uniqueId}`, cidr_block: '10.8.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'vpc created': (r) => r.status === 201 || r.status === 200 });
    if (vpcRes.status !== 201 && vpcRes.status !== 200) {
        console.error(`VPC creation failed: ${vpcRes.status} ${vpcRes.body}`);
        return;
    }
    const vpcId = vpcRes.json('data.id');

    // 2. Create cluster
    const clusterPayload = JSON.stringify({
        name: clusterName,
        vpc_id: vpcId,
        version: k8sVersion,
        workers: 2,
    });
    const clusterRes = http.post(`${BASE_URL}/clusters`, clusterPayload, { headers: authHeaders });
    check(clusterRes, { 'cluster creation accepted': (r) => r.status === 202 });

    if (clusterRes.status !== 202) {
        console.error(`Cluster creation failed: ${clusterRes.status} ${clusterRes.body}`);
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        return;
    }
    const clusterId = clusterRes.json('data.id');

    // 3. Poll for cluster to be running
    const isRunning = waitForCluster(authHeaders, clusterId, 'running');
    check(isRunning, { 'cluster is running': (val) => val === true });

    if (!isRunning) {
        console.error(`Cluster ${clusterId} never reached running state`);
        http.del(`${BASE_URL}/clusters/${clusterId}`, null, { headers: authHeaders });
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        fail('Cluster never reached running state');
        return;
    }

    // 4. Get kubeconfig
    const kubeRes = http.get(`${BASE_URL}/clusters/${clusterId}/kubeconfig`, { headers: authHeaders });
    check(kubeRes, { 'kubeconfig retrieved': (r) => r.status === 200 });

    // 5. Get cluster health
    const healthRes = http.get(`${BASE_URL}/clusters/${clusterId}/health`, { headers: authHeaders });
    check(healthRes, { 'cluster health retrieved': (r) => r.status === 200 });

    // 6. List clusters
    const listRes = http.get(`${BASE_URL}/clusters`, { headers: authHeaders });
    check(listRes, { 'clusters listed': (r) => r.status === 200 });

    // 7. Scale cluster
    const scalePayload = JSON.stringify({ workers: 3 });
    const scaleRes = http.put(`${BASE_URL}/clusters/${clusterId}/scale`, scalePayload, { headers: authHeaders });
    check(scaleRes, { 'cluster scaled': (r) => r.status === 200 });

    // 8. Add node group
    const ngPayload = JSON.stringify({
        name: `ng-${uniqueId}`,
        instance_type: 'standard-2',
        min_size: 1,
        max_size: 3,
        desired_size: 1,
    });
    const ngRes = http.post(`${BASE_URL}/clusters/${clusterId}/nodegroups`, ngPayload, { headers: authHeaders });
    check(ngRes, { 'node group added': (r) => r.status === 201 || r.status === 200 });
    const ngName = ngRes.status === 201 || ngRes.status === 200 ? ngRes.json('data.name') : null;

    // 9. Update node group
    if (ngName) {
        const updateNgPayload = JSON.stringify({ min_size: 2, max_size: 4, desired_size: 2 });
        const updateNgRes = http.put(`${BASE_URL}/clusters/${clusterId}/nodegroups/${ngName}`, updateNgPayload, { headers: authHeaders });
        check(updateNgRes, { 'node group updated': (r) => r.status === 200 });
    }

    // 10. Delete node group
    if (ngName) {
        const deleteNgRes = http.del(`${BASE_URL}/clusters/${clusterId}/nodegroups/${ngName}`, null, { headers: authHeaders });
        check(deleteNgRes, { 'node group deleted': (r) => r.status === 202 || r.status === 204 || r.status === 200 });
    }

    // 11. Delete cluster
    const delRes = http.del(`${BASE_URL}/clusters/${clusterId}`, null, { headers: authHeaders });
    check(delRes, { 'cluster deleted': (r) => r.status === 202 || r.status === 200 });

    // 12. Cleanup VPC
    sleep(2);
    http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });

    sleep(1);
}
