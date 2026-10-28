import { DEFAULT_CRITERIA, createTrip, parseMessage, selectedIds, cheaperSelection, money, cities } from './model.js';
const config = window.TRAVEL_CONFIG || { mode: 'demo', apiBaseUrl: '/api' };
export const isDemo = config.mode !== 'api';
const tokenKey = 'odin-v1:access_token';
const userKey = 'odin-v1:user_id';
const sessionKey = 'odin-v1:session_id';
const pause = () => new Promise(resolve => setTimeout(resolve, 450));
function tokenStorage() {
    try {
        return sessionStorage;
    }
    catch {
        return null;
    }
}
export function getAccessToken() {
    try {
        return tokenStorage()?.getItem(tokenKey) || '';
    }
    catch {
        return '';
    }
}
export function setAccessToken(token) {
    tokenStorage()?.setItem(tokenKey, token);
}
export function clearAccessToken() {
    try {
        tokenStorage()?.removeItem(tokenKey);
    }
    catch {
        /* sessionStorage can be blocked */
    }
}
function stored(key) {
    try {
        return tokenStorage()?.getItem(key) || '';
    }
    catch {
        return '';
    }
}
function store(key, value) {
    if (value)
        tokenStorage()?.setItem(key, value);
    else
        tokenStorage()?.removeItem(key);
}
export function getUserId() { return stored(userKey); }
export function setUserId(id) { store(userKey, id); }
export function getSessionId() { return stored(sessionKey); }
export function setSessionId(id) { store(sessionKey, id); }
export function clearSessionId() { store(sessionKey, ''); }
function storeAccount(result) {
    if (!result.access_token)
        throw new Error('Сервер не вернул токен.');
    const userId = result.user?.id || '';
    if (userId && getUserId() && userId !== getUserId())
        clearSessionId();
    setAccessToken(result.access_token);
    if (userId)
        setUserId(userId);
    return result;
}
export class UnauthorizedError extends Error {
    constructor(message) {
        super(message || 'Нужно войти снова.');
        this.name = 'UnauthorizedError';
        this.status = 401;
    }
}
function snakeKey(key) {
    return key.replace(/[A-Z]/g, letter => `_${letter.toLowerCase()}`);
}
function toSnake(value) {
    if (Array.isArray(value))
        return value.map(toSnake);
    if (value && typeof value === 'object')
        return Object.fromEntries(Object.entries(value).map(([key, item]) => [snakeKey(key), toSnake(item)]));
    return value;
}
async function request(path, { method = 'GET', body, requestId, anonymous = false } = {}) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 20000);
    const token = anonymous ? '' : getAccessToken();
    try {
        const response = await fetch(`${config.apiBaseUrl.replace(/\/$/, '')}${path}`, {
            method, credentials: 'omit', signal: controller.signal,
            headers: {
                Accept: 'application/json',
                ...(token ? { Authorization: `Bearer ${token}` } : {}),
                ...(body ? { 'Content-Type': 'application/json' } : {}),
                ...(requestId ? { 'Idempotency-Key': requestId } : {})
            },
            ...(body ? { body: JSON.stringify(toSnake(body)) } : {})
        });
        const json = await response.json().catch(() => ({}));
        if (response.status === 401) {
            const message = json.message || 'Нужно войти снова.';
            if (anonymous)
                throw new Error(message);
            clearAccessToken();
            throw new UnauthorizedError(message);
        }
        if (!response.ok)
            throw new Error(json.message || `Сервис временно недоступен (${response.status}). Повторите запрос.`);
        return json;
    }
    catch (error) {
        if (error.name === 'AbortError')
            throw new Error('Сервис долго не отвечает. Попробуйте ещё раз.');
        if (error instanceof TypeError)
            throw new Error('Нет связи с сервером. Проверьте подключение.');
        throw error;
    }
    finally {
        clearTimeout(timeout);
    }
}
function sessionPath(sessionId) {
    if (!sessionId)
        throw new Error('Сначала создайте сессию.');
    return `/sessions/${encodeURIComponent(sessionId)}`;
}
function dialogText(raw) {
    const summary = String(raw.summary || '').trim();
    if (summary)
        return summary;
    const questions = (raw.questions || []).map(question => question.text).filter(Boolean);
    if (questions.length)
        return questions.join('\n');
    const ideas = (raw.advice || []).map(idea => [idea.title || idea.city, idea.reason || idea.description].filter(Boolean).join(' — ')).filter(Boolean);
    if (ideas.length)
        return ideas.join('\n');
    const notes = [...(raw.warnings || []), ...(raw.missing_fields || [])].filter(Boolean);
    if (notes.length)
        return notes.join('\n');
    return 'Сообщение принято. Маршрут появится, когда параметры можно искать.';
}
function chatFromDialog(raw) {
    const questions = (raw.questions || []).filter(question => question.id).map(question => ({
        questionId: question.id,
        field: question.field || '',
        text: question.text || question.field || 'Уточните',
        required: Boolean(question.required),
        inputType: question.input_type || 'text'
    }));
    const actions = (raw.advice || []).filter(idea => idea.id).map(idea => ({
        label: idea.title || idea.city || idea.id,
        value: idea.id,
        ideaId: idea.id
    }));
    return { message: dialogText(raw), questions, actions, readyForSearch: Boolean(raw.ready_for_search) };
}
function frontendTrip(trip, sessionId) {
    if (!trip || !Array.isArray(trip.stops) || !Array.isArray(trip.segments))
        return null;
    return { ...trip, id: sessionId || trip.id };
}
function asMinor(price) {
    const raw = price && typeof price === 'object' ? price.amount_minor : price;
    if (raw === undefined || raw === null || raw === '')
        return null;
    const amount = typeof raw === 'number' ? raw : Number(String(raw).trim());
    return Number.isFinite(amount) ? amount : null;
}
function cityInfo(name) {
    const found = Object.keys(cities).find(key => key.toLowerCase() === String(name || '').trim().toLowerCase());
    if (found)
        return { city: found, country: cities[found].country, lat: cities[found].lat, lng: cities[found].lng };
    return { city: String(name || 'Город'), country: '', lat: null, lng: null };
}
function sameCity(left, right) { return String(left || '').trim().toLowerCase() === String(right || '').trim().toLowerCase(); }
function nightCount(start, end) {
    const from = Date.parse(`${String(start || '').slice(0, 10)}T12:00:00Z`);
    const to = Date.parse(`${String(end || '').slice(0, 10)}T12:00:00Z`);
    if (!Number.isFinite(from) || !Number.isFinite(to))
        return 0;
    return Math.max(0, Math.round((to - from) / 86400000));
}
function modeLabel(mode) {
    return { plane: 'Самолёт', train: 'Поезд', bus: 'Автобус', suburban: 'Электричка', water: 'Водный' }[mode] || 'Транспорт';
}
function transportLegs(part) {
    const from = part.origin_city || part.origin;
    const to = part.destination_city || part.destination;
    if (!from || !to || !part.departure_at || !part.arrival_at)
        return [];
    const outbound = { id: part.id, from, to, departure: part.departure_at, arrival: part.arrival_at, price: asMinor(part.price), currency: part.price?.currency || 'RUB', source: part.source || '', provider: part.provider || '', mode: part.mode || '', transfers: part.transfers || 0, roundTripReturn: false };
    if (!part.return_departure_at || !part.return_arrival_at)
        return [outbound];
    return [outbound, { ...outbound, id: `${part.id}-return`, from: to, to: from, departure: part.return_departure_at, arrival: part.return_arrival_at, price: null, transfers: part.return_transfers || 0, roundTripReturn: true }];
}
function stayOnly(hotels, sessionId) {
    const names = [...new Set(hotels.map(hotel => hotel.city).filter(Boolean))];
    const stops = names.map((city, index) => {
        const options = hotels.filter(hotel => sameCity(hotel.city, city)).map(hotelFromPart);
        const sample = hotels.find(hotel => sameCity(hotel.city, city));
        return { id: `stay-${index}`, ...cityInfo(city), arrival: sample?.check_in || '', departure: sample?.check_out || '', nights: options[0]?.nights || 0, hotelOptions: options, hotel: options[0] || null, activities: [] };
    });
    return finishTrip(stops, [], sessionId, hotels.some(hotel => hotel.source === 'synthetic'), hotels[0]?.price?.currency || 'RUB');
}
function finishTrip(stops, segments, sessionId, synthetic, currency) {
    const shown = [...segments.flatMap(segment => segment.options).filter(ticket => !ticket.roundTripReturn), ...stops.flatMap(stop => stop.hotelOptions)];
    const incomplete = shown.some(item => item.priceMinor == null);
    const transportMinor = segments.reduce((sum, segment) => sum + (segment.ticket.roundTripReturn || segment.ticket.priceMinor == null ? 0 : segment.ticket.priceMinor), 0);
    const hotelsMinor = stops.reduce((sum, stop) => sum + (stop.hotel?.priceMinor == null ? 0 : stop.hotel.priceMinor), 0);
    const warnings = [];
    if (incomplete)
        warnings.push('Сумма неполная: у части услуг нет цены.');
    if (segments.some(segment => segment.ticket.roundTripReturn))
        warnings.push('Тариф туда-обратно посчитан один раз и не прибавляется к обратному билету.');
    if (synthetic)
        warnings.push('Предложения с пометкой «демо» синтетические, это не живое наличие.');
    const destination = stops.find(stop => stop.nights)?.city || stops[1]?.city || stops[0].city;
    const startDate = String(stops[0].departure || stops[0].arrival || '').slice(0, 10);
    const endDate = String(stops.at(-1).arrival || stops.at(-1).departure || startDate).slice(0, 10);
    return {
        id: sessionId, revision: 1, status: 'draft', demo: synthetic, title: `${stops[0].city} — ${destination}`, subtitle: synthetic ? 'Синтетические варианты, не живое наличие.' : 'Варианты из поиска',
        criteria: { origin: stops.find(stop => stop.id === 'origin')?.city || stops[0].city, destination, startDate, nights: stops.reduce((sum, stop) => sum + (stop.nights || 0), 0) || nightCount(startDate, endDate), adults: 1, childrenAges: [], budgetMinor: null, currency: currency || 'RUB', pets: false },
        startDate, endDate, stops, segments,
        totals: { currency: currency || 'RUB', transportMinor, hotelsMinor, transferMinor: 0, activitiesMinor: 0, totalMinor: transportMinor + hotelsMinor },
        warnings, tags: synthetic ? ['демо'] : [], cover: 'city'
    };
}
function ticketFromLeg(leg) {
    return {
        id: leg.id, carrier: leg.provider || 'Перевозчик', number: leg.transfers ? `${modeLabel(leg.mode)} · пересадок ${leg.transfers}` : modeLabel(leg.mode),
        type: leg.mode === 'plane' ? 'flight' : (leg.mode || 'flight'), departure: leg.departure, arrival: leg.arrival, priceMinor: leg.price,
        baggage: null, refundable: null, petsAllowed: false, demo: leg.source === 'synthetic', source: leg.source, roundTripReturn: leg.roundTripReturn
    };
}
function hotelFromPart(part) {
    const nights = nightCount(part.check_in, part.check_out);
    const priceMinor = asMinor(part.price);
    return {
        id: part.id, name: part.name || 'Отель', city: part.city || '', rating: '', stars: part.stars || 0, priceMinor,
        nightlyMinor: priceMinor != null && nights ? Math.round(priceMinor / nights) : null, nights, rooms: 1, petsAllowed: false,
        cancellation: part.free_cancellation ? 'Бесплатная отмена' : 'Условия отмены не указаны', meals: 'Питание не указано', distance: '',
        cover: 'city', demo: part.source === 'synthetic', source: part.source || ''
    };
}
function assembleParts(parts, sessionId) {
    const legs = (parts.transport_parts || []).flatMap(transportLegs).sort((a, b) => String(a.departure).localeCompare(String(b.departure)));
    const hotels = parts.hotel_parts || [];
    if (!legs.length && !hotels.length)
        return null;
    if (!legs.length)
        return stayOnly(hotels, sessionId);
    const originLeg = legs.find(leg => !leg.roundTripReturn) || legs[0];
    const origin = originLeg ? originLeg.from : (hotels[0]?.city || '');
    const stayNames = [...new Set([...(originLeg ? legs.filter(leg => sameCity(leg.from, origin) && !sameCity(leg.to, origin)).map(leg => leg.to) : []), ...hotels.map(hotel => hotel.city)].filter(name => name && !sameCity(name, origin)))];
    const stays = (stayNames.length ? stayNames : hotels.map(hotel => hotel.city).filter(Boolean)).map((city, index) => {
        const incoming = legs.filter(leg => sameCity(leg.to, city));
        const outgoing = legs.filter(leg => sameCity(leg.from, city) && !sameCity(leg.to, city));
        const options = hotels.filter(hotel => sameCity(hotel.city, city)).map(hotelFromPart);
        const nights = options[0]?.nights || nightCount(incoming[0]?.arrival, outgoing[0]?.departure);
        return { id: `stay-${index}`, ...cityInfo(city), arrival: incoming[0]?.arrival || '', departure: outgoing[0]?.departure || incoming[0]?.arrival || '', nights, hotelOptions: options, hotel: options[0] || null, activities: [] };
    });
    const hasReturn = legs.some(leg => sameCity(leg.to, origin) && !sameCity(leg.from, origin));
    const stops = [{ id: 'origin', ...cityInfo(origin), arrival: originLeg?.departure || stays[0]?.arrival || '', departure: originLeg?.departure || '', nights: 0, hotelOptions: [], hotel: null, activities: [] }, ...stays];
    if (hasReturn)
        stops.push({ id: 'return', ...cityInfo(origin), arrival: legs.filter(leg => sameCity(leg.to, origin)).at(-1)?.arrival || '', departure: '', nights: 0, hotelOptions: [], hotel: null, activities: [] });
    const segments = [];
    for (let i = 0; i < stops.length - 1; i++) {
        const options = legs.filter(leg => sameCity(leg.from, stops[i].city) && sameCity(leg.to, stops[i + 1].city)).map(ticketFromLeg);
        if (!options.length)
            return null;
        segments.push({ id: `segment-${i}`, fromStopId: stops[i].id, toStopId: stops[i + 1].id, from: stops[i].city, to: stops[i + 1].city, options, ticket: options[0] });
    }
    return finishTrip(stops, segments, sessionId, [...legs, ...hotels].some(item => item.source === 'synthetic'), legs.find(leg => leg.currency)?.currency || 'RUB');
}
function tripFromSearch(search, sessionId) {
    const ready = frontendTrip(search.trip, sessionId);
    if (ready)
        return ready;
    if (!search.parts)
        return null;
    return assembleParts(search.parts, sessionId);
}
function searchNotes(search) {
    const parts = search.parts || {};
    return [...(parts.limitations || []), ...(parts.errors || []).map(error => error.message || error.code || '').filter(Boolean)];
}
async function searchReady(sessionId, dialog) {
    if (!dialog.readyForSearch)
        return { ...dialog, trip: null };
    await request(`${sessionPath(sessionId)}/confirm`, { method: 'POST', body: {} });
    const search = await request(`${sessionPath(sessionId)}/search`, { method: 'POST', body: {} });
    const trip = tripFromSearch(search, sessionId);
    const notes = searchNotes(search);
    if (trip)
        trip.warnings = [...notes, ...(trip.warnings || [])];
    const message = [dialog.message, notes.join('\n')].filter(Boolean).join('\n');
    return { ...dialog, trip, message };
}
function addDays(iso, nights) {
    const [year, month, day] = String(iso).split('-').map(Number);
    const date = new Date(year, month - 1, day);
    date.setDate(date.getDate() + Number(nights));
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}
function applyCriteria(request, criteria) {
    const next = { ...(request || {}), schema_version: '1.0', locale: request?.locale || 'ru', currency: 'RUB' };
    next.origin = { ...(request?.origin || {}), city: criteria.origin };
    next.destination = { ...(request?.destination || {}), city: criteria.destination };
    next.dates = { ...(request?.dates || {}), departure_date: criteria.startDate, nights: criteria.nights, return_date: addDays(criteria.startDate, criteria.nights) };
    const ages = (criteria.childrenAges || []).filter(age => Number.isInteger(age) && age >= 0);
    next.travelers = { ...(request?.travelers || {}), adults: criteria.adults, children_ages: ages };
    if (criteria.pets)
        next.travelers.pets = [{ kind: 'dog' }];
    else
        delete next.travelers.pets;
    if (criteria.budgetMinor > 0)
        next.budget = { ...(request?.budget || {}), total: criteria.budgetMinor, scope: request?.budget?.scope || 'total' };
    if (!next.transport)
        next.transport = { modes: ['plane', 'train'], round_trip: true };
    if (!next.hotel)
        next.hotel = { needed: true, rooms: 1 };
    return next;
}
export const api = {
    async login({ email, password }) {
        return storeAccount(await request('/auth/login', { method: 'POST', body: { email, password }, anonymous: true }));
    },
    async register({ email, password }) {
        return storeAccount(await request('/auth/register', { method: 'POST', body: { email, password }, anonymous: true }));
    },
    async createSession() {
        const result = await request('/sessions', { method: 'POST', body: {} });
        if (!result.session_id)
            throw new Error('Сервер не вернул сессию.');
        setSessionId(result.session_id);
        return result.session_id;
    },
    async sendMessage({ trip, text, criteria }) {
        if (!isDemo) {
            const sessionId = trip?.id || getSessionId();
            const raw = await request(`${sessionPath(sessionId)}/messages`, { method: 'POST', body: { text } });
            return searchReady(sessionId, chatFromDialog(raw));
        }
        await pause();
        const current = trip?.criteria || DEFAULT_CRITERIA;
        const parsed = criteria ? { patch: criteria, handled: true, cheaper: criteria.budgetMinor < current.budgetMinor } : parseMessage(text, current);
        if (!parsed.handled)
            return { trip, message: 'В демо я умею менять направление, бюджет, даты и состав гостей. Попробуйте «В Сочи до 100 тысяч», «На два дня позже» или «С нами собака». Для свободного диалога команда подключит AI через API.', actions: [] };
        const nextCriteria = { ...current, ...parsed.patch };
        let next = createTrip(nextCriteria, trip ? selectedIds(trip) : {}, trip?.id || `trip-${crypto.randomUUID()}`);
        if (parsed.cheaper)
            next = createTrip(nextCriteria, cheaperSelection(next), next.id);
        next.revision = (trip?.revision || 0) + 1;
        const overBudget = next.totals.totalMinor > next.criteria.budgetMinor;
        const changes = [];
        if (parsed.patch.startDate)
            changes.push('обновил даты');
        if (parsed.patch.destination)
            changes.push('подобрал направление');
        if (parsed.patch.pets !== undefined)
            changes.push('проверил условия для питомца');
        if (parsed.cheaper)
            changes.push('сравнил стоимость вариантов');
        if (criteria)
            changes.push('сохранил параметры');
        const tradeoffs = [next.segments.some(s => !s.ticket.baggage) ? 'Часть билетов включает только ручную кладь.' : '', next.stops.some(s => s.hotel?.cancellation === 'Без возврата') ? 'У части отелей выбран невозвратный тариф.' : ''].filter(Boolean).join(' ');
        return {
            trip: next,
            message: `${changes.length ? changes.join(', ').replace(/^./, c => c.toUpperCase()) : 'Обновил поездку'}. ${next.stops.filter(s => s.nights).map(s => `${s.city} — ${s.nights} ночей`).join(', ')}. Полная стоимость: ${money(next.totals.totalMinor)} на всех гостей, включая трансферы.${overBudget ? ' Вариант выше бюджета: попробуйте сократить поездку или выбрать другое направление.' : ' Можно сравнить билеты и отели в карточках справа.'}${tradeoffs ? ` ${tradeoffs}` : ''}${next.criteria.pets ? ' Доплата за перевозку питомца требует уточнения и пока не включена.' : ''}`,
            actions: [{ label: 'На 2 дня позже', value: 'На два дня позже' }, { label: 'Найти дешевле', value: 'Хочу дешевле' }]
        };
    },
    async submitAnswers(sessionId, answers) {
        const raw = await request(`${sessionPath(sessionId)}/answers`, { method: 'POST', body: { answers } });
        return searchReady(sessionId, chatFromDialog(raw));
    },
    async selectIdea(sessionId, ideaId, comment = '') {
        const raw = await request(`${sessionPath(sessionId)}/ideas`, { method: 'POST', body: { idea_id: ideaId, comment } });
        return searchReady(sessionId, chatFromDialog(raw));
    },
    async getRoutes(sessionId) {
        const search = await request(`${sessionPath(sessionId)}/routes`);
        const notes = searchNotes(search);
        const trip = tripFromSearch(search, sessionId);
        if (trip)
            trip.warnings = [...notes, ...(trip.warnings || [])];
        return { trip, message: notes.join('\n'), actions: [], questions: [] };
    },
    async updateConfig(sessionId, criteria) {
        const current = await request(`${sessionPath(sessionId)}/config`);
        const config = current.config && typeof current.config === 'object' ? { ...current.config } : { version: '1.0' };
        config.version = config.version || '1.0';
        config.request = applyCriteria(config.request, criteria);
        await request(`${sessionPath(sessionId)}/config`, { method: 'PATCH', body: { config } });
        await request(`${sessionPath(sessionId)}/confirm`, { method: 'POST', body: {} });
        const search = await request(`${sessionPath(sessionId)}/search`, { method: 'POST', body: {} });
        const notes = searchNotes(search);
        const trip = tripFromSearch(search, sessionId);
        if (trip)
            trip.warnings = [...notes, ...(trip.warnings || [])];
        return {
            trip,
            message: notes.length ? notes.join('\n') : 'Параметры сохранены. Ищу варианты.',
            actions: [],
            questions: []
        };
    },
    async selectOffer(trip, type, itemId, offerId) {
        if (!isDemo)
            return request(`/trips/${encodeURIComponent(trip.id)}/selection`, { method: 'PATCH', body: { type, itemId, offerId, revision: trip.revision } });
        await pause();
        const item = type === 'hotel' ? trip.stops.find(s => s.id === itemId) : trip.segments.find(s => s.id === itemId);
        const offers = type === 'hotel' ? item?.hotelOptions : item?.options;
        if (!offers?.some(o => o.id === offerId))
            throw new Error('Этот вариант больше недоступен. Обновите поездку.');
        const next = createTrip(trip.criteria, { ...selectedIds(trip), [itemId]: offerId }, trip.id);
        next.revision = trip.revision + 1;
        return { trip: next };
    },
    async book(trip, requestId) {
        if (!isDemo)
            return request(`/trips/${encodeURIComponent(trip.id)}/bookings`, { method: 'POST', body: { revision: trip.revision, demo: true }, requestId });
        await pause();
        return { booking: { id: `DEMO-${requestId.replace(/-/g, '').slice(0, 8).toUpperCase()}`, status: 'confirmed_demo', createdAt: new Date().toISOString(), demo: true }, trip: { ...trip, status: 'booked' } };
    },
    async getProfile() {
        if (!isDemo) {
            const user = await request('/auth/me');
            if (user.id)
                setUserId(user.id);
            return { profile: { name: user.email || 'Путешественник', email: user.email || '', phone: '' } };
        }
        return { profile: readStorage('profile', { name: 'Путешественник', email: '', phone: '' }) };
    },
    async saveProfile(profile) {
        if (!isDemo)
            return request('/me', { method: 'PATCH', body: profile });
        writeStorage('profile', profile);
        return { profile };
    },
    async getHistory() {
        if (!isDemo)
            return request('/me/trips');
        return { trips: readStorage('history', []) };
    },
    async getTrip(id) {
        if (!isDemo)
            return request(`/trips/${encodeURIComponent(id)}`);
        const trip = readStorage('history', []).find(item => item.trip.id === id)?.trip;
        if (!trip)
            throw new Error('Не удалось найти эту поездку.');
        return { trip };
    }
};
export function readStorage(key, fallback) {
    try {
        const value = localStorage.getItem(`odin-v1:${key}`);
        return value ? JSON.parse(value) : fallback;
    }
    catch {
        return fallback;
    }
}
export function writeStorage(key, value) {
    try {
        localStorage.setItem(`odin-v1:${key}`, JSON.stringify(value));
        return true;
    }
    catch {
        return false;
    }
}
export function saveDemoBooking(booking, trip) {
    const history = readStorage('history', []);
    if (!history.some(item => item.booking.id === booking.id))
        history.unshift({ booking, trip });
    return writeStorage('history', history);
}
