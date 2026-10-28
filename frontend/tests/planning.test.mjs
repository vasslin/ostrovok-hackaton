import test from 'node:test';
import assert from 'node:assert/strict';
import { reviewRequest, providerIssue } from '../dist/planning.js';

test('review saves explicit values, preserves preferences and requires transport consent', () => {
    const data = new FormData();
    for (const [key,value] of Object.entries({ origin:'Москва',originCountry:'ru',destination:'Париж',destinationCountry:'fr',departure:'2026-11-04',return:'2026-11-11',adults:'2',children:'7, 12',budget:'80000',currency:'RUB',hotel:'on',roundTrip:'on' })) data.set(key,value);
    data.append('modes','plane');
    const base = { hotel:{ amenities:['wifi'] },transport:{max_transfers:2},travelers:{pets:[{kind:'cat'}]} };
    const config = reviewRequest(base,data);
    assert.equal(config.hotel.needed,true);
    assert.equal(config.destination.country_code,'FR');
    assert.equal(config.budget.total,80000);
    assert.deepEqual(config.travelers.children_ages,[7,12]);
    assert.deepEqual(config.travelers.pets,[{kind:'cat'}]);
    assert.equal(config.transport.max_transfers,2);
    assert.equal(reviewRequest(base,data,true).hotel.needed,false);
    assert.equal(base.hotel.needed,undefined);
});
test('provider states distinguish absent configuration, empty results and errors', () => {
    assert.match(providerIssue({kind:'hotel',code:'not_configured'}),/не подключён/);
    assert.match(providerIssue({kind:'hotel',code:'no_results'}),/не найдены/);
    assert.match(providerIssue({kind:'hotel',code:'http_500'}),/Ошибка/);
});
