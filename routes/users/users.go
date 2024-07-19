package users

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/util"
	"golang.org/x/crypto/bcrypt"
)

type usersHandler struct {
	db *db.Db
}

func GetUsersHandler(db *db.Db) *usersHandler {
	return &usersHandler{
		db: db,
	}
}

func (handler *usersHandler) updateUsers(email string, password string) (db.PlainUser, bool) {
	database, success := handler.db.GetDatabase()
	if !success {
		return db.PlainUser{}, false
	}
	hashPassowrd, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return db.PlainUser{}, false
	}
	id := handler.db.GetNextUserId()
	nextUser := db.User{
		Id:       id,
		Email:    email,
		Password: hashPassowrd,
	}
	_, exists := database.Users[email]
	if exists {
		return db.PlainUser{}, false
	}
	database.Users[email] = nextUser
	success = handler.db.UpdateDatabase(database, "chirp")
	if !success {
		fmt.Println("Problem getting database")
		return db.PlainUser{}, false
	}
	return db.PlainUser{Id: id, Email: email}, true
}

func (handler *usersHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("Posting")
	reqBody, ok := getBody(r)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid email posted")
		return
	}
	response, ok := handler.updateUsers(reqBody.Email, reqBody.Password)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Couldn't update database")
		return
	}
	util.RespondWithJSON(w, 200, response)
}

type reqBody struct {
	Password string
	Email    string
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

func (handler *usersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
