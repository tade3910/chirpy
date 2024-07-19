package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/middleware/apiConfig"
	polka "github.com/tade3910/chirpy/routes/Polka"
	"github.com/tade3910/chirpy/routes/chirp"
	"github.com/tade3910/chirpy/routes/chirps"
	"github.com/tade3910/chirpy/routes/login"
	"github.com/tade3910/chirpy/routes/refresh"
	"github.com/tade3910/chirpy/routes/users"
)

type HealthHandler struct {
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

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	if port == "" || jwtSecret == "" {
		log.Fatal("No Port found in env")
	}
	router := http.NewServeMux()
	apiCfg := apiConfig.GetApiConfig(jwtSecret, polkaKey)
	router.Handle("/app/*", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	router.Handle("/api/healthz", &HealthHandler{})
	db, ok := db.GetDb()
	if !ok {
		log.Fatal("Could not connect to database")
	}
	router.Handle("/api/chirps", apiCfg.EnsureAuthenticated(chirps.GetChirpsHandler(db)))
	router.Handle("/api/chirps/", apiCfg.EnsureAuthenticated(chirp.GetChirpHandler(db)))
	router.Handle("/api/users", apiCfg.EnsureAuthenticated(users.GetUsersHandler(db)))
	router.Handle("/api/login", apiCfg.WithJwtSecret(login.GetLoginHandler(db)))
	router.Handle("/api/refresh", apiCfg.WithJwtSecret(refresh.GetRefreshHandler(db)))
	router.Handle("/api/polka/webhooks", apiCfg.CheckPolkaKey(polka.GetPolkaHandler(db)))
	router.HandleFunc("/admin/metrics", apiCfg.HandleMetrics)
	router.HandleFunc("/api/reset", apiCfg.HandleReset)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	fmt.Printf("Server listening on port %s\n", port)
	server.ListenAndServe()
}
