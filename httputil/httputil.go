package httputil

import (
	"net/http"
	"bytes"
	"io/ioutil"
)

func PostJson(url string, payload []byte) (response []byte, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	resp, err = ioutil.ReadAll(resp.Body)
	return
}
