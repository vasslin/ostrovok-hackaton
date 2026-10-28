export const DEFAULT_CRITERIA = {
    origin: 'Москва', destination: 'Анталья', startDate: '2026-11-10', nights: 7,
    adults: 2, childrenAges: [7], budgetMinor: 15000000, currency: 'RUB', pets: false
};
export const cities = {
    'Москва': { lat: 55.7558, lng: 37.6173, timezone: '+03:00', country: 'Россия' },
    'Стамбул': { lat: 41.0082, lng: 28.9784, timezone: '+03:00', country: 'Турция' },
    'Анталья': { lat: 36.8969, lng: 30.7133, timezone: '+03:00', country: 'Турция' },
    'Сочи': { lat: 43.5855, lng: 39.7231, timezone: '+03:00', country: 'Россия' },
    'Тбилиси': { lat: 41.7151, lng: 44.8271, timezone: '+04:00', country: 'Грузия' },
    'Калининград': { lat: 54.7104, lng: 20.4522, timezone: '+02:00', country: 'Россия' },
    'Казань': { lat: 55.7961, lng: 49.1064, timezone: '+03:00', country: 'Россия' },
    'Санкт-Петербург': { lat: 59.9343, lng: 30.3351, timezone: '+03:00', country: 'Россия' }
};
const destinationInfo = {
    'Анталья': { country: 'Турция', title: 'Два города. Одно маленькое приключение.', subtitle: 'Улицы Стамбула, а потом - море и спокойные дни в Анталье.', tags: ['Море и город', 'Для семьи', '7 ночей'], cover: 'sea', hotel: ['Lara Garden Residence', 'Old Town Family Hotel', 'Azure Coast Resort'], activity: ['Прогулка по старому городу', 'Семейный день у моря'], daily: 7900, flight: 12900 },
    'Сочи': { country: 'Россия', title: 'К морю, ближе к горам.', subtitle: 'Неспешные прогулки по набережной и день среди горных вершин.', tags: ['Без пересадок', 'Море и горы', 'Для семьи'], cover: 'mountain', hotel: ['Зелёный мыс', 'Морской квартал', 'Гранд Парк Сочи'], activity: ['Прогулка по дендрарию', 'День в горах'], daily: 5200, flight: 8200 },
    'Тбилиси': { country: 'Грузия', title: 'Город, который умеет встречать.', subtitle: 'Старые улочки, уютные дворики и новые любимые места.', tags: ['Городские прогулки', 'Гастрономия', 'Культура'], cover: 'city', hotel: ['Courtyard Stories', 'Old Tbilisi Rooms', 'Terrace Boutique'], activity: ['Прогулка по старому Тбилиси', 'Музей и канатная дорога'], daily: 5600, flight: 11900 },
    'Калининград': { country: 'Россия', title: 'Балтика. Дюны. Время для себя.', subtitle: 'Прогулки у воды и короткие поездки к песчаным берегам.', tags: ['Короткий перелёт', 'Балтийское море', 'Прогулки'], cover: 'sea', hotel: ['Балтийские истории', 'Дом у реки', 'Amber Residence'], activity: ['Прогулка по острову Канта', 'День на Балтийском побережье'], daily: 4600, flight: 6900 }
};
export function addDays(date, days) {
    const d = new Date(`${date}T12:00:00Z`);
    d.setUTCDate(d.getUTCDate() + days);
    return d.toISOString().slice(0, 10);
}
export function money(minor, currency = 'RUB') { if (!Number.isFinite(minor)) return 'Цена уточняется'; return new Intl.NumberFormat('ru-RU', { style: 'currency', currency, maximumFractionDigits: 0 }).format(minor / 100); }
export function dateLabel(date, opts = {}) { return new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'short', timeZone: 'UTC', ...opts }).format(new Date(`${date.slice(0, 10)}T12:00:00Z`)).replace('.', ''); }
export function localTime(date) { return date.slice(11, 16); }
export function guestsLabel(c) { return `${c.adults} взр.${c.childrenAges.length ? ` · ${c.childrenAges.length} реб.` : ''}${c.pets ? ' · с питомцем' : ''}`; }
function dateTime(date, time, city) { return `${date}T${time}:00${cities[city].timezone}`; }
export function validateCriteria(c) {
    if (!cities[c.origin])
        throw new Error('Пока доступны вылеты из Москвы, Казани и Санкт-Петербурга.');
    if (!destinationInfo[c.destination])
        throw new Error('В демокаталоге доступны Анталья, Сочи, Тбилиси и Калининград.');
    if (c.origin === c.destination)
        throw new Error('Город отправления и назначения должны отличаться.');
    if (!Number.isInteger(c.nights) || c.nights < 2 || c.nights > 21)
        throw new Error('Для демо выберите от 2 до 21 ночи.');
    if (!Number.isInteger(c.adults) || c.adults < 1 || c.adults > 6)
        throw new Error('Выберите от 1 до 6 взрослых.');
    if (!Array.isArray(c.childrenAges) || c.childrenAges.length > 4 || c.childrenAges.some(a => !Number.isInteger(a) || a < 0 || a > 17))
        throw new Error('Укажите до 4 детей, возраст - от 0 до 17 лет.');
    if (!Number.isFinite(c.budgetMinor) || c.budgetMinor <= 0)
        throw new Error('Бюджет должен быть больше нуля.');
    if (!/^\d{4}-\d{2}-\d{2}$/.test(c.startDate) || !Number.isFinite(Date.parse(c.startDate)) || new Date(`${c.startDate}T12:00:00Z`).toISOString().slice(0, 10) !== c.startDate)
        throw new Error('Укажите существующую дату поездки.');
}
export function createTrip(criteria, selected = {}, id = 'trip-demo') {
    const c = { ...DEFAULT_CRITERIA, ...criteria };
    validateCriteria(c);
    const info = destinationInfo[c.destination];
    const via = c.destination === 'Анталья' && c.nights >= 5;
    const intermediateNights = via ? 2 : 0;
    const endDate = addDays(c.startDate, c.nights);
    const guestFactor = c.adults + c.childrenAges.reduce((s, age) => s + (age < 2 ? .1 : .6), 0);
    const rooms = Math.ceil((c.adults + c.childrenAges.length) / 3);
    const stop = (city, index, day, stayNights) => ({
        id: `stop-${index}`, city, country: cities[city].country, lat: cities[city].lat, lng: cities[city].lng,
        arrival: dateTime(addDays(c.startDate, day), index === 0 ? '07:00' : '12:30', city),
        departure: dateTime(addDays(c.startDate, day + stayNights), '15:00', city),
        nights: stayNights, hotelOptions: [], hotel: null, activities: []
    });
    const stops = [stop(c.origin, 0, 0, 0)];
    if (via)
        stops.push(stop('Стамбул', 1, 0, intermediateNights));
    stops.push(stop(c.destination, stops.length, intermediateNights, c.nights - intermediateNights));
    stops.push(stop(c.origin, stops.length, c.nights, 0));
    for (const s of stops.filter(s => s.nights > 0)) {
        const isIstanbul = s.city === 'Стамбул';
        const names = isIstanbul ? ['Pera Family House', 'Golden Horn Rooms', 'Bosphorus Garden'] : info.hotel;
        const nightly = isIstanbul ? 5900 : info.daily;
        s.hotelOptions = names.map((name, i) => ({
            id: `hotel-${s.city}-${i}`, name, city: s.city, rating: [9.2, 8.7, 9.6][i],
            stars: [4, 3, 5][i], priceMinor: Math.round(nightly * [1, .72, 1.4][i] * s.nights * rooms * 100),
            nightlyMinor: Math.round(nightly * [1, .72, 1.4][i] * rooms * 100), nights: s.nights, rooms,
            petsAllowed: i !== 0, cancellation: i === 1 ? 'Без возврата' : 'Бесплатная отмена за 7 дней',
            meals: i === 2 ? 'Завтрак и ужин' : 'Завтрак включён', distance: ['1,2 км от центра', 'В центре', 'У набережной'][i],
            cover: isIstanbul ? 'city' : info.cover, demo: true
        })).filter(h => !c.pets || h.petsAllowed);
        s.hotel = s.hotelOptions.find(h => h.id === selected[s.id]) || s.hotelOptions[0];
        s.activities = (isIstanbul ? ['Прогулка по Султанахмету', 'Прогулка у Босфора'] : info.activity).map((name, i) => ({ id: `act-${s.id}-${i}`, name, duration: i ? '2 часа' : '1,5 часа', priceMinor: 0, note: 'Самостоятельная прогулка · без билетов', demo: true }));
    }
    const segments = [];
    for (let i = 0; i < stops.length - 1; i++) {
        const from = stops[i], to = stops[i + 1];
        const date = i === 0 ? c.startDate : to.nights === 0 ? endDate : addDays(c.startDate, intermediateNights);
        const base = via ? [10900, 3400, info.flight][i] : info.flight;
        const originExtra = c.origin === 'Москва' ? 0 : 2200;
        const options = [0, 1, 2].map(j => ({
            id: `ticket-${from.city}-${to.city}-${j}`, carrier: ['Demo Air', 'Travel Airlines', 'Sky Routes'][j],
            number: `DM ${1200 + i * 100 + j * 11}`, type: 'flight',
            departure: dateTime(date, ['08:00', '11:00', '14:00'][j], from.city),
            arrival: dateTime(date, ['12:30', '15:30', '18:30'][j], to.city),
            priceMinor: Math.round((base * [1, .83, 1.22][j] + originExtra) * guestFactor * 100),
            baggage: j !== 1, refundable: j === 2, petsAllowed: j !== 1, demo: true
        })).filter(t => !c.pets || t.petsAllowed);
        const segment = { id: `segment-${i}`, fromStopId: from.id, toStopId: to.id, from: from.city, to: to.city, options, ticket: null };
        segment.ticket = options.find(t => t.id === selected[segment.id]) || options[0];
        to.arrival = segment.ticket.arrival;
        from.departure = segment.ticket.departure;
        segments.push(segment);
    }
    const transportMinor = segments.reduce((n, s) => n + s.ticket.priceMinor, 0);
    const hotelsMinor = stops.reduce((n, s) => n + (s.hotel?.priceMinor || 0), 0);
    const transferMinor = (via ? 3600 : 2400) * 100;
    return {
        id, revision: 1, status: 'draft', demo: true, title: info.title, subtitle: info.subtitle,
        criteria: c, startDate: c.startDate, endDate, stops, segments,
        totals: { currency: 'RUB', transportMinor, hotelsMinor, transferMinor, activitiesMinor: 0, totalMinor: transportMinor + hotelsMinor + transferMinor },
        warnings: c.pets ? ['Перевозку питомца нужно согласовать с перевозчиком. Доплата за питомца в демо не рассчитана.'] : [],
        tags: info.tags, cover: info.cover
    };
}
export function selectedIds(trip) { return Object.fromEntries([...trip.stops.filter(s => s.hotel).map(s => [s.id, s.hotel.id]), ...trip.segments.map(s => [s.id, s.ticket.id])]); }
export function cheaperSelection(trip) {
    return Object.fromEntries([...trip.stops.filter(s => s.hotel).map(s => [s.id, [...s.hotelOptions].sort((a, b) => a.priceMinor - b.priceMinor)[0].id]), ...trip.segments.map(s => [s.id, [...s.options].sort((a, b) => a.priceMinor - b.priceMinor)[0].id])]);
}
export function parseMessage(text, current) {
    const t = text.toLowerCase().trim(), patch = {};
    let handled = false;
    const destinations = [['антал', 'Анталья'], ['турци', 'Анталья'], ['сочи', 'Сочи'], ['тбилис', 'Тбилиси'], ['грузи', 'Тбилиси'], ['калининград', 'Калининград'], ['балтик', 'Калининград']];
    for (const [key, value] of destinations)
        if (t.includes(key)) {
            patch.destination = value;
            handled = true;
            break;
        }
    if (/мор[еюя]|пляж/.test(t) && !patch.destination) {
        patch.destination = /без загран|по россии|без пересад/.test(t) ? 'Сочи' : 'Анталья';
        handled = true;
    }
    for (const [regex, city] of [[/из москв/, 'Москва'], [/из казан/, 'Казань'], [/из (питер|петербург|санкт-петербург)/, 'Санкт-Петербург']])
        if (regex.test(t)) {
            patch.origin = city;
            handled = true;
        }
    const budget = t.match(/(?:до|бюджет(?:ом)?|не больше)\s*(\d[\d\s]*(?:[.,]\d+)?)\s*(тыс(?:яч)?\.?|к\b|₽|руб)?/);
    if (budget) {
        const v = Number(budget[1].replace(/\s/g, '').replace(',', '.'));
        patch.budgetMinor = Math.round(v * (budget[2]?.startsWith('тыс') || budget[2] === 'к' ? 1000 : 1) * 100);
        handled = true;
    }
    if (/с (?:нами )?(?:собак|кот|питомц)/.test(t)) {
        patch.pets = true;
        handled = true;
    }
    if (/без (собак|кот|питомц)/.test(t)) {
        patch.pets = false;
        handled = true;
    }
    if (/без дет/.test(t)) {
        patch.childrenAges = [];
        handled = true;
    }
    if (/с реб[её]нком/.test(t)) {
        const age = t.match(/реб[её]н(?:ком|ок|ку|ка)\s*(\d{1,2})/);
        patch.childrenAges = [age ? Number(age[1]) : 7];
        handled = true;
    }
    if (/вдво[её]м/.test(t)) {
        patch.adults = 2;
        patch.childrenAges = [];
        handled = true;
    }
    if (/один|одна|соло/.test(t)) {
        patch.adults = 1;
        patch.childrenAges = [];
        handled = true;
    }
    const adults = t.match(/(\d)\s*взросл/);
    if (adults) {
        patch.adults = Number(adults[1]);
        handled = true;
    }
    const nights = t.match(/(?:на\s*)?(\d{1,2})\s*(ноч|дн)/);
    if (nights && !/позже|раньше/.test(t)) {
        patch.nights = Number(nights[1]);
        handled = true;
    }
    if (/недел/.test(t)) {
        patch.nights = /две|2\s*недел/.test(t) ? 14 : 7;
        handled = true;
    }
    if (/позже|раньше/.test(t)) {
        const delta = t.match(/на\s*(\d+)/);
        const n = delta ? Number(delta[1]) : /два|двое/.test(t) ? 2 : 1;
        patch.startDate = addDays(current.startDate, /раньше/.test(t) ? -n : n);
        handled = true;
    }
    const iso = t.match(/\b(20\d{2}-\d{2}-\d{2})\b/);
    if (iso) {
        patch.startDate = iso[1];
        handled = true;
    }
    const russianDate = t.match(/(?:с\s*)?(\d{1,2})\s*(ноябр|декабр|январ|феврал|март|апрел|ма[яй]|июн|июл|август|сентябр|октябр)/);
    if (russianDate) {
        const months = ['январ', 'феврал', 'март', 'апрел', 'ма', 'июн', 'июл', 'август', 'сентябр', 'октябр', 'ноябр', 'декабр'];
        const month = months.findIndex(m => russianDate[2].startsWith(m)) + 1;
        patch.startDate = `${current.startDate.slice(0, 4)}-${String(month).padStart(2, '0')}-${russianDate[1].padStart(2, '0')}`;
        handled = true;
    }
    return { patch, cheaper: /дешев|эконом|меньше бюджет/.test(t) || !!budget, handled: handled || /дешев|эконом|меньше бюджет/.test(t) };
}
