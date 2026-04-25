import { test, expect } from '../../fixtures/auth';

test.describe('Campaign Detail', () => {
  let campaignId: number;
  let campaignName: string;

  test.beforeEach(async ({ api, page }) => {
    if (!campaignId) {
      const campaign = await api.createCampaign({ name: `Detail Campaign ${Date.now()}`, status: 'planned' });
      campaignId = (campaign as Record<string, unknown>).id as number;
      campaignName = (campaign as Record<string, unknown>).name as string;
    }
    await page.goto(`/campaigns/${campaignId}`);
    await page.waitForLoadState('networkidle');
  });

  test('shows campaign details', async ({ page }) => {
    await expect(page.getByRole('heading', { name: campaignName })).toBeVisible();
  });

  test('edit from detail page', async ({ page }) => {
    await page.getByRole('button', { name: /edit/i }).first().click();
    await page.waitForTimeout(500);
    await expect(page.getByText(/edit campaign/i)).toBeVisible();
    await page.getByRole('button', { name: /cancel/i }).click();
  });

  test('back link returns to list', async ({ page }) => {
    await page.getByRole('button', { name: /back/i }).click();
    await expect(page).toHaveURL('/campaigns');
  });

  test('shows related sections', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /leads/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /opportunities/i })).toBeVisible();
  });

  test('comments and tags sections visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /comments/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /tags/i })).toBeVisible();
  });

  test('subscribe button visible', async ({ page }) => {
    await expect(page.getByRole('button', { name: /subscribe/i })).toBeVisible();
  });
});
