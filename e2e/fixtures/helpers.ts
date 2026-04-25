import type { Page, Locator } from '@playwright/test';

/**
 * Find a form field by its label text, scoped to a container.
 * By default scopes to the modal overlay (.fixed.inset-0) if present,
 * falling back to the full page.
 */
export function formField(page: Page, labelText: string | RegExp, container?: Locator): Locator {
  const scope = container ?? page.locator('.fixed.inset-0').or(page.locator('body')).first();
  const label = scope.locator('label').filter({ hasText: labelText }).first();
  return label.locator('xpath=..').locator('input, select, textarea').first();
}

/**
 * Fill a form field found by label text, scoped to a container.
 */
export async function fillField(page: Page, labelText: string | RegExp, value: string, container?: Locator) {
  const field = formField(page, labelText, container);
  const tagName = await field.evaluate(el => el.tagName.toLowerCase());
  if (tagName === 'select') {
    await field.selectOption(value);
  } else {
    await field.clear();
    await field.fill(value);
  }
}
