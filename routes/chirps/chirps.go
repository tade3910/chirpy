package chirps

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/middleware/apiConfig"
	"github.com/tade3910/chirpy/util"
)

type chirpsHandler struct {
	db *db.Db
}

func GetChirpsHandler(db *db.Db) *chirpsHandler {
	return &chirpsHandler{
		db: db,
	}
}

func (handler *chirpsHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	authorIdString, ok := r.Context().Value(apiConfig.UserId).(string)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "I messed up passing the id")
		return
	}
	authorId, err := strconv.Atoi(authorIdString)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "error converting id to int")
	}
	chrip, ok := handleChirp(r)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid chirp posted")
		return
	}
	response, ok := handler.updateChirps(chrip, authorId)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Couldn't update database")
		return
	}
	util.RespondWithJSON(w, 200, response)
}

func cleanBody(s string) string {
	s = strings.ToLower(s)
	words := strings.Split(s, " ")
	for index, word := range words {
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			words[index] = "****"
		}
	}
	return strings.Join(words, " ")
}

func handleChirp(r *http.Request) (string, bool) {
	type respBody struct {
		Body string
	}
	bodyStruct, ok := util.GetBody(r, &respBody{})
	if !ok {
		return "", false
	}
	if len(bodyStruct.Body) > 140 {
		return "", false
	} else {
		return cleanBody(bodyStruct.Body), true
	}
}

func (handler *chirpsHandler) getFormatedDatabase() ([]db.Chirp, bool) {
	database, success := handler.db.GetDatabase()
	if !success {
		return nil, false
	}
	formatedDatabse := make([]db.Chirp, len(database.Chirps))
	for key, chirp := range database.Chirps {
		formatedDatabse[key] = chirp
	}
	return formatedDatabse, true
}

func (handler *chirpsHandler) handleGet(w http.ResponseWriter) {
	chirps, ok := handler.getFormatedDatabase()
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	util.RespondWithJSON(w, 200, chirps)
}

func (handler *chirpsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w)
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (handler *chirpsHandler) updateChirps(data string, authorId int) (db.Chirp, bool) {
	database, success := handler.db.GetDatabase()
	if !success {
		fmt.Println("Problem getting database")
		return db.Chirp{}, false
	}
	id := handler.db.GetNextId()
	nextChirp := db.Chirp{
		Id:       id,
		Body:     data,
		AuthorId: authorId,
	}
	database.Chirps[id] = nextChirp
	success = handler.db.UpdateDatabase(database, db.ChirpDatabase)
	if !success {
		fmt.Println("Problem getting database")
		return db.Chirp{}, false
	}
	return nextChirp, true
}
