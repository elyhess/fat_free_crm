import { test, expect } from '../../fixtures/auth';
import * as path from 'path';
import * as fs from 'fs';

test.describe('Avatar', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/profile');
  });

  test('avatar section visible', async ({ page }) => {
    await expect(page.getByText(/avatar/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /upload/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /remove/i })).toBeVisible();
  });

  test('upload avatar', async ({ page }) => {
    // Create a minimal test image
    const testImagePath = path.join(__dirname, '..', '..', 'test-results', 'test-avatar.png');
    const dir = path.dirname(testImagePath);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });

    // Minimal valid PNG (1x1 red pixel)
    const png = Buffer.from([
      0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
      0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
      0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
      0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, // 8-bit RGB
      0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, 0x54, // IDAT chunk
      0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00, 0x00, // compressed data
      0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC, 0x33, // checksum
      0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, // IEND chunk
      0xAE, 0x42, 0x60, 0x82,
    ]);
    fs.writeFileSync(testImagePath, png);

    // Upload via file input
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(testImagePath);
    await page.waitForTimeout(1000);

    // Avatar image should update
    const avatarImg = page.locator('img[alt="Avatar"]');
    await expect(avatarImg).toBeVisible();

    // Cleanup
    fs.unlinkSync(testImagePath);
  });

  test('remove avatar', async ({ page }) => {
    await page.getByRole('button', { name: /remove/i }).click();
    await page.waitForTimeout(500);
  });

  test('file type restriction info shown', async ({ page }) => {
    await expect(page.getByText(/png.*jpeg.*gif/i)).toBeVisible();
    await expect(page.getByText(/5mb/i)).toBeVisible();
  });
});
