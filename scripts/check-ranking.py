import json
import os
import urllib.request
from pathlib import Path

root = Path(__file__).resolve().parents[1]
settings = {}
if (root / '.env').exists():
    for line in (root / '.env').read_text().splitlines():
        if '=' in line and not line.lstrip().startswith('#'):
            name, value = line.split('=', 1)
            settings[name] = value.strip().strip('"').strip("'")
headers = {'Content-Type': 'application/json'}
key = os.environ.get('SERVICE_API_KEY', settings.get('SERVICE_API_KEY', ''))
if key:
    headers['Authorization'] = 'Bearer ' + key
body = (root / 'examples/ranking/routes.json').read_bytes()
request = urllib.request.Request(os.environ.get('TEST_BASE_URL', 'http://localhost:8080') + '/v1/ranking/routes', body, headers)
with urllib.request.urlopen(request, timeout=190) as response:
    result = json.load(response)
assert len(result['results']) == 1 and not result['rejected']
route = result['results'][0]
assert route['id'] == 'route-1'
assert 0 <= route['score'] <= 100
assert route['total_price']['amount_minor'] == '6500000'
assert route['price_complete'] and route['requires_verification']
assert result['parts']['transport_parts'][0]['source'] == 'synthetic'
print(json.dumps({'id': route['id'], 'score': route['score'], 'total_price': route['total_price'], 'requires_verification': route['requires_verification']}, ensure_ascii=False, indent=2))
