import { test, expect } from '../../fixtures/auth';
import { fillField } from '../../fixtures/helpers';

test.describe('Accounts List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');
  });

  test('page loads with heading and table', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Accounts' })).toBeVisible();
    await expect(page.locator('table')).toBeVisible();
  });

  test('table has expected columns', async ({ page }) => {
    for (const col of ['Name', 'Email', 'Phone']) {
      await expect(page.locator('th', { hasText: col })).toBeVisible();
    }
  });

  test('create new account via form', async ({ page, api }) => {
    const name = `E2E Create ${Date.now()}`;
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
    await api.deleteEntity('accounts', body.id);
  });

  test('edit existing account', async ({ page, api }) => {
    const account = await api.createAccount({ name: `E2E Edit ${Date.now()}` });
    const id = (account as Record<string, unknown>).id as number;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: /E2E Edit/ });
    await row.getByRole('button', { name: /edit/i }).first().click();

    const newName = `E2E Edited ${Date.now()}`;
    const modal = page.locator('.fixed.inset-0');
    await fillField(page, 'Name', newName, modal);

    const responsePromise = page.waitForResponse(resp =>
      resp.url().includes(`/api/v1/accounts/${id}`) && resp.request().method() === 'PUT'
    );
    await page.getByRole('button', { name: 'Update' }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(200);

    await api.deleteEntity('accounts', id);
  });

  test('delete account', async ({ page, api }) => {
    const account = await api.createAccount({ name: `E2E Delete ${Date.now()}` });
    const id = (account as Record<string, unknown>).id as number;

    await page.reload();
    await page.waitForLoadState('networkidle');
    const row = page.getByRole('row', { name: /E2E Delete/ });
    await row.getByRole('button', { name: /delete/i }).click();

    const responsePromise = page.waitForResponse(resp =>
      resp.url().includes(`/api/v1/accounts/${id}`) && resp.request().method() === 'DELETE'
    );
    await page.getByRole('button', { name: /confirm|delete|yes/i }).last().click();
    const response = await responsePromise;
    expect(response.status()).toBe(200);
  });

  test('clicking account name navigates to detail', async ({ page }) => {
    const firstLink = page.locator('table tbody tr').first().locator('a').first();
    await firstLink.click();
    await expect(page).toHaveURL(/\/accounts\/\d+/);
  });

  test('sorting by column', async ({ page }) => {
    await page.locator('th', { hasText: 'Name' }).click();
    await page.waitForTimeout(500);
  });

  test('pagination controls visible', async ({ page }) => {
    await expect(page.getByText(/total/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
  });

  test('pagination navigates between pages', async ({ page }) => {
    const firstRow = await page.locator('table tbody tr').first().locator('a').first().textContent();
    await page.getByRole('button', { name: /next/i }).click();
    await page.waitForTimeout(500);
    const secondRow = await page.locator('table tbody tr').first().locator('a').first().textContent();
    expect(firstRow).not.toEqual(secondRow);
  });
});
