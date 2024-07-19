package chirp

import (
	"net/http"
	"strconv"
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

func (handler *chirpHandler) getChirp(chripId int) (db.Chirp, bool) {
	readDatabase, success := handler.db.GetDatabase()
	if !success {
		return db.Chirp{}, false
	}
	chirp, ok := readDatabase.Chirps[chripId]
	if !ok {
		return db.Chirp{}, false
	}
	return chirp, true
}

func (handler *chirpHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	url := r.URL.Path
	urlSplit := strings.Split(url, "/")
	if len(url) < 4 {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid params")
		return
	}
	id := urlSplit[3]
	intId, err := strconv.Atoi(id)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid params")
		return
	}
	chirp, ok := handler.getChirp(intId)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, "Could not get id")
		return
	}
	util.RespondWithJSON(w, 200, chirp)
}

func (handler *chirpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
