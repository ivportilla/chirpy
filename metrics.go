package main

import (
	"fmt"
	"net/http"
)

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