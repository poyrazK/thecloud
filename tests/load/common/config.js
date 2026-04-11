// Shared configuration for k6 load tests
// All load test scripts should import from here instead of defining their own

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
export const API_KEY = __ENV.API_KEY || '';
export const DURATION = __ENV.DURATION || '1m';
export const VUS = parseInt(__ENV.VUS || '10', 10);

// Default thresholds for most tests
export const DEFAULT_THRESHOLDS = {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1000'],
};

// More relaxed thresholds for full-lifecycle tests
export const LIFECYCLE_THRESHOLDS = {
    http_req_failed: ['rate<0.10'],
    http_req_duration: ['p(95)<5000'],
};

// Strict thresholds for smoke/summary tests
// Note: http_req_failed is excluded because smoke tests often expect 401/503 responses
//       We rely on explicit checks instead
export const SMOKE_THRESHOLDS = {
    http_req_duration: ['p(95)<500'],
};

// Content types
export const CONTENT_TYPE_JSON = { 'Content-Type': 'application/json' };
export const CONTENT_TYPE_OCTET = { 'Content-Type': 'application/octet-stream' };

// Standard headers factory
export function makeHeaders(apiKey, extra = {}) {
    return {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey,
        ...extra,
    };
}
