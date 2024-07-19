package util

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func RespondWithError(w http.ResponseWriter, code int, msg string) error {
	return RespondWithJSON(w, code, map[string]string{"error": msg})
}

func GetBody[T interface{}](r *http.Request, bodyStruct T) (T, bool) {
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return bodyStruct, false
	}
	err = json.Unmarshal(body, bodyStruct)
	if err != nil {
		return bodyStruct, false
	}
	return bodyStruct, true
}

type authType string

const (
	Bearer authType = "Bearer"
	ApiKey authType = "ApiKey"
)

func GetAuthToken(r *http.Request, auauthType authType) (string, error) {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return "", fmt.Errorf("no Authorization token provided")
	}
	split := strings.Split(bearerToken, " ")
	if len(split) != 2 || split[0] != string(auauthType) {
		return "", fmt.Errorf("malformated Authorization token provided")
	}
	return split[1], nil
}

func CreateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func CreateAcessToken(expiry_time time.Duration, user_id int, jwtSecret string) (string, error) {

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
