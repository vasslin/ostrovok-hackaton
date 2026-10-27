# cbr — курсы валют ЦБ РФ

**Банк России, ежедневные курсы.** Чтобы привести цены в рубли: зарубежные отели, билеты в валюте, бюджет пользователя.

**Статус:** ✓ проверено 2026-10-25, HTTP 200. **Ключ не нужен.**

[Официальное описание XML API Банка России](https://www.cbr.ru/development/SXML/).

## Ручка `daily`

`GET https://www.cbr.ru/scripts/XML_daily.asp`

| Параметр | Обязат. | Формат | Пример |
|---|---|---|---|
| `date_req` | нет | `DD/MM/YYYY`; без него — на последнюю зарегистрированную дату | `25/10/2026` |

Конфиг: `configs/examples/cbr.daily.json`.

```bash
curl -s "https://www.cbr.ru/scripts/XML_daily.asp?date_req=25/10/2026" | iconv -f windows-1251 -t utf-8
```

Ответ — **XML в кодировке windows-1251** (настоящий, сокращён; полный — `testdata/cbr/daily.xml`):

```xml
<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="24.10.2026" name="Foreign Currency Market">
  <Valute ID="R01010"><NumCode>036</NumCode><CharCode>AUD</CharCode><Nominal>1</Nominal>
    <Name>Австралийский доллар</Name><Value>56,4250</Value><VunitRate>56,425</VunitRate></Valute>
  <Valute ID="R01030"><NumCode>012</NumCode><CharCode>DZD</CharCode><Nominal>100</Nominal>
    <Name>Алжирских динаров</Name><Value>59,7463</Value><VunitRate>0,597463</VunitRate></Valute>
</ValCurs>
```

| Поле | Значение |
|---|---|
| `CharCode` | код валюты ISO 4217 (`USD`, `EUR`, `TRY`, `AED`…) |
| `Nominal` | за сколько единиц валюты указан `Value` (**бывает 10, 100**) |
| `Value` | рублей за `Nominal` единиц |
| `VunitRate` | рублей за **одну** единицу — проще брать его |
| `ValCurs/@Date` | дата, на которую действует курс |

## Подводные камни

- **Кодировка windows-1251.** В Go: `golang.org/x/net/html/charset` → `charset.NewReaderLabel` при разборе XML.
- **Десятичная запятая** (`56,4250`): заменить на точку перед разбором числа.
- **Выходные:** на запрос за воскресенье 25.10 пришёл курс от **24.10** — ЦБ отдаёт последний рабочий день. Это нормально.
- **Кэшировать на полдня:** курс меняется раз в день.
