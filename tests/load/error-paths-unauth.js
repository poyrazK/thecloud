import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL } from './common/config.js';

export const options = {
    stages: [
        { duration: '10s', target: 10 },
        { duration: '30s', target: 20 },
        { duration: '10s', target: 0 },
    ],
};

export default function () {
    // 1. Test without authentication (should get 401)
    const unauthRes = http.get(`${BASE_URL}/instances`);
    check(unauthRes, { 'unauthorized access rejected': (r) => r.status === 401 });

    // 2. Test with invalid API key (should get 401)
    const invalidKeyRes = http.get(`${BASE_URL}/instances`, {
        headers: { 'X-API-Key': 'invalid-key-12345' }
    });
    check(invalidKeyRes, { 'invalid API key rejected': (r) => r.status === 401 });

    // 3. Test with invalid UUID format (should get 400 or 404 before auth check)
    const invalidUuidRes = http.get(`${BASE_URL}/instances/not-a-uuid`, {
        headers: { 'X-API-Key': 'invalid-key-12345' }
    });
    check(invalidUuidRes, { 'invalid UUID rejected': (r) => r.status === 400 || r.status === 404 || r.status === 422 });

    // 4. Test wrong content type on auth endpoint
    const wrongContentRes = http.post(
        `${BASE_URL}/auth/register`,
        'not-json',
        { headers: { 'Content-Type': 'text/plain' } }
    );
    check(wrongContentRes, { 'wrong content type rejected': (r) => r.status === 400 || r.status === 415 || r.status === 422 });

    sleep(1);
}
