package main

import (
	"log"
	"net/http"
	"os"

	"github.com/denisok6893-rgb/ai-property-matching/internal/domain"
	httpapi "github.com/denisok6893-rgb/ai-property-matching/internal/http"
	"github.com/denisok6893-rgb/ai-property-matching/internal/matching"
	"github.com/denisok6893-rgb/ai-property-matching/internal/storage"
)

type Config struct {
	Address        string
	PropertiesPath string
	WeightsPath    string
	Storage        string
	DBPath         string
}

func main() {
	cfg := loadConfig()

        var (
            props []domain.Property
            err   error
            store *storage.SQLiteStore
        )

	switch cfg.Storage {
	case "sqlite":

		store, err = storage.OpenSQLite(cfg.DBPath)
		if err != nil {
			log.Fatalf("open sqlite: %v", err)
		}

		if err := store.EnsureSchema(); err != nil {
			log.Fatalf("sqlite schema: %v", err)
		}

		n, err := store.CountProperties()
		if err != nil {
			log.Fatalf("sqlite count: %v", err)
		}

		if n == 0 {
			seed, err := storage.LoadPropertiesFromFile(cfg.PropertiesPath)
			if err != nil {
				log.Fatalf("seed load: %v", err)
			}
			if err := store.UpsertMany(seed); err != nil {
				log.Fatalf("sqlite seed upsert: %v", err)
			}
		}

		props, _, err = store.ListProperties(20, 0)
		if err != nil {
			log.Fatalf("sqlite list: %v", err)
		}

	default: // memory
		props, err = storage.LoadPropertiesFromFile(cfg.PropertiesPath)
		if err != nil {
			log.Fatalf("load properties: %v", err)
		}
	}

	w, err := matching.LoadWeightsFromFile(cfg.WeightsPath)
	if err != nil {
		log.Printf("use default weights (reason: %v)", err)
		w = matching.DefaultWeights()
	}

	engine := matching.NewEngine(w)
	srv := httpapi.NewServer(engine, props)
        if cfg.Storage == "sqlite" && store != nil {
            srv.PropsRepo = &httpapi.SQLitePropertiesRepo{Store: store}
        }

	log.Printf("API listening on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, srv.Routes()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func loadConfig() Config {
	return Config{
		Address:        getEnv("API_ADDRESS", ":8080"),
		PropertiesPath: getEnv("PROPERTIES_PATH", "data/properties.json"),
		WeightsPath:    getEnv("WEIGHTS_PATH", "configs/weights.json"),
		Storage:        getEnv("STORAGE", "memory"), // memory | sqlite
		DBPath:         getEnv("DB_PATH", "data/app.db"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
