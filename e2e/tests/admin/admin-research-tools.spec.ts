import { test, expect } from '../../fixtures/auth';

test.describe('Admin Research Tools', () => {
  test('page loads', async ({ page }) => {
    await page.goto('/admin/research-tools');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: /research tools/i })).toBeVisible();
  });

  test('non-admin sees access denied', async ({ demoPage }) => {
    await demoPage.goto('/admin/research-tools');
    await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
  });
});
