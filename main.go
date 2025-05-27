package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/quockhanhcao/go-server/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	godotenv.Load()
	const filePathRoot = "."
	const port = "8080"
	var apiCfg = apiConfig{}
	// db queries
	dbURL := os.Getenv("DB_URL")
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	dbQueries := database.New(dbConn)
	apiCfg.db = dbQueries
	apiCfg.platform = os.Getenv("PLATFORM")

	mux := http.NewServeMux()

	// file server
	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	// route handler
	mux.HandleFunc("GET /api/healthz", readinessHandler)

	// users
	mux.HandleFunc("POST /api/users", apiCfg.createUsersHandler)
    mux.HandleFunc("POST /api/login", apiCfg.loginHandler)
	// chirps
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
    mux.HandleFunc("GET /api/chirps/{id}", apiCfg.getChirpByIdHandler)

	// admin routes
	mux.HandleFunc("POST /admin/reset", apiCfg.deleteAllUsersHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	// start the server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving files from %s on port: %s\n", filePathRoot, port)
	log.Fatal(server.ListenAndServe())
}
