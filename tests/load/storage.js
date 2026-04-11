import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { BASE_URL, CONTENT_TYPE_OCTET } from './common/config.js';
import { registerAndLogin } from './common/auth.js';
import { loadProfile } from './common/profiles.js';

export const options = {
    ...loadProfile,
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<2000'],
    },
};

export default function () {
    const uniqueId = uuidv4().substring(0, 8);
    const bucketName = `bucket-${uniqueId}`;
    const objectKey = `test-object-${uniqueId}.txt`;
    const objectContent = `Hello from k6 load test ${uniqueId}`;

    const { authHeaders } = registerAndLogin(uniqueId);

    // 1. Create bucket
    const createBucketRes = http.post(`${BASE_URL}/storage/buckets`,
        JSON.stringify({ name: bucketName, is_public: false }),
        { headers: authHeaders }
    );
    check(createBucketRes, { 'bucket created': (r) => r.status === 201 || r.status === 200 });

    if (createBucketRes.status !== 201 && createBucketRes.status !== 200) {
        console.error(`Bucket creation failed: ${createBucketRes.status} ${createBucketRes.body}`);
        return;
    }

    // 2. List buckets
    const listBucketsRes = http.get(`${BASE_URL}/storage/buckets`, { headers: authHeaders });
    check(listBucketsRes, { 'buckets listed': (r) => r.status === 200 });

    // 3. Upload object
    const uploadRes = http.put(
        `${BASE_URL}/storage/${bucketName}/${objectKey}`,
        objectContent,
        { headers: { ...authHeaders, ...CONTENT_TYPE_OCTET } }
    );
    check(uploadRes, { 'object uploaded': (r) => r.status === 200 || r.status === 201 });

    // 4. List objects in bucket
    const listObjsRes = http.get(`${BASE_URL}/storage/${bucketName}`, { headers: authHeaders });
    check(listObjsRes, { 'objects listed': (r) => r.status === 200 });

    // 5. Download object
    const downloadRes = http.get(`${BASE_URL}/storage/${bucketName}/${objectKey}`, { headers: authHeaders });
    check(downloadRes, { 'object downloaded': (r) => r.status === 200 });
    if (downloadRes.status === 200) {
        check(downloadRes, { 'object content matches': (r) => r.body === objectContent });
    }

    // 6. Delete object
    const deleteObjRes = http.del(`${BASE_URL}/storage/${bucketName}/${objectKey}`, null, { headers: authHeaders });
    check(deleteObjRes, { 'object deleted': (r) => r.status === 204 || r.status === 200 });

    // 7. Delete bucket (with force=true since we deleted object)
    const deleteBucketRes = http.del(`${BASE_URL}/storage/buckets/${bucketName}?force=true`, null, { headers: authHeaders });
    check(deleteBucketRes, { 'bucket deleted': (r) => r.status === 204 || r.status === 200 });

    sleep(1);
}
