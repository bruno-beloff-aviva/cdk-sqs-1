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
	singleshot.SingleShotService[testmessage.TestMessage]
	logger    *zapray.Logger
	dbManager dbmanager.DynamoManager
	id        string
}

func NewContinuousService(logger *zapray.Logger, cfg aws.Config, dbManager dbmanager.DynamoManager, id string) ContinuousService {
	handler := ContinuousService{logger: logger, dbManager: dbManager, id: id}
	// service.attachGateway(services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed)
	handler.NewGateway(handler, logger, services.NullEventHasBeenProcessed, services.NullMarkEventAsProcessed)

	return handler
}

// func (m *ContinuousService) attachGateway(eventHasBeenProcessed services.EventHasBeenProcessedFunc, EventAsProcessed services.MarkEventAsProcessedFunc) {
// 	m.Gateway = singleshot.NewSingleshotGateway(m.logger, m, eventHasBeenProcessed, EventAsProcessed)
// }

func (m ContinuousService) Receive(ctx context.Context, record events.SQSMessage) (err error) {
	m.logger.Info("Receive", zap.String("record body", record.Body))

	var message testmessage.TestMessage

	err = json.Unmarshal([]byte(record.Body), &message)
	if err != nil {
		return err
	}

	err = m.Gateway.ProcessOnce(ctx, message)

	if err == nil {
		// TODO: success metrics here
	}

	return err
}

func (m ContinuousService) UniqueID(event testmessage.TestMessage) (policyOrQuoteID string, eventID string, err error) {
	return event.Client, event.Sent, nil
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
