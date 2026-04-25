import { test, expect } from '../../fixtures/auth';

test.describe('Authorization', () => {
  test('unauthenticated user redirected to login', async ({ browser }) => {
    const context = await browser.newContext({ storageState: { cookies: [], origins: [] } });
    const page = await context.newPage();
    await page.goto('/accounts');
    await page.waitForURL('/login');
    await context.close();
  });

  test('non-admin cannot access admin settings', async ({ demoPage }) => {
    await demoPage.goto('/admin/settings');
    await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
  });

  test('non-admin cannot access admin fields', async ({ demoPage }) => {
    await demoPage.goto('/admin/fields');
    await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
  });

  test('non-admin cannot access admin research tools', async ({ demoPage }) => {
    await demoPage.goto('/admin/research-tools');
    await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
  });

  test('private entities not visible to non-owner', async ({ api, demoPage }) => {
    // Create a private account as admin
    const account = await api.createAccount({ name: `Private E2E ${Date.now()}`, access: 'Private' });
    const id = (account as Record<string, unknown>).id as number;

    try {
      await demoPage.goto('/accounts');
      await demoPage.waitForLoadState('networkidle');
      // The private account should not appear in the demo user's list
      await expect(demoPage.getByText(`Private E2E`)).toBeHidden();
    } finally {
      await api.deleteEntity('accounts', id);
    }
  });
});
