package sessions

import (
	gsessions "github.com/gorilla/sessions"
	"net/http"
	"os"
)

var store = gsessions.NewCookieStore([]byte(os.Getenv("token_password")))

func Get(req *http.Request) (*gsessions.Session, error) {
	return store.Get(req, "session-name")
}
