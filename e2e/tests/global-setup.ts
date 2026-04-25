import { test as setup } from '@playwright/test';
import { ApiClient } from '../fixtures/api-client';
import * as fs from 'fs';
import * as path from 'path';

setup('authenticate as admin', async ({ page }) => {
  const { token, user } = await ApiClient.login('admin', 'Dem0P@ssword!!');

  // Set localStorage via the page
  await page.goto('http://localhost:3000/login');
  await page.evaluate(({ token, user }) => {
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(user));
  }, { token, user });

  // Save storage state
  const dir = path.join(__dirname, '..', 'auth-state');
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  await page.context().storageState({ path: path.join(dir, 'admin.json') });
});
