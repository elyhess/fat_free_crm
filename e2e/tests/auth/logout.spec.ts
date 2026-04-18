import { test, expect } from '../../fixtures/auth';

test.describe('Logout', () => {
  test('sign out redirects to login', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: /sign out/i }).click();
    await page.waitForURL('/login');
  });

  test('protected routes redirect to login after logout', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: /sign out/i }).click();
    await page.waitForURL('/login');
    await page.goto('/accounts');
    await page.waitForURL('/login');
  });
});
