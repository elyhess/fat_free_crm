import { test, expect } from '@playwright/test';

test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Registration', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/register');
  });

  test('shows registration form', async ({ page }) => {
    await expect(page.getByLabel(/username/i)).toBeVisible();
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel('Password', { exact: true })).toBeVisible();
    await expect(page.getByLabel(/confirm password/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /sign up|register|create/i })).toBeVisible();
  });

  test('sign in link navigates to login', async ({ page }) => {
    await page.getByRole('link', { name: /sign in/i }).click();
    await page.waitForURL('/login');
  });

  test('password mismatch shows error', async ({ page }) => {
    await page.getByLabel(/username/i).fill('testuser');
    await page.getByLabel(/email/i).fill('test@example.com');
    await page.getByLabel('Password', { exact: true }).fill('StrongP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('differentpassword');
    await page.getByRole('button', { name: /sign up|register|create/i }).click();
    await expect(page.getByText(/match/i)).toBeVisible();
  });
});
