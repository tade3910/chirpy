package apiConfig

import (
	"fmt"
	"net/http"
	"sync"
)

type apiConfig struct {
	fileserverHits int
	JwtSecret      string
	mu             sync.Mutex
}

func GetApiConfig(JwtSecret string) *apiConfig {
	return &apiConfig{
		JwtSecret: JwtSecret,
	}
}

func (cfg *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.mu.Lock()
		cfg.fileserverHits++
		cfg.mu.Unlock()
		next.ServeHTTP(w, r)
	})
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
