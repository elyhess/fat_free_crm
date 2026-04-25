import { test, expect } from '../../fixtures/auth';

test.describe('Opportunity Detail', () => {
  let oppId: number;
  let oppName: string;

  test.beforeEach(async ({ api, page }) => {
    if (!oppId) {
      const opp = await api.createOpportunity({ name: `Detail Opp ${Date.now()}`, stage: 'prospecting', amount: 50000 });
      oppId = (opp as Record<string, unknown>).id as number;
      oppName = (opp as Record<string, unknown>).name as string;
    }
    await page.goto(`/opportunities/${oppId}`);
    await page.waitForLoadState('networkidle');
  });

  test('shows opportunity details', async ({ page }) => {
    await expect(page.getByRole('heading', { name: oppName })).toBeVisible();
  });

  test('edit from detail page', async ({ page }) => {
    await page.getByRole('button', { name: /edit/i }).first().click();
    await page.waitForTimeout(500);
    await expect(page.getByText(/edit opportunit/i)).toBeVisible();
    await page.getByRole('button', { name: /cancel/i }).click();
  });

  test('back link returns to list', async ({ page }) => {
    await page.getByRole('button', { name: /back/i }).click();
    await expect(page).toHaveURL('/opportunities');
  });

  test('comments and tags sections visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /comments/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /tags/i })).toBeVisible();
  });

  test('subscribe button visible', async ({ page }) => {
    await expect(page.getByRole('button', { name: /subscribe/i })).toBeVisible();
  });
});
