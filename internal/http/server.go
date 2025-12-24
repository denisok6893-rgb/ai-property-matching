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
