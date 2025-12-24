# ai-property-matching
MVP AI Property Matching Platform (learning project)
# ai-property-matching

AI-powered property matching engine (MVP).

Проект предназначен для подбора объектов недвижимости под требования пользователя
на основе комбинации жёстких фильтров и скоринговой модели с объяснениями причин.

## Status

**MVP v0**

Проект:
- собирается
- запускается
- покрыт базовыми тестами
- зафиксирован в GitHub

Источник истины — репозиторий, не чат.

---

## Tech Stack

- Language: Go
- Runtime: Go 1.25.4
- Environment: Termux (Android)
- Storage: JSON (mock data)
- API: HTTP (net/http)

---

## Project Structure

.
├── cmd/api/ # Entry point (HTTP API)
├── internal/
│ ├── domain/ # Domain models
│ ├── matching/ # Matching engine
│ └── storage/ # Data loading layer
├── configs/
│ └── weights.json # Matching weights configuration
├── data/
│ └── properties.json # Mock property data
├── go.mod
└── README.md

---

## Matching Engine

### Hard filters
Объект отбрасывается, если не проходит:
- бюджет
- обязательные amenities

### Soft scoring
Для подходящих объектов считается score (0–100) с учётом:
- весов из `configs/weights.json`
- частичного соответствия параметрам

### Explainability
Каждый результат содержит:
- итоговый score
- список причин (`reasons`), объясняющих расчёт

---

## HTTP API

### Health check

GET /health

Response:
```json
{
  "status": "ok"
}

Property matching
POST /match

Request (пример):
{
  "budget": 5000000,
  "amenities": ["balcony", "parking"]
}

Response (пример):
{
  "matches": [
    {
      "property_id": "p1",
      "score": 87,
      "reasons": [
        "budget fits",
        "has parking",
        "partial amenities match"
      ]
    }
  ]
}

Run locally
go test ./...
go run ./cmd/api
По умолчанию сервер стартует локально и готов принимать HTTP-запросы.

Design Principles

Простота важнее абстракций

Объяснимость важнее «магии»

Каждое изменение — отдельный коммит

MVP сначала, расширение потом

Backward compatibility по возможности

Roadmap (high-level)

Pagination и список объектов недвижимости

CRUD для объектов (properties)

Улучшение matching-алгоритма

Внешний интерфейс (Web / Telegram)

Все шаги реализуются итеративно и фиксируются в GitHub.



