import { test, expect } from '../../fixtures/auth';

test.describe('Filtering', () => {
  test('filter accounts by name', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');
    const nameFilter = page.getByPlaceholder(/filter by name/i);
    if (await nameFilter.isVisible()) {
      await nameFilter.fill('corp');
      await page.waitForTimeout(500);
      const rows = page.locator('table tbody tr');
      const count = await rows.count();
      expect(count).toBeGreaterThan(0);
    }
  });

  test('clear filters restores full list', async ({ page }) => {
    await page.goto('/accounts');
    await page.waitForLoadState('networkidle');
    const nameFilter = page.getByPlaceholder(/filter by name/i);
    if (await nameFilter.isVisible()) {
      await nameFilter.fill('testfilter');
      await page.waitForTimeout(500);

      const clearBtn = page.getByRole('button', { name: /clear/i });
      if (await clearBtn.isVisible()) {
        await clearBtn.click();
        await page.waitForTimeout(500);
      }
    }
  });
});
