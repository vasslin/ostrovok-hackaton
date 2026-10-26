import test from 'node:test';
import assert from 'node:assert/strict';
import { DEFAULT_CRITERIA, createTrip } from '../dist/model.js';
globalThis.window = { TRAVEL_CONFIG: { mode: 'demo' } };
const storage = new Map();
globalThis.localStorage = { getItem: key => storage.get(key) ?? null, setItem: (key, value) => storage.set(key, value) };
const { api, saveDemoBooking, readStorage } = await import('../dist/api.js');

test('Asking for a cheaper trip reduces even an already-within-budget total and discloses tradeoffs',async()=>{
  const trip=createTrip(DEFAULT_CRITERIA);
  const result=await api.sendMessage({trip,text:'Хочу дешевле'});
  assert.ok(result.trip.totals.totalMinor<trip.totals.totalMinor);
  assert.match(result.message,/ручную кладь/);
  assert.match(result.message,/невозвратный/);
});
test('An impossible budget is reported honestly instead of producing fabricated offers',async()=>{
  const trip=createTrip(DEFAULT_CRITERIA);
  const result=await api.sendMessage({trip,text:'Бюджет до 1000 рублей'});
  assert.ok(result.trip.totals.totalMinor>result.trip.criteria.budgetMinor);
  assert.match(result.message,/выше бюджета/);
});
test('Unsupported dialogue does not pretend to be a live AI answer',async()=>{
  const trip=createTrip(DEFAULT_CRITERIA);
  const result=await api.sendMessage({trip,text:'Расскажи историю архитектуры'});
  assert.equal(result.trip,trip);
  assert.match(result.message,/Для свободного диалога/);
});
test('Repeated booking with the same request ID cannot duplicate browser history',async()=>{
  const trip=createTrip(DEFAULT_CRITERIA);
  const first=await api.book(trip,'test-request-one');
  const second=await api.book(trip,'test-request-one');
  assert.equal(first.booking.id,second.booking.id);
  assert.equal(first.trip.status,'booked');
  saveDemoBooking(first.booking,first.trip);saveDemoBooking(second.booking,second.trip);
  assert.equal(readStorage('history',[]).length,1);
});
