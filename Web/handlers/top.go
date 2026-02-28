package handlers

import (
	"encoding/json"
	"net/http"
)

func TopUsers(w http.ResponseWriter, r *http.Request) {
	var Uprajnenie string
	json.NewDecoder(r.Body).Decode(&Uprajnenie)
}
func AddRes(w http.ResponseWriter, r *http.Request) {

}
func GetRes(w http.ResponseWriter, r *http.Request) {

}
