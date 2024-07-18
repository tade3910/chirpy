package chirp

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/util"
)

type chirpHandler struct {
	db *db.Db
}

func GetChirpHandler(db *db.Db) *chirpHandler {
	return &chirpHandler{
		db: db,
	}
}

func (handler *chirpHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	chrip, ok := handleChirp(r)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid chirp posted")
		return
	}
	response, ok := handler.db.UpdateChirps(chrip)
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

func (handler *chirpHandler) handleGet(w http.ResponseWriter) {
	chirps, ok := handler.db.GetFormatedDatabase()
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	util.RespondWithJSON(w, 200, chirps)
}

func (handler *chirpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w)
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
