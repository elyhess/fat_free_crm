import { test, expect } from '../../fixtures/auth';

test.describe('Admin Settings', () => {
  test('settings page loads for admin', async ({ page }) => {
    await page.goto('/admin/settings');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: /settings/i })).toBeVisible();
  });

  test('settings form visible', async ({ page }) => {
    await page.goto('/admin/settings');
    await page.waitForLoadState('networkidle');
    // Should have at least one form section
    await expect(page.locator('input, select, textarea').first()).toBeVisible();
  });

  test('non-admin sees access denied', async ({ demoPage }) => {
    await demoPage.goto('/admin/settings');
    await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
  });
});
