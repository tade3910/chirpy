package users

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/middleware/apiConfig"
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

func (handler *usersHandler) addUser(email string, password string) (db.PlainUser, bool) {
	database, success := handler.db.GetDatabase()
	if !success {
		return db.PlainUser{}, false
	}
	hashPassowrd, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return db.PlainUser{}, false
	}
	id := handler.db.GetNextUserId()
	nextUser := &db.User{
		Password: hashPassowrd,
		PlainUser: db.PlainUser{
			Id:    id,
			Email: email,
		},
	}
	_, exists := database.Users[email]
	if exists {
		return db.PlainUser{}, false
	}
	database.Users[email] = nextUser
	database.IDUsersMap[id] = nextUser
	success = handler.db.UpdateDatabase(database, db.UserDatabase)
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
	authStruct, ok := util.GetBody(r, &authStruct{})
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid email posted")
		return
	}
	response, ok := handler.addUser(authStruct.Email, authStruct.Password)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Couldn't update database")
		return
	}
	util.RespondWithJSON(w, 200, response)
}

type authStruct struct {
	Password string
	Email    string
}

func (handler *usersHandler) updateUser(userId int, email string, password string) (db.PlainUser, bool) {
	database, success := handler.db.GetDatabase()
	if !success {
		return db.PlainUser{}, false
	}
	user, exists := database.IDUsersMap[userId]
	if !exists {
		return db.PlainUser{}, false
	}
	hashPassowrd, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return db.PlainUser{}, false
	}
	delete(database.Users, user.Email)
	user.Email = email
	user.Password = hashPassowrd
	database.IDUsersMap[userId] = user
	database.Users[user.Email] = user
	success = handler.db.UpdateDatabase(database, db.NoDatabase)
	if !success {
		fmt.Println("Problem getting database")
		return db.PlainUser{}, false
	}
	return db.PlainUser{Id: userId, Email: email}, true
}

func (handler *usersHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	userIdString, ok := r.Context().Value(apiConfig.UserId).(string)
	if !ok {
		util.RespondWithError(w, 500, "I didn't pass in the user Id correctly")
		return
	}
	authDetails, ok := util.GetBody(r, &authStruct{})
	if !ok {
		util.RespondWithError(w, 500, "Error parsing req body")
		return
	}
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		util.RespondWithError(w, 500, err.Error())
		return
	}
	response, ok := handler.updateUser(userId, authDetails.Email, authDetails.Password)
	if !ok {
		util.RespondWithError(w, 500, "Error Updating user")
		return
	}
	util.RespondWithJSON(w, 200, response)
}

func (handler *usersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	case http.MethodPut:
		handler.handlePut(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
