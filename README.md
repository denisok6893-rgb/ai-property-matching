h# ai-property-matching (MVP v0)

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

## Match API

`POST /match`

Пример запроса (реальный профиль):

```bash
curl -sS -X POST http://localhost:8080/match \
  -H "Content-Type: application/json" \
  -d '{
    "profile": {
      "name": "Denis test",
      "location_preference": "Valencia",
      "budget_min": 200000,
      "budget_max": 400000,
      "desired_bedrooms": 3,
      "desired_bathrooms": 2,
      "priorities": {
        "quietness": 0.3,
        "sun_exposure": 0.2,
        "wind_protection": 0.0,
        "low_tourism": 0.1,
        "family_friendliness": 0.1,
        "expat_community": 0.0,
        "investment_focus": 0.1,
        "walkability": 0.1,
        "green_areas": 0.1,
        "sea_proximity": 0.0
      },
      "hard_filters": {
        "must_have_amenities": ["parking"]
      }
    },
    "limit": 5
  }'; echo

Пример ответа:

{"results":[{"property":{"id":"es-001","title":"Sunny family apartment in Valencia","location":"Valencia","price":320000,"bedrooms":3,"bathrooms":2,"area_sqm":110,"amenities":["balcony","storage","parking"],"features":{"quietness":0.6,"sun_exposure":0.85,"wind_protection":0.55,"tourism_intensity":0.3,"family_friendly":0.82,"expat_friendly":0.7,"investment_potential":0.62,"distance_to_sea_km":1.5,"walkability":0.8,"green_areas":0.6}},"score":70.1,"reasons":[{"type":"quietness","message":"quietness: good","impact":1},{"type":"sun_exposure","message":"sun exposure: strong match","impact":0.85},{"type":"family_friendliness","message":"family friendly: strong match","impact":0.41},{"type":"low_tourism","message":"low tourism: good","impact":0.39},{"type":"walkability","message":"walkability: strong match","impact":0.31},{"type":"investment_focus","message":"investment potential: good","impact":0.29},{"type":"green_areas","message":"green areas: good","impact":0.2}]}]}

Тесты
go test ./...

Данные и конфигурация

data/properties.json — мок-объекты

configs/weights.json — веса факторов скоринга
