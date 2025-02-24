package singleshotembedded

import (
	"context"
	"sqstest/services"

	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type SingleshotHandler[T any] interface {
	Process(ctx context.Context, event T) (err error)
	UniqueID(event T) (policyOrQuoteID string, eventID string, err error)
}

type SingleshotGateway[T any] struct {
	logger                *zapray.Logger
	eventHasBeenProcessed services.EventHasBeenProcessedFunc
	markEventAsProcessed  services.MarkEventAsProcessedFunc
}

func NewSingleshotGateway[T any](logger *zapray.Logger, eventHasBeenProcessed services.EventHasBeenProcessedFunc, markEventAsProcessed services.MarkEventAsProcessedFunc) SingleshotGateway[T] {
	return SingleshotGateway[T]{
		logger:                logger,
		eventHasBeenProcessed: eventHasBeenProcessed,
		markEventAsProcessed:  markEventAsProcessed,
	}
}

func (g SingleshotGateway[T]) ProcessOnce(ctx context.Context, event T) (err error) {
	g.logger.Debug("ProcessOnce: ", zap.Any("event", event))

	// Check...
	policyOrQuoteID, eventID, err := g.UniqueID(event)
	if err != nil {
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

	err = g.Process(ctx, event)
	if err != nil {
		g.logger.Error("Process error", zap.Error(err))
		return err
	}

	// Mark as processed...
	err = g.markEventAsProcessed(ctx, policyOrQuoteID, eventID)
	if err != nil {
		g.logger.Error("Error marking event as processed", zap.Error(err))
		return err
	}

	return nil
}

func (g SingleshotGateway[T]) Process(ctx context.Context, event T) (err error) {
	g.logger.Error("NULL Process!")

	return nil
}

func (g SingleshotGateway[T]) UniqueID(event T) (policyOrQuoteID string, eventID string, err error) {
	g.logger.Error("NULL UniqueID!")

	return "", "", nil
}
