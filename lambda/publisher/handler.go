package publisher

import (
	"fmt"

	"context"
	"sqstest/lambda/response"
	"sqstest/service"
	"sqstest/service/testmessage"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type PublishHandler struct {
	logger         *zapray.Logger
	publishService service.PublishService
}

func NewPublishHandler(logger *zapray.Logger, publishService service.PublishService) PublishHandler {
	return PublishHandler{
		logger:         logger,
		publishService: publishService,
	}
}

func (h PublishHandler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (apiResponse events.APIGatewayProxyResponse, err error) {
	h.logger.Debug("Handle: ", zap.String("request", fmt.Sprintf("%v", request)))

	var message testmessage.TestMessage
	var resp response.Response

	sourceIP := request.RequestContext.Identity.SourceIP
	message, err = h.publishService.Publish(ctx, sourceIP, request.Path)

	if err != nil {
		h.logger.Error("Publish error", zap.Error(err))
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
