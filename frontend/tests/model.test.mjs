import test from 'node:test';
import assert from 'node:assert/strict';
import { DEFAULT_CRITERIA, createTrip, selectedIds, cheaperSelection, parseMessage, addDays } from '../dist/model.js';

test('The complete party total equals tickets + hotels + transfers + activities',()=>{
  const trip=createTrip(DEFAULT_CRITERIA);
  assert.equal(trip.segments.length,3);
  assert.equal(trip.stops.filter(s=>s.hotel).length,2);
  assert.equal(trip.totals.totalMinor,trip.segments.reduce((s,x)=>s+x.ticket.priceMinor,0)+trip.stops.reduce((s,x)=>s+(x.hotel?.priceMinor||0),0)+trip.totals.transferMinor+trip.totals.activitiesMinor);
  assert.ok(trip.totals.totalMinor<=trip.criteria.budgetMinor);
});
test('Date changes preserve compatible selected hotels and tickets and shift their times',()=>{
  let trip=createTrip(DEFAULT_CRITERIA);
  trip=createTrip(DEFAULT_CRITERIA,cheaperSelection(trip));
  const selection=selectedIds(trip), newDate=addDays(trip.startDate,2);
  const updated=createTrip({...trip.criteria,startDate:newDate},selection);
  assert.deepEqual(selectedIds(updated),selection);
  assert.equal(updated.segments[0].ticket.departure.slice(0,10),newDate);
  assert.equal(updated.endDate,addDays(trip.endDate,2));
});
test('A pet removes incompatible hotel and flight options',()=>{
  const trip=createTrip({...DEFAULT_CRITERIA,pets:true});
  assert.ok(trip.stops.filter(s=>s.hotel).every(s=>s.hotelOptions.every(h=>h.petsAllowed)));
  assert.ok(trip.segments.every(s=>s.options.every(t=>t.petsAllowed)));
  assert.ok(trip.warnings.some(w=>w.includes('Доплата')));
});
test('Changing destination does not reuse a ticket from a different route',()=>{
  const original=createTrip(DEFAULT_CRITERIA);
  const next=createTrip({...DEFAULT_CRITERIA,destination:'Сочи'},selectedIds(original));
  assert.equal(next.segments.length,2);
  assert.ok(next.segments.every(s=>!s.ticket.id.includes('Анталья')&&!s.ticket.id.includes('Стамбул')));
});
test('The cheapest available configuration really reduces the total',()=>{
  const original=createTrip(DEFAULT_CRITERIA);
  const next=createTrip(DEFAULT_CRITERIA,cheaperSelection(original));
  assert.ok(next.totals.totalMinor<original.totals.totalMinor);
});
test('Every flight has valid timezone-aware times and connects the declared stops',()=>{
  for(const destination of ['Анталья','Сочи','Тбилиси','Калининград']){
    const trip=createTrip({...DEFAULT_CRITERIA,destination});
    for(const segment of trip.segments){
      assert.ok(trip.stops.some(s=>s.id===segment.fromStopId));
      assert.ok(trip.stops.some(s=>s.id===segment.toStopId));
      for(const ticket of segment.options){assert.match(ticket.departure,/[+-]\d{2}:\d{2}$/);assert.ok(Date.parse(ticket.arrival)>Date.parse(ticket.departure));}
    }
    for(const stop of trip.stops.filter(s=>s.nights)){assert.ok(Date.parse(stop.departure)>Date.parse(stop.arrival));}
  }
});
test('Demo message parsing updates requested fields without erasing other preferences',()=>{
  assert.equal(parseMessage('В Сочи до 100 тысяч',DEFAULT_CRITERIA).patch.budgetMinor,10000000);
  assert.equal(parseMessage('В Сочи до 100 тысяч',DEFAULT_CRITERIA).patch.destination,'Сочи');
  assert.equal(parseMessage('на два дня позже',DEFAULT_CRITERIA).patch.startDate,'2026-11-12');
  assert.deepEqual(parseMessage('С нами собака',DEFAULT_CRITERIA).patch,{pets:true});
  assert.equal(parseMessage('привет',DEFAULT_CRITERIA).handled,false);
});
test('Invalid dates, party counts, and budgets are rejected',()=>{
  assert.throws(()=>createTrip({...DEFAULT_CRITERIA,startDate:'2026-02-31'}));
  assert.throws(()=>createTrip({...DEFAULT_CRITERIA,adults:0}));
  assert.throws(()=>createTrip({...DEFAULT_CRITERIA,budgetMinor:-1}));
  assert.throws(()=>createTrip({...DEFAULT_CRITERIA,childrenAges:[25]}));
});
