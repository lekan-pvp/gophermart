package handlers

import (
	"encoding/json"
	"github.com/lekan/gophermart/internal/models"
	"net/http"
)

//Signup...
func Signup(w http.ResponseWriter, r *http.Request) {
	creds := &models.Credentials{}
	if err := json.NewDecoder(r.Body).Decode(creds); err != nil {
		w.WriteHeader(400)
		return
	}
}
