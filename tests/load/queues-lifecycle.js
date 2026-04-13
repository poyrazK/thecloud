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
    const queueName = `queue-${uniqueId}`;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `queuetest-${__VU}@loadtest.local`, 'Password123!', `QueueUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create queue
    const queuePayload = JSON.stringify({
        name: queueName,
        visibility_timeout: 30,
        retention_days: 7,
        max_message_size: 262144,
    });
    const queueRes = http.post(`${BASE_URL}/queues`, queuePayload, { headers: authHeaders });
    check(queueRes, { 'queue created': (r) => r.status === 201 || r.status === 200 });

    if (queueRes.status !== 201 && queueRes.status !== 200) {
        console.error(`Queue creation failed: ${queueRes.status} ${queueRes.body}`);
        sleep(1);
        return;
    }
    const queueId = queueRes.json('data.id');

    // 2. Get queue details
    const getQueueRes = http.get(`${BASE_URL}/queues/${queueId}`, { headers: authHeaders });
    check(getQueueRes, { 'queue retrieved': (r) => r.status === 200 });

    // 3. List queues
    const listQueuesRes = http.get(`${BASE_URL}/queues`, { headers: authHeaders });
    check(listQueuesRes, { 'queues listed': (r) => r.status === 200 });

    // 4. Delete queue
    const delQueueRes = http.del(`${BASE_URL}/queues/${queueId}`, null, { headers: authHeaders });
    check(delQueueRes, { 'queue deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });

    sleep(1);
}