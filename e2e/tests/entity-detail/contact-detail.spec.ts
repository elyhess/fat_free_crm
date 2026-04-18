import { test, expect } from '../../fixtures/auth';

test.describe('Contact Detail', () => {
  let contactId: number;

  test.beforeEach(async ({ api, page }) => {
    if (!contactId) {
      const contact = await api.createContact({ first_name: 'DetailFirst', last_name: 'DetailLast', email: 'contact@test.com' });
      contactId = (contact as Record<string, unknown>).id as number;
    }
    await page.goto(`/contacts/${contactId}`);
    await page.waitForLoadState('networkidle');
  });

  test('shows contact details', async ({ page }) => {
    await expect(page.getByText('DetailFirst').first()).toBeVisible();
    await expect(page.getByText('DetailLast').first()).toBeVisible();
  });

  test('edit from detail page', async ({ page }) => {
    await page.getByRole('button', { name: /edit/i }).first().click();
    await page.waitForTimeout(500);
    await expect(page.getByText(/edit contact/i)).toBeVisible();
    await page.getByRole('button', { name: /cancel/i }).click();
  });

  test('back link returns to list', async ({ page }) => {
    await page.getByRole('button', { name: /back/i }).click();
    await expect(page).toHaveURL('/contacts');
  });

  test('shows related opportunities section', async ({ page }) => {
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
