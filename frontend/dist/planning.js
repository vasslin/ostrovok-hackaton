export function providerIssue(issue) {
    const code = issue.code || '';
    const hotel = issue.kind === 'hotel' || ['hotel', 'hotelbeds', 'ratehawk'].includes(issue.provider);
    if (code === 'clarification_required') return issue.message;
    if (hotel && ['not_configured', 'credentials_missing'].includes(code)) return 'Поиск отелей не подключён. Подключите гостиничного провайдера или выберите поиск только транспорта.';
    if (hotel && code === 'no_results') return 'Подходящие отели не найдены. Попробуйте другие даты или требования.';
    if (code === 'missing_field') {
        if (/residency|source_market/.test(issue.message)) return 'Для поиска отеля укажите гражданство гостей.';
        if (/hotel_check|stay/.test(issue.message)) return 'Для отеля нужны подтверждённые даты заезда и выезда.';
        if (/resolved/.test(issue.message)) return 'Не удалось определить коды города. Уточните город и страну в параметрах поездки.';
        return 'Для поиска не хватает параметров. Проверьте карточку поездки.';
    }
    if (code === 'timeout') return hotel ? 'Гостиничный сервис не ответил вовремя. Можно повторить поиск или искать только транспорт.' : 'Транспортный сервис не ответил вовремя. Показаны доступные частичные результаты.';
    if (hotel) return 'Ошибка гостиничного сервиса. Проживание не подобрано; можно повторить поиск или искать только транспорт.';
    if (issue.provider === 'places') return issue.message;
    return 'Один из поставщиков транспорта недоступен. Показаны результаты остальных поставщиков.';
}
export function reviewRequest(base, data, transportOnly = false) {
    const next = structuredClone(base || {});
    next.schema_version = '1.0';
    next.locale ||= 'ru';
    next.currency = String(data.get('currency') || next.currency || 'RUB');
    const city = name => ({ city: String(data.get(name) || '').trim(), ...(String(data.get(`${name}Country`) || '').trim() ? { country_code: String(data.get(`${name}Country`)).trim().toUpperCase() } : {}) });
    next.origin = city('origin');
    next.destination = { ...(next.destination || {}), ...city('destination') };
    if (!next.destination.country_code || !String(data.get('destinationCountry') || '').trim()) delete next.destination.country_code;
    next.dates = { ...(next.dates || {}), departure_date: String(data.get('departure') || '') };
    delete next.dates.departure_month;
    delete next.dates.nights;
    if (data.get('return')) next.dates.return_date = String(data.get('return')); else delete next.dates.return_date;
    const ages = String(data.get('children') || '').trim();
    const children = ages ? ages.split(',').map(value => Number(value.trim())) : [];
    if (children.some(age => !Number.isInteger(age) || age < 0 || age > 17)) throw new Error('Возраст каждого ребёнка должен быть от 0 до 17 лет.');
    next.travelers = { ...(next.travelers || {}), adults: Number(data.get('adults')), children_ages: children };
    const residency = String(data.get('residency') || '').trim().toLowerCase();
    if (residency) next.travelers.residency = residency; else delete next.travelers.residency;
    const budget = Number(data.get('budget'));
    if (budget > 0) next.budget = { ...(next.budget || {}), total: budget, scope: 'total' }; else delete next.budget;
    next.hotel = { ...(next.hotel || {}), needed: !transportOnly && data.get('hotel') === 'on' };
    const modes = data.getAll('modes');
    if (modes.length) next.transport = { ...(next.transport || {}), modes, round_trip: data.get('roundTrip') === 'on' };
    else throw new Error('Выберите хотя бы один вид транспорта.');
    delete next.completeness;
    return next;
}
