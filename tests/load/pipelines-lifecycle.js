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

function waitForBuild(authHeaders, buildId, targetStatus = 'SUCCEEDED') {
    for (let i = 0; i < 60; i++) {
        const getRes = http.get(`${BASE_URL}/pipelines/runs/${buildId}`, { headers: authHeaders });
        if (getRes.status === 200) {
            const status = getRes.json('data.status');
            if (status === targetStatus || status === 'FAILED' || status === 'CANCELED') {
                return status;
            }
        }
        sleep(1);
    }
    return null;
}

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const pipelineName = `pipeline-${uniqueId}`;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `pipelinetest-${__VU}@loadtest.local`, 'Password123!', `PipelineUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create pipeline
    const pipelinePayload = JSON.stringify({
        name: pipelineName,
        repository_url: `https://github.com/test/repo-${uniqueId}`,
        branch: 'main',
        config: {
            stages: [
                {
                    name: 'build',
                    steps: [
                        { name: 'compile', image: 'golang:1.22', commands: ['go build ./...'] },
                    ],
                },
            ],
            environment: {},
        },
    });
    const pipelineRes = http.post(`${BASE_URL}/pipelines`, pipelinePayload, { headers: authHeaders });
    check(pipelineRes, { 'pipeline created': (r) => r.status === 201 || r.status === 200 });

    if (pipelineRes.status !== 201 && pipelineRes.status !== 200) {
        console.error(`Pipeline creation failed: ${pipelineRes.status} ${pipelineRes.body}`);
        return;
    }
    const pipelineId = pipelineRes.json('data.id');

    // 2. Get pipeline details
    const getPipelineRes = http.get(`${BASE_URL}/pipelines/${pipelineId}`, { headers: authHeaders });
    check(getPipelineRes, { 'pipeline retrieved': (r) => r.status === 200 });

    // 3. List pipelines
    const listPipelinesRes = http.get(`${BASE_URL}/pipelines`, { headers: authHeaders });
    check(listPipelinesRes, { 'pipelines listed': (r) => r.status === 200 });

    // 4. Trigger build
    const triggerPayload = JSON.stringify({
        commit_hash: 'abc123',
        trigger_type: 'MANUAL',
    });
    const triggerRes = http.post(`${BASE_URL}/pipelines/${pipelineId}/runs`, triggerPayload, { headers: authHeaders });
    check(triggerRes, { 'build triggered': (r) => r.status === 201 || r.status === 200 });

    let buildStatus = null;
    if (triggerRes.status === 201 || triggerRes.status === 200) {
        const buildId = triggerRes.json('data.id');
        buildStatus = waitForBuild(authHeaders, buildId, 'SUCCEEDED');
        check(buildStatus === 'SUCCEEDED', { 'build succeeded': (val) => val === true });
    }

    // 5. Delete pipeline
    const delPipelineRes = http.del(`${BASE_URL}/pipelines/${pipelineId}`, null, { headers: authHeaders });
    check(delPipelineRes, { 'pipeline deleted': (r) => r.status === 200 || r.status === 204 || r.status === 202 });

    sleep(1);
}