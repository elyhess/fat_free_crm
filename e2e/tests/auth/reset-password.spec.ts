import { test, expect } from '@playwright/test';

test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Reset Password', () => {
  test('shows reset password form with token', async ({ page }) => {
    await page.goto('/reset-password?token=fake-token');
    await expect(page.getByLabel(/new password/i)).toBeVisible();
    await expect(page.getByLabel(/confirm/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /reset|change|update/i })).toBeVisible();
  });

  test('invalid token shows error on submit', async ({ page }) => {
    await page.goto('/reset-password?token=invalid-token');
    await page.getByLabel(/new password/i).fill('NewStr0ngP@ssword!!');
    await page.getByLabel(/confirm/i).fill('NewStr0ngP@ssword!!');
    await page.getByRole('button', { name: /reset|change|update/i }).click();
    await expect(page.getByText(/invalid|expired|error/i)).toBeVisible();
  });
});
