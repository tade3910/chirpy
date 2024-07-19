package apiConfig

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tade3910/chirpy/util"
)

type contextKey string

const (
	UserId    contextKey = "userId"
	JwtSecret contextKey = "jwtSecret"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret      string
	polkaKey       string
	mu             sync.Mutex
}

func GetApiConfig(JwtSecret string, polkaKey string) *apiConfig {
	return &apiConfig{
		jwtSecret: JwtSecret,
		polkaKey:  polkaKey,
	}
}

func (cfg *apiConfig) WithJwtSecret(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), JwtSecret, cfg.jwtSecret)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func excludeRoutes(r *http.Request) bool {
	return r.URL.Path == "/api/users" && r.Method == http.MethodPost
}

func (cfg *apiConfig) EnsureAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If a new user is created they don't have token
		if excludeRoutes(r) {
			next.ServeHTTP(w, r)
			return
		}
		tokenString, err := util.GetAuthToken(r, util.Bearer)
		if err != nil {
			util.RespondWithError(w, 401, err.Error())
			return
		}
		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.jwtSecret), nil
		})
		if err != nil {
			util.RespondWithError(w, 401, "Error parsing auth token")
			return
		}
		userId, err := token.Claims.GetSubject()
		if err != nil {
			util.RespondWithError(w, 401, "userId could not be parsed from token")
			return
		}
		ctx := context.WithValue(r.Context(), UserId, userId)

		// Call the next handler with the modified request context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (cfg *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.mu.Lock()
		cfg.fileserverHits++
		cfg.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) CheckPolkaKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := util.GetAuthToken(r, util.ApiKey)
		if err != nil || tokenString != cfg.polkaKey {
			util.RespondWithError(w, 401, err.Error())
			return
		}
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
