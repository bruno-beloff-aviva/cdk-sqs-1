package service

import (
	"context"
	"encoding/json"
	"errors"
	"sqstest/dynamomanager"
	"sqstest/service/testmessage"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

const sleepSeconds = 20

type SubscribeService struct {
	logger    *zapray.Logger
	dbManager dynamomanager.DynamoManager
	Suspended bool
}

func NewSubscribeService(logger *zapray.Logger, cfg aws.Config, dbManager dynamomanager.DynamoManager) *SubscribeService {
	return &SubscribeService{logger: logger, dbManager: dbManager, Suspended: false}
}

func (m *SubscribeService) Receive(ctx context.Context, record events.SQSMessage) (err error) {
	m.logger.Debug("Receive", zap.String("record body", record.Body))

	var message testmessage.TestMessage
	var reception testmessage.TestReception

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	m.logger.Debug("Receive: ", zap.Any("message", message))

	if m.Suspended && !strings.Contains(message.Path, "resume") {
		return errors.New("Suspended")
	}

	switch {
	case strings.Contains(message.Path, "suspend"):
		m.logger.Warn("*** SUSPEND: ", zap.Any("Path", message.Path))
		m.Suspended = true

	case strings.Contains(message.Path, "resume"):
		m.logger.Warn("*** RESUME: ", zap.Any("Path", message.Path))
		m.Suspended = false

	case strings.Contains(message.Path, "sleep"):
		m.logger.Warn("*** SLEEP: ", zap.Any("Path", message.Path))
		time.Sleep(sleepSeconds * time.Second)

	case strings.Contains(message.Path, "error"):
		m.logger.Warn("*** ERROR: ", zap.Any("Path", message.Path))
		return errors.New(message.Path)

	case strings.Contains(message.Path, "panic"):
		m.logger.Warn("*** PANIC: ", zap.Any("Path", message.Path))
		panic(message.Path)

	default:
		m.logger.Warn("*** OK: ", zap.Any("Path", message.Path))
	}

	// dbManager.Put...
	reception = testmessage.NewTestReception(message)
	m.logger.Info("Receive: ", zap.Any("reception", reception))

	err = m.dbManager.Put(ctx, &reception)
	if err != nil {
		m.logger.Error("Receive: ", zap.Error(err))
	}

	return nil
}
