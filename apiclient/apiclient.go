package apiclient

import (
	"encoding/json"
	"io"
	"net/http"
	"sqstest/lambda/response"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) Client {
	return Client{baseURL: baseURL}
}

func (c Client) Get(test string, function string) (response response.Response) {
	var resp *http.Response
	var err error

	resp, err = http.Get(c.baseURL + test + "/" + function)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}

	return response
}
