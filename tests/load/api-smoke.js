import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    thresholds: {
        http_req_failed: ['rate<0.01'], // http errors should be less than 1%
        http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    },
};

export default function smokeTest() {
    const url = 'http://localhost:8080/health'; // This assumes the app is running locally or in a container
    const res = http.get(url);

    check(res, {
        'status is 200': (r) => r.status === 200,
    });

    sleep(1);
}
