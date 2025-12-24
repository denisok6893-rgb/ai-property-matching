package matching

import (
	"math"
	"sort"
	"strings"

	"github.com/denisok6893-rgb/ai-property-matching/internal/domain"
)

type Engine struct {
	weights Weights
}

func NewEngine(w Weights) *Engine {
	return &Engine{weights: w}
}

// ScoreProperties applies hard filters, computes score (0..100), and returns top results.
func (e *Engine) ScoreProperties(profile domain.ClientProfile, properties []domain.Property, limit int) []domain.ScoreResult {
	type scored struct {
		res    domain.ScoreResult
		reason []domain.ScoreReason
	}
	var out []domain.ScoreResult

	for _, p := range properties {
		if !passesHardFilters(profile, p) {
			continue
		}
		score, reasons := e.scoreOne(profile, p)
		out = append(out, domain.ScoreResult{
			Property: p,
			Score:    score,
			Reasons:  reasons,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if limit <= 0 {
		limit = 5
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func passesHardFilters(profile domain.ClientProfile, p domain.Property) bool {
	// Budget hard filter (if set)
	if profile.BudgetMin > 0 && p.Price < profile.BudgetMin {
		return false
	}
	if profile.BudgetMax > 0 && p.Price > profile.BudgetMax {
		return false
	}
	// Required amenities
	if len(profile.HardFilters.MustHaveAmenities) > 0 {
		have := make(map[string]struct{}, len(p.Amenities))
		for _, a := range p.Amenities {
			have[strings.ToLower(strings.TrimSpace(a))] = struct{}{}
		}
		for _, req := range profile.HardFilters.MustHaveAmenities {
			r := strings.ToLower(strings.TrimSpace(req))
			if r == "" {
				continue
			}
			if _, ok := have[r]; !ok {
				return false
			}
		}
	}
	return true
}

func (e *Engine) scoreOne(profile domain.ClientProfile, p domain.Property) (float64, []domain.ScoreReason) {
	// Soft factors: we compute weighted contributions in 0..1, then scale to 0..100.
	type factor struct {
		key      string
		label    string
		wantHigh bool // true: higher is better; false: lower is better
		weight   float64
		pref     float64
		value    float64
	}

	factors := []factor{
		{"quietness", "quietness", true, e.weights.Quietness, profile.Priorities.Quietness, clamp01(p.Features.Quietness)},
		{"sun_exposure", "sun exposure", true, e.weights.SunExposure, profile.Priorities.SunExposure, clamp01(p.Features.SunExposure)},
		{"wind_protection", "wind protection", true, e.weights.WindProtection, profile.Priorities.WindProtection, clamp01(p.Features.WindProtection)},
		{"low_tourism", "low tourism", false, e.weights.LowTourism, profile.Priorities.LowTourism, clamp01(p.Features.TourismIntensity)},
		{"family_friendliness", "family friendly", true, e.weights.FamilyFriendliness, profile.Priorities.FamilyFriendliness, clamp01(p.Features.FamilyFriendly)},
		{"expat_community", "expat friendly", true, e.weights.ExpatCommunity, profile.Priorities.ExpatCommunity, clamp01(p.Features.ExpatFriendly)},
		{"investment_focus", "investment potential", true, e.weights.InvestmentFocus, profile.Priorities.InvestmentFocus, clamp01(p.Features.InvestmentPotential)},
		{"walkability", "walkability", true, e.weights.Walkability, profile.Priorities.Walkability, clamp01(p.Features.Walkability)},
		{"green_areas", "green areas", true, e.weights.GreenAreas, profile.Priorities.GreenAreas, clamp01(p.Features.GreenAreas)},
		{"sea_proximity", "sea proximity", true, e.weights.SeaProximity, profile.Priorities.SeaProximity, seaProximity01(p.Features.DistanceToSeaKm)},
	}

	var sumW, sum float64
	var contributions []domain.ScoreReason

	for _, f := range factors {
		// If client doesn't care about this factor, skip it.
		if f.pref <= 0 || f.weight <= 0 {
			continue
		}
		w := f.pref * f.weight
		sumW += w

		v := f.value
		if !f.wantHigh {
			// For "low tourism", lower tourism_intensity is better: invert.
			v = 1 - v
		}
		contrib := w * v
		sum += contrib

		contributions = append(contributions, domain.ScoreReason{
			Type:    f.key,
			Message: reasonMessage(f.label, v),
			Impact:  contrib,
		})
	}

	// Soft nudge: budget closeness to max (if set). Adds up to 0.05 of total.
	if profile.BudgetMax > 0 && sumW > 0 {
		close01 := budgetCloseness01(p.Price, profile.BudgetMax)
		w := 0.05 * sumW
		sumW += w
		sum += w * close01
		contributions = append(contributions, domain.ScoreReason{
			Type:    "budget_closeness",
			Message: reasonMessage("budget closeness", close01),
			Impact:  w * close01,
		})
	}

	// If no weights are active, score is neutral 50.
	if sumW <= 0 {
		return 50.0, topReasons(contributions, 5)
	}

	score01 := sum / sumW
	score := math.Round(score01*1000) / 10 // 0.1 precision

	return clamp(score, 0, 100), topReasons(contributions, 7)
}

func topReasons(reasons []domain.ScoreReason, max int) []domain.ScoreReason {
	sort.Slice(reasons, func(i, j int) bool { return reasons[i].Impact > reasons[j].Impact })
	if max <= 0 {
		max = 5
	}
	if len(reasons) > max {
		reasons = reasons[:max]
	}
	// Normalize Impact to share of the best item (0..1) for readability.
	if len(reasons) == 0 {
		return reasons
	}
	best := reasons[0].Impact
	if best <= 0 {
		return reasons
	}
	for i := range reasons {
		reasons[i].Impact = math.Round((reasons[i].Impact/best)*100) / 100
	}
	return reasons
}

func reasonMessage(label string, v float64) string {
	switch {
	case v >= 0.8:
		return label + ": strong match"
	case v >= 0.6:
		return label + ": good"
	case v >= 0.4:
		return label + ": mixed"
	default:
		return label + ": weak"
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func seaProximity01(distanceKm float64) float64 {
	// Convert distance to 0..1 where 0km => 1.0, 10km+ => near 0.
	if distanceKm <= 0 {
		return 1
	}
	// Simple decay curve.
	v := 1 / (1 + distanceKm/2)
	return clamp01(v)
}

func budgetCloseness01(price, budgetMax float64) float64 {
	if budgetMax <= 0 {
		return 0.5
	}
	// Prefer being comfortably within budget; closer to max reduces closeness.
	r := price / budgetMax
	// r <= 0.7 => 1.0; r >= 1.0 => 0.0
	v := (1.0 - (r-0.7)/0.3)
	return clamp01(v)
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}
