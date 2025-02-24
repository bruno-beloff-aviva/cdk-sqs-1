package sub

import (
	"context"
	"encoding/json"
	"sqstest/lambda/handler/singleshot"
	"sqstest/manager/dbmanager"
	"sqstest/service/testmessage"
	"sqstest/service/testreception"
	"sqstest/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joerdav/zapray"
	"go.uber.org/zap"
)

type ContinuousService struct {
	singleshot.SingleshotGateway[testmessage.TestMessage]
	logger    *zapray.Logger
	dbManager dbmanager.DynamoManager
	id        string
}

func NewContinuousService(logger *zapray.Logger, cfg aws.Config, dbManager dbmanager.DynamoManager, id string) ContinuousService {
	handler := ContinuousService{
		singleshot.NewSingleshotGateway[testmessage.TestMessage](logger, services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed),
		logger,
		dbManager,
		id,
	}

	return handler
}

func (m ContinuousService) Receive(ctx context.Context, record events.SQSMessage) (err error) {
	m.logger.Info("Receive", zap.String("record body", record.Body))

	var message testmessage.TestMessage

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	return m.ProcessOnce(ctx, message)
}

func (m ContinuousService) Process(ctx context.Context, event testmessage.TestMessage) (err error) {
	m.logger.Debug("Process: ", zap.Any("event", event))

	// dbManager.Put...
	reception := testreception.NewTestReception(m.id, event)
	m.logger.Info("Process: ", zap.Any("reception", reception))

	err = m.dbManager.Put(ctx, &reception)
	if err != nil {
		m.logger.Error("Put: ", zap.Error(err))
	}

	return err
}

func (m ContinuousService) UniqueID(event testmessage.TestMessage) (policyOrQuoteID string, eventID string, err error) {
	return event.Client, event.Sent, nil
}
