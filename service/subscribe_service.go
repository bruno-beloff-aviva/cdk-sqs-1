package service

import (
	"context"
	"encoding/json"
	"errors"
	"sqstest/dynamomanager"
	"sqstest/service/testmessage"
	"sqstest/sqsmanager"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

const sleepSeconds = 20

type SubscribeService struct {
	logger     *zapray.Logger
	sqsManager sqsmanager.SQSManager
	dbManager  dynamomanager.DynamoManager
}

func NewSubscribeService(logger *zapray.Logger, cfg aws.Config, dbManager dynamomanager.DynamoManager) SubscribeService {
	sqsManager := sqsmanager.NewSQSManager(logger, cfg)

	return SubscribeService{logger: logger, sqsManager: sqsManager, dbManager: dbManager}
}

func (m SubscribeService) Receive(ctx context.Context, record events.SQSMessage) (err error) {
	m.logger.Debug("Receive", zap.String("record body", record.Body))

	var message testmessage.TestMessage
	var reception testmessage.TestReception

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	m.logger.Debug("Receive: ", zap.Any("message", message))

	if strings.Contains(message.Path, "sleep") {
		// sleep...
		m.logger.Warn("*** SLEEP: ", zap.Any("Path", message.Path))
		time.Sleep(sleepSeconds * time.Second)

	} else if strings.Contains(message.Path, "error") {
		// error...
		m.logger.Warn("*** ERROR: ", zap.Any("Path", message.Path))
		return errors.New(message.Path)

	} else if strings.Contains(message.Path, "panic") {
		// panic...
		m.logger.Warn("*** PANIC: ", zap.Any("Path", message.Path))
		panic(message.Path)

	} else {
		// ok...
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
