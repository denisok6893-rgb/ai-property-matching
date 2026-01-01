package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
        "strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/denisok6893-rgb/ai-property-matching/internal/domain"
)

type SQLiteStore struct {
	db *sql.DB
}

func OpenSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// базовые настройки
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) EnsureSchema() error {
	const createTable = `
CREATE TABLE IF NOT EXISTS properties (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  location TEXT NOT NULL,
  price REAL NOT NULL,
  bedrooms INTEGER NOT NULL,
  bathrooms INTEGER NOT NULL,
  area_sqm REAL NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  image_urls_json TEXT NOT NULL DEFAULT '[]',
  amenities_json TEXT NOT NULL DEFAULT '[]',
  features_json TEXT NOT NULL DEFAULT '{}'
);
`
	if _, err := s.db.Exec(createTable); err != nil {
		return err
	}

	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_properties_location ON properties(location);`); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_properties_price ON properties(price);`); err != nil {
		return err
	}

	return nil
}

func (s *SQLiteStore) CountProperties() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM properties`).Scan(&n)
	return n, err
}

// UpsertMany inserts initial dataset without duplicating by id.
func (s *SQLiteStore) UpsertMany(items []domain.Property) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
INSERT OR IGNORE INTO properties
(id, title, location, price, bedrooms, bathrooms, area_sqm, description, image_urls_json, amenities_json, features_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range items {
		img, _ := json.Marshal(p.ImageURLs)
		am, _ := json.Marshal(p.Amenities)
		ft, _ := json.Marshal(p.Features)

		if _, err := stmt.Exec(
			p.ID, p.Title, p.Location, p.Price, p.Bedrooms, p.Bathrooms, p.AreaSQM,
			p.Description, string(img), string(am), string(ft),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) CreateProperty(p domain.Property) (domain.Property, error) {
	if p.ID == "" {
		p.ID = fmt.Sprintf("p-%d", time.Now().UnixNano())
	}
	img, _ := json.Marshal(p.ImageURLs)
	am, _ := json.Marshal(p.Amenities)
	ft, _ := json.Marshal(p.Features)

	_, err := s.db.Exec(`
INSERT INTO properties
(id, title, location, price, bedrooms, bathrooms, area_sqm, description, image_urls_json, amenities_json, features_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		p.ID, p.Title, p.Location, p.Price, p.Bedrooms, p.Bathrooms, p.AreaSQM,
		p.Description, string(img), string(am), string(ft),
	)
	return p, err
}

func (s *SQLiteStore) DeleteProperty(id string) (bool, error) {
	res, err := s.db.Exec(`DELETE FROM properties WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	aff, _ := res.RowsAffected()
	return aff > 0, nil
}

func (s *SQLiteStore) GetProperty(id string) (domain.Property, bool, error) {
	var p domain.Property
	var imgJSON, amJSON, ftJSON string

	err := s.db.QueryRow(`
SELECT id, title, location, price, bedrooms, bathrooms, area_sqm, description, image_urls_json, amenities_json, features_json
FROM properties WHERE id = ?
`, id).Scan(
		&p.ID, &p.Title, &p.Location, &p.Price, &p.Bedrooms, &p.Bathrooms, &p.AreaSQM,
		&p.Description, &imgJSON, &amJSON, &ftJSON,
	)
	if err == sql.ErrNoRows {
		return domain.Property{}, false, nil
	}
	if err != nil {
		return domain.Property{}, false, err
	}

	_ = json.Unmarshal([]byte(imgJSON), &p.ImageURLs)
	_ = json.Unmarshal([]byte(amJSON), &p.Amenities)
	_ = json.Unmarshal([]byte(ftJSON), &p.Features)

	return p, true, nil
}

func (s *SQLiteStore) ListProperties(limit, offset int) ([]domain.Property, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	total, err := s.CountProperties()
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(`
SELECT id, title, location, price, bedrooms, bathrooms, area_sqm, description, image_urls_json, amenities_json, features_json
FROM properties
ORDER BY id
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.Property
	for rows.Next() {
		var p domain.Property
		var imgJSON, amJSON, ftJSON string

		if err := rows.Scan(
			&p.ID, &p.Title, &p.Location, &p.Price, &p.Bedrooms, &p.Bathrooms, &p.AreaSQM,
			&p.Description, &imgJSON, &amJSON, &ftJSON,
		); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal([]byte(imgJSON), &p.ImageURLs)
		_ = json.Unmarshal([]byte(amJSON), &p.Amenities)
		_ = json.Unmarshal([]byte(ftJSON), &p.Features)

		out = append(out, p)
	}
	return out, total, rows.Err()
}

func (s *SQLiteStore) ListPropertiesFiltered(
	limit, offset int,
	location string,
	minPrice, maxPrice float64,
	minBedrooms int,
	sortBy string,
) ([]domain.Property, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// WHERE builder
	where := make([]string, 0, 4)
	args := make([]any, 0, 8)

	if strings.TrimSpace(location) != "" {
		// contains, case-insensitive
		where = append(where, "LOWER(location) LIKE '%' || LOWER(?) || '%'")
		args = append(args, location)
	}
	if minPrice > 0 {
		where = append(where, "price >= ?")
		args = append(args, minPrice)
	}
	if maxPrice > 0 {
		where = append(where, "price <= ?")
		args = append(args, maxPrice)
	}
	if minBedrooms > 0 {
		where = append(where, "bedrooms >= ?")
		args = append(args, minBedrooms)
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	orderSQL := "ORDER BY id"
	switch sortBy {
	case "price_asc":
		orderSQL = "ORDER BY price ASC"
	case "price_desc":
		orderSQL = "ORDER BY price DESC"
	}

	// total count with same WHERE
	countSQL := "SELECT COUNT(*) FROM properties " + whereSQL
	var total int
	if err := s.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// rows
	rowsSQL := `
SELECT id, title, location, price, bedrooms, bathrooms, area_sqm, description, image_urls_json, amenities_json, features_json
FROM properties
` + whereSQL + "\n" + orderSQL + "\nLIMIT ? OFFSET ?"

	rowsArgs := append(append([]any{}, args...), limit, offset)

	rows, err := s.db.Query(rowsSQL, rowsArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.Property
	for rows.Next() {
		var p domain.Property
		var imgJSON, amJSON, ftJSON string
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Location, &p.Price, &p.Bedrooms, &p.Bathrooms, &p.AreaSQM,
			&p.Description, &imgJSON, &amJSON, &ftJSON,
		); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal([]byte(imgJSON), &p.ImageURLs)
		_ = json.Unmarshal([]byte(amJSON), &p.Amenities)
		_ = json.Unmarshal([]byte(ftJSON), &p.Features)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}
