package subscriptionhandler

import (
	"context"
	"fmt"
	"sqstest/service"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SubscriptionHandler struct {
	logger           *zapray.Logger
	subscribeService service.SuspendableService
}

func NewSubscriptionHandler(logger *zapray.Logger, subscribeService service.SuspendableService) SubscriptionHandler {
	return SubscriptionHandler{
		logger:           logger,
		subscribeService: subscribeService,
	}
}

func (h SubscriptionHandler) Handle(ctx context.Context, event events.SQSEvent) (err error) {
	h.logger.Debug("Handle: ", zap.String("ctx", fmt.Sprintf("%v", ctx)), zap.String("event", fmt.Sprintf("%v", event)))
	h.logger.Debug("Handle: ", zap.Int("records", len(event.Records)))

	for _, record := range event.Records {
		err = h.subscribeService.Receive(ctx, record)
		if err != nil {
			h.logger.Info("Handle: ", zap.String("err", err.Error()))
			return err
		}
	}

	return nil
}
