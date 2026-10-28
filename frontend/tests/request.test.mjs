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

test('a ready config is shown without confirming or searching', async () => {
    setAccessToken('token-1');
    const seen = [];
    globalThis.fetch = async (url, options) => { seen.push(url); return jsonResponse({ summary: 'Можно искать', config: { request: { origin: { city: 'Москва' } } }, ready_for_search: true }); };
    const result = await api.sendMessage({ trip: { id: 'session-1' }, text: 'Москва Казань 12 декабря' });
    assert.deepEqual(seen, ['http://localhost:8080/api/sessions/session-1/messages']);
    assert.equal(result.trip, null);
    assert.equal(result.readyForSearch, true);
    assert.equal(result.tripConfig.origin.city, 'Москва');
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

test('criteria save major-unit budget without launching a search', async () => {
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
    assert.equal(patch.body.config.request.budget.total, 150000);
    assert.deepEqual(patch.body.config.request.travelers.pets, [{ kind: 'dog' }]);
    assert.deepEqual(patch.body.config.request.transport.modes, ['train']);
    assert.equal(seen.some(call => call.method === 'POST'), false);
    assert.equal(result.trip, null);
    assert.equal(result.tripConfig.destination.city, 'Казань');
});

test('search parts become a trip with kopeck prices, a single round-trip fare, and no invented zero', async () => {
    setAccessToken('token-1');
    globalThis.fetch = async (url) => {
        if (url.endsWith('/config')) return jsonResponse({ config: {} });
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

test('real dialogue hides internal JSON memory and preserves the adult count in route prices', async () => {
    const config = { schema_version: '1.0', travelers: { adults: 2 }, budget: { total: 80000 } };
    const transport = { id: 'flight', origin_city: 'Москва', destination_city: 'Сочи', departure_at: '2026-11-10T08:00:00+03:00', arrival_at: '2026-11-10T12:00:00+03:00', return_departure_at: '2026-11-17T16:00:00+03:00', return_arrival_at: '2026-11-17T20:00:00+03:00', price: { amount_minor: '500000', currency: 'RUB', scope: 'per_person_round_trip' } };
    const hotel = { id: 'hotel', city: 'Сочи', check_in: '2026-11-10', check_out: '2026-11-17', price: { amount_minor: '3500000', currency: 'RUB', scope: 'stay_all_rooms' } };
    const candidate = { id: 'backend-route', transport: [transport], hotel, price_status: 'estimated' };
    globalThis.fetch = async url => {
        if (url.endsWith('/messages')) return jsonResponse({ summary: '{"notes":["internal memory"]}', config: { request: config }, ready_for_search: true });
        if (url.endsWith('/confirm')) return jsonResponse({ confirmed: true });
        if (url.endsWith('/capabilities')) return jsonResponse({ hotels: { status: 'available' } });
        if (url.endsWith('/config')) return jsonResponse({ config: { request: config } });
        return jsonResponse({ candidates: [candidate], parts: { transport_parts: [transport], hotel_parts: [hotel] } });
    };
    const dialog = await api.sendMessage({ trip: { id: 'session-1' }, text: 'Двое взрослых' });
    assert.equal(dialog.message.includes('internal memory'), false);
    const result = await api.confirmSearch('session-1', dialog.tripConfig);
    assert.equal(result.trip.criteria.adults, 2);
    assert.equal(result.trip.routeId, 'backend-route');
    assert.equal(result.trip.totals.totalMinor, 4500000);
    globalThis.fetch = async (url, options) => {
        assert.equal(url, 'http://localhost:8080/api/orders');
        assert.equal(options.headers['Idempotency-Key'], 'order-key');
        assert.deepEqual(JSON.parse(options.body), { session_id: 'session-1', route_id: 'backend-route' });
        return jsonResponse({ order_id: 'order-1', status: 'demo_booked', mode: 'demo' });
    };
    const booked = await api.book(result.trip, 'order-key');
    assert.equal(booked.booking.id, 'order-1');
    assert.equal(booked.booking.demo, true);
});

test('booking history uses existing order endpoints and never calls the missing me/trips API', async () => {
    const { setUserId } = await import('../dist/api.js');
    setUserId('history-owner');
    const local = new Map();
    globalThis.localStorage = { getItem: key => local.get(key) || null, setItem: (key, value) => local.set(key, value), removeItem: key => local.delete(key) };
    globalThis.fetch = async (url, options) => {
        if (url.endsWith('/orders') && options.method === 'POST') return jsonResponse({ order_id: 'saved-order', status: 'demo_booked', mode: 'demo' });
        assert.equal(url, 'http://localhost:8080/api/orders/saved-order');
        return jsonResponse({ order_id: 'saved-order', status: 'demo_paid' });
    };
    await api.book({ id: 'saved-session', routeId: 'saved-route' }, 'history-key');
    const history = await api.getHistory();
    assert.equal(history.trips.length, 1);
    assert.equal(history.trips[0].booking.status, 'demo_paid');
    assert.equal((await api.getTrip('saved-session')).trip.routeId, 'saved-route');
});

test('hotel unavailable prevents confirmation and never changes the request silently', async () => {
    const seen = [];
    const config = { hotel: { needed: true } };
    globalThis.fetch = async (url, options) => { seen.push(url); return jsonResponse({ mode: 'live', hotels: { status: 'not_configured' } }); };
    await assert.rejects(() => api.confirmSearch('session-1', config), /только транспорт/);
    assert.equal(config.hotel.needed, true);
    assert.deepEqual(seen, ['http://localhost:8080/api/sessions/session-1/capabilities']);
});

test('explicit transport-only confirmation saves then confirms then searches', async () => {
    const seen = [];
    const config = { hotel: { needed: false } };
    globalThis.fetch = async (url, options) => {
        seen.push({ path: url.split('/').at(-1), method: options.method, body: options.body ? JSON.parse(options.body) : null });
        if (url.endsWith('/capabilities')) return jsonResponse({ mode: 'live', hotels: { status: 'not_configured' } });
        return jsonResponse({ config: { request: config }, parts: { errors: [] } });
    };
    await api.confirmSearch('session-1', config);
    assert.deepEqual(seen.map(call => call.path), ['capabilities', 'config', 'confirm', 'search']);
    assert.equal(seen[1].body.config.request.hotel.needed, false);
});

test('partial hotel failure preserves diagnostics and hides internal paths and false totals', async () => {
    const config = { hotel: { needed: true }, travelers: { adults: 1 } };
    const part = { id: 'train', provider: 'yandex_rasp', origin_city: 'Москва', destination_city: 'Париж', departure_at: '2026-11-04T10:00:00+03:00', arrival_at: '2026-11-05T10:00:00+01:00', price: { amount_minor: '500000', scope: 'per_person_one_way', currency: 'RUB' } };
    globalThis.fetch = async url => {
        if (url.endsWith('/config')) return jsonResponse({ config: { request: config } });
        return jsonResponse({ parts: { transport_parts: [part], hotel_parts: [], errors: [{ provider: 'hotel', kind: 'hotel', code: 'missing_field', message: '/travelers/residency' }], limitations: ['route-service: internal'] } });
    };
    const result = await api.getRoutes('session-1');
    assert.equal(result.trip.partial, true);
    assert.equal(result.trip.totals.totalMinor, null);
    assert.equal(result.trip.totals.hotelsMinor, null);
    assert.equal(result.trip.totals.transferMinor, null);
    assert.equal(result.trip.totals.activitiesMinor, null);
    assert.equal(result.message.includes('/travelers/'), false);
    assert.equal(result.message.includes('route-service'), false);
    assert.match(result.message, /гражданство/);
    assert.equal(result.providerIssues[0].code, 'missing_field');
});

test('restore session keeps latest questions and stored config despite read events', async () => {
    const question = { id: 'q-date', text: 'Когда?', field: 'dates.departure_date', required: true };
    const request = { origin: { city: 'Москва' } };
    const turn = { kind: 'planning_turn', response: { config: { request }, questions: [question], summary: 'Нужна дата' } };
    globalThis.fetch = async url => url.endsWith('/config') ? jsonResponse({ config: { request } }) : jsonResponse({ messages: [{ role: 'assistant', text: JSON.stringify(turn) }, { role: 'assistant', text: JSON.stringify({ kind: 'travel_response', response: { config: { request } } }) }] });
    const result = await api.restoreSession('session-1');
    assert.equal(result.questions[0].questionId, 'q-date');
    assert.equal(result.tripConfig.origin.city, 'Москва');
});
