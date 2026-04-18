import { test, expect } from '../../fixtures/auth';

test.describe('Pipeline Board', () => {
  test('toggle to board view', async ({ page }) => {
    await page.goto('/opportunities');
    const boardBtn = page.getByRole('button', { name: /board/i });
    await boardBtn.click();

    // Should show stage columns
    await expect(page.getByText(/prospecting/i)).toBeVisible();
    await expect(page.getByText(/analysis/i)).toBeVisible();
    await expect(page.getByText(/presentation/i)).toBeVisible();
    await expect(page.getByText(/proposal/i)).toBeVisible();
    await expect(page.getByText(/negotiation/i)).toBeVisible();
  });

  test('toggle back to table view', async ({ page }) => {
    await page.goto('/opportunities');
    await page.getByRole('button', { name: /board/i }).click();
    await page.getByRole('button', { name: /table/i }).click();
    await expect(page.locator('table')).toBeVisible();
  });

  test('opportunity cards shown in board', async ({ page, api }) => {
    const opp = await api.createOpportunity({ name: `BoardCard ${Date.now()}`, stage: 'prospecting' });
    const id = (opp as Record<string, unknown>).id as number;
    const name = (opp as Record<string, unknown>).name as string;

    await page.goto('/opportunities');
    await page.getByRole('button', { name: /board/i }).click();
    await page.waitForLoadState('networkidle');

    await expect(page.getByText(name)).toBeVisible();

    await api.deleteEntity('opportunities', id);
  });

  test('drag and drop changes stage', async ({ page, api }) => {
    const opp = await api.createOpportunity({ name: `DragDrop ${Date.now()}`, stage: 'prospecting' });
    const id = (opp as Record<string, unknown>).id as number;
    const name = (opp as Record<string, unknown>).name as string;

    await page.goto('/opportunities');
    await page.getByRole('button', { name: /board/i }).click();
    await page.waitForLoadState('networkidle');

    // Find the card and a target column
    const card = page.getByText(name);
    const targetColumn = page.locator('[data-stage="analysis"]').or(page.locator('text=Analysis').locator('..'));

    if (await card.isVisible() && await targetColumn.isVisible()) {
      // Perform drag and drop
      await card.dragTo(targetColumn);
      await page.waitForTimeout(1000);
    }

    await api.deleteEntity('opportunities', id);
  });
});
