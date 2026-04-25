import { test, expect } from '../../fixtures/auth';
import { fillField } from '../../fixtures/helpers';

test.describe('Admin Fields', () => {
  test('fields page loads', async ({ page }) => {
    await page.goto('/admin/fields');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: /custom fields/i })).toBeVisible();
  });

  test('entity type tabs present', async ({ page }) => {
    await page.goto('/admin/fields');
    await page.waitForLoadState('networkidle');
    // Tabs show entity names with "s" suffix
    for (const entity of ['Account', 'Contact', 'Lead', 'Opportunity', 'Campaign', 'Task']) {
      await expect(page.getByRole('button', { name: `${entity}s` })).toBeVisible();
    }
  });

  test('switching tabs loads different fields', async ({ page }) => {
    await page.goto('/admin/fields');
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: 'Contacts' }).click();
    await page.waitForTimeout(300);
    await page.getByRole('button', { name: 'Accounts' }).click();
    await page.waitForTimeout(300);
  });

  test('add field button opens modal', async ({ page }) => {
    await page.goto('/admin/fields');
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: /add field/i }).first().click();
    await expect(page.getByText(/add custom field/i)).toBeVisible();
    await page.getByRole('button', { name: /cancel/i }).click();
  });

  test('non-admin sees access denied', async ({ demoPage }) => {
    await demoPage.goto('/admin/fields');
    await expect(demoPage.getByText(/admin access required/i)).toBeVisible();
  });
});
