package singleshot

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

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type SingleShotService[T any] struct {
	logger  *zapray.Logger
	Gateway SingleshotGateway[T]
}

func (m *SingleShotService[T]) NewGateway(logger *zapray.Logger, eventHasBeenProcessed services.EventHasBeenProcessedFunc, EventAsProcessed services.MarkEventAsProcessedFunc) {
	m.logger = logger
	m.Gateway = NewSingleshotGateway(logger, m, eventHasBeenProcessed, EventAsProcessed)
}

func (m *SingleShotService[T]) Process(ctx context.Context, event T) (err error) {
	m.logger.Error("NULL Process!")
	return nil
}

func (m *SingleShotService[T]) UniqueID(event T) (policyOrQuoteID string, eventID string, err error) {
	m.logger.Error("NULL Process!")
	return "", "", nil
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

func (g SingleshotGateway[T]) ProcessOnce(ctx context.Context, event T) error {
	g.logger.Debug("ProcessOnce: ", zap.Any("event", event))

	// Check...
	policyOrQuoteID, eventID, err := g.handler.UniqueID(event)
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

	err = g.handler.Process(ctx, event)
	if err != nil {
		g.logger.Error("Process error", zap.Error(err))
		return err
	}

	// Mark as processed...
	err = g.markEventAsProcessed(ctx, policyOrQuoteID, eventID)
	if err != nil {
		g.logger.Error("Error marking event as processed", zap.Error(err))
		return nil
	}

	return nil
}
