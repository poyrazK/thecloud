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

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `imagetest-${__VU}@loadtest.local`, 'Password123!', `ImageUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Register image metadata
    const imagePayload = JSON.stringify({
        name: `img-${uniqueId}`,
        description: `Test image ${uniqueId}`,
        os: 'ubuntu',
        version: '22.04',
        is_public: false,
    });
    const imageRes = http.post(`${BASE_URL}/images`, imagePayload, { headers: authHeaders });
    check(imageRes, { 'image registered': (r) => r.status === 201 || r.status === 200 });

    if (imageRes.status !== 201 && imageRes.status !== 200) {
        console.error(`Image registration failed: ${imageRes.status} ${imageRes.body}`);
        return;
    }
    const imageId = imageRes.json('data.id');

    // 2. Get image details
    const getImageRes = http.get(`${BASE_URL}/images/${imageId}`, { headers: authHeaders });
    check(getImageRes, { 'image retrieved': (r) => r.status === 200 });

    // 3. List images
    const listImagesRes = http.get(`${BASE_URL}/images`, { headers: authHeaders });
    check(listImagesRes, { 'images listed': (r) => r.status === 200 });

    // 4. Delete image
    const delImageRes = http.del(`${BASE_URL}/images/${imageId}`, null, { headers: authHeaders });
    check(delImageRes, { 'image deleted': (r) => r.status === 200 || r.status === 202 || r.status === 204 });

    sleep(1);
}
