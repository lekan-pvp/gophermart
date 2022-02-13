package sendasync

import (
	"net/http"
)

func SendGetAcync(url string, rc chan *http.Response) error {
	response, err := http.Get(url)
	if err != nil {
		rc <- response
	}
	return err
}
