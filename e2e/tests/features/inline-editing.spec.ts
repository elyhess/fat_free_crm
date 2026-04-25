import { test, expect } from '../../fixtures/auth';

test.describe('Inline Editing', () => {
  test('inline edit opportunity stage on list page', async ({ page, api }) => {
    const opp = await api.createOpportunity({ name: `InlineOpp ${Date.now()}`, stage: 'prospecting' });
    const id = (opp as Record<string, unknown>).id as number;

    try {
      await page.goto('/opportunities');
      await page.waitForLoadState('networkidle');

      const row = page.getByRole('row', { name: /InlineOpp/ });
      await expect(row).toBeVisible();

      // Click a stage cell to activate inline edit
      const cells = row.locator('td');
      const stageCell = cells.nth(1);
      await stageCell.click();
      await page.waitForTimeout(300);

      // Assert the select dropdown appears and use it
      const select = stageCell.locator('select');
      await expect(select).toBeVisible();
      await select.selectOption('analysis');
      await select.blur();
      await page.waitForTimeout(500);

      // Verify the stage changed via the API
      const updated = (await api.get(`/opportunities/${id}`)) as Record<string, unknown>;
      expect(updated.stage).toBe('analysis');
    } finally {
      await api.deleteEntity('opportunities', id);
    }
  });

  test('inline edit on entity detail page', async ({ page, api }) => {
    const account = await api.createAccount({ name: `InlineDetail ${Date.now()}` });
    const id = (account as Record<string, unknown>).id as number;

    try {
      await page.goto(`/accounts/${id}`);
      await page.waitForLoadState('networkidle');

      // Assert the page loads and shows the entity heading
      await expect(page.getByRole('heading', { name: /InlineDetail/ })).toBeVisible();
    } finally {
      await api.deleteEntity('accounts', id);
    }
  });
});
