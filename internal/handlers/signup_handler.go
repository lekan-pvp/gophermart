package handlers

import (
	"context"
	"encoding/json"
	"github.com/lekan/gophermart/internal/models"
	"github.com/lekan/gophermart/internal/utils"
	"log"
	"net/http"
	"time"
)

var CreateAccount = func(w http.ResponseWriter, r *http.Request) {
	log.Println("In CreateAccount")
	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()
	log.Println(account.Login)
	resp := account.CreateUser(ctx)
	log.Println("After CreateUser")
	utils.Respond(w, resp)
}
