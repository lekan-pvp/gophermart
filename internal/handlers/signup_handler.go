package handlers

import (
	"encoding/json"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/utils"
	"net/http"
)

var CreateAccount = func(w http.ResponseWriter, r *http.Request) {
	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	resp := account.Create()

	utils.Respond(w, resp)
}
