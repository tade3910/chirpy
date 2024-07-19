package util

import (
	"encoding/json"
	"io"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func RespondWithError(w http.ResponseWriter, code int, msg string) error {
	return RespondWithJSON(w, code, map[string]string{"error": msg})
}

func GetBody[T interface{}](r *http.Request, bodyStruct T) (T, bool) {
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return bodyStruct, false
	}
	err = json.Unmarshal(body, bodyStruct)
	if err != nil {
		return bodyStruct, false
	}
	return bodyStruct, true
}
