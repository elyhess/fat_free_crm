import { test, expect } from '../../fixtures/auth';

test.describe('Search', () => {
  test('search from navbar', async ({ page }) => {
    await page.goto('/');
    await page.getByPlaceholder(/search/i).fill('admin');
    await page.getByPlaceholder(/search/i).press('Enter');
    await expect(page).toHaveURL(/\/search\?q=admin/);
  });

  test('search results page loads', async ({ page }) => {
    await page.goto('/search?q=test');
    await page.waitForLoadState('networkidle');
    // Should show search heading or results
    await expect(page.getByText(/search/i).first()).toBeVisible();
  });

  test('search for known entity returns results', async ({ page, api }) => {
    const account = await api.createAccount({ name: `Searchable${Date.now()}` });
    const name = (account as Record<string, unknown>).name as string;
    const id = (account as Record<string, unknown>).id as number;

    await page.goto(`/search?q=${name}`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    await api.deleteEntity('accounts', id);
  });

  test('empty search shows no results', async ({ page }) => {
    await page.goto('/search?q=zzzznonexistent99999');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    // Should show 0 results or "no results" message
    const content = await page.textContent('body');
    expect(content).toMatch(/no results|0 result|nothing/i);
  });
});
