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
	logger    *zapray.Logger
	dbManager dbmanager.DynamoManager
	id        string
	gateway   singleshot.SingleshotGateway[testmessage.TestMessage]
}

func NewContinuousService(logger *zapray.Logger, cfg aws.Config, dbManager dbmanager.DynamoManager, id string) ContinuousService {
	self := ContinuousService{logger: logger, dbManager: dbManager, id: id}

	self.gateway = singleshot.NewSingleshotGateway(logger, self, services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed)

	return self
}

func (m ContinuousService) Receive(ctx context.Context, record events.SQSMessage) (err error) {
	m.logger.Info("Receive", zap.String("record body", record.Body))

	var message testmessage.TestMessage

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	return m.gateway.Handle(ctx, message)
}

func (m ContinuousService) ProcessOnce(ctx context.Context, event testmessage.TestMessage) (err error) {
	m.logger.Info("ProcessOnce: ", zap.Any("event", event))

	// dbManager.Put...
	reception := testreception.NewTestReception(m.id, event)
	m.logger.Info("Reception: ", zap.Any("reception", reception))

	err = m.dbManager.Put(ctx, &reception)
	if err != nil {
		m.logger.Error("Put: ", zap.Error(err))
	}

	return err
}

func (m ContinuousService) UniqueID(event testmessage.TestMessage) (policyOrQuoteID string, eventID string, err error) {
	return event.Client, event.Sent, nil
}
