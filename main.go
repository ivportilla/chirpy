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
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(res, req)
	})
}

func metricsHandler(apiCfg *apiConfig) func(http.ResponseWriter, *http.Request) {
	htmlResponse := `
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Header().Add("Content-Type", "text/html")
		body := fmt.Sprintf(htmlResponse, apiCfg.fileServerHits.Load())
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
	mux.HandleFunc("GET /api/healthz", healthCheckHandler)
	mux.HandleFunc("GET /admin/metrics", metricsHandler(&apiCfg))
	mux.Handle("POST /admin/reset", apiCfg.middlewareMetricsReset(http.HandlerFunc(metricsHandler(&apiCfg))))

	fmt.Printf("Server listening on port %d\n", port)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error creating server: %s", err.Error())
		return
	}
}
