# ai-property-matching (MVP v0)

HTTP-сервис на Go для подбора объектов недвижимости под профиль клиента и базового управления объектами (in-memory).

## Требования
- Go

## Запуск

По умолчанию слушает `:8080`.

```bash
go run ./cmd/api

Переопределение адреса/портов и путей:
API_ADDRESS=:8083 PROPERTIES_PATH=data/properties.json WEIGHTS_PATH=configs/weights.json go run ./cmd/api

Health
curl -sS http://localhost:8080/health; echo

Ответ:

{"status":"ok"}

Properties API (in-memory)
Список (GET /properties)
curl -sS "http://localhost:8080/properties?limit=5&offset=0"; echo

Пример ответа:

{"limit":5,"offset":0,"total":1,"items":[{"id":"es-001","title":"Sunny family apartment in Valencia","location":"Valencia","price":320000,"bedrooms":3,"bathrooms":2,"area_sqm":110,"amenities":["balcony","storage","parking"]}]}

Создать (POST /properties)
curl -sS -X POST http://localhost:8080/properties \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test flat",
    "location": "Simferopol",
    "price": 6500000,
    "bedrooms": 2,
    "bathrooms": 1,
    "area_sqm": 54.5,
    "amenities": ["parking","elevator"],
    "features": {}
  }'; echo
Пример ответа (id генерируется как p-N):

{"id":"p-2","title":"Test flat","location":"Simferopol","price":6500000,"bedrooms":2,"bathrooms":1,"area_sqm":54.5,"amenities":["parking","elevator"],"features":{"quietness":0,"sun_exposure":0,"wind_protection":0,"tourism_intensity":0,"family_friendly":0,"expat_friendly":0,"investment_potential":0,"distance_to_sea_km":0,"walkability":0,"green_areas":0}}

Получить по id (GET /properties/{id})
curl -sS http://localhost:8080/properties/es-001; echo

Удалить (DELETE /properties/{id})
curl -sS -X DELETE http://localhost:8080/properties/p-2; echo
Ответ:

{"status":"deleted"}


Проверка:

curl -sS http://localhost:8080/properties/p-2; echo


Ответ:

{"error":"not_found"}
Match API

POST /match

curl -sS -X POST http://localhost:8080/match \
  -H "Content-Type: application/json" \
  -d '{
    "profile": {},
    "limit": 5
  }'; echo

Тесты
go test ./...

Данные и конфигурация

data/properties.json — мок-объекты

configs/weights.json — веса факторов скоринга
