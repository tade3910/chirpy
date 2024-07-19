package polka

import (
	"fmt"
	"net/http"

	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/util"
)

type polkaHandler struct {
	db *db.Db
}

func GetPolkaHandler(db *db.Db) *polkaHandler {
	return &polkaHandler{
		db: db,
	}
}

func (handler *polkaHandler) upgradeUser(user_id int) (int, error) {
	datbase, ok := handler.db.GetDatabase()
	if !ok {
		return 500, fmt.Errorf("error getting database")

	}
	user, exists := datbase.IDUsersMap[user_id]
	if !exists {
		return 404, fmt.Errorf("user with provided id doesn't exist")
	}
	user.Is_chirpy_red = true
	datbase.Users[user.Email].Is_chirpy_red = true
	handler.db.UpdateDatabase(datbase, db.NoDatabase)
	return 204, nil
}

func (handler *polkaHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	type respBody struct {
		Event string
		Data  struct {
			User_id int
		}
	}
	bodyStruct, ok := util.GetBody(r, &respBody{})
	if !ok {
		util.RespondWithError(w, 500, "Error parsing body")
		return
	}
	if bodyStruct.Event == "user.upgraded" {
		status, err := handler.upgradeUser(bodyStruct.Data.User_id)
		if err != nil {
			util.RespondWithError(w, status, err.Error())
			return
		}
	}
	util.RespondWithJSON(w, 204, nil)
}

func (handler *polkaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
