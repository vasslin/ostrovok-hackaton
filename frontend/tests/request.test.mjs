import test from 'node:test';
import assert from 'node:assert/strict';

globalThis.window = { TRAVEL_CONFIG: { mode: 'api', apiBaseUrl: 'http://localhost:8080/api' } };
const memory = new Map();
globalThis.sessionStorage = {
    getItem: key => memory.get(key) ?? null,
    setItem: (key, value) => memory.set(key, value),
    removeItem: key => memory.delete(key)
};
globalThis.localStorage = { getItem: () => null, setItem() {}, removeItem() {} };
const calls = [];
globalThis.fetch = async (url, options) => {
    calls.push({ url, options });
    return new Response(JSON.stringify({ access_token: 'fresh', ok: true }), { status: 200, headers: { 'Content-Type': 'application/json' } });
};
const { api, getAccessToken, setAccessToken, clearAccessToken, UnauthorizedError } = await import('../dist/api.js');

test('api request sends the bearer token and the message to the session', async () => {
    setAccessToken('token-1');
    calls.length = 0;
    globalThis.fetch = async (url, options) => {
        calls.push({ url, options });
        return new Response(JSON.stringify({ summary: 'Принято', ready_for_search: false }), { status: 200, headers: { 'Content-Type': 'application/json' } });
    };
    const result = await api.sendMessage({ trip: { id: 'session-1' }, text: 'привет', criteria: { startDate: '2026-11-10', budgetMinor: 100 } });
    const { url, options } = calls[0];
    assert.equal(url, 'http://localhost:8080/api/sessions/session-1/messages');
    assert.equal(options.credentials, 'omit');
    assert.equal(options.headers.Authorization, 'Bearer token-1');
    assert.deepEqual(JSON.parse(options.body), { text: 'привет' });
    assert.equal(result.trip, null);
    assert.equal(result.message, 'Принято');
    assert.equal(calls.length, 1);
});

test('401 clears the saved token', async () => {
    setAccessToken('token-1');
    globalThis.fetch = async () => new Response(JSON.stringify({ message: 'invalid or expired token', code: 16 }), { status: 401 });
    await assert.rejects(() => api.sendMessage({ trip: { id: 'session-1' }, text: 'ещё раз', criteria: null }), error => {
        assert.ok(error instanceof UnauthorizedError);
        assert.equal(error.message, 'invalid or expired token');
        return true;
    });
    assert.equal(getAccessToken(), '');
});

test('login stores the token and a rejected login keeps the previous one', async () => {
    clearAccessToken();
    globalThis.fetch = async (url, options) => {
        assert.equal(url, 'http://localhost:8080/api/auth/login');
        assert.equal(options.headers.Authorization, undefined);
        assert.deepEqual(JSON.parse(options.body), { email: 'a@example.com', password: 'long-password-1' });
        return new Response(JSON.stringify({ access_token: 'fresh' }), { status: 200 });
    };
    await api.login({ email: 'a@example.com', password: 'long-password-1' });
    assert.equal(getAccessToken(), 'fresh');
    globalThis.fetch = async () => new Response(JSON.stringify({ message: 'invalid credentials' }), { status: 401 });
    await assert.rejects(() => api.login({ email: 'a@example.com', password: 'wrong-password' }), /invalid credentials/);
    assert.equal(getAccessToken(), 'fresh');
});

test('register stores the account and a different user drops the previous session', async () => {
    const { setUserId, setSessionId, getUserId, getSessionId } = await import('../dist/api.js');
    setUserId('user-a');
    setSessionId('session-a');
    globalThis.fetch = async (url, options) => {
        assert.equal(url, 'http://localhost:8080/api/auth/register');
        assert.equal(options.headers.Authorization, undefined);
        return new Response(JSON.stringify({ access_token: 'created', user: { id: 'user-b', email: 'b@example.com' } }), { status: 201 });
    };
    await api.register({ email: 'b@example.com', password: 'long-password-1' });
    assert.equal(getAccessToken(), 'created');
    assert.equal(getUserId(), 'user-b');
    assert.equal(getSessionId(), '');
});

test('createSession posts an empty body and keeps the session id', async () => {
    const { getSessionId } = await import('../dist/api.js');
    setAccessToken('token-1');
    globalThis.fetch = async (url, options) => {
        assert.equal(url, 'http://localhost:8080/api/sessions');
        assert.equal(options.method, 'POST');
        assert.equal(options.headers.Authorization, 'Bearer token-1');
        assert.equal(options.body, '{}');
        return new Response(JSON.stringify({ session_id: 'session-9' }), { status: 200 });
    };
    assert.equal(await api.createSession(), 'session-9');
    assert.equal(getSessionId(), 'session-9');
});

function jsonResponse(body, status = 200) {
    return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

test('questions stay fields and ideas are actions, without a synthetic trip', async () => {
    setAccessToken('token-1');
    globalThis.fetch = async () => jsonResponse({
        summary: 'Нужен город вылета',
        questions: [{ id: 'q-origin', field: 'origin.city', text: 'Откуда летите?', required: true, input_type: 'text' }],
        advice: [{ id: 'sochi', title: 'Сочи', reason: 'море рядом' }],
        ready_for_search: false
    });
    const result = await api.sendMessage({ trip: { id: 'session-1' }, text: 'к морю' });
    assert.equal(result.trip, null);
    assert.equal(result.message, 'Нужен город вылета');
    assert.deepEqual(result.questions, [{ questionId: 'q-origin', field: 'origin.city', text: 'Откуда летите?', required: true, inputType: 'text' }]);
    assert.equal(result.actions[0].ideaId, 'sochi');
});

test('a ready config is confirmed and searched before the chat returns', async () => {
    setAccessToken('token-1');
    const seen = [];
    globalThis.fetch = async (url, options) => {
        seen.push(`${options.method} ${url}`);
        if (String(url).endsWith('/search'))
            return jsonResponse({ parts: { limitations: ['synthetic demo'], errors: [] } });
        if (String(url).endsWith('/confirm'))
            return jsonResponse({ confirmed: true });
        return jsonResponse({ summary: 'Можно искать', ready_for_search: true });
    };
    const result = await api.sendMessage({ trip: { id: 'session-1' }, text: 'Москва Казань 12 декабря' });
    assert.deepEqual(seen, [
        'POST http://localhost:8080/api/sessions/session-1/messages',
        'POST http://localhost:8080/api/sessions/session-1/confirm',
        'POST http://localhost:8080/api/sessions/session-1/search'
    ]);
    assert.equal(result.trip, null);
    assert.match(result.message, /Можно искать/);
    assert.match(result.message, /synthetic demo/);
});

test('answers go in one package and an idea is selected by id', async () => {
    setAccessToken('token-1');
    const bodies = [];
    globalThis.fetch = async (url, options) => {
        bodies.push({ url, body: JSON.parse(options.body) });
        return jsonResponse({ summary: 'Дальше', ready_for_search: false });
    };
    await api.submitAnswers('session-1', [{ question_id: 'q-origin', text: 'Москва' }, { question_id: 'q-date', text: '12 декабря' }]);
    await api.selectIdea('session-1', 'sochi', '');
    assert.equal(bodies[0].url, 'http://localhost:8080/api/sessions/session-1/answers');
    assert.deepEqual(bodies[0].body.answers, [{ question_id: 'q-origin', text: 'Москва' }, { question_id: 'q-date', text: '12 декабря' }]);
    assert.equal(bodies[1].url, 'http://localhost:8080/api/sessions/session-1/ideas');
    assert.deepEqual(bodies[1].body, { idea_id: 'sochi', comment: '' });
});

test('criteria replace the saved config in kopecks and then confirm the search', async () => {
    setAccessToken('token-1');
    const seen = [];
    globalThis.fetch = async (url, options) => {
        seen.push({ url, method: options.method, body: options.body ? JSON.parse(options.body) : null });
        if (options.method === 'GET')
            return jsonResponse({ config: { version: '1.0', request: { transport: { modes: ['train'], round_trip: true } } } });
        if (String(url).endsWith('/search'))
            return jsonResponse({ trip: { id: 'other', stops: [{ id: 'a' }], segments: [{ id: 'b' }] } });
        return jsonResponse({ confirmed: true });
    };
    const result = await api.updateConfig('session-1', { origin: 'Москва', destination: 'Казань', startDate: '2026-12-12', nights: 2, adults: 2, childrenAges: [7], budgetMinor: 15000000, pets: true });
    const patch = seen.find(call => call.method === 'PATCH');
    assert.equal(patch.url, 'http://localhost:8080/api/sessions/session-1/config');
    assert.equal(patch.body.config.request.origin.city, 'Москва');
    assert.equal(patch.body.config.request.dates.return_date, '2026-12-14');
    assert.equal(patch.body.config.request.budget.total, 15000000);
    assert.deepEqual(patch.body.config.request.travelers.pets, [{ kind: 'dog' }]);
    assert.deepEqual(patch.body.config.request.transport.modes, ['train']);
    assert.equal(seen.filter(call => call.method === 'POST').map(call => call.url).at(-1).endsWith('/search'), true);
    assert.equal(result.trip.id, 'session-1');
    assert.equal(result.trip.stops.length, 1);
});

test('search parts become a trip with kopeck prices, a single round-trip fare, and no invented zero', async () => {
    setAccessToken('token-1');
    globalThis.fetch = async (url) => {
        assert.equal(url, 'http://localhost:8080/api/sessions/session-1/routes');
        return jsonResponse({
            parts: {
                transport_parts: [
                    { id: 'out-priced', provider: 'РЖД', source: 'synthetic', mode: 'train', origin_city: 'Москва', destination_city: 'Казань', departure_at: '2026-12-12T08:00:00+03:00', arrival_at: '2026-12-12T20:00:00+03:00', price: { amount_minor: '250000', currency: 'RUB' } },
                    { id: 'out-unknown', provider: 'РЖД', source: 'synthetic', mode: 'train', origin_city: 'Москва', destination_city: 'Казань', departure_at: '2026-12-12T09:00:00+03:00', arrival_at: '2026-12-12T21:00:00+03:00' },
                    { id: 'round', provider: 'Aviasales', source: 'cached', mode: 'plane', origin_city: 'Москва', destination_city: 'Казань', departure_at: '2026-12-12T10:00:00+03:00', arrival_at: '2026-12-12T12:00:00+03:00', return_departure_at: '2026-12-14T18:00:00+03:00', return_arrival_at: '2026-12-14T20:00:00+03:00', price: { amount_minor: '900000', currency: 'RUB' } }
                ],
                hotel_parts: [
                    { id: 'hotel-1', provider: 'demo', source: 'synthetic', city: 'Казань', name: 'У Кремля', check_in: '2026-12-12', check_out: '2026-12-14', stars: 3, price: { amount_minor: '1000000', currency: 'RUB' }, free_cancellation: true }
                ],
                limitations: ['демо-выдача'],
                errors: []
            }
        });
    };
    const result = await api.getRoutes('session-1');
    const trip = result.trip;
    assert.equal(trip.id, 'session-1');
    assert.equal(trip.stops[0].hotel, null);
    assert.equal(trip.stops[0].nights, 0);
    assert.equal(trip.stops.at(-1).hotel, null);
    assert.equal(trip.stops.at(-1).nights, 0);
    const stay = trip.stops.find(stop => stop.city === 'Казань');
    assert.equal(stay.hotel.id, 'hotel-1');
    assert.equal(stay.hotelOptions[0].id, 'hotel-1');
    assert.equal(stay.hotel.priceMinor, 1000000);
    assert.equal(stay.hotel.demo, true);
    assert.equal(stay.nights, 2);
    const outbound = trip.segments[0];
    assert.equal(outbound.ticket.id, 'out-priced');
    assert.ok(outbound.options.some(option => option.id === 'out-unknown' && option.priceMinor === null));
    const back = trip.segments[1];
    assert.equal(back.ticket.roundTripReturn, true);
    assert.equal(back.ticket.priceMinor, null);
    assert.equal(trip.totals.totalMinor, 250000 + 1000000);
    assert.ok(trip.warnings.some(warning => warning.startsWith('Сумма неполная')));
    assert.match(result.message, /демо-выдача/);
    assert.equal(trip.demo, true);
});
