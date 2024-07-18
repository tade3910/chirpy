package users

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/util"
)

type usersHandler struct {
	db *db.Db
}

func GetUsersHandler(db *db.Db) *usersHandler {
	return &usersHandler{
		db: db,
	}
}

func (handler *usersHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Posting")
	email, ok := getEmail(r)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid email posted")
		return
	}
	response, ok := handler.db.UpdateUsers(email)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Couldn't update database")
		return
	}
	util.RespondWithJSON(w, 200, response)
}

func getEmail(r *http.Request) (string, bool) {
	type respBody struct {
		Email string
	}
	bodyStruct := &respBody{}
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return "", false
	}
	err = json.Unmarshal(body, bodyStruct)
	if err != nil {
		return "", false
	}
	return bodyStruct.Email, true
}

func (handler *usersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
