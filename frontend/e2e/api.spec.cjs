const { test, expect } = require('playwright/test');

const tripRequest = {
  origin: { city: 'Москва', country_code: 'RU' }, destination: { city: 'Казань', country_code: 'RU' },
  dates: { departure_date: '2026-12-12', return_date: '2026-12-14' },
  travelers: { adults: 2, children_ages: [], residency: 'ru' }, currency: 'RUB',
  budget: { total: 100000, scope: 'total' }, hotel: { needed: true, rooms: 1 },
  transport: { modes: ['train'], round_trip: false }
};

test.beforeEach(async ({ page }) => {
  await page.route(/fonts\.google/, route => route.abort());
});

async function mockBackend(page) {
  const calls = [], errors = [];
  page.on('pageerror', error => errors.push(error.message));
  await page.route('**/api/**', async route => {
    const request = route.request(), path = new URL(request.url()).pathname;
    const body = request.postDataJSON();
    calls.push({ path, method: request.method(), body, token: request.headers().authorization });
    let json;
    if (path === '/api/auth/login' && body.password === 'wrong-password-123') {
      return route.fulfill({ status: 401, json: { message: 'Неверный пароль' } });
    }
    if (path === '/api/auth/login' || path === '/api/auth/register') json = { access_token: 'browser-fixture-token', user: { id: 'fixture-user', email: body.email } };
    else if (path === '/api/auth/me') json = { id: 'fixture-user', email: 'traveler@example.com' };
    else if (path === '/api/sessions') json = { session_id: 'fixture-session' };
    else if (path.endsWith('/messages')) json = { summary: 'Проверьте параметры поездки', config: { request: tripRequest }, ready_for_search: true };
    else if (path.endsWith('/capabilities')) json = { mode: 'demo', hotels: { status: 'available' }, transport: { aviasales: false, yandex_rasp: true } };
    else if (path.endsWith('/config')) json = { config: body?.config || { request: tripRequest } };
    else if (path.endsWith('/confirm')) json = { confirmed: true };
    else if (path.endsWith('/search')) json = { parts: {
      transport_parts: [{ id: 'train-1', mode: 'train', provider: 'fixture', source: 'synthetic', origin_city: 'Москва', destination_city: 'Казань', departure_at: '2026-12-12T08:00:00+03:00', arrival_at: '2026-12-12T20:00:00+03:00', price: { amount_minor: '250000', currency: 'RUB', scope: 'total' } }],
      hotel_parts: [{ id: 'hotel-1', provider: 'fixture', source: 'synthetic', city: 'Казань', name: 'Тестовый отель', check_in: '2026-12-12', check_out: '2026-12-14', stars: 3, price: { amount_minor: '1000000', currency: 'RUB' } }]
    } };
    else throw new Error(`Unexpected API request: ${request.method()} ${path}`);
    await route.fulfill({ json });
  });
  return { calls, errors };
}

async function login(page) {
  await page.goto('/');
  await expect(page.locator('#login-form')).toBeVisible();
  await page.locator('#login-form [name=email]').fill('traveler@example.com');
  await page.locator('#login-form [name=password]').fill('browser-test-password');
  await page.locator('#login-form [type=submit]').click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
}

async function fitsScreen(page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)).toBe(true);
}

test('API login error and registration keep the form usable and unlock the page', async ({ page }) => {
  const { calls, errors } = await mockBackend(page);
  await page.goto('/');
  await expect(page.locator('.app-shell')).toHaveAttribute('inert', '');
  await fitsScreen(page);
  await page.locator('#login-form [name=email]').fill('traveler@example.com');
  await page.locator('#login-form [name=password]').fill('wrong-password-123');
  await page.locator('#login-form [type=submit]').click();
  await expect(page.locator('#login-error')).toHaveText('Неверный пароль');
  await page.locator('[data-action=auth-mode]').click();
  await expect(page.locator('#login-form [name=email]')).toHaveValue('traveler@example.com');
  await page.locator('#login-form [name=password]').fill('browser-test-password');
  await page.locator('#login-form [type=submit]').click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await expect(page.locator('body')).not.toHaveClass(/overlay-open/);
  await expect(page.locator('.app-shell')).not.toHaveAttribute('inert');
  expect(calls.some(call => call.path === '/api/auth/register')).toBe(true);
  expect(errors).toEqual([]);
});

test('API planning keeps review edits across mobile views and searches only after confirmation', async ({ page }) => {
  const { calls, errors } = await mockBackend(page);
  await login(page);
  const phone = await page.locator('.mobile-switch').isVisible();
  if (phone) await page.locator('.mobile-switch [data-action=mobile-chat]').click();
  await page.locator('#chat-input').fill('Москва Казань 12 декабря');
  await page.locator('.send-button').click();
  await expect(page.locator('.thinking')).toHaveCount(0);
  if (phone) await page.locator('.mobile-switch [data-action=mobile-route]').click();
  await expect(page.locator('#review-form')).toBeVisible();
  await fitsScreen(page);
  await page.locator('#review-form [name=budget]').fill('120000');
  if (phone) {
    await page.locator('.mobile-switch [data-action=mobile-chat]').click();
    await page.locator('.mobile-switch [data-action=mobile-route]').click();
  }
  await expect(page.locator('#review-form [name=budget]')).toHaveValue('120000');
  expect(calls.some(call => call.path.endsWith('/search'))).toBe(false);
  await page.locator('#review-form [value=search]').click();
  await expect(page.locator('.trip-total')).toBeVisible();
  await expect(page.locator('#review-form')).toHaveCount(0);
  await expect(page.locator('#route-map .map-marker')).not.toHaveCount(0);
  await fitsScreen(page);
  const saved = calls.find(call => call.method === 'PATCH' && call.path.endsWith('/config'));
  expect(saved.body.config.request.budget.total).toBe(120000);
  expect(saved.token).toBe('Bearer browser-fixture-token');
  expect(calls.filter(call => call.path.endsWith('/confirm'))).toHaveLength(1);
  expect(calls.filter(call => call.path.endsWith('/search'))).toHaveLength(1);
  await page.locator('[data-action=edit-criteria]').last().click();
  await expect(page.locator('#review-form [name=budget]')).toHaveValue('120000');
  expect(errors).toEqual([]);
});
