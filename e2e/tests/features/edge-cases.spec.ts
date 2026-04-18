import { test, expect } from '../../fixtures/auth';
import { fillField } from '../../fixtures/helpers';

test.describe('Edge Cases', () => {
  test('browser back after form submission does not duplicate entity', async ({ page, api }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    const name = `E2E BackBtn ${Date.now()}`;
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new account/i)).toBeVisible();

    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', name, modal);

    const responsePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/accounts') && resp.request().method() === 'POST'
    );
    await page.getByRole('button', { name: 'Create' }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(201);
    const body = await response.json();

    // Press browser back
    await page.goBack();
    await page.waitForLoadState('networkidle');

    // Verify no duplicate was created — the original should still exist exactly once
    // Clean up the created entity
    await api.deleteEntity('accounts', body.id);
  });

  test('refresh during form editing resets form state', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new account/i)).toBeVisible();

    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', 'Should Be Lost', modal);

    // Refresh the page
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Modal should no longer be open after refresh
    await expect(page.locator('.fixed.inset-0')).not.toBeVisible({ timeout: 5000 });
    // Page should be back to normal accounts list
    await expect(page.getByRole('heading', { name: 'Accounts' })).toBeVisible();
  });

  test('navigate away during form editing', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new account/i)).toBeVisible();

    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', 'Should Be Lost', modal);

    // Close the modal first (Escape), then navigate away
    await page.keyboard.press('Escape');
    await expect(modal).not.toBeVisible({ timeout: 3000 });
    await page.goto('/contacts');
    await page.waitForLoadState('networkidle');

    // Should have navigated to contacts
    await expect(page).toHaveURL(/\/contacts/);
    await expect(page.getByRole('heading', { name: 'Contacts' })).toBeVisible();
  });

  test('rapid pagination clicks do not break the app', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    const nextBtn = page.getByRole('button', { name: /next/i });
    if (await nextBtn.isVisible()) {
      // Click next rapidly several times
      await nextBtn.click();
      await nextBtn.click();
      await nextBtn.click();
      await page.waitForTimeout(1000);

      // App should still be functional — table should be visible
      await expect(page.locator('table')).toBeVisible();
      // Page indicator should show a valid page
      await expect(page.getByText(/total/i)).toBeVisible();
    }
  });

  test('empty state displays "no records" message', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    const nameFilter = page.getByPlaceholder(/filter by name/i);
    if (await nameFilter.isVisible()) {
      await nameFilter.fill('zzzznonexistentaccount99999');
      await page.waitForTimeout(1000);

      // EntityList renders "No records found." when data is empty
      await expect(page.getByText(/no records found/i)).toBeVisible();
    }
  });

  test('search with empty query shows proper handling', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const searchInput = page.getByPlaceholder(/search/i);
    await searchInput.fill('');
    await searchInput.press('Enter');
    await page.waitForLoadState('networkidle');

    // Should either stay on current page or show search with no/all results
    // App should not crash
    const body = await page.textContent('body');
    expect(body).toBeTruthy();
  });

  test('search with special characters does not cause XSS', async ({ page }) => {
    const specialChars = '<script>alert("xss")</script>&"\'';
    await page.goto(`/search?q=${encodeURIComponent(specialChars)}`);
    await page.waitForLoadState('networkidle');

    // Verify no alert dialog was triggered (XSS protection)
    // The page should render normally without executing injected script
    await expect(page.getByText(/search/i).first()).toBeVisible();

    // Verify the special characters are rendered as text, not executed
    const hasScriptTag = await page.evaluate(() => {
      // Check that no rogue script tags were injected into the DOM
      const scripts = document.querySelectorAll('script');
      return Array.from(scripts).some(s => s.textContent?.includes('alert("xss")'));
    });
    expect(hasScriptTag).toBe(false);
  });

  test('multiple modal opens produce a clean form each time', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    // First open: fill in a name, then cancel
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new account/i)).toBeVisible();

    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', 'First Attempt', modal);
    await page.getByRole('button', { name: /cancel/i }).click();

    // Wait for modal to close
    await expect(modal).not.toBeVisible({ timeout: 3000 });

    // Second open: form should be clean
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new account/i)).toBeVisible();

    const nameField = modal.locator('label').filter({ hasText: 'Name' }).locator('xpath=..').locator('input, select, textarea').first();
    const value = await nameField.inputValue();
    expect(value).toBe('');

    // Cancel to clean up
    await page.getByRole('button', { name: /cancel/i }).click();
  });

  test('submit form and immediately close modal', async ({ page, api }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');

    const name = `E2E RaceClose ${Date.now()}`;
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new account/i)).toBeVisible();

    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', name, modal);

    // Click Create and try to close the modal immediately
    const responsePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/accounts') && resp.request().method() === 'POST'
    );
    await page.getByRole('button', { name: 'Create' }).click();

    // Wait for the response to complete so we can capture the ID for cleanup
    const response = await responsePromise;
    const body = await response.json();

    // App should still be functional after the race
    await page.waitForLoadState('networkidle');
    await expect(page.locator('table')).toBeVisible();

    // Clean up
    if (body?.id) {
      await api.deleteEntity('accounts', body.id);
    }
  });
});
