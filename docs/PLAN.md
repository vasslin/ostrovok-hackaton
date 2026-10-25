# Battle plan: AI travel agent for Ostrovok

**Ostrovok Hackathon 2026 · track «О!дин запрос» · team godnota**
Written 2026-10-25. A living doc: edit it as you make decisions.

> **Update (2026-10-25):**
> - **Backend:** the team chose Go services (travel-service, ai-service, provider-service, user-service, gateway), not the Python monolith in §5 and §8. The ideas still apply; the layout doesn't.
> - **Provider APIs and configs:** see [providers/PLAN.md](providers/PLAN.md).
> - **Free LLM for the ai-service: GigaChat.** The free package for individuals lasts 12 months: Lite 250M tokens, Pro 40M, Max 25M, GigaChat-3-Ultra 50M. It handles one request at a time, and tool calls use OpenAI's older `functions` format (details in §2).

---

## 0. TL;DR

- **What we build:** a web app where a traveler writes one message in their own words and an agent takes them all the way: inspiration → destination → hotels, flights or trains, and transfers → comparison → one plan with checked connections → cart → mock payment → after-purchase help. When conditions change, it updates the plan instead of starting over.
- **"AI agent or API?"** Both. We call an LLM (GigaChat) through its API. The *agent* is our Python code: a loop that lets the model call our functions (search, compare, check, book) until the job is done. Nothing to train.
- **Stack:** Python, FastAPI and an LLM SDK for the agent and tools; React for the web UI (Streamlit if nobody does frontend); SQLite; Docker Compose.
- **Data:** none is provided. Use free real sources (OSM hotels, Open-Meteo weather and sea temperature, Yandex Rasp schedules) plus generated prices and rates, and label every value real or synthetic.
- **What wins:**
  1. It works on **any** query the jury invents: data for any city, a 50+ query test set, graceful failures.
  2. The agent visibly **acts**: plan canvas, live activity feed, comparison tables, checks. Not a chat that retells search results.
  3. **Trust and control**: editable "what I understood" chips, why / trade-offs / risks on every option, connections and total cost checked in code.
  4. **Re-planning as a diff**: "+2 days → flights changed, hotel kept, +4 100 ₽".
  5. **Numbers**: your own eval with completion rate, number of questions, time to first recommendation, 0 connection errors.

---

## 1. What the jury will actually check

### 1.1 Requirements → our feature → demo proof

| # | Requirement | Our feature | Demo proof |
|---|---|---|---|
| 1 | Understand the request; ask only the questions you can't move forward without | Trip brief extracted from free text: explicit and hidden needs (budget, travelers, constraints, mood). At most 2 questions, with quick-reply buttons. Everything else becomes a visible, editable assumption | The jury's vague query gets 1 question, not a 15-question form |
| 2 | Inspiration and destination | Destination cards: why it fits, total cost for the whole party, weather and sea temperature for those dates, door-to-door travel time, visa | «Тёплое море в ноябре» → 3 options with numbers |
| 3 | Selection and comparison | Hotels, flights or trains, and transfers, ranked by what matters to *this* traveler: cancellation, meals, location, travel time, fit with dates and times | Comparison table with a score per criterion |
| 4 | Explaining the choice | Each option shows why it fits, what you give up, and risks. Change a criterion and the ranking updates | «Расстояние до пляжа не важно» → instant re-rank |
| 5 | Assembly and purchase | One timeline and cart. Checks: arrival → transfer → check-in, check-out → transfer → departure. Mock checkout | The agent catches a 06:00 arrival vs a 14:00 check-in and offers early check-in |
| 6 | Re-planning | Patch the brief, recompute only what changed, show a diff | «А если на два дня позже?», «с нами едет собака» |

Bonuses: after-purchase help (early check-in request, date change with the penalty from the rate's rules, reminders, a transfer for a hotel already bought) · memory between sessions · quality metrics · mobile web or messenger (responsive web plus Telegram).

### 1.2 Between the lines

- **Queries you haven't seen at the defense.** Working on anything matters more than one perfect happy path.
- **"Retelling search results in chat is not a solution."** Show the agent acting: a live activity feed («Проверяю стыковки…»), structured cards, a plan canvas, a cart.
- **"If data is mocked, say so."** Put a small real / cached / synthetic badge on every number, plus one slide about data.
- **Three user stories, not one.** Traveler: a clear plan. Ostrovok: the whole trip in one cart (conversion, average order, repeat visits). **Hotel:** recommended only to guests it suits, so fewer cancellations. Most teams will forget the hotel story. Cover it with «Лучше всего для… / Не подойдёт, если…» per hotel and guest type.
- **How earlier editions were judged** (Tproger article about this hackathon): online; an LLM pre-screened submissions and dropped empty or raw ones; then experts judged technical and product effectiveness; teams had about a week. **So the README counts, and `docker compose up` must just work.**
- **Ostrovok's own view** (managing director Daria Kochetkova, March 2026): travelers use AI but still want to keep control. They double-check, compare, read reviews. In Ostrovok's survey about 26% don't trust AI to organize a trip. That gives the pitch line: **the agent drives, the traveler keeps the wheel.**

---

## 2. AI agents in 10 minutes

An agent is **LLM + tools + loop + state**:

- **LLM** (GigaChat over its API): reads text, decides the next step, writes explanations.
- **Tools:** ordinary Python functions described with a JSON schema (`search_hotels`, `build_plan`, `checkout`…). The model can't run them. It *asks* for a call; your code runs the function and sends the result back.
- **Loop:** repeat "model → tool calls → results → model" until the model replies without calling a tool.
- **State:** your own objects (trip brief, plan versions, cart, bookings), stored in the backend, not in the model's memory.

```
user:  «хочу к тёплому морю в ноябре, недорого, с ребёнком»
model: tool_use update_brief{...} + tool_use suggest_destinations{...}   ← asks for 2 calls
code:  runs both → tool_result (JSON)
model: tool_use search_hotels{...}
code:  runs it → tool_result
model: text answer, stop_reason = "end_turn"      ← the UI already shows the cards
```

The core loop, in pseudo-code (the same shape in any language or SDK):

```text
messages = [system prompt, user message, today's date + compact trip state]
loop:
    reply = llm.chat(messages, functions=TOOLS)
    append reply to messages
    if reply has no function call: show reply.text to the user; stop
    result = run our Go/Python function(reply.function_call.name, reply.function_call.arguments)
    append {role: "function", name, content: result as JSON} to messages
    # on an error, send the error text as the result: the model usually recovers
```

**Which LLM: GigaChat** (free for individuals, works from Russia without a VPN):

- Free package for 12 months: Lite 250M tokens, Pro 40M, Max 25M, GigaChat-3-Ultra 50M. Use Max or Ultra for the agent, Lite for cheap jobs (synthetic reviews, simulated users in evals).
- Access: register a GigaChat API project on developers.sber.ru, take the authorization key, exchange it for an access token (`https://ngw.devices.sberbank.ru:9443/api/v2/oauth`, scope `GIGACHAT_API_PERS`). The token lives 30 minutes; the official `gigachat` SDK refreshes it for you.
- The OpenAI-compatible endpoint is `https://api.giga.chat/v1`. It needs the Russian Минцифры root certificate.

**GigaChat gotchas:**

- **One request at a time** on the free package. Queue LLM calls in the ai-service; evals run sequentially.
- **One function call per reply**, in OpenAI's older format: `functions` in the request, `function_call` in the reply, and the result goes back as a `role: "function"` message. So make tools chunky: one call should do a lot of work in code.
- If you use LiteLLM: model `gigachat/GigaChat-2-Max`, credentials in `GIGACHAT_CREDENTIALS`, pass `ssl_verify=False` (or install the certificate). Version 1.106.4 fails on `tools` without the `mcp` package: `pip install mcp`.
- Keep the system prompt and the function list identical between calls, and put today's date and the trip state in the newest message.
- Keep every LLM call in one module, so switching models later touches nothing else.

**Framework?** Not needed. LangGraph, Google ADK or CrewAI add concepts and latency. One agent with about 12 good tools plus deterministic Python is faster to build, easier to debug and more reliable. Multi-agent setups look good on slides but break more often on unseen queries.

**Golden rule:** *the LLM handles understanding, decisions, questions and explanations; code handles facts, prices, dates, math, filters and checks.* The model never invents a hotel, price or time. Every fact it states comes from a tool result, and you can check that by id.

---

## 3. Has anyone solved this already?

**Not this exact task.** The track is new in 2026, this year's team repos are private (yours too), and last year's public team repos (`github.com/ostrovok-hackathon-2025`) are for a different track ("Secret Guest"). These are the closest projects and research to borrow from:

| What | Borrow | Note |
|---|---|---|
| Google ADK **Travel Concierge** sample (`google/adk-samples`, now renamed `google/adk-recipes`) | The same stage split as this track: inspiration → planning → booking → pre-trip → in-trip → post-trip; mock data; a user profile as memory | Multi-agent on Google's stack. Copy the structure, not the framework. Its path after the rename is unconfirmed |
| `vercel-labs/gemini-chatbot` (~1.3k★) | Generative UI: flight search → pick a flight → seat → payment → boarding pass, all as interactive widgets in the chat | Next.js. Good for UI ideas |
| `nirbar1985/ai-travel-agent` (~800★) | LangGraph agent for flights and hotels; asks a human before risky actions | Uses SerpApi |
| `serpapi/travel-planning-agent` | Plans its searches, validates structured inputs (IATA codes), cites sources | Small and readable |
| TripWeaver (LangGraph + MCP + Gradio) | Each data source is a separate MCP server; the UI streams a live trace of tool calls | |
| Mastra travel planner (TypeScript) | Typed tool schemas (Zod) | |
| **TravelPlanner** benchmark (OSU NLP, ICML 2024) | GPT-4-era agents met every constraint in under 5% of plans. **Check constraints in code; don't trust the LLM with them** | Paper + dataset |
| **Flex-TravelPlanner** (2025) | When re-planning over several turns, models favor a newly added low-priority wish over an earlier hard constraint. **Keep constraints in explicit state, marked hard or soft** | |
| **Otello** AI assistant (Russian booking service) | Free-text hotel search; results as cards in the chat; book from the card | This is the "cards in chat" baseline. We have to clearly beat it |
| Booking.com AI Trip Planner write-ups | Moderation, intent classification, answers grounded in data | |

---

## 4. Data: what's real, what's generated

Strategy: **real where it's free and reliable, generated where it isn't, labeled everywhere, cached on disk.**

| Need | Source | Real? | Key? | Notes |
|---|---|---|---|---|
| City → coordinates, country | Nominatim (OpenStreetMap) | real | no | 1 request/s, needs a User-Agent; cache |
| Hotels: names, locations, stars, tags | OSM Overpass (`tourism=hotel`, `stars=*`) | real | no | public servers are rate-limited, so pre-fetch and cache |
| Prices, availability, rooms, meal plans, cancellation rules, check-in times, kid and pet policies | our generator, seeded by hotel id (city price level × stars × season × noise) | synthetic | — | stable across calls; badge «демо-цены» |
| Reviews of Russian hotels | Yandex Geo Reviews Dataset 2023 (~500k reviews of Russian businesses, hotels included) | real | no | check the license |
| Reviews elsewhere | LLM-generated per guest type, consistent with the hotel's attributes | synthetic | — | generate once, store |
| Air temperature and rain for a month | Open-Meteo Historical Weather (ERA5, 1940 to now) | real | no | average the same month over recent years |
| Forecast for trips within ~2 weeks | Open-Meteo Forecast | real | no | |
| **Sea temperature** | Open-Meteo Marine API, `sea_surface_temperature` | real | no | history from 2022 → «море в ноябре: X °C» from real data |
| Flight and train schedules from and within Russia | Yandex Rasp API v3 (Яндекс Расписания) | real | free key | use the `api.rasp.yandex-net.ru` domain; attribution required |
| Flight prices (Russian market) | Travelpayouts / Aviasales Data API (cheapest prices users found in the last 2–7 days) | real, cached | free partner token | check the access rules; otherwise generate |
| Flight duration and price fallback | great-circle distance between airports → duration and price model | synthetic | — | |
| Airports, IATA codes | OurAirports CSV | real | no | public domain |
| Transfer drive time and distance | OSRM (OSM routing) | real | no | the demo server is rate-limited; cache |
| Transfer prices | per-km tariff by city | synthetic | — | |
| Visa rules for a Russian passport | passport-index-dataset (GitHub), or a hand-made table for ~40 destinations | real / curated | no | check it's up to date |
| Holidays (crowds, price peaks) | Nager.Date API | real | no | |
| Exchange rates | Central Bank of Russia daily XML | real | no | |
| Destination text and photos | Wikivoyage / Wikipedia, Wikimedia Commons | real | no | CC BY-SA attribution |
| **Destination catalog** ("where to go") | hand-curated 50–80 places popular with travelers from Russia, tagged: sea, ski, city, nature, kids, visa, direct flights | curated | — | the inspiration stage depends on it |

**Dead ends, don't start there:**
- **Amadeus Self-Service APIs** shut down on **2026-07-17**. Tutorials built on them no longer work.
- ETG / RateHawk B2B API (docs.emergingtravel.com) is real but needs a contract. Use its flow as the template for our mock (below).
- Scraping booking sites: slow, fragile and legally grey. Synthetic data with honest labels is explicitly allowed.

**Make *any* destination work** (the key to unseen queries): city → geocode → real OSM hotels nearby → generated rates, stable per hotel → flights from distance or real schedules → climate from Open-Meteo. Pre-cache the ~30 likeliest destinations so the demo is fast and doesn't depend on rate-limited public servers.

**Mock booking flow modeled on a real B2B hotel API.** ETG's B2B API v3 (RateHawk / ZenHotels) documents this sequence: search (`/search/serp/region/`, `/geo/` or `/hotels/`; 1–2 cheapest rates per hotel) → hotel page (`/search/hp/`, all rates) → prebook (`/hotel/prebook`, refreshes the price; any change must be shown to the user) → booking form → booking finish → status polling → order info / cancel. Copy that shape into `providers/hotels_mock.py`, including an occasional price change at prebook that the agent has to handle. Pitch line: "our mock follows a real B2B booking flow, so plugging in a real supplier is an adapter, not a rewrite." Don't present it as Ostrovok's own API.

---

## 5. Architecture

### 5.1 The picture

```
┌──────────────────── Web UI (React) ───────────────────┐   ┌─ Telegram (bonus) ─┐
│  Chat + quick replies   │  Trip canvas: brief chips,  │   │  same backend API  │
│  + live activity feed   │  cards, compare, timeline,  │   └─────────┬──────────┘
│                         │  cart, checkout, my trips   │             │
└────────────┬────────────┴─────────────────────────────┘             │
             │  SSE: text chunks + UI events (cards, checks, diffs)   │
┌────────────▼────────────────────────── FastAPI backend ────────────▼──────────┐
│  Agent loop (LLM API, function calls)                                          │
│    system prompt (rules)  +  today's date & trip state each turn               │
│    └─ tools ──► Planner core: plain, deterministic, unit-tested Python         │
│                  brief · destinations · hotels · transport · transfers        │
│                  ranking & total trip cost · checks · plan versions & diffs   │
│                  cart · mock booking (prebook → book) · after-purchase        │
│  Providers: Open-Meteo, OSM, Yandex Rasp, Travelpayouts, synthetic  (cached)   │
│  Storage: SQLite (sessions, plan versions, bookings, user memory)              │
└────────────────────────────────────────────────────────────────────────────────┘
          ▲
          └── eval runner: simulated users call the same backend (no UI)
```

Keep the backend UI-agnostic: it emits events. The web app, the Telegram bot and the eval runner are all just clients.

### 5.2 Trip state (Pydantic sketch)

```python
from datetime import date
from pydantic import BaseModel, Field

class Travelers(BaseModel):
    adults: int = 2
    children_ages: list[int] = []
    pets: list[str] = []                     # "собака, 6 кг"

class Assumption(BaseModel):
    field: str                               # "origin", "nights", "budget"
    value: str
    reason: str                              # "из прошлой поездки", "типично для семейного отдыха"
    confirmed: bool = False

class TripBrief(BaseModel):
    origin: str | None = None
    destination: str | None = None           # None → inspiration stage
    date_from: date | None = None
    date_to: date | None = None
    flexible_days: int = 0
    nights: int | None = None
    travelers: Travelers = Field(default_factory=Travelers)
    budget_rub: int | None = None
    budget_is_hard: bool = False
    hard_constraints: list[str] = []         # "max_flight_hours<=5", "pet_friendly"
    preferences: dict[str, float] = {}       # weights: {"beach": 0.9, "all_inclusive": 0.7}
    mood: str | None = None                  # "спокойный семейный пляжный отдых"
    assumptions: list[Assumption] = []

class Plan(BaseModel):                       # TransportOption, HotelRate, ... defined the same way
    version: int
    brief: TripBrief                         # snapshot of what this plan was built for
    destination: str
    outbound: "TransportOption"
    inbound: "TransportOption"
    hotel: "HotelRate"                       # hotel + room + meal plan + cancellation terms
    transfers: list["TransferOption"]
    checks: list["Check"]                    # ok / warn / error, message, suggested fix
    total: "PriceBreakdown"                  # hotel, transport, transfers, fees, meals estimate
    tradeoffs: list[str]
    risks: list[str]
```

Show `assumptions` as chips in the UI. A click edits the brief, which triggers re-planning.

### 5.3 Tools (about 12; a few chunky tools beat many tiny ones)

| Tool | What it does (in code) |
|---|---|
| `update_brief(patch)` | Merges explicit and inferred needs, records assumptions. Returns missing blocking fields and which parts of the plan are now stale. Called on every new user input |
| `ask_user(question, options)` | Shows a question with quick-reply buttons; the turn ends |
| `suggest_destinations(...)` | Scores the catalog: climate and sea temperature for the dates, door-to-door time from the origin, estimated total for the party, visa, tags → top 3–5 with reasons |
| `search_transport(origin, dest, dates, pax, mode)` | Flights and/or trains with times, duration, price, baggage, pet rules |
| `search_hotels(dest, dates, guests, filters)` | Top N with score breakdown and total trip cost |
| `get_hotel(hotel_id, dates)` | All rates (meals × cancellation), policies, "best for / not for", review summary. Like a B2B "hotel page" |
| `search_transfers(from, to, at, pax, extras)` | Taxi / shuttle / private (child seat, pets), with time and price |
| `build_plan(selection)` | Assembles the itinerary, runs every check, computes totals, saves a new plan version and its diff from the previous one |
| `price_calendar(plan_id, days)` | Total trip cost for nearby start dates. Powers «а если на два дня позже?» |
| `checkout(plan_id)` | Prebook (may return a price change the user must confirm) → mock payment → booking |
| `manage_booking(booking_id, action, ...)` | Early check-in request, date change (penalty from the rate's rules), add a transfer, reminders (.ics) |
| `remember(fact)` | Saves a preference to the user's profile |

Each tool also pushes a UI event (`destinations`, `hotels`, `plan`, `diff`, `cart`, `question`), so the canvas fills in while the model is still writing. Tool results sent back to the model stay compact (top options, ids, key fields); the full data goes only to the UI.

**Make tools do the heavy lifting.** `suggest_destinations` should check climate, travel time and price for 30 candidates inside one call, in parallel, in code. That's one model round trip instead of 30, so it's faster and more reliable.

### 5.4 Checks done in code (and unit-tested)

| Check | Rule (tune the numbers) | If it fails |
|---|---|---|
| Early arrival | arrival + transfer + 30 min is ≥3 h before check-in | offer early check-in (fee), luggage storage, or booking the night before |
| Late arrival | after 23:00 local time | tell the hotel about the late arrival; needs 24h reception |
| Overnight flight | lands the next day, so hotel dates shift | error: fix the dates |
| Departure buffer | check-out + transfer leaves < 2 h (international) / 1.5 h (domestic) before the flight | error: earlier transfer |
| Long gap before departure | > 4 h between check-out and leaving for the airport | offer late check-out or luggage storage |
| Nights | hotel nights ≠ trip nights | error |
| Room fits the party | adults + children (+ cot for infants) | error |
| Kids | adults-only hotel with children | error |
| Pets | hotel / airline / train / transfer rules, weight limits | error, with alternatives |
| Budget | total > budget (hard) or > +10% (soft) | cheaper swaps, saying what you lose |
| Documents | visa needed; passport validity; time to get a visa before departure | warning |
| Weather | rainy season; cool sea for small kids | info / warning |
| Time zones | every time shown in local time at its place | — |

### 5.5 Ranking and "total trip cost"

- `score = Σ weight × criterion_score`, each criterion in [0, 1]: price fit, location (beach / center / venue), meals, cancellation flexibility, family fit, pet fit, quietness, rating, door-to-door time, timing fit (arrival vs check-in).
- Weights start from the brief (family → family fit and meals up; tight budget → price up). The user can edit them in the UI.
- **Total trip cost** = hotel + transport for everyone + transfers + fees (early check-in, pets, baggage) + an estimate for meals not included. That answers the task's «не съест ли трансфер всю экономию на отеле» with numbers: «Отель B дешевле на 6 000 ₽, но трансфер туда-обратно 7 400 ₽ — итого B дороже».
- Explanations: the LLM writes 2–3 lines from the score breakdown: top reasons (why), weakest criterion (trade-off), failed checks (risks). Grounded, short, consistent.

### 5.6 Re-planning (task 6)

Patch the brief, then a dependency map decides what to recompute:

| Change | Recompute | Keep |
|---|---|---|
| Dates | transport both ways, hotel availability and price, transfer times | destination, chosen hotel if still available, preferences |
| Budget | ranking; maybe hotel, meal plan, refundable vs non-refundable | destination, dates; transport unless the budget forces a change |
| Party (+child / +dog / +person) | rooms, policies, transport rules, prices | destination and dates where still valid |
| Destination | everything except the brief | brief |
| Preference («ближе к пляжу») | ranking only | everything else |

Each change creates plan v(n+1) plus a **diff card**: each component unchanged / changed / removed, the price delta, and the reason. For date changes, show a **price calendar** with the total trip cost for start dates ±3 days. That's the best possible answer to «а если на два дня позже?».

### 5.7 Memory and after-purchase (bonuses 1 and 2)

- **Memory:** a SQLite profile per user (cookie or Telegram id): home city, travelers (children stored by birth year, so ages stay correct next year), likes and dislikes (night flights, all-inclusive), usual budget, past trips. Shown in a «Что я о вас помню» panel where each item can be deleted. Loaded into the trip state at session start.
- **After purchase:** "My trips" with actions. Early check-in request: the agent writes the message to the hotel and the request gets a status. Date change: re-plan plus the penalty from the booked rate's cancellation rules. Add a transfer to the bought hotel, timed from the flight arrival. Reminders: an `.ics` file with alarms (online check-in, passport, weather 3 days before).

---

## 6. Conversation design

### 6.1 Clarifying questions

- Pull everything you can from the first message, then fill gaps with sensible defaults: home city from memory; «ноябрь» → flexible dates in November, 7 nights; «недорого» → a budget band for this party and destination.
- Ask only when a missing value blocks a sensible search or changes the price a lot: origin if unknown, rough dates if none, and children's ages when «с ребёнком» comes without an age (it changes prices and rules).
- At most 1–2 questions per turn, each with quick-reply buttons and a «подбери без этого» option.
- Show every assumption as an editable chip: «Вылет из Москвы ✎», «7 ночей ✎», «До 150 000 ₽ ✎».
- Act first, refine later: show 3 destinations right after at most one round of questions. Target time to first recommendation: under ~30–40 s.
- Never pay without an explicit «Оплатить» from the user.
- Resolve dates in code, not in the LLM: «через две недели», «на майские», «на выходных». Today's date is sent every turn.

### 6.2 Hidden requirements → explicit

Put this table in the system prompt, and use the same presets in code:

| They say | It implies |
|---|---|
| «с ребёнком» | ask the age; family room; kids' club or pool; gentle sandy beach; no night flights; child seat in the transfer |
| «без долгих перелётов» | ≤ ~4–5 h; direct flights; door-to-door time matters |
| «недорого» | a value band, not the cheapest at any cost; show total trip cost |
| «с собакой» | pet-friendly hotel (weight limits, fee); airline or train pet rules; a car that allows pets; vet papers when crossing borders |
| «медовый месяц» | adults-only or quiet; sea view; romantic extras |
| «с бабушкой» | accessibility, elevator, short walks, no early flights |
| «на выходные» | Friday evening to Sunday; a train may beat a plane door to door |
| «командировка» | near the venue; flexible cancellation; breakfast; invoice documents |
| «хочу тусоваться» / «тихо» | nightlife nearby vs away from it |

### 6.3 Example: the jury's own query

«Хочу куда-нибудь к тёплому морю в ноябре, недорого, с ребёнком, без долгих перелётов»

1. **Brief:** warm sea (hard); November, flexible dates; budget "low" (soft, needs a number); 2 adults (assumed) + 1 child (age unknown); flight ≤ ~5 h, direct, daytime (hidden); mood: relaxed family beach.
2. **One question:** «Откуда летите и сколько лет ребёнку?» [Москва] [Санкт-Петербург] [другое] · [до 2] [2–6] [7–12] [13+]. The assumption chips are shown next to it.
3. **3 destinations**, for example Hurghada (warmest sea, ~5 h, cheapest all-inclusive) · Antalya (shortest flight, but the sea is cooling) · Dubai (warm, pricier). Each comes with the family's total, real sea and air temperatures, flight time and visa rules. Real numbers come from the tools; these are illustrative.
4. **Pick one** → hotels compared on family fit, kids' club, beach distance, meals and free cancellation, each with why / trade-offs / risks; daytime direct flights; a transfer.
5. **Plan timeline with checks** («прилёт 14:30 + 25 мин трансфер → заселение 15:00 ✓») → cart → prebook (any price change is shown) → mock payment → confirmation.
6. **«А если на два дня позже?»** → diff plus price calendar. **«С нами едет собака»** → swap to a pet-friendly hotel with an explanation, airline pet rules added as a risk, a pet-friendly transfer.

### 6.4 System prompt outline (short and stable, so it caches)

- Role: Ostrovok's travel agent. Reply in the user's language (Russian by default), briefly; the cards carry the details.
- Process: brief → (destinations) → options → plan → checks → cart → purchase → after-purchase. Call `update_brief` whenever the user adds or changes anything.
- The question policy (6.1) and the hidden-requirements table (6.2).
- Grounding: only state hotels, prices, times and policies that came from tools. If a tool has no data, say so.
- For every recommendation: why it fits, what you give up, risks.
- Don't move to checkout with failing checks unless the user explicitly accepts the risk.
- Say plainly that prices and availability are demo data.

---

## 7. UI

- **Top:** a stage progress bar in the task's own words: Вдохновение → Направление → Подбор → Сравнение → Сборка → Покупка → После покупки. The jury sees coverage at a glance.
- **Left:** chat with quick-reply buttons and a live activity feed («Ищу отели в Хургаде на 14–21 ноября для 2+1…», «Проверяю стыковки…»).
- **Right, the trip canvas:**
  - «Что я понял»: brief chips, editable.
  - Destination cards: photo, why, total cost, air and sea temperature, door-to-door time, visa.
  - Hotel cards plus a comparison table: score bars per criterion, «Лучше всего для / Не подойдёт, если», reviews by guest type, and a map (Leaflet + OSM: hotel, beach, airport, route).
  - Plan timeline: departure → arrival → transfer → check-in → … → departure, each step green / yellow / red with a "fix" button.
  - Cart with the total trip cost breakdown → checkout (mock card form, prebook price-change dialog) → confirmation.
  - My trips (after-purchase actions) and the memory panel.
- **Data badges** (real / cached / synthetic) on numbers, and a «Источники данных» page.
- **Mobile:** responsive layout with the canvas as tabs. Together with Telegram (a bot, or a Mini App that opens the same web UI) that covers bonus 4.

**Fresh ideas** (the organizers say they value them; pick 2–3): door-to-door time instead of flight time, so plane vs train is compared fairly · price calendar for date shifts · "not for you if…" anti-recommendations · a share link so the family can decide together · voice input in the mobile UI · the tools exposed as an MCP server ("any AI assistant can sell a whole trip through it"; a couple of hours, since the tools are already Python functions).

---

## 8. Repo layout

```
backend/
  app/main.py          FastAPI: POST /chat (SSE stream), GET /state, /trips, /metrics
  agent/llm.py         the only module that talks to the LLM API
  agent/loop.py        the agent loop
  agent/prompts.py     system prompt
  agent/tools.py       tool schemas + dispatch to core/
  core/models.py       TripBrief, Plan, Cart, Booking (Pydantic)
  core/destinations.py core/hotels.py core/transport.py core/transfers.py
  core/scoring.py      weights, score breakdown, total trip cost
  core/checks.py       connection and policy checks (unit-tested!)
  core/replan.py       dependency map, plan versions, diffs, price calendar
  core/booking.py      cart, mock prebook / book / cancel (B2B-API-like flow)
  core/memory.py       user profiles
  providers/           open_meteo.py osm.py rasp.py travelpayouts.py synthetic.py cache.py
  data/                destinations.yaml, cached API responses, generated hotels
  tests/               pytest for core/ (checks, scoring, replan)
frontend/              React + Vite + TypeScript + Tailwind (+ shadcn/ui, Leaflet)
eval/
  scenarios.yaml       50+ test queries with personas and expectations
  run_eval.py          simulated users + metrics → eval/results/*.json
docker-compose.yml  .env.example  README.md
```

---

## 9. Roadmap

Assumes about a week; earlier editions ran about a week, but **confirm the deadline**. Rule: **get the whole path working end to end, ugly, by day 2; then make each stage better.** Write the test queries *before* building, so you don't tune the agent to your own favorite example.

### Next 24 hours

- [ ] Confirm the deadline, the defense format (live or online, how long, does the jury type the queries?) and what to submit (repo, video, slides). Ask in the organizers' Telegram bot or channel.
- [ ] Get a working LLM API key with a spending limit; make a hello-world call.
- [ ] Agree on the stack and repo layout; create the skeleton and `.env.example`.
- [ ] Write the Pydantic models, tool contracts and SSE event types. These are the interfaces between teammates.
- [ ] Destination catalog v1 (20 places) + synthetic hotels for 5 demo cities.
- [ ] Agent loop with 3 tools (`update_brief`, `search_hotels`, `checkout`) working in the terminal.
- [ ] 20 test queries written down (grow to 50+ later).

### Day by day

| Day | Goal | Done when |
|---|---|---|
| 1 | Skeleton + loop + first tools | you can chat in the terminal and a hotel lands in the cart |
| 2 | **End to end, ugly (P0):** destinations with Open-Meteo, flights (synthetic first), transfers, `build_plan` with 3–4 checks, mock checkout, a minimal web UI with cards | first message → mock payment works. Commit and record a backup video right away |
| 3 | **Smart:** question policy and assumption chips, why / trade-offs / risks, score breakdown, total trip cost, re-planning with diffs, real data swapped in (OSM, marine, Rasp, Travelpayouts) | all 6 main tasks can be demoed |
| 4 | **Robust:** eval set + simulated users + metrics; fix failures; edge cases; pre-caching; latency work (parallel tools, streaming, effort level) | completion rate ≥ 90% on your set |
| 5 | **Bonuses that score:** after-purchase actions, memory, mobile + Telegram, metrics page | each bonus can be shown in under 30 s |
| 6 | **Polish + pitch:** UI polish, data badges, map, README with GIFs, public URL, `docker compose up`, slides, rehearsals | feature freeze; bug fixes only |
| 7 | Buffer; rehearse with friends throwing random queries; submit early | submitted |

### Priorities when time runs out

- **P0, must:** the full path, the question policy, destination cards, comparison of hotels + transport + transfers, explanations, plan with checks, cart + mock payment, re-planning with diffs, honest data labels, the README.
- **P1, differentiators:** editable brief chips, total trip cost, price calendar, prebook price change, eval metrics, real data sources, streaming + activity feed, memory.
- **P2, nice to have:** the full after-purchase set, Telegram, map, hotel "best for / not for", tools as an MCP server, voice input.

### Team split

Agree on the interfaces on day 1, then build in parallel against mocks.

- **A, agent:** loop, prompts, brief, question policy, re-planning.
- **B, data and core:** providers, synthetic generator, scoring, checks, booking flow, tests.
- **C, frontend:** chat, canvas, cards, checkout, mobile, Telegram.
- **D, quality and story:** eval runner, metrics, README, slides, demo script. Also the person who tries to break the agent.

With 2 people: A+B and C+D. Solo: follow the day plan in order and use Streamlit for the UI.

---

## 10. Quality metrics (bonus 3, and how you get robust)

**Test set** (`eval/scenarios.yaml`), 50+ queries in about 10 groups:

- Vague inspiration: «хочу к морю в ноябре недорого с ребёнком», «куда поехать на новогодние, чтобы был снег и не очень дорого»
- Exact booking: «Казань, 12–14 декабря, двое взрослых, до 8 000 ₽ за ночь, рядом с Кремлём»
- Families: infants, teenagers, three generations
- Pets: «с собакой 15 кг на машине по Золотому кольцу»
- Tight or impossible: «Мальдивы, 5 звёзд, 30 000 ₽ на двоих на неделю»
- Trains and weekends: «на выходные из Питера на поезде, куда-нибудь с красивой природой»
- Multi-city or long trips: «10 дней по Узбекистану»
- Business: «командировка в Екатеринбург, нужен отель рядом с Экспо и закрывающие документы»
- Accessibility: wheelchair users, elderly travelers
- Mid-flow changes: dates, budget, +dog, +person, a different destination
- Edge cases: an English query, typos and slang, off-topic, rudeness, past dates, «завтра», a group of 20

**Simulated user:** an LLM playing a persona with hidden facts (origin, dates, ages, budget). It answers only what it's asked, accepts reasonable options, asks for one change, then pays. Cap it at ~15 turns.

**Measure**, and show "v1 → final" in the pitch:

| Metric | Target |
|---|---|
| Completion rate (reached a confirmed mock booking) | ≥ 90% |
| Clarifying questions before the first recommendation | ≤ 2 |
| Time to first recommendation | < 30–40 s |
| Connection and policy errors left in final plans | 0 |
| Hard-constraint violations (budget, pets, kids, flight time) | 0 |
| Grounding errors (a hotel or price in the text that no tool returned; checked automatically) | 0 |
| Plan reuse after a change (% of components kept) | report it |
| Cost per conversation, tokens, p50 / p95 latency | report it |
| Explanation quality (LLM judge: why, trade-off and risk present and correct) | report it |

Simulated-user runs cost real tokens. Run the full set a few times, not after every edit.

---

## 11. Demo, pitch, README

**Demo script (5–7 min), then the jury's own queries:**

1. Hook (20 s): the 10-tabs pain, plus Ostrovok's own finding that travelers want control → "the agent drives, you keep the wheel".
2. Live: the vague query → 1 question → 3 destinations with numbers → hotels with why / trade-offs / risks → a plan with a caught problem (early arrival → early check-in offered) → cart → prebook → payment → confirmation.
3. Re-planning: «а если на два дня позже?» → diff and price calendar; «с нами едет собака» → pet-friendly swap.
4. After purchase: «добавь трансфер» / «можно ранний заезд?» on the booked trip.
5. Numbers: N scenarios, completion %, questions asked, time to first recommendation, 0 connection errors.
6. Architecture and honesty: what's real, what's synthetic, the B2B-flow mock.
7. Business: the whole trip in one cart (bigger average order), fewer abandoned searches, better hotel fit (fewer cancellations); roadmap.

**Defense tips:** pre-warm the caches; use a "demo mode" seed so prices stay stable; keep a recorded backup video and a public URL. If the agent can't do something, it should say so and offer an alternative. Juries reward honest, graceful behavior.

**README checklist** (an LLM may read it first):

- [ ] One-paragraph pitch, a GIF or screenshots, links to the live demo and the video.
- [ ] A table mapping features to main tasks 1–6 and bonuses 1–4.
- [ ] How to run: `docker compose up` + `.env.example`.
- [ ] Architecture diagram, the list of tools, the data sources table (real / synthetic).
- [ ] Metrics results and known limitations.

---

## 12. Traps to avoid

- A 15-question form at the start.
- Letting the LLM write the itinerary, prices included, from its own head.
- Starting with a multi-agent framework.
- Spending two days scraping real prices.
- Calling public APIs live during the defense without a cache.
- "Re-planning" by editing the conversation history (and the API returns 400 for edited history on newer accounts).
- Setting `temperature`; following Amadeus tutorials.
- An empty README.

## 13. Risks

| Risk | Mitigation |
|---|---|
| No LLM API access or budget | solve on day 1; one module for LLM calls |
| Public APIs rate-limited or down | disk cache, pre-warmed destinations, synthetic fallback |
| Slow responses | parallel tool calls, chunky tools, streaming + activity feed, lower effort for simple turns, precomputed data |
| Unseen queries break it | 50+ eval queries, data for any city, a graceful «так не получится, но вот вариант» |
| Hallucinated facts | facts only from tools, plus an automatic grounding check |
| Scope creep | P0 first; feature freeze on day 6 |

## 14. Open questions (answer today)

1. The exact deadline, the defense format and length, and what to submit.
2. Team size, and who does frontend → React or Streamlit.
3. Which LLM API you can pay for and use from where you are, and the budget.
4. Is a public URL required, or is the repo enough?

---

## Sources

- Track description: https://docs.ostrovok.tech/s/0ee4d8b2-ae50-4dda-a2b5-112fcdb5398d
- How the Ostrovok hackathon is run (Tproger): https://tproger.ru/articles/go--llm-i-1-4-mln-ohvata--kak-ostrovok-delaet-edinstvennyj-regul
- Ostrovok on AI and trip planning (March 2026): https://eanews.ru/ekaterinburg/20260318140250/sverdlovchane-stali-chasche-polzovatsya-iskusstvennym-intellektom-pri-planirovanii-puteshestviy
- Otello's AI assistant: https://hi-tech.mail.ru/news/146386-servis-otello-zapustil-ii-assistenta-dlya-planirovaniya-puteshestvij/
- Last year's public team repos: https://github.com/ostrovok-hackathon-2025
- Google ADK Travel Concierge: https://developers.google.com/workspace/add-ons/samples/travel-concierge · https://docs.langdb.ai/guides/building-agents/building-travel-concierge-with-google-adk
- Generative UI chatbot: https://github.com/vercel-labs/gemini-chatbot
- https://github.com/nirbar1985/ai-travel-agent · https://github.com/serpapi/travel-planning-agent · https://mastra.ai/blog/travel-ai
- TravelPlanner: https://arxiv.org/abs/2402.01622 · https://github.com/OSU-NLP-Group/TravelPlanner · Flex-TravelPlanner: https://arxiv.org/abs/2506.04649
- Amadeus Self-Service shutdown: https://www.phocuswire.com/amadeus-shut-down-self-service-apis-portal-developers
- ETG API v3 docs: https://docs.emergingtravel.com/docs/integration-guide/
- Open-Meteo Marine API: https://open-meteo.com/en/docs/marine-weather-api
- Yandex Rasp API: https://yandex.ru/dev/rasp/raspapi
- Travelpayouts / Aviasales Data API: https://support.travelpayouts.com/hc/en-us/articles/203956163
- YandexGPT function calling: https://yandex.cloud/en/docs/ai-studio/concepts/generation/function-call
