import { test, expect } from '../../fixtures/auth';

test.describe('Custom Fields', () => {
  test('admin fields page loads', async ({ page }) => {
    await page.goto('/admin/fields');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: /custom fields/i })).toBeVisible();
  });

  test('entity type tabs work', async ({ page }) => {
    await page.goto('/admin/fields');
    await page.waitForLoadState('networkidle');
    for (const entity of ['Contacts', 'Leads', 'Accounts']) {
      await page.getByRole('button', { name: entity }).click();
      await page.waitForTimeout(300);
    }
  });

  test('add field modal opens and closes', async ({ page }) => {
    await page.goto('/admin/fields');
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: /add field/i }).first().click();
    await expect(page.getByText(/add custom field/i)).toBeVisible();
    await page.getByRole('button', { name: /cancel/i }).click();
  });
});
