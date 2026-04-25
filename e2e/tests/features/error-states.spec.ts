import { test, expect } from '../../fixtures/auth';

test.describe('Error States', () => {
  test('unknown route redirects to dashboard', async ({ page }) => {
    await page.goto('/nonexistent-page');
    await page.waitForLoadState('networkidle');
    // App.tsx has a catch-all: <Route path="*" element={<Navigate to="/" replace />} />
    await expect(page).toHaveURL(/\/$/);
  });

  test('entity detail with non-existent ID shows error', async ({ page }) => {
    await page.goto('/accounts/999999');
    await page.waitForLoadState('networkidle');
    // EntityDetailPage renders error in a red alert box when API returns 404
    const errorBanner = page.locator('.bg-red-50');
    const loadingText = page.getByText('Loading...');
    // Should show error, not be stuck on loading spinner
    await expect(loadingText).not.toBeVisible({ timeout: 10000 });
    await expect(errorBanner).toBeVisible();
  });

  test('entity detail with string ID shows error', async ({ page }) => {
    await page.goto('/accounts/abc');
    await page.waitForLoadState('networkidle');
    const errorBanner = page.locator('.bg-red-50');
    const loadingText = page.getByText('Loading...');
    await expect(loadingText).not.toBeVisible({ timeout: 10000 });
    await expect(errorBanner).toBeVisible();
  });

  test('API error is displayed to user on list page', async ({ page }) => {
    // Intercept API call and force a server error
    await page.route('**/api/v1/accounts*', (route) => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' }),
      });
    });
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');
    // EntityList renders error in a red alert box
    const errorBanner = page.locator('.bg-red-50');
    await expect(errorBanner).toBeVisible({ timeout: 10000 });
  });

  test('API error is displayed on form submission', async ({ page, api }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    // Open the create form
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new account/i)).toBeVisible();

    // Intercept the POST to simulate a server error
    await page.route('**/api/v1/accounts', (route, request) => {
      if (request.method() === 'POST') {
        route.fulfill({
          status: 422,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Validation failed' }),
        });
      } else {
        route.continue();
      }
    });

    const modal = page.locator('.fixed.inset-0');
    const { fillField } = await import('../../fixtures/helpers');
    await fillField(page, 'Name', 'Will Fail', modal);
    await page.getByRole('button', { name: 'Create' }).click();
    await page.waitForTimeout(1000);

    // Modal should still be open (form not closed on error)
    await expect(modal).toBeVisible();
  });
});
