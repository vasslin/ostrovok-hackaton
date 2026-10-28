import { icon } from './icons.js';
const TILE_SIZE = 256;
let countries = null;
const geographyReady = fetch(new URL('./assets/countries.geojson', import.meta.url)).then(r => { if (!r.ok)
    throw new Error('Map data unavailable'); return r.json(); }).then(data => { countries = data.features; }).catch(() => { });
function project(lng, lat, zoom) {
    const size = TILE_SIZE * 2 ** zoom, sin = Math.sin(Math.max(-85, Math.min(85, lat)) * Math.PI / 180);
    return { x: (lng + 180) / 360 * size, y: (.5 - Math.log((1 + sin) / (1 - sin)) / (4 * Math.PI)) * size };
}
function escape(text) { return String(text).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c])); }
const land = [
    [[-11, 36], [-9, 43], [-5, 44], [-4, 48], [-1, 49], [2, 51], [6, 54], [8, 57], [12, 57], [13, 54], [19, 55], [23, 60], [27, 66], [31, 70], [43, 69], [60, 70], [65, 55], [55, 47], [47, 46], [44, 42], [40, 41], [41, 39], [36, 37], [29, 36], [27, 40], [24, 41], [24, 38], [22, 36], [20, 40], [18, 42], [16, 41], [17, 39], [15, 38], [12, 42], [10, 44], [7, 44], [5, 43], [3, 42], [0, 40], [-1, 37]],
    [[26, 41], [29, 42], [34, 42], [36, 41], [40, 41], [42, 42], [46, 42], [50, 39], [48, 36], [44, 35], [41, 36], [36, 36], [35, 35], [33, 36], [30, 36], [28, 37], [27, 39]],
    [[-18, 29], [-10, 35], [-6, 36], [1, 36], [10, 37], [12, 33], [20, 32], [25, 32], [32, 31], [35, 28], [40, 18], [-18, 18]],
    [[-8, 50], [-6, 54], [-5, 58], [-3, 59], [0, 55], [2, 52], [0, 50]],
    [[7, 58], [5, 62], [10, 68], [17, 71], [24, 71], [28, 69], [25, 64], [20, 60], [17, 56], [13, 55], [11, 58]],
    [[8, 41], [9, 43], [10, 42], [10, 39], [9, 39]], [[12, 38], [15, 38], [16, 37], [13, 37]], [[23, 35], [26, 35], [26, 36], [23, 36]], [[32, 35], [34, 35], [35, 35.5], [33, 35.5]]
];
const seas = [{ name: 'Чёрное море', lng: 34, lat: 43.1 }, { name: 'Средиземное море', lng: 24, lat: 33.9 }, { name: 'Каспийское море', lng: 51, lat: 41 }];
const labels = [{ name: 'РОССИЯ', lng: 41, lat: 54 }, { name: 'ТУРЦИЯ', lng: 34, lat: 39.1 }, { name: 'УКРАИНА', lng: 31, lat: 49 }, { name: 'ГРУЗИЯ', lng: 44, lat: 42.6 }, { name: 'БОЛГАРИЯ', lng: 25, lat: 42.8 }, { name: 'ГРЕЦИЯ', lng: 22, lat: 39.3 }, { name: 'РУМЫНИЯ', lng: 25, lat: 46 }];
export class RouteMap {
    constructor(element, { onStop, onSegment }) {
        this.el = element;
        this.onStop = onStop;
        this.onSegment = onSegment;
        this.zoom = 3;
        this.offset = { x: 0, y: 0 };
        this.trip = null;
        this.autoFit = true;
        this.dragging = false;
        this.el.addEventListener('click', e => {
            if (this.moved) {
                this.moved = false;
                return;
            }
            const stop = e.target.closest('[data-map-stop]');
            if (stop)
                this.onStop(stop.dataset.mapStop);
            const segment = e.target.closest('[data-map-segment]');
            if (segment)
                this.onSegment(segment.dataset.mapSegment);
            const control = e.target.closest('[data-map-control]');
            if (control) {
                const a = control.dataset.mapControl;
                if (a === 'fit')
                    this.autoFit = true;
                else {
                    this.autoFit = false;
                    this.zoom = Math.max(2, Math.min(9, this.zoom + (a === 'in' ? 1 : -1)));
                    this.offset = { x: 0, y: 0 };
                }
                this.draw();
            }
        });
        this.el.addEventListener('pointerdown', e => {
            if (e.target.closest('button,a'))
                return;
            this.start = { x: e.clientX, y: e.clientY, ox: this.offset.x, oy: this.offset.y };
            this.dragging = true;
            this.moved = false;
        });
        this.move = e => { if (!this.dragging)
            return; const dx = e.clientX - this.start.x, dy = e.clientY - this.start.y; if (Math.abs(dx) + Math.abs(dy) > 6) {
            this.moved = true;
            this.autoFit = false;
            this.offset = { x: this.start.ox + dx, y: this.start.oy + dy };
            this.draw();
        } };
        this.up = () => { this.dragging = false; };
        window.addEventListener('pointermove', this.move);
        window.addEventListener('pointerup', this.up);
        this.observer = new ResizeObserver(() => this.draw());
        this.observer.observe(element);
        geographyReady.then(() => this.draw());
    }
    update(trip, selectedStop) {
        const signature = trip?.stops.map(s => s.city).join('|');
        if (signature !== this.signature) {
            this.autoFit = true;
            this.offset = { x: 0, y: 0 };
            this.signature = signature;
        }
        this.trip = trip;
        this.selectedStop = selectedStop;
        this.draw();
    }
    draw() {
        if (!this.trip || !this.el.clientWidth)
            return;
        const w = this.el.clientWidth, h = this.el.clientHeight;
        const points = this.trip.stops;
        const center = { lng: (Math.min(...points.map(s => s.lng)) + Math.max(...points.map(s => s.lng))) / 2, lat: (Math.min(...points.map(s => s.lat)) + Math.max(...points.map(s => s.lat))) / 2 };
        if (this.autoFit) {
            this.offset = { x: 0, y: 0 };
            this.zoom = 2;
            for (let z = 3; z <= 7; z++) {
                const p = points.map(s => project(s.lng, s.lat, z));
                if (Math.max(...p.map(a => a.x)) - Math.min(...p.map(a => a.x)) > w - 170 || Math.max(...p.map(a => a.y)) - Math.min(...p.map(a => a.y)) > h - 150)
                    break;
                this.zoom = z;
            }
        }
        const mid = project(center.lng, center.lat, this.zoom), left = mid.x - w / 2 - this.offset.x, top = mid.y - h / 2 - this.offset.y;
        const point = (lng, lat) => { const p = project(lng, lat, this.zoom); return { x: p.x - left, y: p.y - top }; };
        const path = coords => coords.map(([lng, lat], i) => { const p = point(lng, lat); return `${i ? 'L' : 'M'}${p.x.toFixed(1)},${p.y.toFixed(1)}`; }).join(' ') + 'Z';
        const dots = points.map(s => ({ ...s, ...point(s.lng, s.lat) }));
        const unique = dots.filter((s, i) => dots.findIndex(a => a.city === s.city) === i);
        const lines = this.trip.segments.map((s, i) => { const a = dots.find(d => d.id === s.fromStopId), b = dots.find(d => d.id === s.toStopId); const dx = b.x - a.x, dy = b.y - a.y, len = Math.hypot(dx, dy) || 1; const bend = Math.min(80, len * .22) * (i % 2 ? -1 : 1); const cx = (a.x + b.x) / 2 - dy / len * bend, cy = (a.y + b.y) / 2 + dx / len * bend; const mx = .25 * a.x + .5 * cx + .25 * b.x, my = .25 * a.y + .5 * cy + .25 * b.y; return `<path class="route-halo" d="M${a.x},${a.y} Q${cx},${cy} ${b.x},${b.y}"/><path class="route-line" d="M${a.x},${a.y} Q${cx},${cy} ${b.x},${b.y}"/><path tabindex="0" role="button" aria-label="Переезд ${escape(s.from)} — ${escape(s.to)}" class="route-hit" data-map-segment="${s.id}" d="M${a.x},${a.y} Q${cx},${cy} ${b.x},${b.y}"/><g class="plane-dot" transform="translate(${mx},${my})" aria-hidden="true"><circle r="13"/><path d="m-6 1 12-5-4 11-2-5Z"/></g>`; }).join('');
        const mapLand = countries ? countries.map(feature => {
            const polygons = feature.geometry.type === 'Polygon' ? [feature.geometry.coordinates] : feature.geometry.type === 'MultiPolygon' ? feature.geometry.coordinates : [];
            return polygons.map(rings => `<path d="${rings.map(path).join(' ')}" fill="#fafbf7" stroke="#d5dfd8" stroke-width=".8" fill-rule="evenodd"/>`).join('');
        }).join('') : land.map(c => `<path d="${path(c)}" fill="#fafbf7" stroke="#d5dfd8" stroke-width="1.3"/>`).join('');
        this.el.innerHTML = `<svg class="geographic-map" viewBox="0 0 ${w} ${h}" aria-label="Географическая карта маршрута" role="img"><defs><pattern id="map-grid" width="40" height="40" patternUnits="userSpaceOnUse"><path d="M40 0H0V40" fill="none" stroke="#dce6e9" stroke-width=".5"/></pattern></defs><rect width="100%" height="100%" fill="#e8f0f3"/><rect width="100%" height="100%" fill="url(#map-grid)"/>${mapLand}${seas.map(l => { const p = point(l.lng, l.lat); return `<text class="sea-label" x="${p.x}" y="${p.y}">${l.name}</text>`; }).join('')}${labels.map(l => { const p = point(l.lng, l.lat); return `<text class="country-label" x="${p.x}" y="${p.y}">${l.name}</text>`; }).join('')}${lines}</svg><div class="map-markers">${unique.map((s, i) => `<button class="map-marker ${s.id === this.selectedStop ? 'active' : ''}" style="left:${s.x}px;top:${s.y}px" data-map-stop="${s.id}" aria-label="${escape(s.city)}, подробности остановки"><span class="marker-dot">${i === 0 ? '<span class="home-dot"></span>' : i}</span><span class="marker-label">${escape(s.city)}${s.nights ? `<small>${s.nights} ноч.</small>` : ''}</span></button>`).join('')}</div><div class="map-top-badge">${icon('map')} Маршрут на карте <span>· ${unique.length} города</span></div><div class="map-controls"><button data-map-control="in" aria-label="Приблизить карту">${icon('plus')}</button><button data-map-control="out" aria-label="Отдалить карту">${icon('minus')}</button><button data-map-control="fit" aria-label="Показать весь маршрут">${icon('refresh')}</button></div><div class="map-caption">${countries ? '© Natural Earth · ' : ''}Линии — связи городов, не траектории рейсов</div>`;
        this.el.querySelectorAll('.route-hit').forEach(el => el.addEventListener('keydown', e => { if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            this.onSegment(el.dataset.mapSegment);
        } }));
    }
    destroy() { this.observer.disconnect(); window.removeEventListener('pointermove', this.move); window.removeEventListener('pointerup', this.up); }
}
