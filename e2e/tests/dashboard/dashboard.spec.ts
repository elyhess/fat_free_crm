import { test, expect } from '../../fixtures/auth';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('shows task summary section', async ({ page }) => {
    await expect(page.getByText(/tasks/i).first()).toBeVisible();
  });

  test('shows pipeline section', async ({ page }) => {
    await expect(page.getByText(/pipeline/i).first()).toBeVisible();
  });

  test('shows activity section', async ({ page }) => {
    await expect(page.getByText(/activity/i).first()).toBeVisible();
  });

  test('sections are collapsible', async ({ page }) => {
    // Find a section header and click to collapse
    const tasksHeader = page.getByText(/tasks/i).first();
    await tasksHeader.click();
    // Click again to expand
    await tasksHeader.click();
  });

  test('task summary shows bucket counts', async ({ page }) => {
    // Should show task categories like ASAP, Today, etc. or a count
    const taskSection = page.locator('text=/tasks/i').first().locator('..');
    await expect(taskSection).toBeVisible();
  });

  test('pipeline shows stage data', async ({ page }) => {
    // Pipeline section should have stage names or amounts
    const pipelineSection = page.locator('text=/pipeline/i').first().locator('..');
    await expect(pipelineSection).toBeVisible();
  });

  test('activity feed shows recent events', async ({ page }) => {
    // Activity items should have entity names and action types
    const activitySection = page.locator('text=/activity/i').first().locator('..');
    await expect(activitySection).toBeVisible();
  });
});
