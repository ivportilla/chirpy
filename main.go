package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	apiCfg := apiConfig{}
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
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error creating server: %s", err.Error())
		return
	}
}
