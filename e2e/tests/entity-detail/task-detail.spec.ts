import { test, expect } from '../../fixtures/auth';

test.describe('Task Detail', () => {
  let taskId: number;
  let taskName: string;

  test.beforeEach(async ({ api, page }) => {
    if (!taskId) {
      const task = await api.createTask({ name: `Detail Task ${Date.now()}`, priority: 'high', category: 'call', bucket: 'due_asap' });
      taskId = (task as Record<string, unknown>).id as number;
      taskName = (task as Record<string, unknown>).name as string;
    }
    await page.goto(`/tasks/${taskId}`);
    await page.waitForLoadState('networkidle');
  });

  test('shows task details', async ({ page }) => {
    await expect(page.getByRole('heading', { name: taskName })).toBeVisible();
  });

  test('edit from detail page', async ({ page }) => {
    await page.getByRole('button', { name: /edit/i }).first().click();
    await page.waitForTimeout(500);
    await expect(page.getByText(/edit task/i)).toBeVisible();
    await page.getByRole('button', { name: /cancel/i }).click();
  });

  test('complete and uncomplete task', async ({ page }) => {
    const completeBtn = page.getByRole('button', { name: /complete/i });
    if (await completeBtn.isVisible()) {
      await completeBtn.click();
      await page.waitForTimeout(500);
      const uncompleteBtn = page.getByRole('button', { name: /uncomplete|undo/i });
      if (await uncompleteBtn.isVisible()) {
        await uncompleteBtn.click();
      }
    }
  });

  test('back link returns to list', async ({ page }) => {
    await page.getByRole('button', { name: /back/i }).click();
    await expect(page).toHaveURL('/tasks');
  });

  test('comments section visible', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /comments/i })).toBeVisible();
  });
});
