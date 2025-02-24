package singleshot

import (
	"context"
	"sqstest/services"

	"go.uber.org/zap"
)

type SingleshotHandler[T any] interface {
	Process(ctx context.Context, event T) (err error)
	UniqueID(event T) (policyOrQuoteID string, eventID string, err error)
	Logger() *zap.Logger
}

type SingleshotGateway[T any] struct {
	handler               SingleshotHandler[T]
	eventHasBeenProcessed services.EventHasBeenProcessedFunc
	markEventAsProcessed  services.MarkEventAsProcessedFunc
}

func NewSingleshotGateway[T any](handler SingleshotHandler[T], eventHasBeenProcessed services.EventHasBeenProcessedFunc, markEventAsProcessed services.MarkEventAsProcessedFunc) SingleshotGateway[T] {
	return SingleshotGateway[T]{
		handler:               handler,
		eventHasBeenProcessed: eventHasBeenProcessed,
		markEventAsProcessed:  markEventAsProcessed,
	}
}

func (g SingleshotGateway[T]) ProcessOnce(ctx context.Context, event T) error {
	g.handler.Logger().Debug("ProcessOnce: ", zap.Any("event", event))

	// Check...
	policyOrQuoteID, eventID, err := g.handler.UniqueID(event)
	if err != nil {
		g.handler.Logger().Error("Error getting UniqueID", zap.Error(err))
		return err
	}

	eventHasBeenProcessed, err := g.eventHasBeenProcessed(ctx, policyOrQuoteID, eventID)
	if err != nil {
		g.handler.Logger().Error("Error checking if event has been processed", zap.Error(err))
		return err
	}

	if eventHasBeenProcessed {
		g.handler.Logger().Info("Event has already been processed")
		return nil
	}

	processErr := g.handler.Process(ctx, event)

	// Mark as processed...
	err = g.markEventAsProcessed(ctx, policyOrQuoteID, eventID)
	if err != nil {
		g.handler.Logger().Error("Error marking event as processed", zap.Error(err))
		return err
	}

	return processErr
}
