import json
import os
import time
import urllib.error
import urllib.request
from pathlib import Path

root = Path(__file__).resolve().parents[1]
settings = {}
if (root / '.env').exists():
    for line in (root / '.env').read_text().splitlines():
        if '=' in line and not line.lstrip().startswith('#'):
            key, value = line.split('=', 1)
            settings[key] = value.strip().strip('\"').strip("'")
base = os.environ.get('TEST_BASE_URL', 'http://localhost:8080')
key = os.environ.get('SERVICE_API_KEY', settings.get('SERVICE_API_KEY', ''))
results = []

def call(operation, body, label):
    headers = {'Content-Type': 'application/json'}
    if key:
        headers['Authorization'] = 'Bearer ' + key
    req = urllib.request.Request(base + '/v1/planning/' + operation + ('?include_metadata=true' if operation != 'recommendations' else ''), json.dumps(body, ensure_ascii=False).encode(), headers)
    start = time.monotonic()
    try:
        with urllib.request.urlopen(req, timeout=190) as response:
            out = json.load(response)
    except urllib.error.HTTPError as error:
        raise SystemExit(f'{label}: HTTP {error.code}: {error.read().decode()}') from None
    if operation != 'recommendations':
        result = out['result']
        meta = out['metadata']
        assert set(result) == {'status', 'config', 'advice', 'questions'}
        out = {**result, **meta, 'config': result['config']}
    seconds = round(time.monotonic() - start, 2)
    results.append({'scenario': label, 'seconds': seconds, 'output': out})
    print(f'{label}: {out.get("status", "recommendations")}, {seconds}s', flush=True)
    print(json.dumps(out, ensure_ascii=False, indent=2), flush=True)
    return out

def example(name):
    return json.loads((root / ('examples/stage1/' + name + '.json')).read_text())

full = call('analyze', example('known_trip'), 'known_trip')
assert full['status'] == 'generate' and not full['missing_fields']
assert full['config']['origin']['city'] == 'Москва'
assert full['config']['travelers']['adults'] == 2
partial = call('analyze', example('missing_fields'), 'missing_fields')
assert partial['status'] == 'questions' and 1 <= len(partial['questions']) <= 2
assert partial['missing_fields']
assert partial['config']['travelers']['adults'] == 0
answers = [{'question_id': q['id'], 'text': 'С 10 по 17 ноября 2026 года, двое взрослых'} for q in partial['questions']]
answered = call('update', {'config': partial['config'], 'context': {'summary': partial['summary']}, 'questions': partial['questions'], 'answers': answers, 'text': 'Поедем с 10 по 17 ноября 2026, двое взрослых.', 'current_date': '2026-10-28'}, 'answer_batch')
assert answered['status'] == 'generate'
assert answered['config']['dates']['departure_date'] == '2026-11-10'
advice = call('analyze', example('inspiration'), 'inspiration')
assert advice['status'] == 'advice'
recommendations = call('recommendations', {'text': example('inspiration')['text'], 'config': advice['config'], 'context': {'summary': advice['summary']}, 'last_sessions': [], 'current_date': '2026-10-28'}, 'recommendations')
advice['advice'] = recommendations['advice']
assert 2 <= len(advice['advice']) <= 3
assert not (advice['config'].get('destination') or {}).get('city')
assert all(i['source'] == 'ai_generated_idea' for i in advice['advice'])
selected = call('update', {'config': advice['config'], 'context': {'summary': advice['summary']}, 'selected_idea': advice['advice'][0], 'text': 'Выбираю эту идею. Бюджет всей поездки уменьшим до 60000 рублей.', 'current_date': '2026-10-28'}, 'idea_selection_and_revision')
assert selected['status'] == 'generate'
assert selected['config']['budget']['total'] == 60000
assert selected['config']['origin']['city'] == 'Москва'
forced = call('refine', {'config': {}, 'context': {'summary': 'Пользователь пока не сообщил параметры поездки.'}, 'current_date': '2026-10-28'}, 'force_with_no_facts')
assert forced['status'] == 'questions' and forced['missing_fields']
report = Path(os.environ.get('STAGE1_REPORT', '/tmp/godnota-stage1-smoke.json'))
report.write_text(json.dumps(results, ensure_ascii=False, indent=2) + '\n')
print(f'All seven stage-1 scenarios passed; report: {report}', flush=True)
