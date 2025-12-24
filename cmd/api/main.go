package main

import (
	"log"
	"net/http"
	"os"

	httpapi "github.com/denisok6893-rgb/ai-property-matching/internal/http"
	"github.com/denisok6893-rgb/ai-property-matching/internal/matching"
	"github.com/denisok6893-rgb/ai-property-matching/internal/storage"
)

type Config struct {
	Address        string
	PropertiesPath string
	WeightsPath    string
}

func main() {
	cfg := loadConfig()

	props, err := storage.LoadPropertiesFromFile(cfg.PropertiesPath)
	if err != nil {
		log.Fatalf("load properties: %v", err)
	}

	w, err := matching.LoadWeightsFromFile(cfg.WeightsPath)
	if err != nil {
		log.Printf("use default weights (reason: %v)", err)
		w = matching.DefaultWeights()
	}

	engine := matching.NewEngine(w)
	srv := httpapi.NewServer(engine, props)

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
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
