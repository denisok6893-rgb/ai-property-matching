package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/denisok6893-rgb/ai-property-matching/internal/domain"
	"github.com/denisok6893-rgb/ai-property-matching/internal/matching"
)

type Server struct {
	Engine     *matching.Engine
	Properties []domain.Property
}

func NewServer(engine *matching.Engine, properties []domain.Property) *Server {
	return &Server{Engine: engine, Properties: properties}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/match", s.handleMatch)
	mux.HandleFunc("/demo", s.handleDemo)
	mux.HandleFunc("/properties", s.handlePropertiesList)
	mux.HandleFunc("/properties/", s.handlePropertiesGetByID)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type MatchRequest struct {
	Profile domain.ClientProfile `json:"profile"`
	Limit   int                  `json:"limit"`
}

type MatchResponse struct {
	Results []domain.ScoreResult `json:"results"`
}

func (s *Server) handleMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	limit := req.Limit
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}
	if limit <= 0 {
		limit = 5
	}

	results := s.Engine.ScoreProperties(req.Profile, s.Properties, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(MatchResponse{Results: results}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// ---- Properties API (read-only v1) ----

type PropertySummary struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Location  string  `json:"location"`
	Price     float64 `json:"price"`
	Bedrooms  int     `json:"bedrooms"`
	Bathrooms int     `json:"bathrooms"`
	AreaSQM   float64 `json:"area_sqm"`
	Amenities []string `json:"amenities,omitempty"`
}

type PropertiesListResponse struct {
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
	Total  int               `json:"total"`
	Items  []PropertySummary `json:"items"`
}

func (s *Server) handlePropertiesList(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.handlePropertiesCreate(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit, offset := parseLimitOffset(r, 20, 0)

	total := len(s.Properties)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	items := make([]PropertySummary, 0, end-offset)
	for _, p := range s.Properties[offset:end] {
		items = append(items, PropertySummary{
			ID:        p.ID,
			Title:     p.Title,
			Location:  p.Location,
			Price:     p.Price,
			Bedrooms:  p.Bedrooms,
			Bathrooms: p.Bathrooms,
			AreaSQM:   p.AreaSQM,
			Amenities: p.Amenities,
		})
	}

	writeJSON(w, http.StatusOK, PropertiesListResponse{
		Limit:  limit,
		Offset: offset,
		Total:  total,
		Items:  items,
	})
}

func (s *Server) handlePropertiesGetByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/properties/"):]
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		for _, p := range s.Properties {
			if p.ID == id {
				writeJSON(w, http.StatusOK, p)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return

	case http.MethodDelete:
		for i, p := range s.Properties {
			if p.ID == id {
				// remove element by index
				s.Properties = append(s.Properties[:i], s.Properties[i+1:]...)
				writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

type CreatePropertyRequest struct {
	Title     string          `json:"title"`
	Location  string          `json:"location"`
	Price     float64         `json:"price"`
	Bedrooms  int             `json:"bedrooms"`
	Bathrooms int             `json:"bathrooms"`
	AreaSQM   float64         `json:"area_sqm"`
       	Description string   `json:"description"`
	ImageURLs   []string `json:"image_urls"`
	Amenities []string        `json:"amenities"`
	Features  domain.Features `json:"features"`
}

func (s *Server) handlePropertiesCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// minimal validation
	if req.Title == "" || req.Location == "" {
		http.Error(w, "title and location are required", http.StatusBadRequest)
		return
	}
	if req.Price <= 0 {
		http.Error(w, "price must be > 0", http.StatusBadRequest)
		return
	}

	id := "p-" + strconv.FormatInt(int64(len(s.Properties)+1), 10)

	p := domain.Property{
		ID:        id,
		Title:     req.Title,
		Location:  req.Location,
		Price:     req.Price,
		Bedrooms:  req.Bedrooms,
		Bathrooms: req.Bathrooms,
		AreaSQM:   req.AreaSQM,
		Description: req.Description,
                ImageURLs:   req.ImageURLs,
                Amenities: req.Amenities,
		Features:  req.Features,
	}

	s.Properties = append(s.Properties, p)
	writeJSON(w, http.StatusCreated, p)
}


func parseLimitOffset(r *http.Request, defLimit, defOffset int) (int, int) {
	q := r.URL.Query()

	limit := defLimit
	if v := q.Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}
	if limit <= 0 {
		limit = defLimit
	}
	// safety cap
	if limit > 200 {
		limit = 200
	}

	offset := defOffset
	if v := q.Get("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			offset = parsed
		}
	}
	if offset < 0 {
		offset = defOffset
	}

	return limit, offset
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) handleDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	html := `<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>ai-property-matching — demo</title>
  <style>
    body { font-family: system-ui, -apple-system, Segoe UI, Roboto, Arial, sans-serif; margin: 16px; }
    textarea { width: 100%; min-height: 220px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
    button { padding: 10px 14px; font-size: 16px; }
    pre { white-space: pre-wrap; word-wrap: break-word; background: #f6f6f6; padding: 12px; border-radius: 10px; }
    .grid { display: grid; gap: 12px; }
    .cols { display: grid; gap: 12px; grid-template-columns: 1fr; }
    @media (min-width: 900px) { .cols { grid-template-columns: 1fr 1fr; } }
    .card { border: 1px solid #e6e6e6; border-radius: 12px; padding: 12px; }
    .list { display: grid; gap: 10px; }
    .item { border: 1px solid #eaeaea; border-radius: 12px; padding: 10px; cursor: pointer; }
    .item:hover { background: #fafafa; }
    .muted { color: #666; font-size: 14px; }
    .imgs { display: grid; gap: 10px; grid-template-columns: 1fr; }
    @media (min-width: 700px) { .imgs { grid-template-columns: 1fr 1fr; } }
    img { width: 100%; height: auto; border-radius: 12px; border: 1px solid #eee; }
    .row { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; }
    input { padding: 10px; font-size: 16px; width: 220px; }
  </style>
</head>
<body>
  <h2>ai-property-matching — demo</h2>
  <div class="muted">Сервер: <code>` + r.Host + `</code></div>

  <div class="cols" style="margin-top:12px;">
    <div class="grid">
      <div class="card">
        <div class="row">
          <button id="btnProps">List properties</button>
          <input id="idInput" placeholder="property id (например es-001)"/>
          <button id="btnDetails">Details</button>
        </div>
        <div class="muted" style="margin-top:8px;">Кликни по объекту в списке — откроются детали справа.</div>
      </div>

      <div class="card">
        <div><b>Запрос (JSON) → POST /match</b></div>
        <textarea id="payload"></textarea>
        <div style="margin-top:10px;">
          <button id="btnMatch">Match</button>
        </div>
      </div>

      <div class="card">
        <div><b>Список объектов (GET /properties)</b></div>
        <div id="list" class="list" style="margin-top:10px;">Нажми List properties…</div>
      </div>
    </div>

    <div class="grid">
      <div class="card">
        <div><b>Детали объекта (GET /properties/{id})</b></div>
        <div id="details" class="muted" style="margin-top:10px;">Выбери объект…</div>
        <div id="images" class="imgs" style="margin-top:10px;"></div>
      </div>

      <div class="card">
        <div><b>Ответ /match</b></div>
        <pre id="out">Нажми Match…</pre>
      </div>
    </div>
  </div>

<script>
const defaultPayload = {
  profile: {
    name: "Demo",
    location_preference: "Valencia",
    budget_min: 200000,
    budget_max: 400000,
    desired_bedrooms: 3,
    desired_bathrooms: 2,
    priorities: {
      quietness: 0.3,
      sun_exposure: 0.2,
      wind_protection: 0.0,
      low_tourism: 0.1,
      family_friendliness: 0.1,
      expat_community: 0.0,
      investment_focus: 0.1,
      walkability: 0.1,
      green_areas: 0.1,
      sea_proximity: 0.0
    },
    hard_filters: { must_have_amenities: ["parking"] }
  },
  limit: 5
};

const ta = document.getElementById("payload");
const out = document.getElementById("out");
const listEl = document.getElementById("list");
const detailsEl = document.getElementById("details");
const imagesEl = document.getElementById("images");
const idInput = document.getElementById("idInput");

ta.value = JSON.stringify(defaultPayload, null, 2);

function money(n) {
  if (typeof n !== "number") return n;
  return n.toLocaleString("ru-RU");
}

function renderImages(urls) {
  imagesEl.innerHTML = "";
  if (!Array.isArray(urls) || urls.length === 0) return;
  for (const u of urls) {
    const img = document.createElement("img");
    img.src = u;
    img.alt = "property image";
    imagesEl.appendChild(img);
  }
}

function renderDetails(p) {
  if (!p) {
    detailsEl.textContent = "Не найдено";
    imagesEl.innerHTML = "";
    return;
  }
  const amenities = Array.isArray(p.amenities) ? p.amenities.join(", ") : "";
  detailsEl.innerHTML =
    "<div><b>" + (p.title || "") + "</b></div>" +
    "<div class='muted'>ID: <code>" + (p.id || "") + "</code></div>" +
    "<div style='margin-top:8px;'>Локация: <b>" + (p.location || "") + "</b></div>" +
    "<div>Цена: <b>" + money(p.price) + "</b></div>" +
    "<div>Комнат: <b>" + (p.bedrooms ?? "") + "</b>, санузлов: <b>" + (p.bathrooms ?? "") + "</b>, площадь: <b>" + (p.area_sqm ?? "") + " м²</b></div>" +
    (p.description ? "<div style='margin-top:8px;'>" + p.description + "</div>" : "") +
    (amenities ? "<div style='margin-top:8px;'><span class='muted'>Amenities:</span> " + amenities + "</div>" : "");
  renderImages(p.image_urls);
}

async function loadProperties() {
  listEl.textContent = "Загрузка...";
  try {
    const res = await fetch("/properties?limit=50&offset=0");
    const data = await res.json();
    const items = (data && data.items) ? data.items : [];
    listEl.innerHTML = "";
    if (items.length === 0) {
      listEl.textContent = "Пусто";
      return;
    }
    for (const it of items) {
      const div = document.createElement("div");
      div.className = "item";
      div.innerHTML =
        "<div><b>" + (it.title || "") + "</b></div>" +
        "<div class='muted'>ID: <code>" + (it.id || "") + "</code> • " + (it.location || "") + "</div>" +
        "<div class='muted'>Цена: " + money(it.price) + " • " + (it.bedrooms ?? "") + " bd • " + (it.bathrooms ?? "") + " ba • " + (it.area_sqm ?? "") + " м²</div>";
      div.addEventListener("click", async () => {
        idInput.value = it.id || "";
        await loadDetails(it.id);
      });
      listEl.appendChild(div);
    }
  } catch (e) {
    listEl.textContent = "Ошибка: " + e.message;
  }
}

async function loadDetails(id) {
  if (!id) return;
  detailsEl.textContent = "Загрузка...";
  imagesEl.innerHTML = "";
  try {
    const res = await fetch("/properties/" + encodeURIComponent(id));
    const text = await res.text();
    try {
      const obj = JSON.parse(text);
      if (obj && obj.error) {
        renderDetails(null);
        return;
      }
      renderDetails(obj);
    } catch {
      detailsEl.textContent = text;
    }
  } catch (e) {
    detailsEl.textContent = "Ошибка: " + e.message;
  }
}

document.getElementById("btnProps").addEventListener("click", loadProperties);

document.getElementById("btnDetails").addEventListener("click", async () => {
  const id = idInput.value.trim();
  await loadDetails(id);
});

document.getElementById("btnMatch").addEventListener("click", async () => {
  out.textContent = "Запрос...";
  let payload;
  try { payload = JSON.parse(ta.value); } catch(e) {
    out.textContent = "Ошибка JSON: " + e.message;
    return;
  }
  try {
    const res = await fetch("/match", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(payload)
    });
    const text = await res.text();
    out.textContent = text;
  } catch (e) {
    out.textContent = "Ошибка запроса: " + e.message;
  }
});

// Auto-load list on open
loadProperties();
</script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}
