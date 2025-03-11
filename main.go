package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsReset() {
	cfg.fileserverHits.Store(0)
}

func main() {
	var apiCfg apiConfig
	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	serveMux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{Hits: %d}`, apiCfg.fileserverHits.Load())
	})
	serveMux.HandleFunc("POST /reset", func(w http.ResponseWriter, r *http.Request) {
		apiCfg.middlewareMetricsReset()
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(":8080", serveMux); err != nil {
		fmt.Println(err)
	}
}
