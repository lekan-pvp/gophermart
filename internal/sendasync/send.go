package sendasync

import (
	"net/http"
)

func SendGetAcync(url string, rc chan *http.Response) error {
	response, err := http.Post(url, "application/json", nil)
	if err == nil {
		rc <- response
	}
	return err
}
