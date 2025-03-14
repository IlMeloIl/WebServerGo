package main

import (
	"GoServer/internal/database"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	Secret         string
}

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"password"`
	AccessToken    string    `json:"token"`
	RefreshToken   string    `json:"refresh_token"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	var apiCfg apiConfig = apiConfig{
		fileserverHits: atomic.Int32{},
		DB:             database.New(db),
		Platform:       os.Getenv("PLATFORM"),
		Secret:         os.Getenv("SECRET_KEY"),
	}
	serveMux := http.NewServeMux()
	middleware := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.Handle("/app/", middleware)
	serveMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.middlewareMetricsReset)
	serveMux.HandleFunc("POST /api/users", apiCfg.createUser)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.createChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpByID)
	serveMux.HandleFunc("POST /api/login", apiCfg.login)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.refresh)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.revoke)
	serveMux.HandleFunc("PUT /api/users", apiCfg.updateUser)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirp)
	if err := http.ListenAndServe(":8080", serveMux); err != nil {
		fmt.Println(err)
	}
}
