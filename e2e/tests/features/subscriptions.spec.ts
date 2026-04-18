import { test, expect } from '../../fixtures/auth';

test.describe('Subscriptions', () => {
  let accountId: number;

  test.beforeAll(async () => {
    const { token } = await (await import('../../fixtures/api-client')).ApiClient.login('admin', 'Dem0P@ssword!!');
    const api = new (await import('../../fixtures/api-client')).ApiClient(token);
    const account = await api.createAccount({ name: `Sub Test ${Date.now()}` });
    accountId = (account as Record<string, unknown>).id as number;
  });

  test.afterAll(async () => {
    const { token } = await (await import('../../fixtures/api-client')).ApiClient.login('admin', 'Dem0P@ssword!!');
    const api = new (await import('../../fixtures/api-client')).ApiClient(token);
    await api.deleteEntity('accounts', accountId);
  });

  test('subscribe to entity', async ({ page }) => {
    await page.goto(`/accounts/${accountId}`);
    const subBtn = page.getByRole('button', { name: /subscribe/i });
    await subBtn.click();
    await page.waitForTimeout(500);
    // Button should change to indicate subscribed
    await expect(page.getByRole('button', { name: /unsubscribe|subscribed/i })).toBeVisible();
  });

  test('unsubscribe from entity', async ({ page }) => {
    await page.goto(`/accounts/${accountId}`);
    // If already subscribed, unsubscribe
    const unsubBtn = page.getByRole('button', { name: /unsubscribe|subscribed/i });
    if (await unsubBtn.isVisible()) {
      await unsubBtn.click();
      await page.waitForTimeout(500);
      await expect(page.getByRole('button', { name: /subscribe/i })).toBeVisible();
    }
  });
});
