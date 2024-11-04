package main

import "net/http"

func healthCheckHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(200)
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.Write([]byte("OK"))
}