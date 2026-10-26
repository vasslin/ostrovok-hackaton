import { DEFAULT_CRITERIA, createTrip, parseMessage, selectedIds, cheaperSelection, money } from './model.js';
const config = window.TRAVEL_CONFIG || { mode: 'demo', apiBaseUrl: '/api' };
export const isDemo = config.mode !== 'api';
const pause = () => new Promise(resolve => setTimeout(resolve, 450));
async function request(path, { method = 'GET', body, requestId } = {}) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 20000);
    try {
        const response = await fetch(`${config.apiBaseUrl.replace(/\/$/, '')}${path}`, {
            method, credentials: 'same-origin', signal: controller.signal,
            headers: { Accept: 'application/json', ...(body ? { 'Content-Type': 'application/json' } : {}), ...(requestId ? { 'Idempotency-Key': requestId } : {}) },
            ...(body ? { body: JSON.stringify(body) } : {})
        });
        const json = await response.json().catch(() => ({}));
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
export const api = {
    async sendMessage({ trip, text, criteria }) {
        if (!isDemo)
            return request('/chat/messages', { method: 'POST', body: { tripId: trip?.id || null, text, criteria: criteria || null }, requestId: crypto.randomUUID() });
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
        if (!isDemo)
            return request('/me');
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
