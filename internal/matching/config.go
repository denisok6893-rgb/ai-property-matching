package matching

import (
	"encoding/json"
	"fmt"
	"os"
)

// Weights defines coefficients for each preference factor.
type Weights struct {
	Quietness          float64 `json:"quietness"`
	SunExposure        float64 `json:"sun_exposure"`
	WindProtection     float64 `json:"wind_protection"`
	LowTourism         float64 `json:"low_tourism"`
	FamilyFriendliness float64 `json:"family_friendliness"`
	ExpatCommunity     float64 `json:"expat_community"`
	InvestmentFocus    float64 `json:"investment_focus"`
	Walkability        float64 `json:"walkability"`
	GreenAreas         float64 `json:"green_areas"`
	SeaProximity       float64 `json:"sea_proximity"`
}

// DefaultWeights returns a reasonable baseline for MVP.
func DefaultWeights() Weights {
	return Weights{
		Quietness:          1.0,
		SunExposure:        0.9,
		WindProtection:     0.6,
		LowTourism:         1.0,
		FamilyFriendliness: 0.9,
		ExpatCommunity:     0.7,
		InvestmentFocus:    0.85,
		Walkability:        0.7,
		GreenAreas:         0.6,
		SeaProximity:       0.8,
	}
}

// LoadWeightsFromFile loads weights from JSON file, falling back to defaults on file read errors.
func LoadWeightsFromFile(path string) (Weights, error) {
	w := DefaultWeights()
	b, err := os.ReadFile(path)
	if err != nil {
		return w, fmt.Errorf("read weights file: %w", err)
	}
	if err := json.Unmarshal(b, &w); err != nil {
		return w, fmt.Errorf("unmarshal weights: %w", err)
	}
	return w, nil
}
