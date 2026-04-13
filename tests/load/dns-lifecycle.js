import http from 'k6/http';
import { check, sleep, fail } from 'k6';
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
    const zoneName = `zone-${uniqueId}.com`;
    const recordName = `www.${zoneName}`;

    // Use cached auth to avoid rate limiting
    const auth = getOrCreateApiKey(__VU, `dns-${__VU}@loadtest.local`, 'Password123!', `DNSUser ${__VU}`);
    if (!auth || !auth.apiKey) {
        sleep(1);
        return;
    }
    const { authHeaders } = auth;

    // 1. Create VPC (required for DNS zone)
    const vpcPayload = JSON.stringify({ name: `vpc-dns-${uniqueId}`, cidr_block: '10.10.0.0/16' });
    const vpcRes = http.post(`${BASE_URL}/vpcs`, vpcPayload, { headers: authHeaders });
    check(vpcRes, { 'vpc created': (r) => r.status === 201 || r.status === 200 });
    if (vpcRes.status !== 201 && vpcRes.status !== 200) {
        console.error(`VPC creation failed: ${vpcRes.status} ${vpcRes.body}`);
        sleep(1);
        return;
    }
    const vpcId = vpcRes.json('data.id');

    // 2. Create DNS zone
    const zonePayload = JSON.stringify({
        name: zoneName,
        description: `Test DNS zone ${uniqueId}`,
        vpc_id: vpcId,
    });
    const zoneRes = http.post(`${BASE_URL}/dns/zones`, zonePayload, { headers: authHeaders });
    check(zoneRes, { 'zone created': (r) => r.status === 201 || r.status === 200 });

    if (zoneRes.status !== 201 && zoneRes.status !== 200) {
        console.error(`Zone creation failed: ${zoneRes.status} ${zoneRes.body}`);
        http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });
        sleep(1);
        return;
    }
    const zoneId = zoneRes.json('data.id');

    // 3. List zones
    const listZonesRes = http.get(`${BASE_URL}/dns/zones`, { headers: authHeaders });
    check(listZonesRes, { 'zones listed': (r) => r.status === 200 });

    // 4. Get zone details
    const getZoneRes = http.get(`${BASE_URL}/dns/zones/${zoneId}`, { headers: authHeaders });
    check(getZoneRes, { 'zone retrieved': (r) => r.status === 200 });

    // 5. Create DNS record (A type)
    const recordPayload = JSON.stringify({
        name: recordName,
        type: 'A',
        content: '192.168.1.1',
        ttl: 3600,
    });
    const recordRes = http.post(`${BASE_URL}/dns/zones/${zoneId}/records`, recordPayload, { headers: authHeaders });
    check(recordRes, { 'record created': (r) => r.status === 201 || r.status === 200 });

    let recordId = null;
    if (recordRes.status === 201 || recordRes.status === 200) {
        recordId = recordRes.json('data.id');

        // 6. List records in zone
        const listRecordsRes = http.get(`${BASE_URL}/dns/zones/${zoneId}/records`, { headers: authHeaders });
        check(listRecordsRes, { 'records listed': (r) => r.status === 200 });

        // 7. Get record details
        const getRecordRes = http.get(`${BASE_URL}/dns/records/${recordId}`, { headers: authHeaders });
        check(getRecordRes, { 'record retrieved': (r) => r.status === 200 });

        // 8. Update record
        const updatePayload = JSON.stringify({
            content: '192.168.1.2',
            ttl: 7200,
        });
        const updateRes = http.put(`${BASE_URL}/dns/records/${recordId}`, updatePayload, { headers: authHeaders });
        check(updateRes, { 'record updated': (r) => r.status === 200 });

        // 9. Delete record
        const deleteRecordRes = http.del(`${BASE_URL}/dns/records/${recordId}`, null, { headers: authHeaders });
        check(deleteRecordRes, { 'record deleted': (r) => r.status === 200 || r.status === 204 || r.status === 202 });
    }

    // 10. Delete zone
    const deleteZoneRes = http.del(`${BASE_URL}/dns/zones/${zoneId}`, null, { headers: authHeaders });
    check(deleteZoneRes, { 'zone deleted': (r) => r.status === 200 || r.status === 204 || r.status === 202 });

    // 11. Cleanup VPC
    sleep(2);
    http.del(`${BASE_URL}/vpcs/${vpcId}`, null, { headers: authHeaders });

    sleep(1);
}