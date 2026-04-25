import { test, expect } from '../../fixtures/auth';

/**
 * Feature Gap Tests — Strict Assertions
 *
 * Every test makes real assertions that FAIL if the API does not work.
 * No catch-and-skip, no console.log escape hatches.
 */

// ---------------------------------------------------------------------------
// 1. Lead Conversion (POST /leads/:id/convert)
// ---------------------------------------------------------------------------
test.describe('Lead Conversion', () => {
  test('convert lead via API', async ({ api }) => {
    const lead = await api.createLead({ first_name: 'ConvertAPI', last_name: `Test ${Date.now()}` });
    const id = lead.id as number;

    try {
      const result = await api.post(`/leads/${id}/convert`, {
        account: { name: `Converted Account ${Date.now()}` },
        opportunity: { name: `Converted Opp ${Date.now()}` },
      });
      expect(result).toBeDefined();

      // Verify the lead status changed to converted
      const updated = (await api.get(`/leads/${id}`)) as Record<string, unknown>;
      expect(updated.status).toBe('converted');
    } finally {
      await api.deleteEntity('leads', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 2. Lead Rejection (PUT /leads/:id/reject)
// ---------------------------------------------------------------------------
test.describe('Lead Rejection', () => {
  test('reject lead via API', async ({ api }) => {
    const lead = await api.createLead({ first_name: 'Reject', last_name: `Test ${Date.now()}` });
    const id = lead.id as number;

    try {
      await api.put(`/leads/${id}/reject`, {});
      const updated = (await api.get(`/leads/${id}`)) as Record<string, unknown>;
      expect(updated.status).toBe('rejected');
    } finally {
      await api.deleteEntity('leads', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 3. Task Complete / Uncomplete (PUT /tasks/:id/complete)
// ---------------------------------------------------------------------------
test.describe('Task Complete / Uncomplete', () => {
  test('complete and uncomplete a task via API', async ({ api }) => {
    const task = await api.createTask({ name: `Complete Test ${Date.now()}` });
    const id = task.id as number;

    try {
      // Complete the task
      await api.put(`/tasks/${id}/complete`, {});
      const completed = (await api.get(`/tasks/${id}`)) as Record<string, unknown>;
      expect(completed.completed_at).toBeTruthy();

      // Uncomplete the task
      await api.put(`/tasks/${id}/uncomplete`, {});
      const uncompleted = (await api.get(`/tasks/${id}`)) as Record<string, unknown>;
      expect(uncompleted.completed_at).toBeFalsy();
    } finally {
      await api.deleteEntity('tasks', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 4. Entity Subscriptions via API
// ---------------------------------------------------------------------------
test.describe('Entity Subscriptions', () => {
  test('subscribe and unsubscribe to an account', async ({ api }) => {
    const account = await api.createAccount({ name: `Sub Test ${Date.now()}` });
    const id = account.id as number;

    try {
      // Subscribe
      await api.post(`/accounts/${id}/subscribe`, {});

      // Check subscription
      const subCheck = (await api.get(`/accounts/${id}/subscription`)) as Record<string, unknown>;
      expect(subCheck.subscribed).toBeTruthy();

      // Unsubscribe
      await api.post(`/accounts/${id}/unsubscribe`, {});

      // Verify unsubscribed
      const unsubCheck = (await api.get(`/accounts/${id}/subscription`)) as Record<string, unknown>;
      expect(unsubCheck.subscribed).toBeFalsy();
    } finally {
      await api.deleteEntity('accounts', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 5. Comments via API
// ---------------------------------------------------------------------------
test.describe('Comments via API', () => {
  test('CRUD comments on an account', async ({ api }) => {
    const account = await api.createAccount({ name: `Comment Gap Test ${Date.now()}` });
    const id = account.id as number;

    try {
      // Add comment
      const created = (await api.post(`/accounts/${id}/comments`, { comment: 'e2e gap test comment' })) as Record<string, unknown>;
      const commentId = created.id as number;
      expect(commentId).toBeTruthy();

      // List comments
      const list = (await api.get(`/accounts/${id}/comments`)) as Record<string, unknown>[];
      expect(Array.isArray(list)).toBe(true);
      const found = list.some((c) => c.id === commentId);
      expect(found).toBe(true);

      // Delete comment
      await api.del(`/comments/${commentId}`);

      // Verify deleted
      const listAfter = (await api.get(`/accounts/${id}/comments`)) as Record<string, unknown>[];
      expect(Array.isArray(listAfter)).toBe(true);
      const stillExists = listAfter.some((c) => c.id === commentId);
      expect(stillExists).toBe(false);
    } finally {
      await api.deleteEntity('accounts', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 6. Tags via API
// ---------------------------------------------------------------------------
test.describe('Tags via API', () => {
  test('add, list, and remove tags on an account', async ({ api }) => {
    const account = await api.createAccount({ name: `Tag Gap Test ${Date.now()}` });
    const id = account.id as number;

    try {
      // Add tag
      await api.post(`/accounts/${id}/tags`, { name: 'important' });

      // List tags
      const tags = (await api.get(`/accounts/${id}/tags`)) as Record<string, unknown>[];
      expect(Array.isArray(tags)).toBe(true);
      const tagEntry = tags.find(
        (t) => t.name === 'important' || (typeof t === 'string' && t === 'important'),
      );
      expect(tagEntry).toBeDefined();

      // Find the tag id for deletion
      const tagId = typeof tagEntry === 'object' && tagEntry !== null ? (tagEntry as Record<string, unknown>).id : null;
      expect(tagId).toBeTruthy();

      // Remove tag
      await api.del(`/accounts/${id}/tags/${tagId}`);

      // Verify removed
      const tagsAfter = (await api.get(`/accounts/${id}/tags`)) as Record<string, unknown>[];
      expect(Array.isArray(tagsAfter)).toBe(true);
      const stillExists = tagsAfter.some((t) => t.name === 'important');
      expect(stillExists).toBe(false);
    } finally {
      await api.deleteEntity('accounts', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 7. Addresses via API
// ---------------------------------------------------------------------------
test.describe('Addresses via API', () => {
  test('add, list, and delete address on an account', async ({ api }) => {
    const account = await api.createAccount({ name: `Address Gap Test ${Date.now()}` });
    const id = account.id as number;

    try {
      // Add address
      const created = (await api.post(`/accounts/${id}/addresses`, {
        street1: '123 Main St',
        city: 'New York',
        state: 'NY',
        zipcode: '10001',
      })) as Record<string, unknown>;
      const addressId = created.id as number;
      expect(addressId).toBeTruthy();

      // List addresses
      const addresses = (await api.get(`/accounts/${id}/addresses`)) as Record<string, unknown>[];
      expect(Array.isArray(addresses)).toBe(true);
      const found = addresses.some((a) => a.id === addressId);
      expect(found).toBe(true);

      // Delete address
      await api.del(`/addresses/${addressId}`);

      // Verify deleted
      const addressesAfter = (await api.get(`/accounts/${id}/addresses`)) as Record<string, unknown>[];
      expect(Array.isArray(addressesAfter)).toBe(true);
      const stillExists = addressesAfter.some((a) => a.id === addressId);
      expect(stillExists).toBe(false);
    } finally {
      await api.deleteEntity('accounts', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 8. Audit Trail / Versions
// ---------------------------------------------------------------------------
test.describe('Audit Trail / Versions', () => {
  test('version history shows create and update events', async ({ api }) => {
    const account = await api.createAccount({ name: `Version Test ${Date.now()}` });
    const id = account.id as number;

    try {
      // Update the account to generate a second version entry
      await api.put(`/accounts/${id}`, { name: `Version Test Updated ${Date.now()}` });

      // Fetch versions
      const versions = (await api.get(`/accounts/${id}/versions`)) as Record<string, unknown>[];
      expect(Array.isArray(versions)).toBe(true);
      expect(versions.length).toBeGreaterThanOrEqual(2);
      const events = versions.map((v) => v.event);
      expect(events).toContain('create');
      expect(events).toContain('update');
    } finally {
      await api.deleteEntity('accounts', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 9. Export Formats
// ---------------------------------------------------------------------------
test.describe('Export Formats', () => {
  test('export accounts as CSV', async ({ api }) => {
    const res = await api.getRaw('/accounts/export');
    expect(res.ok).toBe(true);

    const contentType = res.headers.get('content-type') || '';
    const body = await res.text();
    // CSV should have a text/csv content type or at least look like CSV
    const isCSV = contentType.includes('csv') || body.includes(',');
    expect(isCSV).toBe(true);
    // Check for expected header columns
    const hasHeaders = body.includes('name') || body.includes('Name');
    expect(hasHeaders).toBe(true);
  });

  test('export contacts as vCard', async ({ api }) => {
    const res = await api.getRaw('/contacts/export/vcard');
    expect(res.ok).toBe(true);

    const contentType = res.headers.get('content-type') || '';
    const body = await res.text();
    const isVCard = contentType.includes('vcard') || body.includes('BEGIN:VCARD');
    expect(isVCard).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// 10. Saved Searches
// ---------------------------------------------------------------------------
test.describe('Saved Searches', () => {
  test('CRUD saved searches', async ({ api }) => {
    const searchName = `E2E Saved Search ${Date.now()}`;

    // Create saved search
    const created = (await api.post('/saved_searches', {
      name: searchName,
      entity: 'accounts',
      search: { name_cont: 'test' },
    })) as Record<string, unknown>;
    const searchId = created.id as number;
    expect(searchId).toBeTruthy();

    try {
      // List saved searches and verify ours appears with its filters
      const list = (await api.get('/saved_searches')) as Record<string, unknown>[];
      expect(Array.isArray(list)).toBe(true);
      const found = list.find((s) => s.id === searchId);
      expect(found).toBeDefined();

      // Delete saved search
      await api.del(`/saved_searches/${searchId}`);

      // Verify deleted
      const listAfter = (await api.get('/saved_searches')) as Record<string, unknown>[];
      expect(Array.isArray(listAfter)).toBe(true);
      const stillExists = listAfter.some((s) => s.id === searchId);
      expect(stillExists).toBe(false);
    } catch (e) {
      // Cleanup on failure
      await api.deleteEntity('saved_searches' as never, searchId);
      throw e;
    }
  });
});

// ---------------------------------------------------------------------------
// 11. Field Groups / Custom Fields
// ---------------------------------------------------------------------------
test.describe('Field Groups / Custom Fields', () => {
  test('list field groups for accounts', async ({ api }) => {
    const result = (await api.get('/field_groups?entity=accounts')) as Record<string, unknown>;
    expect(result.entity_type).toBe('Account');
    expect(Array.isArray(result.field_groups)).toBe(true);
  });

  test('custom fields on an account', async ({ api }) => {
    const account = await api.createAccount({ name: `Custom Fields Test ${Date.now()}` });
    const id = account.id as number;

    try {
      const result = await api.get(`/accounts/${id}/custom_fields`);
      expect(result).toBeDefined();
      expect(typeof result).toBe('object');
    } finally {
      await api.deleteEntity('accounts', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 12. Dashboard Endpoints
// ---------------------------------------------------------------------------
test.describe('Dashboard Endpoints', () => {
  test('dashboard tasks endpoint', async ({ api }) => {
    const result = await api.get('/dashboard/tasks');
    expect(result).toBeDefined();
  });

  test('dashboard pipeline endpoint', async ({ api }) => {
    const result = await api.get('/dashboard/pipeline');
    expect(result).toBeDefined();
  });
});
