import json
import os
import time
import urllib.error
import urllib.request
from pathlib import Path

settings = {}
path = Path(__file__).resolve().parents[1] / '.env'
if path.exists():
    for line in path.read_text().splitlines():
        if '=' in line and not line.lstrip().startswith('#'):
            key, value = line.split('=', 1)
            settings[key.strip()] = value.strip().strip('\"').strip("'")
base = os.environ.get('TEST_BASE_URL', 'http://localhost:8080')
key = os.environ.get('SERVICE_API_KEY', settings.get('SERVICE_API_KEY', ''))

def request(path, body):
    headers = {'Content-Type': 'application/json'}
    if key:
        headers['Authorization'] = 'Bearer ' + key
    req = urllib.request.Request(base + path, json.dumps(body, ensure_ascii=False).encode(), headers)
    started = time.monotonic()
    try:
        with urllib.request.urlopen(req, timeout=190) as response:
            result = json.load(response)
    except urllib.error.HTTPError as error:
        raise SystemExit(f'{path}: HTTP {error.code}: {error.read().decode()}') from None
    print(f'{path}: {time.monotonic() - started:.1f}s')
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return result

result = request('/v1/config', {'message': 'Из Москвы в Сочи с 10 по 17 ноября 2026 года. Два взрослых. Бюджет всей поездки 80000 рублей. Нужен тихий отель рядом с морем.', 'config': {}})
config = result['config']
assert config['adults'] == 2
assert config['departure_date'] == '2026-11-10'
assert config['return_date'] == '2026-11-17'
assert config['budget'] == 80000
assert not result['missing_fields']
result = request('/v1/rank', {'config': config, 'hotels': [
    {'id': 'sea', 'name': 'Тихая бухта', 'details': {'total_price': 35000, 'currency': 'RUB', 'check_in': '2026-11-10', 'check_out': '2026-11-17', 'distance_to_beach_m': 150, 'quiet_area': True, 'adults': 2}},
    {'id': 'city', 'name': 'Центр', 'details': {'total_price': 95000, 'currency': 'RUB', 'check_in': '2026-11-10', 'check_out': '2026-11-17', 'distance_to_beach_m': 5000, 'quiet_area': False, 'adults': 2}}
]})
assert {row['id'] for row in result['results']} == {'sea', 'city'}
assert result['results'][0]['id'] == 'sea'
print('Local config extraction and hotel ranking passed')
