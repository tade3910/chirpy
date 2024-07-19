package refresh

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/middleware/apiConfig"
	"github.com/tade3910/chirpy/util"
)

type refreshHandler struct {
	db *db.Db
}

func GetRefreshHandler(db *db.Db) *refreshHandler {
	return &refreshHandler{
		db: db,
	}
}

func (handler *refreshHandler) refreshTokenToSession(oldRefreshToken string) (*db.Session, *db.Database, int, error) {
	database, ok := handler.db.GetDatabase()
	if !ok {
		return nil, nil, 500, fmt.Errorf("couldn't get database")
	}
	session, ok := database.Sessions[oldRefreshToken]
	if !ok {
		return nil, nil, 401, fmt.Errorf("refresh token doesn't exist in database")
	} else if session.Expires.Before(time.Now().UTC()) {
		return nil, nil, 401, fmt.Errorf("refresh token has expired")
	}
	return &session, database, 200, nil
}
func (handler *refreshHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	oldRefreshToken, err := util.GetAuthToken(r, util.Bearer)
	if err != nil {
		util.RespondWithError(w, 500, err.Error())
		return
	}
	_, database, errorCode, err := handler.refreshTokenToSession(oldRefreshToken)
	if err != nil {
		util.RespondWithError(w, errorCode, err.Error())
		return
	}
	delete(database.Sessions, oldRefreshToken)
	handler.db.UpdateDatabase(database, db.NoDatabase)
	util.RespondWithJSON(w, 201, nil)
}

func (handler *refreshHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	oldRefreshToken, err := util.GetAuthToken(r, util.Bearer)
	if err != nil {
		util.RespondWithError(w, 500, err.Error())
		return
	}
	session, database, errorCode, err := handler.refreshTokenToSession(oldRefreshToken)
	if err != nil {
		util.RespondWithError(w, errorCode, err.Error())
		return
	}
	refreshToken, err := util.CreateRefreshToken()
	if err != nil {
		util.RespondWithError(w, 500, "Could not create new refresh token")
	}
	database.Sessions[refreshToken] = db.GetNewSession(session.User)
	delete(database.Sessions, oldRefreshToken)
	// need to generate new access token
	jwtSecret, ok := r.Context().Value(apiConfig.JwtSecret).(string)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "I messed up sharing the secret context")
		return
	}
	expiry_time := 1 * time.Hour
	token, err := util.CreateAcessToken(expiry_time, session.User.Id, jwtSecret)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Could not create access token")
		return
	}
	handler.db.UpdateDatabase(database, db.NoDatabase)
	util.RespondWithJSON(w, 200, map[string]string{"token": token})
}

func (handler *refreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	case http.MethodDelete:
		handler.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
