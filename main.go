package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/routes/chirp"
	"github.com/tade3910/chirpy/routes/chirps"
	"github.com/tade3910/chirpy/routes/users"
)

type HealthHandler struct {
}

type apiConfig struct {
	fileserverHits int
	mu             sync.Mutex
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.mu.Lock()
		cfg.fileserverHits++
		cfg.mu.Unlock()
		next.ServeHTTP(w, r)
	})

}

func (handler *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg.mu.Lock()
	responseBody := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
	`, cfg.fileserverHits)
	cfg.mu.Unlock()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseBody))
}

func (cfg *apiConfig) HandleReset(w http.ResponseWriter, r *http.Request) {
	cfg.mu.Lock()
	cfg.fileserverHits = 0
	cfg.mu.Unlock()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	responseBody := fmt.Sprintf("Hits: %d", cfg.fileserverHits)
	w.Write([]byte(responseBody))
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("No Port found in env")
	}
	router := http.NewServeMux()
	apiCfg := &apiConfig{}
	router.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	router.Handle("/api/healthz", &HealthHandler{})
	db, ok := db.GetDb()
	if !ok {
		log.Fatal("Could not connect to database")
	}
	router.Handle("/api/chirps", chirp.GetChirpHandler(db))
	router.Handle("/api/chirps/", chirps.GetChirpsHandler(db))
	router.Handle("/api/users", users.GetUsersHandler(db))
	router.HandleFunc("/admin/metrics", apiCfg.HandleMetrics)
	router.HandleFunc("/api/reset", apiCfg.HandleReset)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	fmt.Printf("Server listening on port %s\n", port)
	server.ListenAndServe()
}
