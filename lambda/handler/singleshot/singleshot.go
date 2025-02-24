package singleshot

import (
	"context"
	"sqstest/services"

	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SingleshotHandler[T any] interface {
	ProcessOnce(ctx context.Context, event T) (err error)
	UniqueID(event T) (policyOrQuoteID string, eventID string, err error)
}

type SingleshotGateway[T any] struct {
	logger                *zapray.Logger
	handler               SingleshotHandler[T]
	eventHasBeenProcessed services.EventHasBeenProcessedFunc
	markEventAsProcessed  services.MarkEventAsProcessedFunc
}

func NewSingleshotGateway[T any](logger *zapray.Logger, handler SingleshotHandler[T], eventHasBeenProcessed services.EventHasBeenProcessedFunc, markEventAsProcessed services.MarkEventAsProcessedFunc) SingleshotGateway[T] {
	return SingleshotGateway[T]{
		logger:                logger,
		handler:               handler,
		eventHasBeenProcessed: eventHasBeenProcessed,
		markEventAsProcessed:  markEventAsProcessed,
	}
}

func (g SingleshotGateway[T]) Handle(ctx context.Context, event T) error {
	// Check...
	policyOrQuoteID, eventID, err := g.handler.UniqueID(event)
	if err == nil {
		g.logger.Error("Error getting UniqueID", zap.Error(err))
		return err
	}

	eventHasBeenProcessed, err := g.eventHasBeenProcessed(ctx, policyOrQuoteID, eventID)
	if err != nil {
		g.logger.Error("Error checking if event has been processed", zap.Error(err))
		return err
	}

	if eventHasBeenProcessed {
		g.logger.Info("Event has already been processed")
		return nil
	}

	processErr := g.handler.ProcessOnce(ctx, event)

	// Mark as processed...
	err = g.markEventAsProcessed(ctx, policyOrQuoteID, eventID)
	if err != nil {
		g.logger.Error("Error marking event as processed", zap.Error(err))
		return err
	}

	return processErr
}
