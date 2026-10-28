import { view, loadViews } from './templates.js';
import { icon } from './icons.js';
import { DEFAULT_CRITERIA, cities, createTrip, money, dateLabel, localTime, guestsLabel } from './model.js';
import { api, isDemo, readStorage, writeStorage, saveDemoBooking, UnauthorizedError, getAccessToken, getSessionId } from './api.js';
import { RouteMap } from './map.js';
const $ = selector => document.querySelector(selector);
const esc = value => String(value ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const app = $('#app');
const starterMessage = { role: 'assistant', text: 'Привет! Куда хочется отправиться? Можно начать с идеи: «К морю с ребёнком, до 150 тысяч». Справа — пример: из Москвы, 10 ноября, на 7 ночей, 2 взрослых и ребёнок 7 лет. Все цены и варианты демонстрационные.', actions: [{ label: 'К морю с ребёнком', value: 'К морю с ребёнком до 150 тысяч' }, { label: 'В Сочи до 100 тысяч', value: 'В Сочи до 100 тысяч' }] };
let initialTrip = isDemo ? createTrip(DEFAULT_CRITERIA) : null;
const persisted = isDemo ? readStorage('session', null) : null;
if (persisted?.trip && Array.isArray(persisted.trip.stops) && Array.isArray(persisted.trip.segments)) {
    try {
        initialTrip = persisted.trip;
    }
    catch { }
}
const state = {
    page: 'planner', tab: 'overview', trip: initialTrip,
    messages: persisted?.messages?.length ? persisted.messages : [starterMessage],
    profile: { name: 'Путешественник', email: '', phone: '' }, history: [],
    busy: false, bookingBusy: false, mobileView: 'route', drawer: null, modal: null,
    booking: persisted?.booking || null, error: null, draftText: '', profileBusy: false,
    loginEmail: '', loginError: '', loginBusy: false, authMode: 'login', pendingSend: null,
    sessionId: isDemo ? '' : getSessionId(), answerDraft: {}
};
let map, lastFocus, toastTimer;
const price = amount => Number.isFinite(amount) ? money(amount, state.trip?.totals.currency || 'RUB') : 'цена неизвестна';
const button = (text, action, cls = '', attrs = '') => view('button-button', {
    cls: cls,
    action: action,
    attrs: attrs,
    text: text
});
function notify(text) { $('#toast').textContent = text; $('#toast').classList.add('visible'); clearTimeout(toastTimer); toastTimer = setTimeout(() => $('#toast').classList.remove('visible'), 4000); }
function persist() { if (isDemo && !writeStorage('session', { trip: state.trip, messages: state.messages.slice(-60), booking: state.booking }))
    notify('Память браузера недоступна: изменения сохранятся только до закрытия страницы.'); }
function draftTrip() {
    if (isDemo)
        return Promise.resolve(state.trip?.status === 'booked' ? { ...state.trip, id: `trip-${crypto.randomUUID()}`, status: 'draft' } : state.trip);
    const needsNew = !getSessionId() || state.trip?.status === 'booked';
    const ready = needsNew ? api.createSession() : Promise.resolve(getSessionId());
    return ready.then(id => {
        state.sessionId = id;
        if (state.trip?.status === 'booked')
            return { ...state.trip, id, status: 'draft' };
        if (state.trip)
            return state.trip.id === id ? state.trip : { ...state.trip, id };
        return { id };
    });
}
function formatRange(trip = state.trip) { return trip ? `${dateLabel(trip.startDate)} — ${dateLabel(trip.endDate)}` : 'Даты поездки'; }
function render() {
    map?.destroy();
    map = null;
    const planner = state.page === 'planner';
    app.innerHTML = view('render-app-shell', {
        plannerButton: button(icon('spark'), 'planner', `rail-button ${planner ? 'active' : ''}`, 'aria-label="Составить поездку" title="Составить поездку"'),
        historyButton: button(icon('bag'), 'history', `rail-button ${state.page === 'history' ? 'active' : ''}`, 'aria-label="Мои поездки" title="Мои поездки"'),
        profileButton: button(icon('user'), 'profile', `rail-button ${state.page === 'profile' ? 'active' : ''}`, 'aria-label="Личный кабинет" title="Личный кабинет"'),
        profileButton2: button(esc((state.profile.name || 'П').slice(0, 1).toUpperCase()), 'profile', 'avatar', 'aria-label="Открыть профиль"'),
        conditional: isDemo ? 'Демо · без реальных списаний' : 'Подключён API',
        profileButton3: button(view('render-span', {
            userIcon: icon('user'),
            name: esc(state.profile.name)
        }), 'profile', 'profile-link', 'aria-label="Открыть личный кабинет"'),
        conditional2: planner ? renderPlanner() : state.page === 'profile' ? renderProfile() : state.page === 'history' ? renderHistory() : renderSummary(),
        conditional3: isDemo ? 'Демонстрационные данные · бронирование имитируется' : 'Данные от API команды',
        plannerButton2: button(view('render-span-2', {
            sparkIcon: icon('spark')
        }), 'planner', planner ? 'active' : ''),
        historyButton2: button(view('render-span-3', {
            bagIcon: icon('bag')
        }), 'history', state.page === 'history' ? 'active' : ''),
        profileButton4: button(view('render-span-4', {
            userIcon: icon('user')
        }), 'profile', state.page === 'profile' ? 'active' : ''),
        renderedDrawer: renderDrawer(),
        renderedModal: renderModal()
    });
    if (planner && state.trip) {
        map = new RouteMap($('#route-map'), { onStop: id => openDrawer('stop', id), onSegment: id => openDrawer('segment', id) });
        map.update(state.trip, state.drawer?.type === 'stop' ? state.drawer.id : null);
    }
    const messages = $('#messages');
    if (messages)
        messages.scrollTop = messages.scrollHeight;
    if (state.modal || state.drawer) {
        const panel = $('[role="dialog"]');
        panel?.querySelector('input,button,select')?.focus();
    }
}
function renderPlanner() {
    return view('render-planner-workspace', {
        conditional: state.mobileView === 'chat' ? 'mobile-chat' : '',
        mobileChatButton: button(`${icon('chat')} Чат`, 'mobile-chat', state.mobileView === 'chat' ? 'active' : ''),
        mobileRouteButton: button(`${icon('map')} Маршрут`, 'mobile-route', state.mobileView === 'route' ? 'active' : ''),
        sparkIcon: icon('spark'),
        conditional2: isDemo ? 'Демо-помощник · готов к идеям' : 'Помощник по путешествиям',
        stateMessages: state.messages.map((m, i) => view('render-planner-message', {
            mRole: m.role,
            conditional: m.role === 'assistant' ? view('render-planner-message-avatar', {
                sparkIcon: icon('spark')
            }) : '',
            replace: esc(m.text).replace(/\n/g, '<br>'),
            conditional2: [
                m.questions?.length && i === state.messages.findLastIndex(item => item.questions?.length) ? view('render-planner-questions', {
                    index: i,
                    disabled: state.busy ? 'disabled' : '',
                    fields: m.questions.map(question => view('render-planner-question', {
                        text: esc(question.text),
                        questionId: esc(question.questionId),
                        value: esc(state.answerDraft[question.questionId] || ''),
                        required: question.required ? 'required' : ''
                    })).join('')
                }) : '',
                m.actions?.length ? view('render-planner-message-actions', {
                    mActions: m.actions.map((a, j) => button(esc(a.label), 'chat-action', 'suggestion', `data-message="${i}" data-index="${j}" ${state.busy ? 'disabled' : ''}`)).join('')
                }) : ''
            ].join(''),
            conditional3: m.role === 'assistant' ? (isDemo ? 'Демо-помощник' : 'Помощник') : 'Вы'
        })).join(''),
        conditional3: state.busy ? view('render-planner-thinking', {}) : '',
        conditional4: state.error ? view('render-planner-inline-error', {
            infoIcon: icon('info'),
            stateError: esc(state.error),
            retryButton: button('Повторить', 'retry', 'text-button')
        }) : '',
        quickButton: button(`${icon('wallet')} Дешевле`, 'quick', 'suggestion', 'data-text="Хочу дешевле"'),
        quickButton2: button(`${icon('calendar')} Даты +2`, 'quick', 'suggestion', 'data-text="На два дня позже"'),
        quickButton3: button(`${icon('pet')} С питомцем`, 'quick', 'suggestion', 'data-text="С нами собака"'),
        conditional5: state.busy ? 'disabled' : '',
        stateDraftText: esc(state.draftText),
        conditional6: state.busy ? 'disabled' : '',
        arrowIcon: icon('arrow'),
        shieldIcon: icon('shield'),
        conditional7: isDemo ? 'Без AI-вызовов. Настоящая модель подключается через backend.' : 'Ваши пожелания помогают собрать подходящую поездку.',
        conditional8: state.trip ? renderTrip() : view('render-planner-empty-state', {
            mapIcon: icon('map'),
            mobileChatButton: button('Открыть чат', 'mobile-chat', 'primary')
        })
    });
}
function renderTrip() {
    const t = state.trip, c = t.criteria;
    return view('render-trip-trip-intro', {
        tTitle: esc(t.title),
        tSubtitle: esc(t.subtitle),
        newTripButton: button(icon('plus') + ' Новая поездка', 'new-trip', 'button-light new-trip'),
        editCriteriaButton: button(`${icon('calendar')} ${formatRange()}`, 'edit-criteria', 'criteria-chip'),
        editCriteriaButton2: button(`${icon('user')} ${guestsLabel(c)}`, 'edit-criteria', 'criteria-chip'),
        editCriteriaButton3: button(`${icon('wallet')} ${Number.isFinite(c.budgetMinor) ? `до ${price(c.budgetMinor)}` : 'бюджет не задан'}`, 'edit-criteria', 'criteria-chip'),
        editCriteriaButton4: button('Изменить', 'edit-criteria', 'criteria-edit'),
        conditional: state.tab !== 'overview' ? 'active' : '',
        planeIcon: icon('plane'),
        value: [['overview', 'Маршрут'], ['tickets', 'Билеты'], ['hotels', 'Отели']].map(([id, label]) => button(`${label}${id === 'tickets' ? view('render-trip-span', {
            length: t.segments.length
        }) : id === 'hotels' ? view('render-trip-span-2', {
            length: t.stops.filter(s => s.hotel).length
        }) : ''}`, 'tab', state.tab === id ? 'active' : '', `role="tab" aria-selected="${state.tab === id}" data-tab="${id}"`)).join(''),
        nights: t.criteria.nights,
        length: t.stops.filter(s => s.nights).length,
        conditional2: t.stops.filter(s => s.nights).length === 1 ? 'направление' : 'направления',
        conditional3: state.tab === 'overview' ? renderItinerary() : state.tab === 'tickets' ? renderTickets() : renderHotels(),
        renderedBudgetWarning: renderBudgetWarning(),
        tWarnings: t.warnings.map(w => view('render-trip-notice', {
            infoIcon: icon('info'),
            w: esc(w)
        })).join(''),
        expression: c.adults + c.childrenAges.length,
        price: price(t.totals.totalMinor),
        showCostButton: button('Что включено?', 'show-cost', 'text-button'),
        summaryButton: button(`К оформлению ${icon('arrow')}`, 'summary', 'primary'),
        infoIcon: icon('info')
    });
}
function renderBudgetWarning() { if (!state.trip || !Number.isFinite(state.trip.criteria.budgetMinor) || !Number.isFinite(state.trip.totals.totalMinor) || state.trip.totals.totalMinor <= state.trip.criteria.budgetMinor || state.trip.warnings.some(warning => warning.startsWith('Сумма неполная')))
    return ''; return view('render-budget-warning-notice', {
    infoIcon: icon('info'),
    price: price(state.trip.totals.totalMinor - state.trip.criteria.budgetMinor),
    quickButton: button('Найти дешевле', 'quick', 'text-button', 'data-text="Хочу дешевле"')
}); }
function renderItinerary() { const t = state.trip; return view('render-itinerary-itinerary', {
    tStops: t.stops.map((s, i) => view('render-itinerary-stop-card', {
        conditional: i > 0 ? view('render-itinerary-transport-line', {
            id: t.segments[i - 1].id,
            planeIcon: icon('plane'),
            carrier: esc(t.segments[i - 1].ticket.carrier),
            number: esc(t.segments[i - 1].ticket.number),
            duration: duration(t.segments[i - 1].ticket),
            price: t.segments[i - 1].ticket.roundTripReturn ? 'В тарифе туда-обратно' : price(t.segments[i - 1].ticket.priceMinor),
            chevronIcon: icon('chevron')
        }) : '',
        conditional2: s.hotel ? 'with-hotel' : 'compact-stop',
        conditional3: i === 0 ? 'origin' : '',
        conditional4: i === 0 ? icon('pin') : i === t.stops.length - 1 ? icon('check') : i,
        sCity: esc(s.city),
        sCountry: esc(s.country),
        conditional5: s.nights ? `${dateLabel(s.arrival)} — ${dateLabel(s.departure)} · ${s.nights} ноч.` : i === 0 ? `Отправление ${dateLabel(s.departure)}, ${localTime(s.departure)}` : `Возвращение ${dateLabel(s.arrival)}, ${localTime(s.arrival)}`,
        stopButton: button(s.nights ? 'Подробнее ' + icon('chevron') : icon('chevron'), 'stop', 'icon-link', `data-id="${s.id}" aria-label="Подробнее: ${esc(s.city)}"`),
        conditional6: s.hotel ? view('render-itinerary-hotel-mini', {
            cover: s.hotel.cover,
            landscape: landscape(s.hotel.cover),
            hotelIcon: icon('hotel'),
            name: esc(s.hotel.name),
            rating: s.hotel.rating,
            meals: esc(s.hotel.meals),
            cPetBadge: cPetBadge(s.hotel),
            price: price(s.hotel.priceMinor),
            sNights: s.nights,
            sunIcon: icon('sun'),
            sActivities: s.activities.map(a => esc(a.name)).join(' · ')
        }) : ''
    })).join('')
}); }
function cPetBadge(h) { return state.trip.criteria.pets && h.petsAllowed ? ' · С питомцем' : ''; }
function duration(ticket) { const minutes = Math.round((Date.parse(ticket.arrival) - Date.parse(ticket.departure)) / 60000); if (!Number.isFinite(minutes) || minutes < 0)
    return ''; return `${Math.floor(minutes / 60)} ч${minutes % 60 ? ` ${minutes % 60} мин` : ''}`; }
function renderTickets() { return view('render-tickets-offer-groups', {
    segments: state.trip.segments.map(s => view('render-tickets-offer-group', {
        sFrom: esc(s.from),
        arrowIcon: icon('arrow'),
        sTo: esc(s.to),
        date: dateLabel(s.ticket.departure),
        sOptions: s.options.map(o => ticketOffer(s, o)).join('')
    })).join('')
}); }
function ticketOffer(s, o) { const selected = s.ticket.id === o.id; return view('ticket-offer-ticket-offer', {
    conditional: selected ? 'selected' : '',
    sId: s.id,
    oId: o.id,
    conditional2: state.busy ? 'disabled' : '',
    selected: selected,
    planeIcon: icon('plane'),
    oCarrier: esc(o.carrier),
    oNumber: esc(o.number),
    conditional3: o.baggage === true ? 'Багаж включён' : o.baggage === false ? 'Только ручная кладь' : 'Багаж не указан',
    localTime: localTime(o.departure),
    localTime2: localTime(o.arrival),
    duration: duration(o),
    price: o.roundTripReturn ? 'В тарифе туда-обратно' : price(o.priceMinor),
    conditional4: selected ? icon('check') : ''
}); }
function renderHotels() { return view('render-hotels-hotel-groups', {
    stops: state.trip.stops.filter(s => s.hotel).map(s => view('render-hotels-offer-group', {
        sCity: esc(s.city),
        sNights: s.nights,
        rooms: s.hotel.rooms,
        conditional: s.hotel.rooms > 1 ? 'а' : '',
        sHotelOptions: s.hotelOptions.map(o => hotelOffer(s, o)).join('')
    })).join('')
}); }
function hotelOffer(s, o) { const selected = s.hotel.id === o.id; return view('hotel-offer-hotel-offer', {
    conditional: selected ? 'selected' : '',
    oCover: o.cover,
    landscape: landscape(o.cover),
    oRating: o.rating,
    repeat: '★'.repeat(o.stars),
    oName: esc(o.name),
    pinIcon: icon('pin'),
    oDistance: esc(o.distance),
    checkIcon: icon('check'),
    oMeals: esc(o.meals),
    conditional2: o.cancellation === 'Без возврата' ? 'muted' : 'green',
    shieldIcon: icon('shield'),
    oCancellation: esc(o.cancellation),
    conditional3: state.trip.criteria.pets ? view('hotel-offer-p', {
        petIcon: icon('pet')
    }) : '',
    price: price(o.priceMinor),
    sNights: s.nights,
    selectHotelButton: button(selected ? `${icon('check')} Выбран` : 'Выбрать', 'select-hotel', selected ? 'button-selected' : 'button-light', `data-item="${s.id}" data-offer="${o.id}" ${state.busy ? 'disabled' : ''} aria-pressed="${selected}"`)
}); }
function landscape(type) {
    if (type === 'city')
        return view('landscape-svg', {});
    if (type === 'mountain')
        return view('landscape-svg-2', {});
    return view('landscape-svg-3', {});
}
function renderSummary() { const t = state.trip; if (!t)
    return view('render-summary-page-content', {
        plannerButton: button('В планировщик', 'planner', 'primary')
    }); return view('render-summary-page-content-2', {
    plannerButton: button('← К маршруту', 'planner', 'text-button'),
    tCover: t.cover,
    landscape: landscape(t.cover),
    origin: esc(t.criteria.origin),
    destination: esc(t.criteria.destination),
    tStops: esc(t.stops.filter(s => s.nights).map(s => s.city).join(' + ')),
    dateRange: formatRange(),
    nights: t.criteria.nights,
    guestsLabel: guestsLabel(t.criteria),
    tSegments: t.segments.map(s => view('render-summary-summary-item', {
        planeIcon: icon('plane'),
        sFrom: esc(s.from),
        sTo: esc(s.to),
        date: dateLabel(s.ticket.departure),
        localTime: localTime(s.ticket.departure),
        localTime2: localTime(s.ticket.arrival),
        carrier: esc(s.ticket.carrier),
        price: price(s.ticket.priceMinor)
    })).join(''),
    tStops2: t.stops.filter(s => s.hotel).map(s => view('render-summary-summary-item-2', {
        hotelIcon: icon('hotel'),
        name: esc(s.hotel.name),
        sCity: esc(s.city),
        sNights: s.nights,
        meals: esc(s.hotel.meals),
        price: price(s.hotel.priceMinor)
    })).join(''),
    bagIcon: icon('bag'),
    price: price(t.totals.transferMinor),
    sunIcon: icon('sun'),
    renderedBudgetWarning: renderBudgetWarning(),
    tWarnings: t.warnings.map(w => view('render-summary-notice', {
        infoIcon: icon('info'),
        w: esc(w)
    })).join(''),
    costBreakdown: costBreakdown(),
    price2: price(t.totals.totalMinor),
    guestsLabel2: guestsLabel(t.criteria),
    nights2: t.criteria.nights,
    bookButton: button(t.status === 'booked' ? `${icon('check')} Заказ оформлен` : 'Забронировать поездку', 'book', 'primary full', `${state.bookingBusy || t.status === 'booked' ? 'disabled' : ''}`),
    shieldIcon: icon('shield'),
    conditional: state.booking && t.status === 'booked' ? view('render-summary-booking-number', {
        id: esc(state.booking.id)
    }) : ''
}); }
function costBreakdown() { const t = state.trip.totals; return view('cost-breakdown-cost-breakdown', {
    price: price(t.transportMinor),
    price2: price(t.hotelsMinor),
    price3: price(t.transferMinor),
    price4: price(t.activitiesMinor)
}); }
function renderProfile() { return view('render-profile-page-content', {
    toUpperCase: esc((state.profile.name || 'П').slice(0, 1).toUpperCase()),
    name: esc(state.profile.name),
    conditional: isDemo ? 'Демо-профиль в этом браузере' : 'Профиль путешественника',
    name2: esc(state.profile.name),
    email: esc(state.profile.email),
    phone: esc(state.profile.phone),
    conditional2: state.profileBusy ? 'disabled' : '',
    conditional3: state.profileBusy ? 'Сохраняем…' : 'Сохранить изменения',
    infoIcon: icon('info'),
    conditional4: isDemo ? 'Данные хранятся локально в вашем браузере. Аккаунт и авторизация в демо не создаются.' : 'Профиль и авторизация предоставляются backend команды.',
    bagIcon: icon('bag'),
    historyButton: button('Открыть мои поездки ' + icon('arrow'), 'history', 'button-light')
}); }
function renderHistory() { const history = state.history; return view('render-history-page-content', {
    conditional: isDemo ? 'Сохранены в этом браузере.' : 'Ваши сохранённые путешествия.',
    newTripButton: button(icon('plus') + ' Новая поездка', 'new-trip', 'primary'),
    conditional2: state.trip?.status === 'draft' ? view('render-history-draft-banner', {
        mapIcon: icon('map'),
        origin: esc(state.trip.criteria.origin),
        destination: esc(state.trip.criteria.destination),
        dateRange: formatRange(),
        plannerButton: button('Продолжить ' + icon('arrow'), 'planner', 'button-light')
    }) : '',
    conditional3: history.length ? view('render-history-history-grid', {
        history: history.map(item => view('render-history-history-card', {
            cover: item.trip.cover,
            landscape: landscape(item.trip.cover),
            checkIcon: icon('check'),
            conditional: item.booking.demo ? 'Демобронь' : 'Оформлена',
            dateRange: formatRange(item.trip),
            stops: esc(item.trip.stops.filter(s => s.nights).map(s => s.city).join(' + ')),
            guestsLabel: guestsLabel(item.trip.criteria),
            nights: item.trip.criteria.nights,
            amount: money(item.trip.totals.totalMinor, item.trip.totals.currency),
            openHistoryButton: button('Посмотреть ' + icon('arrow'), 'open-history', 'text-button', `data-id="${esc(item.trip.id)}"`),
            id: esc(item.booking.id)
        })).join('')
    }) : view('render-history-empty-state', {
        bagIcon: icon('bag'),
        plannerButton: button('Составить поездку', 'planner', 'primary')
    })
}); }
function openDrawer(type, id) { lastFocus = document.activeElement; state.drawer = { type, id }; render(); }
function renderDrawer() {
    if (!state.drawer || !state.trip)
        return '';
    const t = state.trip, { type, id } = state.drawer;
    let content = '', title = '';
    if (type === 'stop') {
        const s = t.stops.find(s => s.id === id);
        if (!s)
            return '';
        title = s.city;
        content = view('render-drawer-eyebrow', {
            sCountry: esc(s.country),
            conditional: s.nights ? `${s.nights} ноч.` : 'Точка маршрута',
            planeIcon: icon('plane'),
            date: dateLabel(s.arrival),
            localTime: localTime(s.arrival),
            slice: esc(s.arrival.slice(-6)),
            arrowIcon: icon('arrow'),
            date2: dateLabel(s.departure),
            localTime2: localTime(s.departure),
            slice2: esc(s.departure.slice(-6)),
            conditional2: s.hotel ? view('render-drawer-section-label', {
                cover: s.hotel.cover,
                landscape: landscape(s.hotel.cover),
                name: esc(s.hotel.name),
                rating: s.hotel.rating,
                meals: esc(s.hotel.meals),
                distance: esc(s.hotel.distance),
                cancellation: esc(s.hotel.cancellation),
                drawerHotelsButton: button('Сравнить отели', 'drawer-hotels', 'button-light full'),
                sActivities: s.activities.map(a => view('render-drawer-activity-detail', {
                    sunIcon: icon('sun'),
                    aName: esc(a.name),
                    aDuration: esc(a.duration),
                    aNote: esc(a.note)
                })).join('')
            }) : view('render-drawer-notice', {
                infoIcon: icon('info'),
                conditional: s.id === 'stop-0' ? 'начало' : 'конец'
            })
        });
    }
    else if (type === 'segment') {
        const s = t.segments.find(s => s.id === id);
        if (!s)
            return '';
        title = `${s.from} → ${s.to}`;
        content = view('render-drawer-eyebrow-2', {
            date: dateLabel(s.ticket.departure),
            sOptions: s.options.map(o => ticketOffer(s, o)).join(''),
            infoIcon: icon('info')
        });
    }
    else {
        title = 'Что включено';
        content = view('render-drawer-muted', {
            dateRange: formatRange(),
            guestsLabel: guestsLabel(t.criteria),
            costBreakdown: costBreakdown(),
            price: price(t.totals.totalMinor),
            infoIcon: icon('info')
        });
    }
    return view('render-drawer-overlay', {
        title: esc(title),
        closeOverlayButton: button(icon('close'), 'close-overlay', 'icon-button', 'aria-label="Закрыть подробности"'),
        content: content
    });
}
function renderModal() {
    if (!state.modal)
        return '';
    let content = '';
    if (state.modal === 'criteria') {
        const c = state.trip?.criteria || DEFAULT_CRITERIA;
        content = view('render-modal-eyebrow', {
            value: ['Москва', 'Казань', 'Санкт-Петербург'].map(city => view('render-modal-option', {
                conditional: city === c.origin ? 'selected' : '',
                city: city
            })).join(''),
            value2: ['Анталья', 'Сочи', 'Тбилиси', 'Калининград'].map(city => view('render-modal-option-2', {
                conditional: city === c.destination ? 'selected' : '',
                city: city
            })).join(''),
            cStartDate: c.startDate,
            cNights: c.nights,
            cAdults: c.adults,
            cChildrenAges: c.childrenAges.join(', '),
            round: Math.round(c.budgetMinor / 100),
            conditional: c.pets ? 'checked' : '',
            conditional2: state.busy ? 'disabled' : '',
            arrowIcon: icon('arrow')
        });
    }
    else if (state.modal === 'booking') {
        content = view('render-modal-modal-illustration', {
            bagIcon: icon('bag'),
            origin: esc(state.trip.criteria.origin),
            destination: esc(state.trip.criteria.destination),
            dateRange: formatRange(),
            guestsLabel: guestsLabel(state.trip.criteria),
            price: price(state.trip.totals.totalMinor),
            shieldIcon: icon('shield'),
            confirmBookButton: button(state.bookingBusy ? 'Оформляем…' : 'Подтвердить демобронь', 'confirm-book', 'primary full', state.bookingBusy ? 'disabled' : '')
        });
    }
    else if (state.modal === 'success') {
        content = view('render-modal-success-icon', {
            checkIcon: icon('check'),
            id: esc(state.booking.id),
            dateRange: formatRange(),
            price: price(state.trip.totals.totalMinor),
            historyButton: button('Мои поездки ' + icon('arrow'), 'history', 'primary full')
        });
    }
    else if (state.modal === 'new-trip') {
        content = view('render-modal-eyebrow-2', {
            closeOverlayButton: button('Оставить текущую', 'close-overlay', 'button-light'),
            confirmNewButton: button('Начать новую', 'confirm-new', 'primary')
        });
    }
    else if (state.modal === 'login') {
        const register = state.authMode === 'register';
        content = view('render-modal-login', {
            email: esc(state.loginEmail),
            loginError: esc(state.loginError),
            disabled: state.loginBusy ? 'disabled' : '',
            title: register ? 'Создать аккаунт' : 'Нужно войти',
            hint: register ? 'Пароль не короче 12 символов. После входа можно отправить сообщение.' : 'Войдите, чтобы отправить сообщение. Набранный текст останется в поле чата.',
            submitLabel: register ? 'Создать аккаунт' : 'Войти',
            switchLabel: register ? 'Уже есть аккаунт' : 'Создать аккаунт'
        });
    }
    return view('render-modal-overlay', {
        conditional: state.modal === 'success' ? 'success-modal' : '',
        closeOverlayButton: button(icon('close'), 'close-overlay', 'modal-close icon-button', 'aria-label="Закрыть окно"'),
        content: content
    });
}
async function send(text, criteria) {
    if (state.busy || (!text.trim() && !criteria))
        return;
    if (!isDemo && !getAccessToken()) {
        state.draftText = text;
        state.pendingSend = { text, criteria: criteria || null };
        state.error = null;
        state.loginError = '';
        state.modal = 'login';
        render();
        return;
    }
    state.error = null;
    state.lastRequest = { text, criteria };
    state.busy = true;
    state.draftText = '';
    state.messages.push({ role: 'user', text });
    render();
    try {
        const result = await api.sendMessage({ trip: await draftTrip(), text, criteria });
        if (result.trip) {
            if (!isDemo && state.sessionId)
                result.trip.id = state.sessionId;
            state.trip = result.trip;
            state.booking = null;
            state.bookingRequestId = null;
        }
        state.messages.push({ role: 'assistant', text: result.message, actions: result.actions || [], questions: result.questions || [] });
        persist();
    }
    catch (e) {
        state.error = e.message;
        state.draftText = text;
        showLogin(e);
    }
    finally {
        state.busy = false;
        render();
    }
}
async function chooseIdea(ideaId, label) {
    if (state.busy)
        return;
    if (!getAccessToken()) {
        state.pendingSend = null;
        state.modal = 'login';
        render();
        return;
    }
    state.error = null;
    state.busy = true;
    state.messages.push({ role: 'user', text: label });
    render();
    try {
        const trip = await draftTrip();
        const result = await api.selectIdea(trip.id, ideaId, '');
        if (result.trip) {
            if (state.sessionId)
                result.trip.id = state.sessionId;
            state.trip = result.trip;
            state.booking = null;
            state.bookingRequestId = null;
        }
        state.messages.push({ role: 'assistant', text: result.message, actions: result.actions || [], questions: result.questions || [] });
    }
    catch (error) {
        state.error = error.message;
        showLogin(error);
    }
    finally {
        state.busy = false;
        render();
    }
}
async function select(type, itemId, offerId) { if (state.busy)
    return; state.busy = true; render(); try {
    const result = await api.selectOffer(await draftTrip(), type, itemId, offerId);
    state.trip = result.trip;
    state.booking = null;
    state.bookingRequestId = null;
    persist();
    notify(type === 'hotel' ? 'Отель выбран. Стоимость поездки обновлена.' : 'Билет выбран. Стоимость поездки обновлена.');
}
catch (e) {
    showLogin(e);
    notify(e.message);
}
finally {
    state.busy = false;
    render();
} }
async function loadHistory() { try {
    const result = await api.getHistory();
    state.history = result.trips || [];
}
catch (e) {
    showLogin(e);
    notify(e.message);
} }
async function navigate(page) { state.page = page; state.drawer = null; state.modal = null; state.error = null; if (page === 'history')
    await loadHistory(); if (page === 'planner' && !isDemo && state.trip && getSessionId()) {
    try {
        const result = await api.getRoutes(getSessionId());
        if (result.trip)
            state.trip = result.trip;
    }
    catch (error) {
        showLogin(error);
    }
} render(); window.scrollTo({ top: 0, behavior: 'instant' }); }
function closeOverlay() { if (state.bookingBusy || state.loginBusy)
    return; state.drawer = null; state.modal = null; render(); lastFocus?.focus?.(); }
function showLogin(error) { if (!(error instanceof UnauthorizedError))
    return; state.loginError = ''; state.modal = 'login'; }
app.addEventListener('click', async (e) => {
    const target = e.target.closest('[data-action]');
    if (!target || target.disabled)
        return;
    if (target.classList.contains('overlay') && e.target !== target)
        return;
    const action = target.dataset.action;
    if (state.busy && ['new-trip', 'confirm-new', 'edit-criteria', 'open-history'].includes(action)) {
        notify('Дождитесь обновления поездки.');
        return;
    }
    if (['planner', 'profile', 'history', 'summary'].includes(action)) {
        await navigate(action);
        return;
    }
    if (action === 'auth-mode') {
        const email = document.querySelector('#login-form [name="email"]');
        if (email)
            state.loginEmail = email.value.trim();
        state.authMode = state.authMode === 'register' ? 'login' : 'register';
        state.loginError = '';
        render();
        return;
    }
    if (action === 'tab') {
        state.tab = target.dataset.tab;
        render();
    }
    if (action === 'mobile-chat' || action === 'mobile-route') {
        state.mobileView = action === 'mobile-chat' ? 'chat' : 'route';
        render();
    }
    if (action === 'quick') {
        await send(target.dataset.text);
    }
    if (action === 'chat-action') {
        const a = state.messages[Number(target.dataset.message)]?.actions?.[Number(target.dataset.index)];
        if (!a)
            return;
        if (a.ideaId && !isDemo)
            await chooseIdea(a.ideaId, a.label);
        else
            await send(a.value);
    }
    if (action === 'retry' && state.lastRequest) {
        state.messages.pop();
        await send(state.lastRequest.text, state.lastRequest.criteria);
    }
    if (action === 'edit-criteria') {
        lastFocus = target;
        state.modal = 'criteria';
        render();
    }
    if (action === 'new-trip') {
        lastFocus = target;
        state.modal = 'new-trip';
        render();
    }
    if (action === 'confirm-new') {
        state.booking = null;
        state.bookingRequestId = null;
        state.drawer = null;
        state.page = 'planner';
        state.tab = 'overview';
        state.error = null;
        if (isDemo) {
            state.trip = createTrip({ ...DEFAULT_CRITERIA }, {}, `trip-${crypto.randomUUID()}`);
            state.messages = [starterMessage];
            state.modal = null;
            persist();
            render();
            return;
        }
        state.busy = true;
        state.modal = null;
        render();
        try {
            state.sessionId = await api.createSession();
            state.trip = null;
            state.messages = [starterMessage];
        }
        catch (error) {
            showLogin(error);
            notify(error.message);
        }
        finally {
            state.busy = false;
            render();
        }
    }
    if (action === 'stop')
        openDrawer('stop', target.dataset.id);
    if (action === 'segment')
        openDrawer('segment', target.dataset.id);
    if (action === 'show-cost')
        openDrawer('cost', '');
    if (action === 'close-overlay')
        closeOverlay();
    if (action === 'drawer-hotels') {
        state.drawer = null;
        state.tab = 'hotels';
        render();
        $('.route-toolbar')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
    if (action === 'select-ticket')
        await select('ticket', target.dataset.item, target.dataset.offer);
    if (action === 'select-hotel')
        await select('hotel', target.dataset.item, target.dataset.offer);
    if (action === 'book') {
        if (state.busy) {
            notify('Дождитесь обновления поездки.');
            return;
        }
        lastFocus = target;
        state.modal = 'booking';
        render();
    }
    if (action === 'confirm-book') {
        if (state.bookingBusy || state.trip.status === 'booked')
            return;
        state.bookingBusy = true;
        state.bookingRequestId = state.bookingRequestId || crypto.randomUUID();
        render();
        try {
            const result = await api.book(state.trip, state.bookingRequestId);
            state.booking = result.booking;
            state.trip = result.trip;
            state.modal = 'success';
            if (isDemo && !saveDemoBooking(result.booking, result.trip))
                notify('Заказ создан, но память браузера недоступна. Сохраните номер заказа.');
            state.bookingRequestId = null;
            persist();
            await loadHistory();
        }
        catch (error) {
            state.bookingError = error.message;
            showLogin(error);
            notify(error.message);
        }
        finally {
            state.bookingBusy = false;
            render();
            if (state.bookingError && $('#booking-error'))
                $('#booking-error').textContent = state.bookingError;
            state.bookingError = null;
        }
    }
    if (action === 'open-history') {
        try {
            const result = await api.getTrip(target.dataset.id);
            state.trip = result.trip;
            state.booking = state.history.find(h => h.trip.id === result.trip.id)?.booking || null;
            state.page = 'summary';
            persist();
            render();
        }
        catch (error) {
            showLogin(error);
            notify(error.message);
            render();
        }
    }
});
app.addEventListener('submit', async (e) => {
    e.preventDefault();
    const form = e.target;
    if (form.id === 'chat-form') {
        const text = new FormData(form).get('message').trim();
        await send(text);
    }
    if (form.id === 'criteria-form') {
        const data = new FormData(form), ages = String(data.get('childrenAges')).trim();
        const criteria = { origin: data.get('origin'), destination: data.get('destination'), startDate: data.get('startDate'), nights: Number(data.get('nights')), adults: Number(data.get('adults')), childrenAges: ages ? ages.split(',').map(a => Number(a.trim())) : [], budgetMinor: Math.round(Number(data.get('budget')) * 100), pets: data.get('pets') === 'on' };
        state.modal = null;
        if (isDemo) {
            await send('Обновить параметры поездки', criteria);
            return;
        }
        state.error = null;
        state.busy = true;
        state.messages.push({ role: 'user', text: 'Обновил параметры поездки' });
        render();
        try {
            const trip = await draftTrip();
            const result = await api.updateConfig(trip.id, criteria);
            if (result.trip && state.sessionId)
                result.trip.id = state.sessionId;
            if (result.trip)
                state.trip = result.trip;
            state.messages.push({ role: 'assistant', text: result.message, actions: result.actions || [], questions: result.questions || [] });
        }
        catch (error) {
            state.error = error.message;
            showLogin(error);
        }
        finally {
            state.busy = false;
            render();
        }
    }
    if (form.id === 'answers-form') {
        if (state.busy)
            return;
        const message = state.messages[Number(form.dataset.message)];
        if (!message?.questions?.length)
            return;
        const data = new FormData(form);
        const answers = message.questions.map(question => ({ question_id: question.questionId, text: String(data.get(question.questionId) || '').trim() }));
        if (answers.some((answer, index) => message.questions[index].required && !answer.text)) {
            notify('Заполните обязательные поля.');
            return;
        }
        state.error = null;
        state.busy = true;
        state.messages.push({ role: 'user', text: answers.map(answer => answer.text).join(', ') });
        render();
        try {
            const trip = await draftTrip();
            const result = await api.submitAnswers(trip.id, answers);
            if (result.trip && state.sessionId)
                result.trip.id = state.sessionId;
            if (result.trip)
                state.trip = result.trip;
            state.messages.push({ role: 'assistant', text: result.message, actions: result.actions || [], questions: result.questions || [] });
        }
        catch (error) {
            state.error = error.message;
            showLogin(error);
        }
        finally {
            state.busy = false;
            render();
        }
    }
    if (form.id === 'login-form') {
        if (state.loginBusy)
            return;
        const data = new FormData(form);
        state.loginEmail = String(data.get('email')).trim();
        const password = String(data.get('password'));
        if (password.length < 12) {
            state.loginError = 'Пароль не короче 12 символов.';
            render();
            return;
        }
        state.loginError = '';
        state.loginBusy = true;
        render();
        try {
            const submit = state.authMode === 'register' ? api.register : api.login;
            await submit({ email: state.loginEmail, password });
            state.modal = null;
            const pending = state.pendingSend;
            state.pendingSend = null;
            state.loginBusy = false;
            if (pending) {
                await send(pending.text, pending.criteria);
                return;
            }
            notify(state.authMode === 'register' ? 'Аккаунт создан.' : 'Вы вошли.');
        }
        catch (error) {
            state.loginError = error.message;
        }
        finally {
            state.loginBusy = false;
            if (!state.busy)
                render();
        }
    }
    if (form.id === 'profile-form') {
        if (state.profileBusy)
            return;
        const data = new FormData(form);
        const profile = { name: String(data.get('name')).trim(), email: String(data.get('email')).trim(), phone: String(data.get('phone')).trim() };
        if (!profile.name) {
            notify('Введите имя.');
            return;
        }
        state.profileBusy = true;
        render();
        try {
            const result = await api.saveProfile(profile);
            state.profile = result.profile;
            notify('Изменения сохранены.');
        }
        catch (error) {
            showLogin(error);
            notify(error.message);
        }
        finally {
            state.profileBusy = false;
            render();
        }
    }
});
app.addEventListener('input', e => {
    if (e.target.id === 'chat-input')
        state.draftText = e.target.value;
    if (e.target.closest('#answers-form'))
        state.answerDraft = { ...state.answerDraft, [e.target.name]: e.target.value };
});
app.addEventListener('keydown', e => { if (e.target.id === 'chat-input' && e.key === 'Enter' && !e.shiftKey && !e.isComposing) {
    e.preventDefault();
    e.target.form.requestSubmit();
} });
document.addEventListener('keydown', e => { if (e.key === 'Escape' && (state.modal || state.drawer))
    closeOverlay(); if (e.key === 'Tab' && (state.modal || state.drawer)) {
    const nodes = [...$('[role="dialog"]').querySelectorAll('button:not(:disabled),input,select,textarea,a[href]')].filter(el => el.offsetParent !== null);
    if (!nodes.length)
        return;
    const first = nodes[0], last = nodes.at(-1);
    if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
    }
    else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
    }
} });
window.addEventListener('hashchange', () => { if (location.hash === '#planner')
    navigate('planner'); });
loadViews().then(() => {
    if (!isDemo && !getAccessToken())
        state.modal = 'login';
    render();
    if (isDemo || getAccessToken()) {
        api.getProfile().then(result => {
            state.profile = result.profile;
            render();
        }).catch(error => { showLogin(error); notify(error.message); render(); });
    }
}).catch(error => notify(error.message));
