import { test, expect } from '../../fixtures/auth';
import { fillField } from '../../fixtures/helpers';

test.describe('Contacts List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/contacts');
    await page.waitForLoadState('networkidle');
  });

  test('page loads with heading and table', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Contacts' })).toBeVisible();
    await expect(page.locator('table')).toBeVisible();
  });

  test('create new contact', async ({ page, api }) => {
    const ts = Date.now();
    await page.getByRole('button', { name: /new/i }).click();
    await expect(page.getByText(/new contact/i)).toBeVisible();
    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'First Name', 'E2EFirst', modal);
    await fillField(page, 'Last Name', `Last${ts}`, modal);

    const responsePromise = page.waitForResponse(resp =>
      resp.url().includes('/api/v1/contacts') && resp.request().method() === 'POST'
    );
    await page.getByRole('button', { name: 'Create' }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(201);

    // Cleanup via API
    const body = await response.json();
    await api.deleteEntity('contacts', body.id);
  });

  test('edit existing contact', async ({ page, api }) => {
    const contact = await api.createContact();
    const id = (contact as Record<string, unknown>).id as number;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: /E2E/ }).first();
    await row.getByRole('button', { name: /edit/i }).click();

    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'First Name', 'Updated', modal);
    await page.getByRole('button', { name: 'Update' }).click();
    await page.waitForTimeout(1000);

    await api.deleteEntity('contacts', id);
  });

  test('delete contact', async ({ page, api }) => {
    const contact = await api.createContact({ first_name: 'DeleteMe', last_name: `Test${Date.now()}` });
    const id = (contact as Record<string, unknown>).id as number;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: /DeleteMe/ });
    await row.getByRole('button', { name: /delete/i }).click();
    await page.getByRole('button', { name: /confirm|delete|yes/i }).last().click();
    await page.waitForTimeout(500);
    await expect(page.getByRole('link', { name: /DeleteMe/ })).toBeHidden();
  });

  test('clicking contact navigates to detail', async ({ page }) => {
    const firstLink = page.locator('table tbody tr').first().locator('a').first();
    await firstLink.click();
    await expect(page).toHaveURL(/\/contacts\/\d+/);
  });
});
