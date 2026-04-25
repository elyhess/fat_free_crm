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
    await page.waitForLoadState('networkidle');
    const subBtn = page.getByRole('button', { name: 'Subscribe to notifications' });
    await expect(subBtn).toBeVisible();
    await subBtn.click();
    // Button should change to indicate subscribed
    await expect(page.getByRole('button', { name: 'Subscribed to notifications' })).toBeVisible();
  });

  test('unsubscribe from entity', async ({ page }) => {
    // First subscribe so we have a known state
    await page.goto(`/accounts/${accountId}`);
    await page.waitForLoadState('networkidle');

    // The button may already show "Subscribed" from prior test; if not, subscribe first
    const subscribedBtn = page.getByRole('button', { name: 'Subscribed to notifications' });
    const subscribeBtn = page.getByRole('button', { name: 'Subscribe to notifications' });

    if (await subscribeBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await subscribeBtn.click();
      await expect(subscribedBtn).toBeVisible();
    }

    // Now click to unsubscribe
    await subscribedBtn.click();

    // Assert we're back to the subscribe state
    await expect(subscribeBtn).toBeVisible();
  });
});
