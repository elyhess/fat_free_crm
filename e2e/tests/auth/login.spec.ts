import { test, expect } from '@playwright/test';

test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Login', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('shows login form', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /fat free crm/i })).toBeVisible();
    await expect(page.getByLabel(/username or email/i)).toBeVisible();
    await expect(page.getByLabel(/password/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
  });

  test('successful login redirects to dashboard', async ({ page }) => {
    await page.getByLabel(/username or email/i).fill('admin');
    await page.getByLabel(/password/i).fill('Dem0P@ssword!!');
    await page.getByRole('button', { name: /sign in/i }).click();
    await page.waitForURL('/', { timeout: 10000 });
  });

  test('invalid credentials shows error', async ({ page }) => {
    await page.getByLabel(/username or email/i).fill('admin');
    await page.getByLabel(/password/i).fill('wrongpassword');
    await page.getByRole('button', { name: /sign in/i }).click();
    await expect(page.getByText('Invalid login or password')).toBeVisible({ timeout: 15000 });
  });

  test('forgot password link navigates correctly', async ({ page }) => {
    await page.getByRole('link', { name: /forgot/i }).click();
    await page.waitForURL('/forgot-password');
  });

  test('sign up link navigates correctly', async ({ page }) => {
    await page.getByRole('link', { name: /sign up/i }).click();
    await page.waitForURL('/register');
  });
});
