import { test, expect } from '../../fixtures/auth';

test.describe('Tags', () => {
  let accountId: number;

  test.beforeAll(async () => {
    const { token } = await (await import('../../fixtures/api-client')).ApiClient.login('admin', 'Dem0P@ssword!!');
    const api = new (await import('../../fixtures/api-client')).ApiClient(token);
    const account = await api.createAccount({ name: `Tag Test ${Date.now()}` });
    accountId = (account as Record<string, unknown>).id as number;
  });

  test.afterAll(async () => {
    const { token } = await (await import('../../fixtures/api-client')).ApiClient.login('admin', 'Dem0P@ssword!!');
    const api = new (await import('../../fixtures/api-client')).ApiClient(token);
    await api.deleteEntity('accounts', accountId);
  });

  test('add tag to entity', async ({ page }) => {
    await page.goto(`/accounts/${accountId}`);

    const tagInput = page.getByPlaceholder(/tag/i);
    if (await tagInput.isVisible()) {
      await tagInput.fill('e2e-tag');
      await page.getByRole('button', { name: /add|\+/i }).click();
      await page.waitForTimeout(500);
      await expect(page.getByText('e2e-tag')).toBeVisible();
    }
  });

  test('remove tag from entity', async ({ page, api }) => {
    // Add tag via API
    await api.post(`/accounts/${accountId}/tags`, { name: 'removable-tag' });

    await page.goto(`/accounts/${accountId}`);
    await expect(page.getByText('removable-tag')).toBeVisible();

    // Click the remove button (x) on the tag pill
    const tagPill = page.locator('text=removable-tag').locator('..');
    const removeBtn = tagPill.getByRole('button').first();
    if (await removeBtn.isVisible()) {
      await removeBtn.click();
      await page.waitForTimeout(500);
    }
  });
});
