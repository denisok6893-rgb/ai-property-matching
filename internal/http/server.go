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
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Path looks like: /properties/{id}
	id := r.URL.Path[len("/properties/"):]
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_id"})
		return
	}

	for _, p := range s.Properties {
		if p.ID == id {
			writeJSON(w, http.StatusOK, p)
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
}
type CreatePropertyRequest struct {
	Title     string          `json:"title"`
	Location  string          `json:"location"`
	Price     float64         `json:"price"`
	Bedrooms  int             `json:"bedrooms"`
	Bathrooms int             `json:"bathrooms"`
	AreaSQM   float64         `json:"area_sqm"`
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
