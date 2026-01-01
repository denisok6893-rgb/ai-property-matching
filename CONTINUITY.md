# CONTINUITY — ai-property-matching

## Текущий статус (2025-12-31)
- MVP v0 стабилен: `go test ./...` и `go run ./cmd/api` работают.
- Адрес по умолчанию: `API_ADDRESS` (дефолт `:8080`).
- Данные и веса:
  - `data/properties.json`
  - `configs/weights.json`

## Что проверено вручную (curl)
- Health:
  - `GET /health` -> `{"status":"ok"}`
- Properties API (in-memory, без записи в файл):
  - `GET /properties?limit=...&offset=...` возвращает список и `total`
  - `POST /properties` создаёт объект, возвращает полный `domain.Property` (id вида `p-N`)
  - `GET /properties/{id}` работает
  - `DELETE /properties/{id}` -> `{"status":"deleted"}`, затем `{"error":"not_found"}`

## Что сделано сегодня
- Обновлён `README.md` с точными примерами запуска и curl (health/properties/match).
- Изменения закоммичены и запушены в `main`.

## Следующий шаг
- B: уточнить `domain.ClientProfile` и добавить в README реальный пример запроса `POST /match` под фактические поля профиля.

Слелано 31.12.25
- Проверен `POST /match` с реальным `ClientProfile`, получен результат со `score` и `reasons`.
- README обновлён: добавлен реальный пример запроса/ответа `/match`.
-Учтён location_preference как soft-boost (location_match в reasons).

Проверено вручную через /match на порту :8081.
Добавлен /demo: список объектов, детали по клику, отображение description и image_urls.

Последние изменения:

## 2026-01-01 — Properties: repo abstraction + SQLite фильтры

- В `internal/http` введён интерфейс `PropertiesRepo`; текущая логика сохранена как `InMemoryPropertiesRepo` (фильтр → сортировка → пагинация).
- Добавлен `SQLitePropertiesRepo` и подключение в `cmd/api/main.go` при `STORAGE=sqlite`.
- В `internal/storage/sqlite_store.go` добавлен `ListPropertiesFiltered()`:
  - WHERE: location (contains, case-insensitive), min_price, max_price, min_bedrooms
  - ORDER BY: price_asc|price_desc
  - LIMIT/OFFSET + COUNT(*) с тем же WHERE
- `GET /properties` в sqlite-режиме теперь читает из БД, а не из in-memory `props`.
- Добавлен тест `internal/http/server_properties_sql_test.go` на фильтры/сортировку.

Команды проверки:
- go test ./...
- API_ADDRESS=:8083 STORAGE=sqlite DB_PATH=./data/app.db PROPERTIES_PATH=./data/properties.json go run ./cmd/api
- curl "http://127.0.0.1:8083/properties?location=valencia"
- curl "http://127.0.0.1:8083/properties?min_price=400000"
- curl "http://127.0.0.1:8083/properties?min_bedrooms=4&sort=price_desc"

## 2026-01-01 — GET /properties: strict query validation

- Добавлена строгая валидация query-параметров в `GET /properties`:
  - limit: int > 0 (иначе 400 invalid_limit)
  - offset: int >= 0 (иначе 400 invalid_offset)
  - sort: только price_asc|price_desc (иначе 400 invalid_sort)
  - min_price/max_price: float >= 0 (иначе 400 invalid_min_price / invalid_max_price)
  - min_bedrooms: int >= 0 (иначе 400 invalid_min_bedrooms)
  - проверка min_price <= max_price (иначе 400 min_price_gt_max_price)
- Контракт успешных ответов не изменён, добавлены только 400 для мусорных входов.

Проверка:
- go test ./...
- curl -i "http://127.0.0.1:8083/properties?min_price=abc"
- curl -i "http://127.0.0.1:8083/properties?sort=bad"
- curl -i "http://127.0.0.1:8083/properties?limit=-1"
- curl -i "http://127.0.0.1:8083/properties?min_price=10&max_price=1"

