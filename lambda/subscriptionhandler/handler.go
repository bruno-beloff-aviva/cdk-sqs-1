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
	logger              *zapray.Logger
	subscriptionService service.SubscriptionService
}

func NewSubscriptionHandler(logger *zapray.Logger, subscriptionService service.SubscriptionService) SubscriptionHandler {
	return SubscriptionHandler{
		logger:              logger,
		subscriptionService: subscriptionService,
	}
}

func (h SubscriptionHandler) Handle(ctx context.Context, event events.SQSEvent) (err error) {
	h.logger.Debug("Handle: ", zap.String("ctx", fmt.Sprintf("%v", ctx)), zap.String("event", fmt.Sprintf("%v", event)))
	h.logger.Debug("Handle: ", zap.Int("records", len(event.Records)))

	for _, record := range event.Records {
		err = h.subscriptionService.Receive(ctx, record)
		if err != nil {
			h.logger.Info("Handle: ", zap.String("err", err.Error()))
			return err
		}
	}

	return nil
}
