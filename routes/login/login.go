package login

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/util"
	"golang.org/x/crypto/bcrypt"
)

type loginHandler struct {
	db        *db.Db
	jwtSecret []byte
}

func GetLoginHandler(db *db.Db, jwtSecret string) *loginHandler {
	return &loginHandler{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

type reqBody struct {
	Password           string
	Email              string
	Expires_in_seconds *int
}

func getBody(r *http.Request) (reqBody, bool) {
	bodyStruct := &reqBody{}
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return reqBody{}, false
	}
	err = json.Unmarshal(body, bodyStruct)
	if err != nil {
		return reqBody{}, false
	}
	return *bodyStruct, true
}

func (handler *loginHandler) createToken(expiry_time time.Duration, user_id int) (string, error) {
	// Create claims with multiple fields populated
	claims := jwt.RegisteredClaims{
		// A usual scenario is to set the expiration time relative to the current time
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry_time)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Issuer:    "chirpy",
		Subject:   fmt.Sprintf("%d", user_id),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(handler.jwtSecret)
}

func (handler *loginHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	body, ok := getBody(r)
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
		token, err := handler.createToken(expiry_time, user.Id)
		if err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, "Could not create token")
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
