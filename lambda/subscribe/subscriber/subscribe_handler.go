package subscriber

import (
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

func (h SubscribeHandler) Handle(event events.SQSEvent) (err error) {
	h.logger.Debug("Handle: ", zap.String("event", fmt.Sprintf("%v", event)))

	for _, record := range event.Records {
		err = h.subscribeService.Receive(record)
		if err != nil {
			h.logger.Info("Handle: ", zap.String("err", err.Error()))
			return err
		}
	}

	return nil
}
