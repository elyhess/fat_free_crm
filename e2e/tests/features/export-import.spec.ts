import { test, expect } from '../../fixtures/auth';

test.describe('Export & Import', () => {
  test('import template endpoint works', async ({ api }) => {
    const res = await api.getRaw('/accounts/import/template');
    expect(res.ok).toBeTruthy();
  });

  test('export accounts', async ({ api }) => {
    const res = await api.getRaw('/accounts/export');
    expect(res.ok).toBeTruthy();
  });

  test('export contacts', async ({ api }) => {
    const res = await api.getRaw('/contacts/export');
    expect(res.ok).toBeTruthy();
  });

  test('export leads', async ({ api }) => {
    const res = await api.getRaw('/leads/export');
    expect(res.ok).toBeTruthy();
  });

  test('export opportunities', async ({ api }) => {
    const res = await api.getRaw('/opportunities/export');
    expect(res.ok).toBeTruthy();
  });

  test('export campaigns', async ({ api }) => {
    const res = await api.getRaw('/campaigns/export');
    expect(res.ok).toBeTruthy();
  });

  test('export tasks', async ({ api }) => {
    const res = await api.getRaw('/tasks/export');
    expect(res.ok).toBeTruthy();
  });
});
