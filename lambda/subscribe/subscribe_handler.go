// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html

package subscribe

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

func (h SubscribeHandler) Handle(event events.SQSEvent) error {
	h.logger.Info("Handle: ", zap.String("event", fmt.Sprintf("%v", event)))

	for _, record := range event.Records {
		err := h.subscribeService.Process(record)

		if err != nil {
			return err
		}
	}
	h.logger.Info("Handle: done")

	return nil
}
