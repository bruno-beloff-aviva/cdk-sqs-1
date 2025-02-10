package response

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type Response struct {
	StatusCode   int
	ErrorMessage string
	Body         string
}

func NewOKResponse(body string) Response {
	return Response{
		StatusCode:   http.StatusOK,
		ErrorMessage: "",
		Body:         body,
	}
}

func NewErrorResponse(err error, body string) Response {
	return Response{
		StatusCode:   http.StatusInternalServerError,
		ErrorMessage: err.Error(),
		Body:         body,
	}
}

func (r Response) APIResponse() (apiResponse events.APIGatewayProxyResponse, marshalErr error) {
	var jsonBody []byte

	jsonBody, marshalErr = json.Marshal(r)

	apiResponse = events.APIGatewayProxyResponse{
		StatusCode: r.StatusCode,
		Body:       string(jsonBody),
	}

	return apiResponse, marshalErr
}
