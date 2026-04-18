import { test, expect } from '@playwright/test';

test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Forgot Password', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/forgot-password');
  });

  test('shows forgot password form', async ({ page }) => {
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /reset|send/i })).toBeVisible();
  });

  test('submitting email shows success message', async ({ page }) => {
    await page.getByLabel(/email/i).fill('admin@fatfreecrm.local');
    await page.getByRole('button', { name: /reset|send/i }).click();
    // Should show success regardless of whether email exists (enumeration prevention)
    await expect(page.getByText(/check your email|instructions|sent/i)).toBeVisible();
  });

  test('back to sign in link works', async ({ page }) => {
    await page.getByRole('link', { name: /sign in|back|login/i }).click();
    await page.waitForURL('/login');
  });
});
