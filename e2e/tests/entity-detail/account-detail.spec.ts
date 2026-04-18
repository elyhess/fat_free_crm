import { test, expect } from '../../fixtures/auth';

test.describe('Account Detail', () => {
  let accountId: number;
  let accountName: string;

  test.beforeEach(async ({ api, page }) => {
    if (!accountId) {
      const account = await api.createAccount({ name: `E2E Detail Account ${Date.now()}`, email: 'detail@test.com', phone: '555-1234' });
      accountId = (account as Record<string, unknown>).id as number;
      accountName = (account as Record<string, unknown>).name as string;
    }
    await page.goto(`/accounts/${accountId}`);
    await page.waitForLoadState('networkidle');
  });

  test('shows account details', async ({ page }) => {
    await expect(page.getByRole('heading', { name: accountName })).toBeVisible();
    await expect(page.getByText('detail@test.com')).toBeVisible();
  });

  test('edit from detail page', async ({ page }) => {
    await page.getByRole('button', { name: /edit/i }).first().click();
    await page.waitForTimeout(500);
    // Modal should open with form
    await expect(page.getByText(/edit account/i)).toBeVisible();
    await page.getByRole('button', { name: /cancel/i }).click();
  });

  test('back link returns to list', async ({ page }) => {
    await page.getByRole('button', { name: /back/i }).click();
    await expect(page).toHaveURL('/accounts');
  });

  test('shows related sections', async ({ page }) => {
    // Should have related entities sections
    await expect(page.getByRole('heading', { name: /contacts/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /opportunities/i })).toBeVisible();
  });

  test('comments section visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /comments/i })).toBeVisible();
  });

  test('tags section visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /tags/i })).toBeVisible();
  });

  test('subscribe button visible', async ({ page }) => {
    await expect(page.getByRole('button', { name: /subscribe/i })).toBeVisible();
  });

  test('history section visible', async ({ page }) => {
    await expect(page.getByText(/history/i)).toBeVisible();
  });

  test('delete from detail page', async ({ page, api }) => {
    const account = await api.createAccount({ name: `E2E Delete Detail ${Date.now()}` });
    const id = (account as Record<string, unknown>).id as number;
    await page.goto(`/accounts/${id}`);
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: /delete/i }).first().click();
    await page.getByRole('button', { name: /confirm|delete|yes/i }).last().click();
    await expect(page).toHaveURL('/accounts');
  });
});
