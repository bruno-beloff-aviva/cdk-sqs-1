package subhandler

import (
	"context"
	"fmt"
	"sqstest/service/sub"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SubHandler struct {
	logger     *zapray.Logger
	subService sub.SubService
}

func NewSubHandler(logger *zapray.Logger, subService sub.SubService) SubHandler {
	return SubHandler{
		logger:     logger,
		subService: subService,
	}
}

func (h SubHandler) Handle(ctx context.Context, event events.SQSEvent) (err error) {
	h.logger.Debug("Handle: ", zap.String("ctx", fmt.Sprintf("%v", ctx)), zap.String("event", fmt.Sprintf("%v", event)))
	h.logger.Debug("Handle: ", zap.Int("records", len(event.Records)))

	for _, record := range event.Records {
		err = h.subService.Handle(ctx, record)
		if err != nil {
			h.logger.Info("Handle: ", zap.String("err", err.Error()))
			return err
		}
	}

	return nil
}
