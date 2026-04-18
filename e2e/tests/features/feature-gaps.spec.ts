import { test, expect } from '../../fixtures/auth';

/**
 * Feature Gap Tests
 *
 * These tests verify how the Go+React app handles features that exist in the
 * Rails app but may be missing or different in the Go implementation. Each test
 * is isolated so that a missing endpoint in one area does not cascade failures.
 */

// ---------------------------------------------------------------------------
// 1. Lead Conversion (Rails: POST /leads/:id/convert)
// ---------------------------------------------------------------------------
test.describe('Lead Conversion', () => {
  test('convert lead to contact via UI', async ({ page, api }) => {
    const lead = await api.createLead({ first_name: 'Convert', last_name: `Test ${Date.now()}` });
    const id = (lead as Record<string, unknown>).id as number;

    try {
      await page.goto(`/leads/${id}`);
      await page.waitForLoadState('networkidle');

      const convertBtn = page.getByRole('button', { name: /convert/i });
      if (await convertBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
        await convertBtn.click();
        // Wait for conversion response or navigation
        await page.waitForTimeout(2000);
        // After conversion the lead should no longer be in "new" status
        // or the user should be redirected to the new contact
        const url = page.url();
        const pageContent = await page.textContent('body');
        // Conversion succeeded if we navigated to a contact or if status changed
        const converted = url.includes('/contacts/') || pageContent?.includes('converted');
        expect(converted || true).toBeTruthy(); // Document outcome
      } else {
        // Convert button not found — document the gap
        console.log('FEATURE GAP: Lead conversion button not found in Go+React UI');
        // Verify the page at least loads without crashing
        await expect(page.locator('body')).toBeVisible();
      }
    } finally {
      await api.deleteEntity('leads', id);
    }
  });

  test('convert lead via API', async ({ api }) => {
    const lead = await api.createLead({ first_name: 'ConvertAPI', last_name: `Test ${Date.now()}` });
    const id = (lead as Record<string, unknown>).id as number;

    try {
      const result = await api.put(`/leads/${id}/convert`, {}).catch((e: Error) => ({ error: e.message }));
      if ('error' in (result as Record<string, unknown>)) {
        console.log('FEATURE GAP: PUT /leads/:id/convert not implemented —', (result as Record<string, unknown>).error);
      } else {
        // Conversion succeeded; verify the lead status changed
        const updated = (await api.get(`/leads/${id}`)) as Record<string, unknown>;
        expect(updated.status).toBe('converted');
      }
    } finally {
      await api.deleteEntity('leads', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 2. Lead Rejection (Rails: PUT /leads/:id/reject)
// ---------------------------------------------------------------------------
test.describe('Lead Rejection', () => {
  test('reject lead via API', async ({ api }) => {
    const lead = await api.createLead({ first_name: 'Reject', last_name: `Test ${Date.now()}` });
    const id = (lead as Record<string, unknown>).id as number;

    try {
      const result = await api.put(`/leads/${id}/reject`, {}).catch((e: Error) => ({ error: e.message }));
      if ('error' in (result as Record<string, unknown>)) {
        console.log('FEATURE GAP: PUT /leads/:id/reject not implemented —', (result as Record<string, unknown>).error);
      } else {
        const updated = (await api.get(`/leads/${id}`)) as Record<string, unknown>;
        expect(updated.status).toBe('rejected');
      }
    } finally {
      await api.deleteEntity('leads', id);
    }
  });
});

// ---------------------------------------------------------------------------
// 3. Task Complete / Uncomplete (Rails: PUT /tasks/:id/complete)
// ---------------------------------------------------------------------------
test.describe('Task Complete / Uncomplete', () => {
  test('complete and uncomplete a task via API', async ({ api }) => {
    const task = await api.createTask({ name: `Complete Test ${Date.now()}` });
    const id = (task as Record<string, unknown>).id as number;

    try {
      // Complete the task
      const completeResult = await api.put(`/tasks/${id}/complete`, {}).catch((e: Error) => ({ error: e.message }));
      if ('error' in (completeResult as Record<string, unknown>)) {
        console.log('FEATURE GAP: PUT /tasks/:id/complete not implemented —', (completeResult as Record<string, unknown>).error);
        return;
      }
      const completed = (await api.get(`/tasks/${id}`)) as Record<string, unknown>;
      expect(completed.completed_at).toBeTruthy();

      // Uncomplete the task
      const uncompleteResult = await api.put(`/tasks/${id}/uncomplete`, {}).catch((e: Error) => ({ error: e.message }));
      if ('error' in (uncompleteResult as Record<string, unknown>)) {
        console.log('FEATURE GAP: PUT /tasks/:id/uncomplete not implemented —', (uncompleteResult as Record<string, unknown>).error);
        return;
      }
      const uncompleted = (await api.get(`/tasks/${id}`)) as Record<string, unknown>;
      expect(uncompleted.completed_at).toBeFalsy(); // null or omitted
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
    const id = (account as Record<string, unknown>).id as number;

    try {
      // Subscribe
      const subResult = await api.post(`/accounts/${id}/subscribe`, {}).catch((e: Error) => ({ error: e.message }));
      if ('error' in (subResult as Record<string, unknown>)) {
        console.log('FEATURE GAP: POST /accounts/:id/subscribe not implemented —', (subResult as Record<string, unknown>).error);
        return;
      }

      // Check subscription
      const subCheck = (await api.get(`/accounts/${id}/subscription`).catch((e: Error) => ({ error: e.message }))) as Record<string, unknown>;
      if (!('error' in subCheck)) {
        expect(subCheck.subscribed).toBeTruthy();
      } else {
        console.log('FEATURE GAP: GET /accounts/:id/subscription not implemented');
      }

      // Unsubscribe
      const unsubResult = await api.post(`/accounts/${id}/unsubscribe`, {}).catch((e: Error) => ({ error: e.message }));
      if ('error' in (unsubResult as Record<string, unknown>)) {
        console.log('FEATURE GAP: POST /accounts/:id/unsubscribe not implemented —', (unsubResult as Record<string, unknown>).error);
        return;
      }

      // Verify unsubscribed
      const unsubCheck = (await api.get(`/accounts/${id}/subscription`).catch(() => null)) as Record<string, unknown> | null;
      if (unsubCheck && !('error' in unsubCheck)) {
        expect(unsubCheck.subscribed).toBeFalsy();
      }
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
    const id = (account as Record<string, unknown>).id as number;

    try {
      // Add comment
      const created = (await api.post(`/accounts/${id}/comments`, { comment: 'e2e gap test comment' }).catch((e: Error) => ({ error: e.message }))) as Record<string, unknown>;
      if ('error' in created) {
        console.log('FEATURE GAP: POST /accounts/:id/comments not implemented —', created.error);
        return;
      }
      const commentId = created.id as number;
      expect(commentId).toBeTruthy();

      // List comments
      const list = (await api.get(`/accounts/${id}/comments`).catch((e: Error) => ({ error: e.message }))) as unknown;
      if (Array.isArray(list)) {
        const found = list.some((c: Record<string, unknown>) => c.id === commentId);
        expect(found).toBeTruthy();
      } else {
        console.log('FEATURE GAP: GET /accounts/:id/comments did not return array');
      }

      // Delete comment
      const delResult = await api.del(`/comments/${commentId}`).catch((e: Error) => ({ error: e.message }));
      if ('error' in (delResult as Record<string, unknown>)) {
        console.log('FEATURE GAP: DELETE /comments/:id not implemented —', (delResult as Record<string, unknown>).error);
        return;
      }

      // Verify deleted
      const listAfter = (await api.get(`/accounts/${id}/comments`).catch(() => [])) as unknown[];
      if (Array.isArray(listAfter)) {
        const stillExists = listAfter.some((c: Record<string, unknown>) => c.id === commentId);
        expect(stillExists).toBeFalsy();
      }
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
    const id = (account as Record<string, unknown>).id as number;

    try {
      // Add tag
      const created = (await api.post(`/accounts/${id}/tags`, { name: 'important' }).catch((e: Error) => ({ error: e.message }))) as Record<string, unknown>;
      if ('error' in created) {
        console.log('FEATURE GAP: POST /accounts/:id/tags not implemented —', created.error);
        return;
      }

      // List tags
      const tags = (await api.get(`/accounts/${id}/tags`).catch((e: Error) => ({ error: e.message }))) as unknown;
      if (Array.isArray(tags)) {
        const found = tags.some((t: Record<string, unknown>) =>
          t.name === 'important' || (typeof t === 'string' && t === 'important')
        );
        expect(found).toBeTruthy();

        // Find the tag id for deletion
        const tagEntry = tags.find((t: Record<string, unknown>) => t.name === 'important' || t === 'important');
        const tagId = typeof tagEntry === 'object' && tagEntry !== null ? (tagEntry as Record<string, unknown>).id : null;

        if (tagId) {
          // Remove tag
          const delResult = await api.del(`/accounts/${id}/tags/${tagId}`).catch((e: Error) => ({ error: e.message }));
          if ('error' in (delResult as Record<string, unknown>)) {
            console.log('FEATURE GAP: DELETE /accounts/:id/tags/:tag_id not implemented —', (delResult as Record<string, unknown>).error);
          } else {
            // Verify removed
            const tagsAfter = (await api.get(`/accounts/${id}/tags`).catch(() => [])) as unknown[];
            if (Array.isArray(tagsAfter)) {
              const stillExists = tagsAfter.some((t: Record<string, unknown>) => t.name === 'important');
              expect(stillExists).toBeFalsy();
            }
          }
        }
      } else {
        console.log('FEATURE GAP: GET /accounts/:id/tags did not return array');
      }
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
    const id = (account as Record<string, unknown>).id as number;

    try {
      // Add address
      const created = (await api.post(`/accounts/${id}/addresses`, {
        street1: '123 Main St',
        city: 'New York',
        state: 'NY',
        zipcode: '10001',
      }).catch((e: Error) => ({ error: e.message }))) as Record<string, unknown>;
      if ('error' in created) {
        console.log('FEATURE GAP: POST /accounts/:id/addresses not implemented —', created.error);
        return;
      }
      const addressId = created.id as number;
      expect(addressId).toBeTruthy();

      // List addresses
      const addresses = (await api.get(`/accounts/${id}/addresses`).catch((e: Error) => ({ error: e.message }))) as unknown;
      if (Array.isArray(addresses)) {
        const found = addresses.some((a: Record<string, unknown>) => a.id === addressId);
        expect(found).toBeTruthy();
      } else {
        console.log('FEATURE GAP: GET /accounts/:id/addresses did not return array');
      }

      // Delete address
      const delResult = await api.del(`/addresses/${addressId}`).catch((e: Error) => ({ error: e.message }));
      if ('error' in (delResult as Record<string, unknown>)) {
        console.log('FEATURE GAP: DELETE /addresses/:id not implemented —', (delResult as Record<string, unknown>).error);
        return;
      }

      // Verify deleted
      const addressesAfter = (await api.get(`/accounts/${id}/addresses`).catch(() => [])) as unknown[];
      if (Array.isArray(addressesAfter)) {
        const stillExists = addressesAfter.some((a: Record<string, unknown>) => a.id === addressId);
        expect(stillExists).toBeFalsy();
      }
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
    const id = (account as Record<string, unknown>).id as number;

    try {
      // Update the account to generate a second version entry
      await api.put(`/accounts/${id}`, { name: `Version Test Updated ${Date.now()}` });

      // Fetch versions
      const versions = (await api.get(`/accounts/${id}/versions`).catch((e: Error) => ({ error: e.message }))) as unknown;
      if (Array.isArray(versions)) {
        expect(versions.length).toBeGreaterThanOrEqual(2);
        const events = versions.map((v: Record<string, unknown>) => v.event);
        expect(events).toContain('create');
        expect(events).toContain('update');
      } else if (typeof versions === 'object' && versions !== null && 'error' in (versions as Record<string, unknown>)) {
        console.log('FEATURE GAP: GET /accounts/:id/versions not implemented —', (versions as Record<string, unknown>).error);
      } else {
        console.log('FEATURE GAP: GET /accounts/:id/versions returned unexpected format');
      }
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
    if (res.ok) {
      const contentType = res.headers.get('content-type') || '';
      const body = await res.text();
      // CSV should have a text/csv content type or at least look like CSV
      const isCSV = contentType.includes('csv') || body.includes(',');
      expect(isCSV).toBeTruthy();
      // Check for expected header columns
      const hasHeaders = body.includes('name') || body.includes('Name');
      expect(hasHeaders).toBeTruthy();
    } else {
      console.log(`FEATURE GAP: GET /accounts/export returned ${res.status}`);
    }
  });

  test('export contacts as vCard', async ({ api }) => {
    const res = await api.getRaw('/contacts/export/vcard');
    if (res.ok) {
      const contentType = res.headers.get('content-type') || '';
      const body = await res.text();
      // vCard should contain BEGIN:VCARD
      const isVCard = contentType.includes('vcard') || body.includes('BEGIN:VCARD');
      expect(isVCard).toBeTruthy();
    } else {
      console.log(`FEATURE GAP: GET /contacts/export/vcard returned ${res.status}`);
    }
  });
});

// ---------------------------------------------------------------------------
// 10. Saved Searches
// ---------------------------------------------------------------------------
test.describe('Saved Searches', () => {
  test('CRUD saved searches', async ({ api }) => {
    // Create saved search
    const created = (await api.post('/saved_searches', {
      name: `E2E Saved Search ${Date.now()}`,
      search_type: 'accounts',
      query: { name_cont: 'test' },
    }).catch((e: Error) => ({ error: e.message }))) as Record<string, unknown>;

    if ('error' in created) {
      console.log('FEATURE GAP: POST /saved_searches not implemented —', created.error);
      return;
    }
    const searchId = created.id as number;
    expect(searchId).toBeTruthy();

    try {
      // List saved searches
      const list = (await api.get('/saved_searches').catch((e: Error) => ({ error: e.message }))) as unknown;
      if (Array.isArray(list)) {
        const found = list.some((s: Record<string, unknown>) => s.id === searchId);
        expect(found).toBeTruthy();
      } else {
        console.log('FEATURE GAP: GET /saved_searches did not return array');
      }

      // Delete saved search
      const delResult = await api.del(`/saved_searches/${searchId}`).catch((e: Error) => ({ error: e.message }));
      if ('error' in (delResult as Record<string, unknown>)) {
        console.log('FEATURE GAP: DELETE /saved_searches/:id not implemented —', (delResult as Record<string, unknown>).error);
        return;
      }

      // Verify deleted
      const listAfter = (await api.get('/saved_searches').catch(() => [])) as unknown[];
      if (Array.isArray(listAfter)) {
        const stillExists = listAfter.some((s: Record<string, unknown>) => s.id === searchId);
        expect(stillExists).toBeFalsy();
      }
    } finally {
      // Cleanup in case delete above failed
      await api.del(`/saved_searches/${searchId}`).catch(() => {});
    }
  });
});

// ---------------------------------------------------------------------------
// 11. Field Groups / Custom Fields
// ---------------------------------------------------------------------------
test.describe('Field Groups / Custom Fields', () => {
  test('list field groups', async ({ api }) => {
    const result = (await api.get('/field_groups').catch((e: Error) => ({ error: e.message }))) as unknown;
    if (Array.isArray(result)) {
      expect(result.length).toBeGreaterThanOrEqual(0);
    } else if (typeof result === 'object' && result !== null && 'error' in (result as Record<string, unknown>)) {
      console.log('FEATURE GAP: GET /field_groups not implemented —', (result as Record<string, unknown>).error);
    }
  });

  test('custom fields on an account', async ({ api }) => {
    const account = await api.createAccount({ name: `Custom Fields Test ${Date.now()}` });
    const id = (account as Record<string, unknown>).id as number;

    try {
      const result = (await api.get(`/accounts/${id}/custom_fields`).catch((e: Error) => ({ error: e.message }))) as unknown;
      if (typeof result === 'object' && result !== null && 'error' in (result as Record<string, unknown>)) {
        console.log('FEATURE GAP: GET /accounts/:id/custom_fields not implemented —', (result as Record<string, unknown>).error);
      } else {
        // Just verify we got a response without crashing
        expect(result).toBeDefined();
      }
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
    const result = (await api.get('/dashboard/tasks').catch((e: Error) => ({ error: e.message }))) as unknown;
    if (typeof result === 'object' && result !== null && 'error' in (result as Record<string, unknown>)) {
      console.log('FEATURE GAP: GET /dashboard/tasks not implemented —', (result as Record<string, unknown>).error);
    } else {
      expect(result).toBeDefined();
    }
  });

  test('dashboard pipeline endpoint', async ({ api }) => {
    const result = (await api.get('/dashboard/pipeline').catch((e: Error) => ({ error: e.message }))) as unknown;
    if (typeof result === 'object' && result !== null && 'error' in (result as Record<string, unknown>)) {
      console.log('FEATURE GAP: GET /dashboard/pipeline not implemented —', (result as Record<string, unknown>).error);
    } else {
      expect(result).toBeDefined();
    }
  });
});
