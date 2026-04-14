// Shared stage profiles for k6 load tests
// Use these in your script's options instead of hardcoding stages

// Quick smoke test profile (1-2 minutes)
export const smokeProfile = {
    stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 10 },
        { duration: '30s', target: 0 },
    ],
};

// Standard load test profile (5-10 minutes)
export const loadProfile = {
    stages: [
        { duration: '30s', target: 20 },
        { duration: '1m', target: 50 },
        { duration: '2m', target: 50 },
        { duration: '30s', target: 0 },
    ],
};

// Heavy load test profile (10+ minutes)
export const stressProfile = {
    stages: [
        { duration: '1m', target: 50 },
        { duration: '2m', target: 100 },
        { duration: '2m', target: 200 },
        { duration: '2m', target: 300 },
        { duration: '1m', target: 0 },
    ],
};

// Realistic user journey profile
export const lifecycleProfile = {
    stages: [
        { duration: '30s', target: 20 },
        { duration: '1m', target: 100 },
        { duration: '2m', target: 100 },
        { duration: '30s', target: 0 },
    ],
};

// Scalability test profile (9+ minutes)
export const scalabilityProfile = {
    stages: [
        { duration: '1m', target: 50 },
        { duration: '2m', target: 100 },
        { duration: '2m', target: 200 },
        { duration: '2m', target: 300 },
        { duration: '2m', target: 0 },
    ],
};

// Soak test profile (long duration)
export const soakProfile = {
    stages: [
        { duration: '2m', target: 50 },
        { duration: '4h', target: 50 },
        { duration: '2m', target: 0 },
    ],
};

// CI-optimized soak test (shortened)
export const soakProfileCI = {
    stages: [
        { duration: '1m', target: 50 },
        { duration: '5m', target: 50 },
        { duration: '1m', target: 0 },
    ],
};

// Burst profile for rate limit testing
export const burstProfile = {
    stages: [
        { duration: '5s', target: 100 },
        { duration: '30s', target: 100 },
        { duration: '5s', target: 0 },
    ],
};
