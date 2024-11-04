package main

import "net/http"

func (cfg *apiConfig) middlewareMetricsReset(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits.Store(0)
		next.ServeHTTP(res, req)
	})
}
