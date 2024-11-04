package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req * http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(res, req)
	})
}

func metricsHandler(apiCfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		body := fmt.Sprintf("Hits: %d", apiCfg.fileServerHits.Load())
		res.Write([]byte(body))
	}
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
	mux.HandleFunc("GET /healthz", healthCheckHandler)
	mux.HandleFunc("GET /metrics", metricsHandler(&apiCfg))
	mux.Handle("POST /reset", apiCfg.middlewareMetricsReset(http.HandlerFunc(metricsHandler(&apiCfg))))

	fmt.Printf("Server listening on port %d\n", port)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error creating server: %s", err.Error())
		return
	}
}
