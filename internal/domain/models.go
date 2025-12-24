package domain

type ClientProfile struct {
	Name               string            `json:"name"`
	LocationPreference string            `json:"location_preference"`
	BudgetMin          float64           `json:"budget_min"`
	BudgetMax          float64           `json:"budget_max"`
	DesiredBedrooms    int               `json:"desired_bedrooms"`
	DesiredBathrooms   int               `json:"desired_bathrooms"`
	Priorities         PreferenceWeights `json:"priorities"`
	HardFilters        HardFilters       `json:"hard_filters"`
}

type HardFilters struct {
	MustHaveAmenities []string `json:"must_have_amenities"`
}

type PreferenceWeights struct {
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

type Property struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Location  string   `json:"location"`
	Price     float64  `json:"price"`
	Bedrooms  int      `json:"bedrooms"`
	Bathrooms int      `json:"bathrooms"`
	AreaSQM   float64  `json:"area_sqm"`
	Amenities []string `json:"amenities"`
	Features  Features `json:"features"`
}

type Features struct {
	Quietness           float64 `json:"quietness"`
	SunExposure         float64 `json:"sun_exposure"`
	WindProtection      float64 `json:"wind_protection"`
	TourismIntensity    float64 `json:"tourism_intensity"`
	FamilyFriendly      float64 `json:"family_friendly"`
	ExpatFriendly       float64 `json:"expat_friendly"`
	InvestmentPotential float64 `json:"investment_potential"`
	DistanceToSeaKm     float64 `json:"distance_to_sea_km"`
	Walkability         float64 `json:"walkability"`
	GreenAreas          float64 `json:"green_areas"`
}

type ScoreResult struct {
	Property Property      `json:"property"`
	Score    float64       `json:"score"`
	Reasons  []ScoreReason `json:"reasons"`
}

type ScoreReason struct {
	Type    string  `json:"type"`
	Message string  `json:"message"`
	Impact  float64 `json:"impact"`
}
