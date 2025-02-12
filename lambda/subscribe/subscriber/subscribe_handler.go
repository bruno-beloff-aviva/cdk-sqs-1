package subscriber

import (
	"context"
	"fmt"
	"sqstest/service"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SubscribeHandler struct {
	logger           *zapray.Logger
	subscribeService service.SubscribeService
}

func NewSubscribeHandler(logger *zapray.Logger, subscribeService service.SubscribeService) SubscribeHandler {
	return SubscribeHandler{
		logger:           logger,
		subscribeService: subscribeService,
	}
}

func (h SubscribeHandler) Handle(ctx context.Context, event events.SQSEvent) (err error) {
	h.logger.Info("Handle: ", zap.String("ctx", fmt.Sprintf("%v", ctx)), zap.String("event", fmt.Sprintf("%v", event)))

	for _, record := range event.Records {
		err = h.subscribeService.Receive(ctx, record)
		if err != nil {
			h.logger.Info("Handle: ", zap.String("err", err.Error()))
			return err
		}
	}

	return nil
}
