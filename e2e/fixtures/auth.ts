import { test as base, type Page } from '@playwright/test';
import { ApiClient } from './api-client';

type Fixtures = {
  api: ApiClient;
  adminApi: ApiClient;
  demoPage: Page;
};

export const test = base.extend<Fixtures>({
  api: async ({}, use) => {
    const { token } = await ApiClient.login('admin', 'Dem0P@ssword!!');
    await use(new ApiClient(token));
  },
  adminApi: async ({}, use) => {
    const { token } = await ApiClient.login('admin', 'Dem0P@ssword!!');
    await use(new ApiClient(token));
  },
  demoPage: async ({ browser, api }, use) => {
    // Ensure the non-admin test user exists
    const { username, password } = await api.ensureNonAdminUser();
    const { token, user } = await ApiClient.login(username, password);
    const context = await browser.newContext({
      storageState: {
        cookies: [],
        origins: [{
          origin: 'http://localhost:3000',
          localStorage: [
            { name: 'token', value: token },
            { name: 'user', value: JSON.stringify(user) },
          ],
        }],
      },
    });
    const page = await context.newPage();
    await use(page);
    await context.close();
  },
});

export { expect } from '@playwright/test';
