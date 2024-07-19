package login

import (
	"net/http"
	"time"

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
	Password string
	Email    string
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
		expiry_time := 1 * time.Hour
		token, err := util.CreateAcessToken(expiry_time, user.Id, jwtSecret)
		if err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, "Could not create token")
			return
		}
		refreshToken, err := util.CreateRefreshToken()
		session := db.GetNewSession(user)
		if err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, "Could not create refresh token")
			return
		}
		database.Sessions[refreshToken] = session
		handler.db.UpdateDatabase(database, db.NoDatabase)
		responseBody := struct {
			Token        string
			RefreshToken string
			db.PlainUser
		}{
			Token:        token,
			RefreshToken: refreshToken,
			PlainUser: db.PlainUser{
				Email:         user.Email,
				Id:            user.Id,
				Is_chirpy_red: user.Is_chirpy_red,
			},
		}
		util.RespondWithJSON(w, 200, responseBody)
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
