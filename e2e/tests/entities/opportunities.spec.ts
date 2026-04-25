import { test, expect } from '../../fixtures/auth';
import { fillField } from '../../fixtures/helpers';

test.describe('Opportunities List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/opportunities');
    await page.waitForLoadState('networkidle');
  });

  test('page loads with heading and table', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Opportunities' })).toBeVisible();
    await expect(page.locator('table')).toBeVisible();
  });

  test('create new opportunity', async ({ page, api }) => {
    const name = `E2E Opp ${Date.now()}`;
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new opportunit/i)).toBeVisible();
    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', name, modal);
    await fillField(page, 'Stage', 'prospecting', modal);

    const responsePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/opportunities') && resp.request().method() === 'POST'
    );
    await page.getByRole('button', { name: 'Create' }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(201);

    // Cleanup via API
    const body = await response.json();
    await api.deleteEntity('opportunities', body.id);
  });

  test('edit existing opportunity', async ({ page, api }) => {
    const opp = await api.createOpportunity();
    const id = (opp as Record<string, unknown>).id as number;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: /E2E Opp/ }).first();
    await row.getByRole('button', { name: /edit/i }).click();

    const newName = `E2E Edited Opp ${Date.now()}`;
    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', newName, modal);
    await page.getByRole('button', { name: 'Update' }).click();
    await page.waitForTimeout(1000);

    await api.deleteEntity('opportunities', id);
  });

  test('delete opportunity', async ({ page, api }) => {
    const opp = await api.createOpportunity({ name: `DelOpp ${Date.now()}` });
    const id = (opp as Record<string, unknown>).id as number;
    const name = (opp as Record<string, unknown>).name as string;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: new RegExp(name) });
    await row.getByRole('button', { name: /delete/i }).click();
    await page.getByRole('button', { name: /confirm|delete|yes/i }).last().click();
    await page.waitForTimeout(500);
    await expect(page.getByRole('link', { name })).toBeHidden();
  });

  test('table/board view toggle', async ({ page }) => {
    const boardBtn = page.getByRole('button', { name: /board/i });
    if (await boardBtn.isVisible()) {
      await boardBtn.click();
      await page.waitForTimeout(500);
      await expect(page.getByText(/prospecting/i)).toBeVisible();
      await expect(page.getByText(/negotiation/i)).toBeVisible();

      await page.getByRole('button', { name: /table/i }).click();
      await expect(page.locator('table')).toBeVisible();
    }
  });

  test('clicking opportunity navigates to detail', async ({ page }) => {
    const firstLink = page.locator('table tbody tr').first().locator('a').first();
    await firstLink.click();
    await expect(page).toHaveURL(/\/opportunities\/\d+/);
  });
});
