package chirp

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/tade3910/chirpy/db"
	"github.com/tade3910/chirpy/middleware/apiConfig"
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

func (handler *chirpHandler) deleteChirp(chripId int, auhtorId int) (int, error) {
	readDatabase, success := handler.db.GetDatabase()
	if !success {
		return 500, fmt.Errorf("could not read from database")
	}
	chirp, ok := readDatabase.Chirps[chripId]
	if !ok {
		return 500, fmt.Errorf("chirp with id %d doesn't exist in database", chripId)
	}
	if chirp.AuthorId != auhtorId {
		return 403, fmt.Errorf("user does not have delete access to this chirp")
	}
	delete(readDatabase.Chirps, chripId)
	handler.db.UpdateDatabase(readDatabase, db.NoDatabase)
	return 204, nil
}

func (handler *chirpHandler) handleGetParamsId(r *http.Request) (int, error) {
	url := r.URL.Path
	urlSplit := strings.Split(url, "/")
	if len(url) < 4 {
		return 0, fmt.Errorf("invalid params")
	}
	id := urlSplit[3]
	intId, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("id must be an int")
	}
	return intId, nil
}

func (handler *chirpHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Getting params")
	chripId, err := handler.handleGetParamsId(r)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Println("Getting author")
	authorId, err := strconv.Atoi(r.Context().Value(apiConfig.UserId).(string))
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Println("Deleting chirp")
	statusCode, err := handler.deleteChirp(chripId, authorId)
	if err != nil {
		util.RespondWithError(w, statusCode, err.Error())
		return
	}
	fmt.Println("Succesfully deleted")
	util.RespondWithJSON(w, statusCode, nil)
}

func (handler *chirpHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	chripId, err := handler.handleGetParamsId(r)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	chirp, ok := handler.getChirp(chripId)
	if !ok {
		util.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Could not get chrip with id, %d", chripId))
		return
	}
	util.RespondWithJSON(w, 200, chirp)
}

func (handler *chirpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r)
	case http.MethodDelete:
		handler.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
