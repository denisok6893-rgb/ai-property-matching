package httpapi

import (
	"context"
	"strconv"

	"github.com/denisok6893-rgb/ai-property-matching/internal/storage"
)

type SQLitePropertiesRepo struct {
	Store *storage.SQLiteStore
}

func (r *SQLitePropertiesRepo) List(ctx context.Context, p ListParams) ([]PropertySummary, int) {
	if r == nil || r.Store == nil {
		return nil, 0
	}

	minPrice, _ := strconv.ParseFloat(p.MinPrice, 64)
	maxPrice, _ := strconv.ParseFloat(p.MaxPrice, 64)
	minBedrooms, _ := strconv.Atoi(p.MinBedrooms)

	props, total, err := r.Store.ListPropertiesFiltered(
		p.Limit,
		p.Offset,
		p.Location,
		minPrice,
		maxPrice,
		minBedrooms,
		p.Sort,
	)
	if err != nil {
		// Контракт сейчас не возвращает 500 на ошибки репозитория (у нас нет ошибок в сигнатуре).
		// Чтобы не менять API/handler в этом шаге — просто "пусто".
		return nil, 0
	}

	out := make([]PropertySummary, 0, len(props))
	for _, prop := range props {
		out = append(out, PropertySummary{
			ID:        prop.ID,
			Title:     prop.Title,
			Location:  prop.Location,
			Price:     prop.Price,
			Bedrooms:  prop.Bedrooms,
			Bathrooms: prop.Bathrooms,
			AreaSQM:   prop.AreaSQM,
			Amenities: prop.Amenities,
		})
	}
	return out, total
}
