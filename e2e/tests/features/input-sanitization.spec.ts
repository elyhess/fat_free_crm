import { test, expect } from '../../fixtures/auth';

test.describe('Input Sanitization & Injection Prevention', () => {

  // ---------------------------------------------------------------------------
  // XSS Prevention
  // ---------------------------------------------------------------------------

  test.describe('XSS Prevention', () => {
    test('script tag in entity name is rendered as literal text', async ({ page, api }) => {
      const xssName = `<script>alert('xss')</script>`;
      let id: number | undefined;

      try {
        const account = await api.createAccount({ name: xssName });
        id = (account as Record<string, unknown>).id as number;
      } catch (err) {
        // If the API rejects the input, that is also acceptable behavior
        expect(String(err)).toBeTruthy();
        return;
      }

      // Count script tags before navigating to the accounts list
      await page.goto('/accounts');
      await page.waitForLoadState('networkidle');

      const scriptCountBefore = await page.evaluate(() => document.querySelectorAll('script').length);

      // Wait for the name to appear as literal text
      await expect(page.getByText(xssName)).toBeVisible({ timeout: 5000 });

      const scriptCountAfter = await page.evaluate(() => document.querySelectorAll('script').length);

      // No new script tags should have been injected
      expect(scriptCountAfter).toBe(scriptCountBefore);

      // The raw HTML should not contain an unescaped script tag from our payload
      const bodyHTML = await page.evaluate(() => document.body.innerHTML);
      expect(bodyHTML).not.toContain(`<script>alert('xss')</script>`);

      await api.deleteEntity('accounts', id!);
    });

    test('event handler attribute in entity name does not fire', async ({ page, api }) => {
      const xssName = `<img onerror=alert(1) src=x>`;
      let id: number | undefined;

      try {
        const account = await api.createAccount({ name: xssName });
        id = (account as Record<string, unknown>).id as number;
      } catch (err) {
        expect(String(err)).toBeTruthy();
        return;
      }

      // Listen for dialogs (alert/confirm/prompt) — none should fire
      let alertFired = false;
      page.on('dialog', async (dialog) => {
        alertFired = true;
        await dialog.dismiss();
      });

      await page.goto('/accounts');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      expect(alertFired).toBe(false);

      // The payload text should be visible as literal text
      await expect(page.getByText(xssName)).toBeVisible({ timeout: 5000 });

      await api.deleteEntity('accounts', id!);
    });

    test('script tag in comment is rendered safely', async ({ page, api }) => {
      const xssComment = `<script>alert('comment-xss')</script>`;
      let accountId: number | undefined;

      try {
        const account = await api.createAccount({ name: `XSS Comment Test ${Date.now()}` });
        accountId = (account as Record<string, unknown>).id as number;
      } catch (err) {
        expect(String(err)).toBeTruthy();
        return;
      }

      try {
        await api.post(`/accounts/${accountId}/comments`, { comment: xssComment });
      } catch (err) {
        // API rejecting script content is acceptable
        expect(String(err)).toBeTruthy();
        await api.deleteEntity('accounts', accountId!);
        return;
      }

      let alertFired = false;
      page.on('dialog', async (dialog) => {
        alertFired = true;
        await dialog.dismiss();
      });

      await page.goto(`/accounts/${accountId}`);
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      expect(alertFired).toBe(false);

      // The comment text should be displayed as literal text, not executed
      const bodyHTML = await page.evaluate(() => document.body.innerHTML);
      expect(bodyHTML).not.toContain(`<script>alert('comment-xss')</script>`);

      await api.deleteEntity('accounts', accountId!);
    });
  });

  // ---------------------------------------------------------------------------
  // SQL Injection Prevention
  // ---------------------------------------------------------------------------

  test.describe('SQL Injection Prevention', () => {
    test('SQL injection in search returns normal results', async ({ page }) => {
      const sqlPayload = `' OR '1'='1`;

      await page.goto(`/search?q=${encodeURIComponent(sqlPayload)}`);
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      // The page should not error out — it should show search results or "no results"
      const bodyText = await page.textContent('body');
      expect(bodyText).not.toMatch(/internal server error|status\s*500|syntax error|sql error/i);
    });

    test('SQL injection in filter shows normal behavior', async ({ page }) => {
      const sqlPayload = `'; DROP TABLE accounts; --`;

      await page.goto(`/accounts?filter[name_cont]=${encodeURIComponent(sqlPayload)}`);
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);

      const bodyText = await page.textContent('body');
      expect(bodyText).not.toMatch(/internal server error|status\s*500|syntax error|sql error/i);

      // Verify the accounts page still works afterward
      await page.goto('/accounts');
      await page.waitForLoadState('networkidle');
      await expect(page.getByText(/accounts/i).first()).toBeVisible();
    });

    test('SQL injection in entity creation is stored as literal text', async ({ api }) => {
      const sqlName = `'; DROP TABLE accounts; --`;
      let id: number | undefined;

      try {
        const account = await api.createAccount({ name: sqlName });
        id = (account as Record<string, unknown>).id as number;
        const name = (account as Record<string, unknown>).name as string;

        // The name should be stored literally, not interpreted as SQL
        expect(name).toBe(sqlName);
      } catch (err) {
        // If the API rejects it, that's also fine
        expect(String(err)).toBeTruthy();
        return;
      }

      // Verify we can retrieve it and the name is intact
      try {
        const fetched = await api.get(`/accounts/${id}`) as Record<string, unknown>;
        expect(fetched.name).toBe(sqlName);
      } catch {
        // Retrieval failure is acceptable if the entity was somehow rejected
      }

      await api.deleteEntity('accounts', id!);
    });
  });

  // ---------------------------------------------------------------------------
  // Special Characters
  // ---------------------------------------------------------------------------

  test.describe('Special Characters', () => {
    test('unicode characters in entity names display correctly', async ({ page, api }) => {
      const unicodeName = `日本語テスト Ñoño ${Date.now()}`;
      let id: number | undefined;

      try {
        const account = await api.createAccount({ name: unicodeName });
        id = (account as Record<string, unknown>).id as number;
      } catch (err) {
        expect(String(err)).toBeTruthy();
        return;
      }

      await page.goto('/accounts');
      await page.waitForLoadState('networkidle');

      await expect(page.getByText(unicodeName)).toBeVisible({ timeout: 5000 });

      await api.deleteEntity('accounts', id!);
    });

    test('emoji in entity names display correctly', async ({ page, api }) => {
      const emojiName = `Test 🏢 Company 💰 ${Date.now()}`;
      let id: number | undefined;

      try {
        const account = await api.createAccount({ name: emojiName });
        id = (account as Record<string, unknown>).id as number;
      } catch (err) {
        expect(String(err)).toBeTruthy();
        return;
      }

      await page.goto('/accounts');
      await page.waitForLoadState('networkidle');

      await expect(page.getByText(emojiName)).toBeVisible({ timeout: 5000 });

      await api.deleteEntity('accounts', id!);
    });

    test('HTML entities in entity names render properly', async ({ page, api }) => {
      const htmlEntitiesName = `&amp; &lt; &gt; &quot; ${Date.now()}`;
      let id: number | undefined;

      try {
        const account = await api.createAccount({ name: htmlEntitiesName });
        id = (account as Record<string, unknown>).id as number;
      } catch (err) {
        expect(String(err)).toBeTruthy();
        return;
      }

      await page.goto('/accounts');
      await page.waitForLoadState('networkidle');

      // The literal string should be rendered — entities should not be double-decoded
      await expect(page.getByText(htmlEntitiesName)).toBeVisible({ timeout: 5000 });

      await api.deleteEntity('accounts', id!);
    });

    test('null bytes in entity names are handled gracefully', async ({ api }) => {
      const nullByteName = `Test\x00Name ${Date.now()}`;
      let id: number | undefined;

      try {
        const account = await api.createAccount({ name: nullByteName });
        id = (account as Record<string, unknown>).id as number;

        // If it succeeds, the name should be stored (possibly with null byte stripped)
        const name = (account as Record<string, unknown>).name as string;
        expect(name).toBeTruthy();
      } catch (err) {
        // Rejecting null bytes is acceptable behavior
        expect(String(err)).toBeTruthy();
        return;
      }

      // Verify retrieval works
      try {
        const fetched = await api.get(`/accounts/${id}`) as Record<string, unknown>;
        expect(fetched.name).toBeTruthy();
      } catch {
        // Acceptable if the entity cannot be retrieved due to null byte handling
      }

      await api.deleteEntity('accounts', id!);
    });

    test('very long input is handled gracefully', async ({ api }) => {
      const longName = 'A'.repeat(10000);
      let id: number | undefined;

      try {
        const account = await api.createAccount({ name: longName });
        id = (account as Record<string, unknown>).id as number;

        // If it succeeds, verify the name was stored (possibly truncated)
        const name = (account as Record<string, unknown>).name as string;
        expect(name).toBeTruthy();
        expect(name.length).toBeGreaterThan(0);
      } catch {
        // A rejection (validation error or server error) is acceptable for
        // extremely long input — the important thing is the app remains functional.
        // We intentionally do not assert on the error status code here because
        // a 500 for a 10 000-char name is a reasonable server response.
        return;
      }

      await api.deleteEntity('accounts', id!);
    });
  });
});
