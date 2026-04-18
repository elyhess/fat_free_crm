import { test, expect } from '../../fixtures/auth';

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('all nav items visible', async ({ page }) => {
    const navLabels = ['Dashboard', 'Tasks', 'Campaigns', 'Leads', 'Accounts', 'Contacts', 'Opportunities'];
    for (const label of navLabels) {
      await expect(page.getByRole('link', { name: label, exact: true })).toBeVisible();
    }
  });

  test('nav links navigate to correct pages', async ({ page }) => {
    const navRoutes = [
      { label: 'Tasks', path: '/tasks' },
      { label: 'Accounts', path: '/accounts' },
      { label: 'Dashboard', path: '/' },
    ];
    for (const { label, path } of navRoutes) {
      await page.getByRole('link', { name: label, exact: true }).click();
      await expect(page).toHaveURL(path);
    }
  });

  test('search navigates to search page', async ({ page }) => {
    await page.getByPlaceholder(/search/i).fill('test query');
    await page.getByPlaceholder(/search/i).press('Enter');
    await expect(page).toHaveURL(/\/search\?q=test%20query/);
  });

  test('admin user sees admin links', async ({ page }) => {
    // Admin links may be in the nav bar
    await expect(page.getByRole('link', { name: /fields/i }).first()).toBeVisible();
    await expect(page.getByRole('link', { name: /settings/i }).first()).toBeVisible();
  });

  test('non-admin user does not see admin links in header', async ({ demoPage }) => {
    await demoPage.goto('/');
    await demoPage.waitForLoadState('networkidle');
    // Check that admin-specific links are not in the nav
    const adminLinks = demoPage.locator('nav').getByRole('link', { name: 'Fields' });
    await expect(adminLinks).toHaveCount(0);
  });

  test('brand text visible', async ({ page }) => {
    await expect(page.getByText('Fat Free CRM')).toBeVisible();
  });
});
