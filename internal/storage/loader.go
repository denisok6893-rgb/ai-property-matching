package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/denisok6893-rgb/ai-property-matching/internal/domain"
)

// LoadPropertiesFromFile reads properties from JSON file and returns a slice of Property.
func LoadPropertiesFromFile(path string) ([]domain.Property, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read properties file: %w", err)
	}

	var props []domain.Property
	if err := json.Unmarshal(b, &props); err != nil {
		return nil, fmt.Errorf("unmarshal properties: %w", err)
	}
	return props, nil
}
