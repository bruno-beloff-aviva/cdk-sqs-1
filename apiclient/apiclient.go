package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sqstest/lambda/response"
	"time"
)

type Client struct {
	baseURL         string
	requestInterval int
}

func NewClient(baseURL string, requestInterval int) Client {
	return Client{baseURL: baseURL, requestInterval: requestInterval}
}

func (c Client) Run(tape Tape) {
	ticker := time.NewTicker(time.Duration(c.requestInterval) * time.Second)

	for {
		r := tape.NextRequest()
		if r.TestId == "" {
			ticker.Stop()
			break
		}

		i := 0
		for t := range ticker.C {
			if i == r.Repeat {
				break
			}
			i++
			fmt.Printf("Tick at: %v\n", t.UTC())
			response := c.Get(r.TestId, r.Function)
			fmt.Println(response.String())
		}
	}
}

func (c Client) Get(testId string, function string) (response response.Response) {
	var resp *http.Response
	var err error

	resp, err = http.Get(c.baseURL + testId + "/" + function)
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
