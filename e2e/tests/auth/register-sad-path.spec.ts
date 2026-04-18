import { test, expect } from '@playwright/test';

test.use({ storageState: { cookies: [], origins: [] } });

const BASE = 'http://localhost:3000/api/v1';

/** Get an admin JWT token for API calls */
async function getAdminToken(): Promise<string> {
  const res = await fetch(`${BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ login: 'admin', password: 'Dem0P@ssword!!' }),
  });
  const data = await res.json();
  return data.token;
}

/** Enable user registration via admin settings API */
async function enableRegistration(token: string) {
  await fetch(`${BASE}/admin/settings`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ user_signup: ':allowed' }),
  });
}

/** Disable user registration via admin settings API */
async function disableRegistration(token: string) {
  await fetch(`${BASE}/admin/settings`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ user_signup: ':not_allowed' }),
  });
}

/** Delete a user by username via admin API */
async function deleteUser(token: string, username: string) {
  // Find user by listing, then delete
  const res = await fetch(`${BASE}/admin/users`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (res.ok) {
    const users = await res.json() as { id: number; username: string }[];
    const user = users.find(u => u.username === username);
    if (user) {
      await fetch(`${BASE}/admin/users/${user.id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });
    }
  }
}

const signUpButton = /sign up/i;

test.describe('Registration — sad paths', () => {
  let adminToken: string;

  test.beforeAll(async () => {
    adminToken = await getAdminToken();
    await enableRegistration(adminToken);
  });

  test.afterAll(async () => {
    await disableRegistration(adminToken);
  });

  test.beforeEach(async ({ page }) => {
    await page.goto('/register');
  });

  // Helper: bypass browser form validation
  async function disableBrowserValidation(page: import('@playwright/test').Page) {
    await page.locator('form').evaluate((form) => form.setAttribute('novalidate', ''));
  }

  test('empty form submission is blocked by browser validation', async ({ page }) => {
    await expect(page.getByLabel(/username/i)).toHaveAttribute('required', '');
    await page.getByRole('button', { name: signUpButton }).click();
    // Still on form, no server error
    await expect(page.getByLabel(/username/i)).toBeVisible();
  });

  test('missing username — server returns error', async ({ page }) => {
    await page.getByLabel(/email/i).fill('user@example.com');
    await page.getByLabel('Password', { exact: true }).fill('ValidP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('ValidP@ssw0rd!!');
    await disableBrowserValidation(page);
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/required/i)).toBeVisible({ timeout: 10000 });
  });

  test('missing email — server returns error', async ({ page }) => {
    await page.getByLabel(/username/i).fill('newuser');
    await page.getByLabel('Password', { exact: true }).fill('ValidP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('ValidP@ssw0rd!!');
    await disableBrowserValidation(page);
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/required/i)).toBeVisible({ timeout: 10000 });
  });

  test('missing password — server returns error', async ({ page }) => {
    await page.getByLabel(/username/i).fill('newuser');
    await page.getByLabel(/email/i).fill('newuser@example.com');
    await disableBrowserValidation(page);
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/required/i)).toBeVisible({ timeout: 10000 });
  });

  test('password too short (5 characters)', async ({ page }) => {
    await page.getByLabel(/username/i).fill('shortpwuser');
    await page.getByLabel(/email/i).fill('shortpw@example.com');
    await page.getByLabel('Password', { exact: true }).fill('Ab1!x');
    await page.getByLabel(/confirm password/i).fill('Ab1!x');
    await disableBrowserValidation(page);
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/password/i).filter({ hasText: /complexity|at least|short|requirements/i })).toBeVisible({ timeout: 10000 });
  });

  test('password missing uppercase letter', async ({ page }) => {
    await page.getByLabel(/username/i).fill('noupperuser');
    await page.getByLabel(/email/i).fill('noupper@example.com');
    await page.getByLabel('Password', { exact: true }).fill('abcdefghijklmn1!');
    await page.getByLabel(/confirm password/i).fill('abcdefghijklmn1!');
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/password/i).filter({ hasText: /complexity|uppercase|requirements/i })).toBeVisible({ timeout: 10000 });
  });

  test('password missing lowercase letter', async ({ page }) => {
    await page.getByLabel(/username/i).fill('noloweruser');
    await page.getByLabel(/email/i).fill('nolower@example.com');
    await page.getByLabel('Password', { exact: true }).fill('ABCDEFGHIJKLMN1!');
    await page.getByLabel(/confirm password/i).fill('ABCDEFGHIJKLMN1!');
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/password/i).filter({ hasText: /complexity|lowercase|requirements/i })).toBeVisible({ timeout: 10000 });
  });

  test('password missing digit', async ({ page }) => {
    await page.getByLabel(/username/i).fill('nodigituser');
    await page.getByLabel(/email/i).fill('nodigit@example.com');
    await page.getByLabel('Password', { exact: true }).fill('Abcdefghijklmn!');
    await page.getByLabel(/confirm password/i).fill('Abcdefghijklmn!');
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/password/i).filter({ hasText: /complexity|digit|requirements/i })).toBeVisible({ timeout: 10000 });
  });

  test('password missing special character', async ({ page }) => {
    await page.getByLabel(/username/i).fill('nosymboluser');
    await page.getByLabel(/email/i).fill('nosymbol@example.com');
    await page.getByLabel('Password', { exact: true }).fill('Abcdefghijklmn1');
    await page.getByLabel(/confirm password/i).fill('Abcdefghijklmn1');
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/password/i).filter({ hasText: /complexity|special|symbol|requirements/i })).toBeVisible({ timeout: 10000 });
  });

  test('password and confirm password do not match', async ({ page }) => {
    await page.getByLabel(/username/i).fill('mismatchuser');
    await page.getByLabel(/email/i).fill('mismatch@example.com');
    await page.getByLabel('Password', { exact: true }).fill('ValidP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('DifferentP@ss1!');
    await page.getByRole('button', { name: signUpButton }).click();
    // Caught client-side
    await expect(page.getByText(/passwords do not match/i)).toBeVisible();
  });

  test('duplicate username shows conflict error', async ({ page }) => {
    await page.getByLabel(/username/i).fill('admin');
    await page.getByLabel(/email/i).fill('unique-reg-test@example.com');
    await page.getByLabel('Password', { exact: true }).fill('ValidP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('ValidP@ssw0rd!!');
    await page.getByRole('button', { name: signUpButton }).click();
    await expect(page.getByText(/already taken/i)).toBeVisible({ timeout: 10000 });
  });

  test('invalid email format is blocked by browser validation', async ({ page }) => {
    await page.getByLabel(/username/i).fill('bademailuser');
    await page.getByLabel(/email/i).fill('notanemail');
    await page.getByLabel('Password', { exact: true }).fill('ValidP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('ValidP@ssw0rd!!');
    await page.getByRole('button', { name: signUpButton }).click();
    // Browser type="email" blocks submission
    await expect(page.getByLabel(/username/i)).toBeVisible();
    await expect(page.getByText(/account created/i)).not.toBeVisible();
  });

  test('XSS payload in username is not executed', async ({ page }) => {
    const xssPayload = '<script>alert(1)</script>';
    await page.getByLabel(/username/i).fill(xssPayload);
    await page.getByLabel(/email/i).fill('xss-test@example.com');
    await page.getByLabel('Password', { exact: true }).fill('ValidP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('ValidP@ssw0rd!!');

    let dialogFired = false;
    page.on('dialog', () => { dialogFired = true; });

    await page.getByRole('button', { name: signUpButton }).click();

    // Wait for any response from server
    await page.waitForTimeout(2000);

    // XSS must not have fired
    expect(dialogFired).toBe(false);
    await expect(page.locator('script:text("alert(1)")')).toHaveCount(0);

    // Cleanup if user was created
    await deleteUser(adminToken, xssPayload);
  });

  test('very long username (1000+ chars)', async ({ page }) => {
    const longUsername = 'a'.repeat(1001);
    await page.getByLabel(/username/i).fill(longUsername);
    await page.getByLabel(/email/i).fill('longuser@example.com');
    await page.getByLabel('Password', { exact: true }).fill('ValidP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('ValidP@ssw0rd!!');

    await page.getByRole('button', { name: signUpButton }).click();

    // Should get some response — either error or success — without crashing
    await page.waitForTimeout(3000);
    // Page should still be functional
    await expect(page.getByRole('heading', { name: /fat free crm/i })).toBeVisible();

    // Cleanup if user was created
    await deleteUser(adminToken, longUsername);
  });

  test('registration disabled shows appropriate error', async ({ page }) => {
    // Temporarily disable registration
    await disableRegistration(adminToken);

    await page.goto('/register');
    await page.getByLabel(/username/i).fill('blockeduser');
    await page.getByLabel(/email/i).fill('blocked@example.com');
    await page.getByLabel('Password', { exact: true }).fill('ValidP@ssw0rd!!');
    await page.getByLabel(/confirm password/i).fill('ValidP@ssw0rd!!');
    await page.getByRole('button', { name: signUpButton }).click();

    await expect(page.getByText(/not allowed/i)).toBeVisible({ timeout: 10000 });

    // Re-enable for remaining tests
    await enableRegistration(adminToken);
  });
});
