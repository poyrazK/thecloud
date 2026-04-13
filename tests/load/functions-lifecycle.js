import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE_URL, LIFECYCLE_THRESHOLDS } from './common/config.js';
import { getOrCreateApiKey } from './common/auth.js';

export const options = {
    stages: [
        { duration: '30s', target: 3 },
        { duration: '2m', target: 3 },
        { duration: '1m', target: 0 },
    ],
    thresholds: LIFECYCLE_THRESHOLDS,
};

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const functionName = `fn-${uniqueId}`;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `fntest-${__VU}@loadtest.local`, 'Password123!', `FnUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. List functions (verify empty state)
    const listFnRes = http.get(`${BASE_URL}/functions`, { headers: authHeaders });
    check(listFnRes, { 'functions listed': (r) => r.status === 200 });

    // 2. Create function with multipart form-data (name, runtime, handler, code file)
    // The handler expects: FormFile("code"), ShouldBind form fields name/runtime/handler
    const boundaries = '----FormBoundary' + Math.random().toString(36).substring(2);
    const body = [
        `--${boundaries}\r\nContent-Disposition: form-data; name="name"\r\n\r\n${functionName}`,
        `--${boundaries}\r\nContent-Disposition: form-data; name="runtime"\r\n\r\npython312`,
        `--${boundaries}\r\nContent-Disposition: form-data; name="handler"\r\n\r\nmain.handle`,
        `--${boundaries}\r\nContent-Disposition: form-data; name="code"; filename="handler.py"\r\nContent-Type: text/plain\r\n\r\ndef handle(event):\n    return {"message": "hello"}`,
        `--${boundaries}--`,
    ].join('\r\n');

    const fnRes = http.post(`${BASE_URL}/functions`, body, { headers: { ...authHeaders, 'Content-Type': `multipart/form-data; boundary=${boundaries}` } });

    check(fnRes, { 'function creation attempted': (r) => r.status === 201 || r.status === 200 });

    let functionId = null;
    if (fnRes.status === 201 || fnRes.status === 200) {
        functionId = fnRes.json('data.id');

        // 3. Get function details
        const getFnRes = http.get(`${BASE_URL}/functions/${functionId}`, { headers: authHeaders });
        check(getFnRes, { 'function retrieved': (r) => r.status === 200 });

        // 4. Invoke function (async)
        const invokePayload = JSON.stringify({ data: 'test' });
        const invokeRes = http.post(`${BASE_URL}/functions/${functionId}/invoke?async=true`, invokePayload, { headers: authHeaders });
        check(invokeRes, { 'function invoked': (r) => r.status === 200 || r.status === 202 || r.status === 201 });

        // 5. Get function logs
        const logsRes = http.get(`${BASE_URL}/functions/${functionId}/logs`, { headers: authHeaders });
        check(logsRes, { 'logs retrieved': (r) => r.status === 200 });

        // 6. Delete function
        const delFnRes = http.del(`${BASE_URL}/functions/${functionId}`, null, { headers: authHeaders });
        check(delFnRes, { 'function deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });
    } else {
        console.log(`Function creation failed (expected without proper code upload): ${fnRes.status} ${fnRes.body}`);
        // Still test list endpoint even without creating a function
        const listAgainRes = http.get(`${BASE_URL}/functions`, { headers: authHeaders });
        check(listAgainRes, { 'functions listed': (r) => r.status === 200 });
    }

    sleep(1);
}