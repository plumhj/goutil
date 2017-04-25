package httputil

import (
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
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

	response, err = ioutil.ReadAll(resp.Body)
	return
}

func ReadJson(r *http.Request, v interface{}) error {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}

	r.Body.Close()

	return nil
}
