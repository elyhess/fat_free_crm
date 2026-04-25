import { test, expect } from '../../fixtures/auth';

test.describe('Comments', () => {
  let accountId: number;

  test.beforeAll(async () => {
    const { token } = await (await import('../../fixtures/api-client')).ApiClient.login('admin', 'Dem0P@ssword!!');
    const api = new (await import('../../fixtures/api-client')).ApiClient(token);
    const account = await api.createAccount({ name: `Comment Test ${Date.now()}` });
    accountId = (account as Record<string, unknown>).id as number;
  });

  test.afterAll(async () => {
    const { token } = await (await import('../../fixtures/api-client')).ApiClient.login('admin', 'Dem0P@ssword!!');
    const api = new (await import('../../fixtures/api-client')).ApiClient(token);
    await api.deleteEntity('accounts', accountId);
  });

  test('add comment to entity', async ({ page }) => {
    await page.goto(`/accounts/${accountId}`);

    const commentInput = page.getByPlaceholder(/comment|write/i);
    await expect(commentInput).toBeVisible();
    await commentInput.fill('E2E test comment');
    // Use the Add button scoped to the comment form
    await commentInput.locator('..').getByRole('button', { name: /add/i }).click();

    await page.waitForTimeout(500);
    await expect(page.getByText('E2E test comment')).toBeVisible();
  });

  test('multiple comments display', async ({ page, api }) => {
    // Add comments via API
    await api.post(`/accounts/${accountId}/comments`, { comment: 'First comment' });
    await api.post(`/accounts/${accountId}/comments`, { comment: 'Second comment' });

    await page.goto(`/accounts/${accountId}`);
    await page.waitForLoadState('networkidle');
    await expect(page.getByText('First comment')).toBeVisible();
    await expect(page.getByText('Second comment')).toBeVisible();
  });

  test('delete comment via API reflects in UI', async ({ page, api }) => {
    const created = (await api.post(`/accounts/${accountId}/comments`, { comment: 'Delete me comment' })) as Record<string, unknown>;
    const commentId = created.id as number;

    await page.goto(`/accounts/${accountId}`);
    await expect(page.getByText('Delete me comment')).toBeVisible();

    // Delete via API (UI does not expose a delete button)
    await api.del(`/comments/${commentId}`);

    // Reload and verify the comment is gone
    await page.reload();
    await page.waitForLoadState('networkidle');
    await expect(page.getByText('Delete me comment')).not.toBeVisible();
  });
});
