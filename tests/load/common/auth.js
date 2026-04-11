// Shared authentication helpers for k6 load tests
import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL, makeHeaders } from './config.js';

// In-memory token cache per VU to avoid re-authenticating on every iteration
const tokenCache = {};

/**
 * Register a new user and return API key.
 * Uses a simple in-memory cache keyed by VU to avoid repeated registrations.
 */
export function getOrCreateApiKey(vuId, email, password, name) {
    if (tokenCache[vuId]) {
        return tokenCache[vuId];
    }

    const headers = { 'Content-Type': 'application/json' };

    // Try to register (may fail if user exists from prior run)
    const regPayload = JSON.stringify({ email, password, name });
    const regRes = http.post(`${BASE_URL}/auth/register`, regPayload, { headers });
    check(regRes, { 'auth-register success': (r) => r.status === 201 || r.status === 200 || r.status === 409 });

    // Try to login regardless (handles both new and existing users)
    const loginPayload = JSON.stringify({ email, password });
    const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers });

    if (loginRes.status !== 200) {
        console.error(`Login failed for ${email}: ${loginRes.status} ${loginRes.body}`);
        return null;
    }

    const apiKey = loginRes.json('data.api_key');
    const tenantId = loginRes.json('data.user.default_tenant_id');
    const authHeaders = makeHeaders(apiKey, { 'X-Tenant-ID': tenantId });
    tokenCache[vuId] = { apiKey, tenantId, authHeaders };
    return tokenCache[vuId];
}

// In-memory cache for registerAndLogin
const registerAndLoginCache = {};

/**
 * Register, login, and return full auth context.
 * Uses cache keyed by uniqueId to avoid repeated registrations.
 */
export function registerAndLogin(uniqueId) {
    if (registerAndLoginCache[uniqueId]) {
        return registerAndLoginCache[uniqueId];
    }

    const email = `user-${uniqueId}@loadtest.local`;
    const password = 'Password123!';
    const name = `User ${uniqueId}`;
    const headers = { 'Content-Type': 'application/json' };

    const regPayload = JSON.stringify({ email, password, name });
    const regRes = http.post(`${BASE_URL}/auth/register`, regPayload, { headers });
    check(regRes, { 'register success': (r) => r.status === 201 || r.status === 200 || r.status === 409 });

    const loginPayload = JSON.stringify({ email, password });
    const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers });
    check(loginRes, { 'login success': (r) => r.status === 200 });

    if (loginRes.status !== 200) {
        console.error(`registerAndLogin: login failed for ${email}: ${loginRes.status} ${loginRes.body}`);
        return null;
    }

    const apiKey = loginRes.json('data.api_key');
    const tenantId = loginRes.json('data.user.default_tenant_id');
    const authHeaders = makeHeaders(apiKey, { 'X-Tenant-ID': tenantId });

    const result = { email, password, apiKey, tenantId, authHeaders };
    registerAndLoginCache[uniqueId] = result;
    return result;
}

/**
 * Get API key via login (assumes user already exists).
 */
export function login(email, password) {
    const headers = { 'Content-Type': 'application/json' };
    const loginPayload = JSON.stringify({ email, password });
    const loginRes = http.post(`${BASE_URL}/auth/login`, loginPayload, { headers });

    if (loginRes.status !== 200) {
        return null;
    }

    const apiKey = loginRes.json('data.api_key');
    const tenantId = loginRes.json('data.tenant_id');
    const authHeaders = makeHeaders(apiKey, { 'X-Tenant-ID': tenantId });
    return {
        apiKey,
        tenantId,
        authHeaders,
    };
}

/**
 * Clear the token cache (useful for setup/teardown in test lifecycle).
 */
export function clearTokenCache() {
    Object.keys(tokenCache).forEach((key) => delete tokenCache[key]);
    Object.keys(registerAndLoginCache).forEach((key) => delete registerAndLoginCache[key]);
}
