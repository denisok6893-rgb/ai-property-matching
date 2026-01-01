package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGETProperties_FiltersAndSort(t *testing.T) {
	t.Parallel()

	// Engine для /properties не нужен — можно nil.
	srv := NewServer(nil, nil)
	ts := httptest.NewServer(srv.Routes())
	defer ts.Close()

	type createReq struct {
		Title     string   `json:"title"`
		Location  string   `json:"location"`
		Price     float64  `json:"price"`
		Bedrooms  int      `json:"bedrooms"`
		Bathrooms int      `json:"bathrooms"`
		AreaSqm   float64  `json:"area_sqm"`
		Amenities []string `json:"amenities"`
	}

	post := func(r createReq) {
		b, _ := json.Marshal(r)
		resp, err := http.Post(ts.URL+"/properties", "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("POST /properties: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("POST /properties status=%d", resp.StatusCode)
		}
	}

	// 3 объекта
	post(createReq{Title: "A", Location: "Valencia", Price: 320000, Bedrooms: 3, Bathrooms: 2, AreaSqm: 110, Amenities: []string{"balcony"}})
	post(createReq{Title: "B", Location: "valencia center", Price: 450000, Bedrooms: 4, Bathrooms: 2, AreaSqm: 140, Amenities: []string{"parking"}})
	post(createReq{Title: "C", Location: "Madrid", Price: 500000, Bedrooms: 4, Bathrooms: 3, AreaSqm: 160, Amenities: []string{"storage"}})

	// location contains (case-insensitive), min_price, min_bedrooms, sort desc
	resp, err := http.Get(ts.URL + "/properties?location=VALENCIA&min_price=400000&min_bedrooms=4&sort=price_desc&limit=20&offset=0")
	if err != nil {
		t.Fatalf("GET /properties: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /properties status=%d", resp.StatusCode)
	}

	var got struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
		Total  int `json:"total"`
		Items  []struct {
			ID       string  `json:"id"`
			Title    string  `json:"title"`
			Location string  `json:"location"`
			Price    float64 `json:"price"`
			Bedrooms int     `json:"bedrooms"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got.Total != 1 {
		t.Fatalf("total=%d want=1", got.Total)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items=%d want=1", len(got.Items))
	}
	if got.Items[0].Title != "B" {
		t.Fatalf("first title=%q want=%q", got.Items[0].Title, "B")
	}
}
