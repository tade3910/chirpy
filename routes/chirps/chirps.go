package chirps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tade3910/chirpy/db"
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
	chrip, ok := handleChirp(r)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid chirp posted")
		return
	}
	response, ok := handler.updateChirps(chrip)
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

func (handler *chirpsHandler) updateChirps(data string) (db.Chirp, bool) {
	database, success := handler.db.GetDatabase()
	if !success {
		fmt.Println("Problem getting database")
		return db.Chirp{}, false
	}
	id := handler.db.GetNextId()
	nextChirp := db.Chirp{
		Id:   id,
		Body: data,
	}
	database.Chirps[id] = nextChirp
	success = handler.db.UpdateDatabase(database, "user")
	if !success {
		fmt.Println("Problem getting database")
		return db.Chirp{}, false
	}
	return nextChirp, true
}
