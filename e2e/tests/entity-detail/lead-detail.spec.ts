import { test, expect } from '../../fixtures/auth';

test.describe('Lead Detail', () => {
  let leadId: number;

  test.beforeEach(async ({ api, page }) => {
    if (!leadId) {
      const lead = await api.createLead({ first_name: 'LeadFirst', last_name: 'LeadLast', status: 'new', company: 'TestCo' });
      leadId = (lead as Record<string, unknown>).id as number;
    }
    await page.goto(`/leads/${leadId}`);
    await page.waitForLoadState('networkidle');
  });

  test('shows lead details', async ({ page }) => {
    await expect(page.getByText('LeadFirst').first()).toBeVisible();
    await expect(page.getByText('LeadLast').first()).toBeVisible();
  });

  test('edit from detail page', async ({ page }) => {
    await page.getByRole('button', { name: /edit/i }).first().click();
    await page.waitForTimeout(500);
    await expect(page.getByText(/edit lead/i)).toBeVisible();
    await page.getByRole('button', { name: /cancel/i }).click();
  });

  test('convert button visible for new leads', async ({ page }) => {
    await expect(page.getByRole('button', { name: /convert/i })).toBeVisible();
  });

  test('action buttons visible', async ({ page }) => {
    // Lead may have reject, convert, or other action buttons depending on status
    const rejectBtn = page.getByRole('button', { name: /reject/i });
    const convertBtn = page.getByRole('button', { name: /convert/i });
    const hasReject = await rejectBtn.isVisible().catch(() => false);
    const hasConvert = await convertBtn.isVisible().catch(() => false);
    expect(hasReject || hasConvert).toBeTruthy();
  });

  test('back link returns to list', async ({ page }) => {
    await page.getByRole('button', { name: /back/i }).click();
    await expect(page).toHaveURL('/leads');
  });

  test('comments and tags sections visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /comments/i })).toBeVisible();
    await expect(page.getByRole('heading', { name: /tags/i })).toBeVisible();
  });
});
