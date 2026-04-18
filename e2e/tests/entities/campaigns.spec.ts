import { test, expect } from '../../fixtures/auth';
import { fillField } from '../../fixtures/helpers';

test.describe('Campaigns List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/campaigns');
    await page.waitForLoadState('networkidle');
  });

  test('page loads with heading and table', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Campaigns' })).toBeVisible();
    await expect(page.locator('table')).toBeVisible();
  });

  test('create new campaign', async ({ page, api }) => {
    const name = `E2E Campaign ${Date.now()}`;
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new campaign/i)).toBeVisible();
    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', name, modal);

    const responsePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/campaigns') && resp.request().method() === 'POST'
    );
    await page.getByRole('button', { name: 'Create' }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(201);

    // Cleanup via API
    const body = await response.json();
    await api.deleteEntity('campaigns', body.id);
  });

  test('edit existing campaign', async ({ page, api }) => {
    const campaign = await api.createCampaign();
    const id = (campaign as Record<string, unknown>).id as number;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: /E2E Campaign/ }).first();
    await row.getByRole('button', { name: /edit/i }).click();

    const newName = `E2E Edited Campaign ${Date.now()}`;
    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', newName, modal);
    await page.getByRole('button', { name: 'Update' }).click();
    await page.waitForTimeout(1000);

    await api.deleteEntity('campaigns', id);
  });

  test('delete campaign', async ({ page, api }) => {
    const campaign = await api.createCampaign({ name: `DelCampaign ${Date.now()}` });
    const id = (campaign as Record<string, unknown>).id as number;
    const name = (campaign as Record<string, unknown>).name as string;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: new RegExp(name) });
    await row.getByRole('button', { name: /delete/i }).click();
    await page.getByRole('button', { name: /confirm|delete|yes/i }).last().click();
    await page.waitForTimeout(500);
    await expect(page.getByRole('link', { name })).toBeHidden();
  });

  test('clicking campaign navigates to detail', async ({ page }) => {
    const firstLink = page.locator('table tbody tr').first().locator('a').first();
    await firstLink.click();
    await expect(page).toHaveURL(/\/campaigns\/\d+/);
  });
});
