import json
import re
from pathlib import Path

root = Path(__file__).resolve().parents[1]
path = root / 'ai-service/web/openapi.json'
spec = json.loads(path.read_text())
source = '\n'.join((root / ('ai-service/internal/service/' + name)).read_text() for name in ['planning_models.go', 'routes_models.go', 'parts.go'])
schemas = spec['components']['schemas']
ref = lambda name: {'$ref': '#/components/schemas/' + name}
for name, body in re.findall(r'type (\w+) struct \{(.*?)\n\}', source, re.S):
    if name == 'planningCompletion':
        continue
    properties = {}
    for line in body.splitlines():
        match = re.search(r'\w+\s+(\*?)(\[\])?(\w+)\s+`json:"([^" ,]+)', line)
        if not match:
            continue
        pointer, repeated, kind, field = match.groups()
        schema = {'type': {'string': 'string', 'bool': 'boolean', 'int': 'integer', 'float64': 'number', 'Minor':'string'}[kind]} if kind in ['string', 'bool', 'int', 'float64', 'Minor'] else ref(kind)
        if repeated:
            schema = {'type': 'array', 'items': schema, 'nullable': True}
        if pointer:
            schema = {'allOf': [schema], 'nullable': True} if '$ref' in schema else {**schema, 'nullable': True}
        if field in ['current_date', 'start_date', 'end_date']:
            schema['format'] = 'date'
        if field == 'status':
            schema['enum'] = ['advice', 'questions', 'generate']
        if field == 'advice':
            schema['maxItems'] = 3
        if field == 'questions' and name == 'PlanningOutput':
            schema['maxItems'] = 2
        properties[field] = schema
    schemas[name] = {'type': 'object', 'additionalProperties': False, 'properties': properties}
    if name in ['PlanningOutput', 'RecommendationsOutput', 'PlanningQuestion', 'PlanningAssumption', 'TravelIdea']:
        schemas[name]['required'] = list(properties)
schemas['PlanningInput']['description'] = 'Нужны text, answers, selected_idea или force_generate. Последнее сообщение и нижний ответ в пачке важнее контекста. Предыдущий config и summary передаёт travel-service. current_date по умолчанию — текущая дата UTC.'
schemas['PlanningInput']['properties']['force_generate']['description'] = 'Попытаться завершить конфиг по имеющемуся контексту. Не разрешает выдумывать обязательные поля.'
schemas['PlanningOutput']['properties']['missing_fields']['description'] = 'Пути обязательных отсутствующих и некорректных полей. generate возвращается только при пустом списке.'
schemas['PlanningOutput']['properties']['status']['description'] = 'advice — вызовите /v1/planning/recommendations для идей без выбора за пользователя; questions — уточнение; generate — конфиг готов к финальной проверке travel-service, но поиск ещё требует подтверждения пользователя.'
schemas['TravelIdea']['properties']['source'] = {'type': 'string', 'enum': ['ai_generated_idea']}
examples = {
 'known_trip': {'summary': 'Пользователь знает параметры', 'value': {'text': 'Из Москвы в Сочи с 10 по 17 ноября 2026, двое взрослых, бюджет 80000 рублей на всю поездку, без пересадок.', 'config': {}, 'current_date': '2026-10-28'}},
 'missing_fields': {'summary': 'Нужны уточнения', 'value': {'text': 'Хочу из Москвы в Сочи. Даты и состав пока уточняем.', 'config': {}, 'current_date': '2026-10-28'}},
 'inspiration': {'summary': 'Пользователь просит идеи', 'value': {'text': 'Предложи два разных направления для спокойного отпуска. Из Москвы, двое взрослых, с 10 по 17 ноября 2026, бюджет 80000 рублей на всю поездку.', 'config': {}, 'current_date': '2026-10-28'}},
 'answers': {'summary': 'Ответы на вопросы', 'value': {'text': '', 'config': {'version': '1.0', 'trip': {'origin': {'city': 'Москва'}, 'destinations': [{'city': 'Сочи'}]}}, 'questions': [{'id': 'trip_dates', 'field': 'trip.date_range', 'text': 'На какие даты?', 'required': True, 'input_type': 'text'}, {'id': 'adults', 'field': 'trip.travelers.adults', 'text': 'Сколько взрослых?', 'required': True, 'input_type': 'number'}], 'answers': [{'question_id': 'trip_dates', 'text': 'С 10 по 17 ноября 2026'}, {'question_id': 'adults', 'text': 'Двое'}], 'current_date': '2026-10-28'}}
}
error_responses = {k: v for k, v in spec['paths']['/v1/config']['post']['responses'].items() if k != '200'}
for operation, title in [('analyze', 'Анализ контекста и сообщения'), ('update', 'Обновить конфиг по правкам или ответам'), ('refine', 'Заполнить пробелы по имеющемуся контексту')]:
    spec['paths']['/v1/planning/' + operation] = {'post': {
      'tags': ['Этап 1'], 'operationId': 'planning_' + operation, 'summary': title,
      'description': 'Использует общий вложенный godnota.common.v1.TripConfig. Один запрос к модели. Конфиг, контекст и выбранную идею хранит вызывающий сервис.',
      'security': [{'serviceKey': []}, {}],
      'requestBody': {'required': True, 'content': {'application/json': {'schema': ref('PlanningInput'), 'examples': examples}}},
      'responses': {'200': {'description': 'Статус, обновлённый конфиг, вопросы/идеи, резюме и допущения', 'content': {'application/json': {'schema': ref('PlanningOutput')}}}, **error_responses}
    }}
spec['paths']['/v1/planning/recommendations'] = {'post': {
 'tags': ['Этап 1'], 'operationId': 'planning_recommendations', 'summary': 'Предложить 2–3 идеи с учётом прошлых сессий',
 'description': 'Идеи не являются проверенными предложениями. Нет подтверждённых цен, наличия или погоды. last_sessions предоставляет session-service через travel-service; максимум пять сессий.',
 'security': [{'serviceKey': []}, {}],
 'requestBody': {'required': True, 'content': {'application/json': {'schema': ref('RecommendationsInput'), 'example': {'text': 'Предложи два разных направления на спокойный отпуск', 'context': {'summary': 'Любит музеи и прогулки'}, 'config': {'version': '1.0', 'trip': {'origin': {'city': 'Москва'}}}, 'last_sessions': [], 'current_date': '2026-10-28'}}}},
 'responses': {'200': {'description': '2–3 разных идеи поездки', 'content': {'application/json': {'schema': ref('RecommendationsOutput')}}}, **error_responses}
}}
if not any(t['name'] == 'Этап 1' for t in spec.get('tags', [])):
    spec.setdefault('tags', []).insert(0, {'name': 'Этап 1'})
spec['info']['version'] = '1.2.0'
spec['info']['description'] = 'Этап 1: вложенный TripConfig, контекст, вопросы, идеи и готовность к поиску. Начните с /v1/planning/analyze. Старые /v1/config и /v1/rank сохранены для совместимости. Модель по умолчанию локальный Qwen; ключ модели не нужен. Для Authorize используется только SERVICE_API_KEY, если он настроен.'

example_dir = root / 'examples/stage1'
example_dir.mkdir(parents=True, exist_ok=True)
for name, entry in examples.items():
    (example_dir / (name + '.json')).write_text(json.dumps(entry['value'], ensure_ascii=False, indent=2) + '\n')

trip_schema = json.loads((root / 'ai-service/schemas/trip_request.schema.json').read_text())
trip_schema.pop('$schema', None)
trip_schema.pop('$id', None)
schemas['TripRequest'] = trip_schema
schemas['WrappedTripConfig'] = {'type':'object','additionalProperties':False,'properties':{'version':{'type':'string','enum':['1.0']},'request':{'type':'object','additionalProperties':True},'resolved':{'type':'object','additionalProperties':True},'trip':ref('PlanningTrip')},'description':'Используйте request или legacy trip; одновременно оба запрещены. resolved задаётся backend, не моделью.'}
schemas['PlanningConfig'] = {'anyOf':[ref('WrappedTripConfig'),{'type':'object','additionalProperties':True}],'description':'Сырой частичный TripRequest либо обёртка {version, request, resolved}. Полный TripRequest проверяется перед generate.'}
strict = json.loads((root / 'ai-service/schemas/ai_response.schema.json').read_text())
strict.pop('$schema',None)
schemas['AIResponse'] = strict
schemas['PlanningMetadata']={'type':'object','properties':{'summary':{'type':'string'},'config_wrapper':ref('WrappedTripConfig'),'missing_fields':{'type':'array','items':{'type':'string'}},'assumptions':{'type':'array','items':ref('PlanningAssumption')},'warnings':{'type':'array','items':{'type':'string'}}}}
schemas['PlanningDebugResponse']={'type':'object','required':['result','metadata'],'properties':{'result':ref('AIResponse'),'metadata':ref('PlanningMetadata')}}
for operation in ['analyze','update','refine']:
    endpoint=spec['paths']['/v1/planning/'+operation]['post']
    endpoint['description']='Принимает канонический TripRequest или TripConfig.request; legacy trip также поддерживается. По умолчанию ровно status, config, advice, questions по ai_response.schema.json. В config возвращается сырой TripRequest. resolved в модель не передаётся и сбрасывается при изменении request. Контекст хранит вызывающий сервис.'
    endpoint['parameters']=[{'in':'query','name':'include_metadata','schema':{'type':'boolean','default':False},'description':'Обёртка {result,metadata}: summary и config_wrapper для следующего вызова. В обычном ответе метаданных нет.'}]
    endpoint['responses']['200']['content']['application/json']['schema']={'oneOf':[ref('AIResponse'),ref('PlanningDebugResponse')]}
for field in ['transport_part_ids','hotel_part_ids']:
    schemas['RouteCandidate']['properties'][field]['nullable']=False
schemas['Money']['properties']['amount_minor']={'type':'string','pattern':'^[0-9]+$','description':'Целое int64 в минимальных единицах валюты. JSON-строка, как в protobuf; без дробных значений.'}
schemas['Money']['properties']['scope']={'type':'string','enum':['per_person_one_way','per_person_round_trip','stay_all_rooms','total']}
schemas['RankRoutesInput']['required']=['config','parts','candidates']
schemas['RankRoutesInput']['properties']['candidates']['minItems']=1
schemas['RankRoutesInput']['properties']['candidates']['maxItems']=20
schemas['RankRoutesInput']['properties']['limit']={'type':'integer','minimum':0,'maximum':20,'default':0}
ranking_example=json.loads((root/'examples/ranking/routes.json').read_text())
for name in ['routes','join-options','final-routes']:
    spec['paths']['/v1/ranking/'+name]={'post':{'tags':['Ранжирование'],'operationId':'rank_'+name.replace('-','_'),'summary':'Ранжировать переданные кандидаты: '+name,'description':'Готовые ID кандидатов и нормализованные provider parts подаются вызывающим сервисом. Сборки маршрутов здесь нет. Все три ручки используют общую проверку ограничений и оценку; join-options не строит стыковки. Неподтверждённые цены/валюты/детские тарифы не превращаются в нулевую стоимость. Ответ сохраняет исходные parts, errors и limitations. limit=0 возвращает все подходящие кандидаты.','security':[{'serviceKey':[]},{}],'requestBody':{'required':True,'content':{'application/json':{'schema':ref('RankRoutesInput'),'example':ranking_example}}},'responses':{'200':{'description':'Оценки, причины отклонения, исходные данные','content':{'application/json':{'schema':ref('RankRoutesOutput')}}},**error_responses}}}
spec.setdefault('tags',[])
if not any(t['name']=='Ранжирование' for t in spec['tags']):spec['tags'].append({'name':'Ранжирование'})
spec['info']['version']='1.3.0'
spec['info']['description']='Канонический TripRequest, строгий AIResponse, нормализованные provider parts и ранжирование готовых кандидатов. Локальный Qwen через Ollama. Для Authorize используется SERVICE_API_KEY, если задан.'
path.write_text(json.dumps(spec,ensure_ascii=False,indent=2)+'\n')
