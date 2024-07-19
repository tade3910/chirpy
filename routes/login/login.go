package login

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/middleware/apiConfig"
	"github.com/tade3910/chirpy/util"
	"golang.org/x/crypto/bcrypt"
)

type loginHandler struct {
	db *db.Db
}

func GetLoginHandler(db *db.Db) *loginHandler {
	return &loginHandler{
		db: db,
	}
}

type reqBody struct {
	Password           string
	Email              string
	Expires_in_seconds *int
}

func createToken(expiry_time time.Duration, user_id int, jwtSecret string) (string, error) {

	// Create claims with multiple fields populated
	claims := jwt.RegisteredClaims{
		// A usual scenario is to set the expiration time relative to the current time
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry_time)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Issuer:    "chirpy",
		Subject:   fmt.Sprintf("%d", user_id),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func (handler *loginHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	jwtSecret, ok := r.Context().Value(apiConfig.JwtSecret).(string)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "I messed up sharing the secret context")
		return
	}
	body, ok := util.GetBody(r, &reqBody{})
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid req body")
		return
	}
	database, success := handler.db.GetDatabase()
	if !success {
		util.RespondWithError(w, http.StatusInternalServerError, "Couldn't read from database")
		return
	}
	user, exists := database.Users[body.Email]
	if !exists {
		util.RespondWithError(w, http.StatusInternalServerError, "No such user exists")
		return
	}
	if bcrypt.CompareHashAndPassword(user.Password, []byte(body.Password)) == nil {
		expiry_time := 24 * time.Hour
		if body.Expires_in_seconds != nil {
			expiry_time = time.Duration(*body.Expires_in_seconds) * time.Second
		}
		token, err := createToken(expiry_time, user.Id, jwtSecret)
		if err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, "Could not create token")
			return
		}
		util.RespondWithJSON(w, 200, map[string]string{"Email": user.Email, "Id": fmt.Sprintf("%d", user.Id), "token": token})
	} else {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid password")
	}
}

func (handler *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
