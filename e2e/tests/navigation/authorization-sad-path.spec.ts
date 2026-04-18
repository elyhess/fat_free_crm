import { test, expect } from '../../fixtures/auth';
import { ApiClient } from '../../fixtures/api-client';

const BASE = 'http://localhost:3000/api/v1';

/** Helper: make an authenticated fetch using a raw token. */
function authedFetch(
  path: string,
  token: string,
  init: RequestInit = {},
): Promise<Response> {
  return fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      ...(init.headers as Record<string, string> ?? {}),
    },
  });
}

test.describe('Authorization sad paths', () => {
  // ---------------------------------------------------------------------------
  // JWT / Token manipulation
  // ---------------------------------------------------------------------------

  test.describe('JWT/Token manipulation', () => {
    test('expired or invalid token returns 401', async () => {
      const res = await fetch(`${BASE}/accounts`, {
        headers: { Authorization: 'Bearer invalid_token_here' },
      });
      expect(res.status).toBe(401);
    });

    test('no token on protected endpoint returns 401', async () => {
      const res = await fetch(`${BASE}/accounts`);
      expect(res.status).toBe(401);
    });

    test('malformed JWT returns 401', async () => {
      const res = await fetch(`${BASE}/accounts`, {
        headers: { Authorization: 'Bearer not.a.jwt' },
      });
      expect(res.status).toBe(401);
    });
  });

  // ---------------------------------------------------------------------------
  // Non-admin accessing admin API endpoints
  // ---------------------------------------------------------------------------

  test.describe('Non-admin accessing admin API endpoints', () => {
    let nonAdminToken: string;

    test.beforeAll(async () => {
      const { token: adminToken } = await ApiClient.login('admin', 'Dem0P@ssword!!');
      const adminClient = new ApiClient(adminToken);
      const { username, password } = await adminClient.ensureNonAdminUser();
      const { token } = await ApiClient.login(username, password);
      nonAdminToken = token;
    });

    test('non-admin POST to admin users endpoint returns 403', async () => {
      const res = await authedFetch('/admin/users', nonAdminToken, {
        method: 'POST',
        body: JSON.stringify({
          username: 'should_not_create',
          email: 'should_not@example.com',
          password: 'Dem0P@ssword!!',
        }),
      });
      expect(res.status).toBe(403);
    });

    test('non-admin GET admin settings returns 403', async () => {
      const res = await authedFetch('/admin/settings', nonAdminToken);
      expect(res.status).toBe(403);
    });

    test('non-admin PUT admin settings returns 403', async () => {
      const res = await authedFetch('/admin/settings', nonAdminToken, {
        method: 'PUT',
        body: JSON.stringify({ base_url: 'http://evil.example.com' }),
      });
      expect(res.status).toBe(403);
    });

    test('non-admin POST admin fields returns 403', async () => {
      const res = await authedFetch('/admin/fields', nonAdminToken, {
        method: 'POST',
        body: JSON.stringify({
          field_group_id: 1,
          label: 'Hacked Field',
          as: 'string',
        }),
      });
      expect(res.status).toBe(403);
    });
  });

  // ---------------------------------------------------------------------------
  // Entity access control
  // ---------------------------------------------------------------------------

  test.describe('Entity access control', () => {
    let adminApi: ApiClient;
    let nonAdminToken: string;

    test.beforeAll(async () => {
      const { token: aToken } = await ApiClient.login('admin', 'Dem0P@ssword!!');
      adminApi = new ApiClient(aToken);
      const { username, password } = await adminApi.ensureNonAdminUser();
      const { token } = await ApiClient.login(username, password);
      nonAdminToken = token;
    });

    test('private account not visible to non-admin in list', async () => {
      const uniqueName = `AuthzPrivate ${Date.now()}`;
      const account = await adminApi.createAccount({ name: uniqueName, access: 'Private' });
      const id = (account as Record<string, unknown>).id as number;

      try {
        const res = await authedFetch('/accounts', nonAdminToken);
        expect(res.status).toBe(200);
        const body = await res.json();
        const accounts = Array.isArray(body) ? body : (body.accounts ?? body.data ?? []);
        const found = accounts.some((a: Record<string, unknown>) => a.name === uniqueName);
        expect(found).toBe(false);
      } finally {
        await adminApi.deleteEntity('accounts', id);
      }
    });

    test('non-admin cannot edit admin private account', async () => {
      const account = await adminApi.createAccount({
        name: `AuthzEditBlock ${Date.now()}`,
        access: 'Private',
      });
      const id = (account as Record<string, unknown>).id as number;

      try {
        const res = await authedFetch(`/accounts/${id}`, nonAdminToken, {
          method: 'PUT',
          body: JSON.stringify({ name: 'Hacked Name' }),
        });
        // Expect 403 (forbidden) or 404 (entity hidden from non-owner)
        expect([403, 404]).toContain(res.status);
      } finally {
        await adminApi.deleteEntity('accounts', id);
      }
    });

    test('non-admin cannot delete admin private account', async () => {
      const account = await adminApi.createAccount({
        name: `AuthzDeleteBlock ${Date.now()}`,
        access: 'Private',
      });
      const id = (account as Record<string, unknown>).id as number;

      try {
        const res = await authedFetch(`/accounts/${id}`, nonAdminToken, {
          method: 'DELETE',
        });
        // Expect 403 (forbidden) or 404 (entity hidden from non-owner)
        expect([403, 404]).toContain(res.status);
      } finally {
        await adminApi.deleteEntity('accounts', id);
      }
    });
  });

  // ---------------------------------------------------------------------------
  // UI authorization - non-admin accessing admin pages
  // ---------------------------------------------------------------------------

  test.describe('UI authorization for admin pages', () => {
    test('non-admin navigating to /admin/settings sees access denied', async ({ demoPage }) => {
      await demoPage.goto('/admin/settings');
      await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
    });

    test('non-admin navigating to /admin/fields sees access denied', async ({ demoPage }) => {
      await demoPage.goto('/admin/fields');
      await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
    });

    test('non-admin navigating to /admin/research sees access denied', async ({ demoPage }) => {
      await demoPage.goto('/admin/research-tools');
      await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
    });
  });
});
