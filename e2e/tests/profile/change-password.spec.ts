import { test, expect } from '../../fixtures/auth';

test.describe('Change Password', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/profile');
    await page.waitForLoadState('networkidle');
  });

  test('change password section visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /change password/i })).toBeVisible();
  });

  test('password form fields present', async ({ page }) => {
    const pwSection = page.locator('form').filter({ hasText: /change/i }).last();
    await expect(pwSection.locator('input[type="password"]').first()).toBeVisible();
  });

  test('password requirements note displayed', async ({ page }) => {
    await expect(page.getByText(/14 characters/i)).toBeVisible();
  });
});
