const { test, expect } = require('playwright/test');

test.beforeEach(async ({ page }) => {
  await page.route(/fonts\.google/, route => route.abort());
  await page.goto('/');
  await expect(page.locator('.trip-total')).toBeVisible();
});
async function noHorizontalOverflow(page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBe(true);
}
async function mobile(page) { return page.locator('.mobile-switch').isVisible(); }
async function openRoute(page) { if (await mobile(page)) await page.locator('[data-action="mobile-route"]').click(); }

test('route, tickets, hotels, profile and history fit the screen', async ({ page }) => {
  await noHorizontalOverflow(page);
  for (const tab of ['tickets', 'hotels', 'overview']) {
    await page.locator(`[data-action="tab"][data-tab="${tab}"]`).click();
    await expect(page.locator(`[data-tab="${tab}"]`)).toHaveAttribute('aria-selected', 'true');
    await noHorizontalOverflow(page);
  }
  const nav = await mobile(page) ? page.locator('.mobile-bottom-nav') : page.locator('.rail-links');
  await nav.locator('[data-action="profile"]').click();
  await expect(page.locator('#profile-form')).toBeVisible();
  await noHorizontalOverflow(page);
  await nav.locator('[data-action="history"]').click();
  await expect(page.locator('.page-title h1')).toHaveText('Мои поездки');
  await noHorizontalOverflow(page);
});

test('chat keeps its draft and send control inside a keyboard-sized viewport', async ({ page }) => {
  test.skip(!await mobile(page), 'Mobile planner layout');
  const original = page.viewportSize();
  await page.locator('[data-action="mobile-chat"]').click();
  const input = page.locator('#chat-input');
  await input.fill('Хочу в Сочи');
  await page.locator('[data-action="mobile-route"]').click();
  await page.locator('[data-action="mobile-chat"]').click();
  await expect(input).toHaveValue('Хочу в Сочи');
  await input.focus();
  await page.setViewportSize({ width: original.width, height: Math.min(original.height, 430) });
  await expect.poll(async () => {
    const rect = await page.locator('.send-button').boundingBox();
    return rect && rect.y >= 0 && rect.y + rect.height <= page.viewportSize().height;
  }).toBe(true);
  expect(await input.evaluate(el => parseFloat(getComputedStyle(el).fontSize))).toBeGreaterThanOrEqual(16);
  await page.locator('.send-button').tap();
  await expect(page.locator('.message.user .bubble').last()).toHaveText('Хочу в Сочи');
  await expect(page.locator('.thinking')).toHaveCount(0);
  await page.setViewportSize(original);
  await noHorizontalOverflow(page);
});

test('criteria sheet is scrollable, locks the background and restores focus', async ({ page }) => {
  const edit = page.locator('[data-action="edit-criteria"]').last();
  await edit.click();
  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible();
  await expect(page.locator('.app-shell')).toHaveAttribute('inert', '');
  await expect(page.locator('body')).toHaveClass(/overlay-open/);
  await noHorizontalOverflow(page);
  await dialog.locator('[name="nights"]').fill('5');
  await dialog.locator('[name="budget"]').fill('180000');
  await dialog.getByRole('button', { name: /Обновить поездку/ }).click();
  await expect(dialog).toHaveCount(0);
  await expect(page.locator('.thinking')).toHaveCount(0);
  await expect(page.locator('.trip-intro')).toBeVisible();
  await edit.click();
  await dialog.getByRole('button', { name: 'Закрыть окно' }).click();
  await expect(page.locator('body')).not.toHaveClass(/overlay-open/);
  await expect(edit).toBeFocused();
});

test('hotel and transport details close without losing the page scroll', async ({ page }) => {
  await page.locator('.stop-card.with-hotel [data-action="stop"]').first().click();
  await expect(page.getByRole('dialog')).toBeVisible();
  await noHorizontalOverflow(page);
  await page.getByRole('button', { name: 'Закрыть подробности' }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await page.locator('[data-action="segment"]').first().click();
  await expect(page.getByRole('dialog')).toBeVisible();
  await noHorizontalOverflow(page);
  await page.keyboard.press('Escape');
  await expect(page.getByRole('dialog')).toHaveCount(0);
});

test('touch controls are large enough and the map responds to zoom', async ({ page }) => {
  test.skip(!await mobile(page), 'Phone touch targets');
  for (const selector of ['[data-action="new-trip"]', '.profile-link', '[data-map-control="in"]', '[data-action="mobile-chat"]', '[data-action="edit-criteria"]']) {
    const box = await page.locator(selector).first().boundingBox();
    expect(box.width, selector).toBeGreaterThanOrEqual(44);
    expect(box.height, selector).toBeGreaterThanOrEqual(44);
  }
  const before = await page.locator('.map-marker').first().getAttribute('style');
  await page.getByRole('button', { name: 'Приблизить карту' }).tap();
  await expect(page.locator('.map-marker').first()).not.toHaveAttribute('style', before);
  await page.getByRole('button', { name: 'Показать весь маршрут' }).tap();
  await expect(page.locator('.map-marker').first()).toHaveAttribute('style', before);
});

test('phone demo checkout creates one booking and opens history', async ({ page }) => {
  await openRoute(page);
  await page.locator('[data-action="summary"]').click();
  await expect(page.locator('.summary-page')).toBeVisible();
  await noHorizontalOverflow(page);
  await page.locator('[data-action="book"]').click();
  await page.locator('[data-action="confirm-book"]').click();
  await expect(page.locator('.success-booking')).toBeVisible();
  await page.getByRole('dialog').locator('[data-action="history"]').click();
  await expect(page.locator('.history-card')).toHaveCount(1);
  await noHorizontalOverflow(page);
});

test('rotating a phone keeps the chat draft and route accessible', async ({ page }) => {
  test.skip(!await mobile(page), 'Phone layout');
  await page.locator('[data-action="mobile-chat"]').click();
  await page.locator('#chat-input').fill('Сочи на выходные');
  const { width, height } = page.viewportSize();
  await page.setViewportSize({ width: height, height: width });
  await expect(page.locator('#chat-input')).toHaveValue('Сочи на выходные');
  await noHorizontalOverflow(page);
  await page.locator('[data-action="mobile-route"]').click();
  await expect(page.locator('.trip-total')).toBeVisible();
  await noHorizontalOverflow(page);
});

test('two-finger map zoom leaves the reset control usable', async ({ page, context, browserName }) => {
  test.skip(browserName !== 'chromium' || !await mobile(page), 'Chromium touch protocol on phone');
  const map = page.locator('#route-map');
  await map.scrollIntoViewIfNeeded();
  const before = await page.locator('.map-marker').first().getAttribute('style');
  const box = await map.boundingBox();
  const x = box.x + box.width / 2, y = box.y + box.height / 2;
  const cdp = await context.newCDPSession(page);
  const points = gap => [{ x: x - gap, y, id: 0 }, { x: x + gap, y, id: 1 }];
  await cdp.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: points(25) });
  for (const gap of [35, 50, 65]) {
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: points(gap) });
  }
  await cdp.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
  await expect(page.locator('.map-marker').first()).not.toHaveAttribute('style', before);
  await page.getByRole('button', { name: 'Показать весь маршрут' }).tap();
  await expect(page.locator('.map-marker').first()).toHaveAttribute('style', before);
  await cdp.detach();
});

test('standalone demo works with the mobile module bundled offline', async ({ page }) => {
  await page.goto('/standalone.html');
  await expect(page.locator('.trip-total')).toBeVisible();
  await noHorizontalOverflow(page);
  if (await mobile(page)) {
    await page.locator('[data-action="mobile-chat"]').click();
    await expect(page.locator('#chat-input')).toBeVisible();
  }
});
