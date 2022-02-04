package utils

import (
	"net/http"
)

func Message(ok bool, status int, message string) map[string]interface{} {
	return map[string]interface{}{"ok": ok, "status": status, "message": message}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	status := data["status"].(int)
	message := data["message"].(string)
	if ok := data["ok"].(bool); !ok {
		http.Error(w, message, status)
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)
	}
	//json.NewEncoder(w).Encode(data)
}
