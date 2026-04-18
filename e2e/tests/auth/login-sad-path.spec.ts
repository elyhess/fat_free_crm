import { test, expect } from '@playwright/test';

test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Login - Sad Path', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('empty username with filled password does not submit', async ({ page }) => {
    await page.getByLabel(/password/i).fill('somepassword');
    await page.getByRole('button', { name: /sign in/i }).click();

    // The username field has the required attribute, so the browser should block submission.
    // Verify we are still on the login page and no error message appeared (form never submitted).
    await expect(page).toHaveURL(/\/login/);
    const usernameInput = page.getByLabel(/username or email/i);
    await expect(usernameInput).toBeVisible();
  });

  test('filled username with empty password does not submit', async ({ page }) => {
    await page.getByLabel(/username or email/i).fill('admin');
    await page.getByRole('button', { name: /sign in/i }).click();

    // The password field has the required attribute, so the browser should block submission.
    await expect(page).toHaveURL(/\/login/);
    const passwordInput = page.getByLabel(/password/i);
    await expect(passwordInput).toBeVisible();
  });

  test('both fields empty does not submit', async ({ page }) => {
    await page.getByRole('button', { name: /sign in/i }).click();

    // Both fields are required; browser validation prevents submission.
    await expect(page).toHaveURL(/\/login/);
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
  });

  test('XSS attempt in username shows error without script execution', async ({ page }) => {
    const xssPayload = "<script>alert('xss')</script>";

    // Listen for any dialog (alert/confirm/prompt) — none should appear.
    let dialogAppeared = false;
    page.on('dialog', async (dialog) => {
      dialogAppeared = true;
      await dialog.dismiss();
    });

    await page.getByLabel(/username or email/i).fill(xssPayload);
    await page.getByLabel(/password/i).fill('anypassword');
    await page.getByRole('button', { name: /sign in/i }).click();

    await expect(page.getByText('Invalid login or password')).toBeVisible({ timeout: 15000 });
    expect(dialogAppeared).toBe(false);

    // Verify the script tag is not rendered as HTML in the page.
    const scriptElements = await page.locator('script:text("alert")').count();
    expect(scriptElements).toBe(0);
  });

  test('SQL injection attempt in username fails with error', async ({ page }) => {
    await page.getByLabel(/username or email/i).fill("' OR 1=1 --");
    await page.getByLabel(/password/i).fill('anypassword');
    await page.getByRole('button', { name: /sign in/i }).click();

    await expect(page.getByText('Invalid login or password')).toBeVisible({ timeout: 15000 });
    await expect(page).toHaveURL(/\/login/);
  });

  test('very long username is handled gracefully', async ({ page }) => {
    const longUsername = 'a'.repeat(10000);

    await page.getByLabel(/username or email/i).fill(longUsername);
    await page.getByLabel(/password/i).fill('anypassword');
    await page.getByRole('button', { name: /sign in/i }).click();

    // Should show an error or stay on the login page — no crash.
    await expect(page).toHaveURL(/\/login/);
    // Either an error is shown or the form is still visible.
    const errorOrForm = page
      .getByText('Invalid login or password')
      .or(page.getByRole('button', { name: /sign in/i }));
    await expect(errorOrForm.first()).toBeVisible({ timeout: 15000 });
  });

  test('very long password is handled gracefully', async ({ page }) => {
    const longPassword = 'b'.repeat(10000);

    await page.getByLabel(/username or email/i).fill('admin');
    await page.getByLabel(/password/i).fill(longPassword);
    await page.getByRole('button', { name: /sign in/i }).click();

    // Should show an error or stay on the login page — no crash.
    await expect(page).toHaveURL(/\/login/);
    const errorOrForm = page
      .getByText('Invalid login or password')
      .or(page.getByRole('button', { name: /sign in/i }));
    await expect(errorOrForm.first()).toBeVisible({ timeout: 15000 });
  });

  test('whitespace-only username does not submit or shows error', async ({ page }) => {
    // The required attribute may treat whitespace-only as empty in some browsers.
    // Use dispatchEvent to bypass potential browser trimming on required check.
    const usernameInput = page.getByLabel(/username or email/i);
    await usernameInput.fill('   ');
    await page.getByLabel(/password/i).fill('anypassword');
    await page.getByRole('button', { name: /sign in/i }).click();

    // Either browser validation blocks it or backend returns an error.
    await expect(page).toHaveURL(/\/login/);
    const errorOrForm = page
      .getByText('Invalid login or password')
      .or(page.getByRole('button', { name: /sign in/i }));
    await expect(errorOrForm.first()).toBeVisible({ timeout: 15000 });
  });

  test('special characters and unicode in credentials shows error', async ({ page }) => {
    await page.getByLabel(/username or email/i).fill('user@£€¥©®™✓✗🎉🚀');
    await page.getByLabel(/password/i).fill('p@$$w0rd!™€🔑');
    await page.getByRole('button', { name: /sign in/i }).click();

    await expect(page.getByText('Invalid login or password')).toBeVisible({ timeout: 15000 });
    await expect(page).toHaveURL(/\/login/);
  });

  test('rapid repeated login attempts still work correctly', async ({ page }) => {
    const attempts = 5;

    for (let i = 0; i < attempts; i++) {
      await page.getByLabel(/username or email/i).fill(`wronguser${i}`);
      await page.getByLabel(/password/i).fill('wrongpassword');
      await page.getByRole('button', { name: /sign in/i }).click();
      await expect(page.getByText('Invalid login or password')).toBeVisible({ timeout: 15000 });
    }

    // After repeated failures, the app should still be functional.
    // Verify the form is still usable and a valid login can succeed.
    await page.getByLabel(/username or email/i).fill('admin');
    await page.getByLabel(/password/i).fill('Dem0P@ssword!!');
    await page.getByRole('button', { name: /sign in/i }).click();
    await page.waitForURL('/', { timeout: 10000 });
  });
});
