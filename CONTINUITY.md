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
