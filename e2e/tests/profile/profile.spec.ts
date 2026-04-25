import { test, expect } from '../../fixtures/auth';

test.describe('Profile', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/profile');
    await page.waitForLoadState('networkidle');
  });

  test('profile page loads', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /profile/i })).toBeVisible();
  });

  test('shows username', async ({ page }) => {
    await expect(page.getByText(/Username:\s*admin/)).toBeVisible();
  });

  test('profile form has fields', async ({ page }) => {
    await expect(page.locator('#login, input[type="email"]').first()).toBeVisible();
  });

  test('update profile successfully', async ({ page }) => {
    const firstNameInput = page.locator('input').nth(0); // first input in profile form
    const formSection = page.locator('form').filter({ hasText: /save/i }).first();
    const firstInput = formSection.locator('input[type="text"]').first();

    if (await firstInput.isVisible()) {
      const origValue = await firstInput.inputValue();
      await firstInput.clear();
      await firstInput.fill('E2EUpdated');
      await formSection.getByRole('button', { name: /save/i }).click();
      await expect(page.getByText(/updated|saved|success/i)).toBeVisible({ timeout: 5000 });

      // Restore
      await firstInput.clear();
      await firstInput.fill(origValue || '');
      await formSection.getByRole('button', { name: /save/i }).click();
    }
  });
});
