package pubhandler

import (
	"fmt"

	"context"
	"sqstest/lambda/response"
	"sqstest/service/pub"
	"sqstest/service/testmessage"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type PubHandler struct {
	logger     *zapray.Logger
	pubService pub.PubService
}

func NewPubHandler(logger *zapray.Logger, pubService pub.PubService) PubHandler {
	return PubHandler{
		logger:     logger,
		pubService: pubService,
	}
}

func (h PubHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (apiResponse events.APIGatewayProxyResponse, err error) {
	h.logger.Debug("Handle: ", zap.String("request", fmt.Sprintf("%v", request)))

	var message testmessage.TestMessage
	var resp response.Response

	sourceIP := request.RequestContext.Identity.SourceIP
	message, err = h.pubService.Publish(ctx, sourceIP, request.Path)

	if err != nil {
		h.logger.Error("Pub error", zap.Error(err))
	}

	if err == nil {
		resp = response.NewOKResponse(message.String())
	} else {
		resp = response.NewErrorResponse(err, message.String())
	}

	apiResponse, err = resp.APIResponse()
	if err != nil {
		h.logger.Error("APIResponse error", zap.Error(err))
	}

	return apiResponse, err
}
