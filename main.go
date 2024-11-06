package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"github.com/ivportilla/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		fmt.Printf("Error opening connection to the db: %v", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)
	apiCfg := apiConfig{
		dbQueries: dbQueries,
	}
	mux := http.NewServeMux()
	port := 8080
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./")))))
	mux.HandleFunc("GET /api/healthz", healthCheckHandler)
	mux.HandleFunc("GET /admin/metrics", metricsHandler(&apiCfg))
	mux.Handle("POST /admin/reset", apiCfg.middlewareMetricsReset(http.HandlerFunc(metricsHandler(&apiCfg))))
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	fmt.Printf("Server listening on port %d\n", port)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error creating server: %s", err.Error())
		os.Exit(1)
	}
}
