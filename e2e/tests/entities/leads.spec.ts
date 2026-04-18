import { test, expect } from '../../fixtures/auth';
import { fillField } from '../../fixtures/helpers';

test.describe('Leads List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/leads');
    await page.waitForLoadState('networkidle');
  });

  test('page loads with heading and table', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Leads' })).toBeVisible();
    await expect(page.locator('table')).toBeVisible();
  });

  test('create new lead', async ({ page, api }) => {
    const ts = Date.now();
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new lead/i)).toBeVisible();
    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'First Name', 'E2ELead', modal);
    await fillField(page, 'Last Name', `Test${ts}`, modal);

    const responsePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/leads') && resp.request().method() === 'POST'
    );
    await page.getByRole('button', { name: 'Create' }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(201);

    // Cleanup via API
    const body = await response.json();
    await api.deleteEntity('leads', body.id);
  });

  test('edit existing lead', async ({ page, api }) => {
    const lead = await api.createLead();
    const id = (lead as Record<string, unknown>).id as number;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: /E2E/ }).first();
    await row.getByRole('button', { name: /edit/i }).click();

    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Company', 'Updated Corp', modal);
    await page.getByRole('button', { name: 'Update' }).click();
    await page.waitForTimeout(1000);

    await api.deleteEntity('leads', id);
  });

  test('delete lead', async ({ page, api }) => {
    const lead = await api.createLead({ first_name: 'DelLead', last_name: `Test${Date.now()}` });
    const id = (lead as Record<string, unknown>).id as number;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: /DelLead/ });
    await row.getByRole('button', { name: /delete/i }).click();
    await page.getByRole('button', { name: /confirm|delete|yes/i }).last().click();
    await page.waitForTimeout(500);
    await expect(page.getByRole('link', { name: /DelLead/ })).toBeHidden();
  });

  test('clicking lead navigates to detail', async ({ page }) => {
    const firstLink = page.locator('table tbody tr').first().locator('a').first();
    await firstLink.click();
    await expect(page).toHaveURL(/\/leads\/\d+/);
  });
});
