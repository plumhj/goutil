package httputil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func PostObjectAsJson(url string, obj interface{}, headers map[string]string, timeout_ms int) (statusCode int, responseBody []byte, err error) {

	data, err := json.Marshal(obj)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	for k,v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: time.Millisecond * time.Duration(timeout_ms),
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	statusCode = resp.StatusCode
	responseBody, err = ioutil.ReadAll(resp.Body)
	return
}

func PostJson(url string, payload []byte, timeout_ms int) (response []byte, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Millisecond * time.Duration(timeout_ms),
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	response, err = ioutil.ReadAll(resp.Body)
	return
}

func PostForm(uri string, formValues url.Values, timeout int) (response []byte, err error) {

	req, err := http.NewRequest("POST", uri, strings.NewReader(formValues.Encode()))

	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: time.Millisecond * time.Duration(timeout),
	}
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
